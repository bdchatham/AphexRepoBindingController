package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexControllerRuntime/pkg/config"
	"github.com/bdchatham/AphexControllerRuntime/pkg/constants"
	"github.com/bdchatham/AphexControllerRuntime/pkg/helpers"
	"github.com/bdchatham/AphexControllerRuntime/pkg/metrics"
	"github.com/bdchatham/AphexControllerRuntime/pkg/validators"
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
)

// RepoBindingReconciler reconciles a RepoBinding object
type RepoBindingReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Log             logr.Logger
	TemplateCatalog *TemplateCatalog
	Config          *config.Config
	statusHelper    *helpers.StatusHelper
	finalizerHelper *helpers.FinalizerHelper
	rbacValidator   *validators.RBACValidator
}

// +kubebuilder:rbac:groups=aphex.io,resources=repobindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aphex.io,resources=repobindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aphex.io,resources=repobindings/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=triggerbindings,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=triggertemplates,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=triggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=triggers.tekton.dev,resources=eventlisteners,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=tekton.dev,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups=argoproj.io,resources=appprojects,verbs=get;list;watch;create;update;patch;delete

// Reconcile handles RepoBinding create/update/delete events
func (r *RepoBindingReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("repobinding", request.NamespacedName)

	// Start metrics timer for reconciliation duration
	metricsCollector := metrics.GetCollector()
	timer := metricsCollector.NewReconcileTimer(constants.ControllerNameRepoBinding)

	// Initialize helpers if not already done (lazy initialization)
	if err := r.ensureHelpers(logger); err != nil {
		logger.Error(err, "Failed to initialize helpers")
		timer.ObserveError(metrics.ClassifyError(err))
		return ctrl.Result{}, err
	}

	// Apply provisioning timeout to the context
	timeout := constants.DefaultProvisioningTimeout
	if r.Config != nil {
		timeout = r.Config.ProvisioningTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	repoBinding, err := r.fetchRepoBinding(ctx, request)
	if err != nil {
		result, fetchErr := r.handleFetchError(logger, err)
		if fetchErr != nil {
			timer.ObserveError(metrics.ClassifyError(fetchErr))
		}
		return result, fetchErr
	}

	// Handle deletion with verified cleanup
	if r.finalizerHelper.IsBeingDeleted(repoBinding) {
		result, err := r.handleDeletionWithHelper(ctx, logger, repoBinding)
		if err != nil {
			timer.ObserveError(metrics.ClassifyError(err))
		} else {
			timer.ObserveSuccess()
			metrics.GetHealthState().RecordSuccess(constants.ControllerNameRepoBinding)
		}
		return result, err
	}

	if !r.finalizerHelper.HasFinalizer(repoBinding) {
		if err := ValidateRepoBinding(repoBinding, r.TemplateCatalog); err != nil {
			logger.Error(err, "Validation failed, not adding finalizer")
			timer.ObserveError("validation")
			return r.updateStatusToFailedWithHelper(ctx, repoBinding, fmt.Sprintf("Validation failed: %s", err.Error()))
		}
		if err := r.finalizerHelper.EnsureFinalizer(ctx, repoBinding); err != nil {
			timer.ObserveError(metrics.ClassifyError(err))
			return ctrl.Result{}, err
		}
		timer.ObserveSuccess()
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.initializeStatusWithHelper(ctx, logger, repoBinding); err != nil {
		timer.ObserveError(metrics.ClassifyError(err))
		return ctrl.Result{}, err
	}

	if r.isAlreadyReady(logger, repoBinding) {
		timer.ObserveSuccess()
		metrics.GetHealthState().RecordSuccess(constants.ControllerNameRepoBinding)
		return ctrl.Result{}, nil
	}

	if err := r.validateAndUpdatePhaseWithHelper(ctx, logger, request, repoBinding); err != nil {
		timer.ObserveError(metrics.ClassifyError(err))
		return ctrl.Result{}, err
	}

	result, err := r.executeProvisioningSteps(ctx, logger, request, repoBinding)
	if err != nil {
		timer.ObserveError(metrics.ClassifyError(err))
	} else {
		timer.ObserveSuccess()
		metrics.GetHealthState().RecordSuccess(constants.ControllerNameRepoBinding)
	}
	return result, err
}

// ensureHelpers initializes the helper components if not already done.
func (r *RepoBindingReconciler) ensureHelpers(logger logr.Logger) error {
	if r.statusHelper == nil {
		retryCount := constants.DefaultStatusRetryCount
		if r.Config != nil {
			retryCount = r.Config.StatusRetryCount
		}
		r.statusHelper = helpers.NewStatusHelper(r.Client, logger, retryCount)
	}

	if r.finalizerHelper == nil {
		r.finalizerHelper = helpers.NewFinalizerHelper(r.Client, logger, constants.RepoBindingFinalizer)
	}

	if r.rbacValidator == nil {
		controllerNS := constants.DefaultPlatformNamespace
		controllerSA := "repobinding-controller"
		if r.Config != nil {
			controllerNS = r.Config.PlatformNamespace
		}
		r.rbacValidator = validators.NewRBACValidator(r.Client, logger, controllerNS, controllerSA)
	}

	// Validate TemplateCatalog is not nil
	if r.TemplateCatalog == nil {
		return fmt.Errorf("TemplateCatalog is nil: controller not properly initialized")
	}

	return nil
}

func (r *RepoBindingReconciler) executeProvisioningSteps(ctx context.Context, logger logr.Logger, request ctrl.Request, repoBinding *platformv1alpha1.RepoBinding) (ctrl.Result, error) {
	logger.Info("Reconciling RepoBinding",
		"repoOrg", repoBinding.Spec.RepoOrg,
		"repoName", repoBinding.Spec.RepoName,
		"pipelineName", repoBinding.Spec.PipelineName,
		"templateRef", repoBinding.Spec.TemplateRef)

	steps := []provisioningStep{
		{name: "namespace", statusField: &repoBinding.Status.NamespaceCreated, provisionFunc: r.provisionNamespace},
		{name: "pipeline", statusField: &repoBinding.Status.PipelineCreated, provisionFunc: r.provisionPipeline},
		{name: "service account", statusField: &repoBinding.Status.ServiceAccountCreated, provisionFunc: r.provisionServiceAccount},
		{name: "RBAC", statusField: &repoBinding.Status.RBACCreated, provisionFunc: r.provisionRBAC},
		{name: "EventListener namespace", statusField: &repoBinding.Status.TriggerBindingCreated, provisionFunc: r.updateEventListenerNamespaces},
		{name: "TriggerTemplate", statusField: &repoBinding.Status.TriggerTemplateCreated, provisionFunc: r.provisionTriggerTemplate},
		{name: "Trigger", statusField: &repoBinding.Status.TriggerCreated, provisionFunc: r.provisionTrigger},
	}

	for _, step := range steps {
		if !*step.statusField {
			if err := r.executeProvisioningStep(ctx, logger, request, repoBinding, step); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	if err := r.finalizeProvisioning(ctx, logger, request, repoBinding); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type provisioningStep struct {
	name          string
	statusField   *bool
	provisionFunc func(context.Context, *platformv1alpha1.RepoBinding) error
}

func (r *RepoBindingReconciler) executeProvisioningStep(ctx context.Context, logger logr.Logger, _ ctrl.Request, repoBinding *platformv1alpha1.RepoBinding, step provisioningStep) error {
	logger.V(1).Info("Provisioning step", "step", step.name)

	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during provisioning step %q: %w", step.name, ctx.Err())
	default:
	}

	metricsCollector := metrics.GetCollector()

	if err := step.provisionFunc(ctx, repoBinding); err != nil {
		logger.Error(err, "Failed to provision", "step", step.name)
		// Record provisioning step failure
		metricsCollector.RecordProvisioningStep(constants.ControllerNameRepoBinding, step.name, "error")
		// Use StatusHelper for status update with retry
		if patchErr := r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
			"phase":             constants.PhaseFailed,
			"message":           fmt.Sprintf("Failed to provision %s: %s", step.name, err.Error()),
			"lastReconcileTime": metav1.Now().Format(time.RFC3339),
		}); patchErr != nil {
			logger.Error(patchErr, "Failed to update status after provisioning failure")
		}
		return err
	}

	// Record provisioning step success
	metricsCollector.RecordProvisioningStep(constants.ControllerNameRepoBinding, step.name, "success")

	// Use StatusHelper to update the status field with retry
	statusPatch := map[string]interface{}{
		"lastReconcileTime": metav1.Now().Format(time.RFC3339),
	}

	// Map step name to status field
	switch step.name {
	case "namespace":
		statusPatch["namespaceCreated"] = true
	case "pipeline":
		statusPatch["pipelineCreated"] = true
	case "service account":
		statusPatch["serviceAccountCreated"] = true
	case "RBAC":
		statusPatch["rbacCreated"] = true
	case "EventListener namespace":
		statusPatch["triggerBindingCreated"] = true
	case "TriggerTemplate":
		statusPatch["triggerTemplateCreated"] = true
	case "Trigger":
		statusPatch["triggerCreated"] = true
	}

	if err := r.statusHelper.PatchStatus(ctx, repoBinding, statusPatch); err != nil {
		logger.Error(err, "Failed to update RepoBinding status")
		return err
	}

	return nil
}

func (r *RepoBindingReconciler) finalizeProvisioning(ctx context.Context, logger logr.Logger, _ ctrl.Request, repoBinding *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during finalization: %w", ctx.Err())
	default:
	}

	logger.V(1).Info("Updating RepoBinding status with webhook information")
	if err := r.updateRepoBindingStatusWithWebhookInfo(ctx, repoBinding); err != nil {
		logger.Error(err, "Failed to update RepoBinding status with webhook info")
		if patchErr := r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
			"phase":             constants.PhaseFailed,
			"message":           fmt.Sprintf("Failed to update webhook info: %s", err.Error()),
			"lastReconcileTime": metav1.Now().Format(time.RFC3339),
		}); patchErr != nil {
			logger.Error(patchErr, "Failed to update status")
		}
		return err
	}

	// Final status update to Ready using StatusHelper
	return r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
		"phase":             constants.PhaseReady,
		"message":           "Tenant resources provisioned successfully",
		"lastReconcileTime": metav1.Now().Format(time.RFC3339),
	})
}

func (r *RepoBindingReconciler) fetchRepoBinding(ctx context.Context, request ctrl.Request) (*platformv1alpha1.RepoBinding, error) {
	repoBinding := &platformv1alpha1.RepoBinding{}
	err := r.Get(ctx, request.NamespacedName, repoBinding)
	return repoBinding, err
}

//nolint:unparam // Result is always empty per controller-runtime pattern
func (r *RepoBindingReconciler) handleFetchError(logger logr.Logger, err error) (ctrl.Result, error) {
	if errors.IsNotFound(err) {
		logger.V(1).Info("RepoBinding resource not found, ignoring since object must be deleted")
		return ctrl.Result{}, nil
	}
	logger.Error(err, "Failed to get RepoBinding")
	return ctrl.Result{}, err
}

// handleDeletionWithHelper performs cleanup using FinalizerHelper with verified cleanup.
//
//nolint:unparam // Result is always empty per controller-runtime pattern
func (r *RepoBindingReconciler) handleDeletionWithHelper(ctx context.Context, _ logr.Logger, repoBinding *platformv1alpha1.RepoBinding) (ctrl.Result, error) {
	if !r.finalizerHelper.NeedsCleanup(repoBinding) {
		return ctrl.Result{}, nil
	}

	cleanupSteps := []helpers.CleanupStep{
		helpers.NewCleanupStep("Trigger", func(ctx context.Context) error {
			return r.cleanupTrigger(ctx, repoBinding)
		}),
		helpers.NewCleanupStep("TriggerTemplate", func(ctx context.Context) error {
			return r.cleanupTriggerTemplate(ctx, repoBinding)
		}),
		helpers.NewCleanupStep("ArgoCD Applications", func(ctx context.Context) error {
			return r.cleanupArgoCDApplications(ctx, repoBinding)
		}),
		helpers.NewCleanupStep("ArgoCD AppProject", func(ctx context.Context) error {
			return r.cleanupArgoCDAppProject(ctx, repoBinding)
		}),
		helpers.NewCleanupStep("Cluster-scoped RBAC", func(ctx context.Context) error {
			return r.cleanupClusterScopedRBAC(ctx, repoBinding)
		}),
		helpers.NewCleanupStep("Namespace", func(ctx context.Context) error {
			return r.cleanupNamespace(ctx, repoBinding)
		}),
	}

	if err := r.finalizerHelper.HandleDeletionWithSteps(ctx, repoBinding, cleanupSteps); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// cleanupTrigger removes the Trigger resource.
func (r *RepoBindingReconciler) cleanupTrigger(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)
	trigger := &triggersv1beta1.Trigger{}
	triggerName := fmt.Sprintf("%s-trigger", rb.Spec.PipelineName)

	if err := r.Get(ctx, client.ObjectKey{Name: triggerName, Namespace: orgNamespace}, trigger); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get Trigger: %w", err)
	}

	if err := r.Delete(ctx, trigger); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Trigger: %w", err)
	}
	return nil
}

// cleanupTriggerTemplate removes the TriggerTemplate resource.
func (r *RepoBindingReconciler) cleanupTriggerTemplate(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)
	triggerTemplate := &triggersv1beta1.TriggerTemplate{}
	templateName := fmt.Sprintf("%s-trigger-template", rb.Spec.PipelineName)

	if err := r.Get(ctx, client.ObjectKey{Name: templateName, Namespace: orgNamespace}, triggerTemplate); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get TriggerTemplate: %w", err)
	}

	if err := r.Delete(ctx, triggerTemplate); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete TriggerTemplate: %w", err)
	}
	return nil
}

// cleanupArgoCDAppProject removes the ArgoCD AppProject.
func (r *RepoBindingReconciler) cleanupArgoCDAppProject(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	appProject := &unstructured.Unstructured{}
	appProject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   constants.ArgoCDGroup,
		Version: constants.ArgoCDVersion,
		Kind:    constants.ArgoCDProjectKind,
	})

	if err := r.Get(ctx, client.ObjectKey{Name: rb.Spec.PipelineName, Namespace: constants.ArgoCDNamespace}, appProject); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get ArgoCD AppProject: %w", err)
	}

	if err := r.Delete(ctx, appProject); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ArgoCD AppProject: %w", err)
	}
	return nil
}

// cleanupArgoCDApplications removes ArgoCD Applications labeled with this pipeline.
func (r *RepoBindingReconciler) cleanupArgoCDApplications(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	appList := &unstructured.UnstructuredList{}
	appList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   constants.ArgoCDGroup,
		Version: constants.ArgoCDVersion,
		Kind:    constants.ArgoCDAppKind,
	})

	if err := r.List(ctx, appList, client.InNamespace(rb.Spec.PipelineName), client.MatchingLabels{
		constants.LabelPipeline: rb.Spec.PipelineName,
	}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to list ArgoCD Applications: %w", err)
	}

	for _, app := range appList.Items {
		if err := r.Delete(ctx, &app); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete ArgoCD Application %s: %w", app.GetName(), err)
		}
	}
	return nil
}

// cleanupClusterScopedRBAC removes ClusterRoles and ClusterRoleBindings labeled with this pipeline.
func (r *RepoBindingReconciler) cleanupClusterScopedRBAC(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	clusterRoleName := fmt.Sprintf("pipeline-runner-%s", rb.Spec.PipelineName)

	// Delete ClusterRoleBinding
	crb := &unstructured.Unstructured{}
	crb.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRoleBinding",
	})
	if err := r.Get(ctx, client.ObjectKey{Name: clusterRoleName}, crb); err == nil {
		if err := r.Delete(ctx, crb); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete ClusterRoleBinding: %w", err)
		}
	}

	// Delete ClusterRole
	cr := &unstructured.Unstructured{}
	cr.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRole",
	})
	if err := r.Get(ctx, client.ObjectKey{Name: clusterRoleName}, cr); err == nil {
		if err := r.Delete(ctx, cr); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete ClusterRole: %w", err)
		}
	}

	return nil
}

// cleanupNamespace deletes the managed pipeline namespace.
func (r *RepoBindingReconciler) cleanupNamespace(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: rb.Spec.PipelineName}, ns); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get namespace: %w", err)
	}

	if err := r.Delete(ctx, ns); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	return nil
}

// initializeStatusWithHelper initializes the RepoBinding status using StatusHelper.
func (r *RepoBindingReconciler) initializeStatusWithHelper(ctx context.Context, _ logr.Logger, repoBinding *platformv1alpha1.RepoBinding) error {
	if repoBinding.Status.Phase == "" {
		return r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
			"phase":             constants.PhasePending,
			"message":           "Starting onboarding process",
			"lastReconcileTime": metav1.Now().Format(time.RFC3339),
		})
	}
	return nil
}

func (r *RepoBindingReconciler) isAlreadyReady(logger logr.Logger, repoBinding *platformv1alpha1.RepoBinding) bool {
	if repoBinding.Status.Phase == constants.PhaseReady {
		logger.V(1).Info("RepoBinding already in Ready state, skipping reconciliation")
		return true
	}
	return false
}

// validateAndUpdatePhaseWithHelper validates the RepoBinding and updates phase using StatusHelper.
func (r *RepoBindingReconciler) validateAndUpdatePhaseWithHelper(ctx context.Context, logger logr.Logger, _ ctrl.Request, repoBinding *platformv1alpha1.RepoBinding) error {
	if repoBinding.Status.Phase == constants.PhasePending {
		if err := ValidateRepoBinding(repoBinding, r.TemplateCatalog); err != nil {
			logger.Error(err, "Validation failed")
			_, updateErr := r.updateStatusToFailedWithHelper(ctx, repoBinding, fmt.Sprintf("Validation failed: %s", err.Error()))
			if updateErr != nil {
				return updateErr
			}
			return err
		}

		return r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
			"phase":             constants.PhaseProvisioning,
			"message":           "Provisioning tenant resources",
			"lastReconcileTime": metav1.Now().Format(time.RFC3339),
		})
	}
	return nil
}

// updateStatusToFailedWithHelper updates the status to Failed using StatusHelper.
//
//nolint:unparam // Result is always empty per controller-runtime pattern
func (r *RepoBindingReconciler) updateStatusToFailedWithHelper(ctx context.Context, repoBinding *platformv1alpha1.RepoBinding, message string) (ctrl.Result, error) {
	err := r.statusHelper.PatchStatus(ctx, repoBinding, map[string]interface{}{
		"phase":             constants.PhaseFailed,
		"message":           message,
		"lastReconcileTime": metav1.Now().Format(time.RFC3339),
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	// Don't requeue failed resources
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *RepoBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return r.SetupWithManagerAndOptions(mgr, nil)
}

// SetupWithManagerAndOptions sets up the controller with the Manager and custom options.
// This allows configuring rate limiting and max concurrent reconciles.
func (r *RepoBindingReconciler) SetupWithManagerAndOptions(mgr ctrl.Manager, opts *controller.Options) error {
	ctrlOpts := controller.Options{}
	if opts != nil {
		ctrlOpts = *opts
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.RepoBinding{}).
		WithOptions(ctrlOpts).
		Complete(r)
}
