package controller

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexControllerRuntime/pkg/constants"
	"github.com/bdchatham/AphexControllerRuntime/pkg/provisioners"
	"github.com/bdchatham/AphexControllerRuntime/pkg/validators"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
)

// provisionNamespace creates or updates the pipeline namespace using idempotent provisioning.
func (r *RepoBindingReconciler) provisionNamespace(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during namespace provisioning: %w", ctx.Err())
	default:
	}

	repoLabel := fmt.Sprintf("%s-%s", rb.Spec.RepoOrg, rb.Spec.RepoName)

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: rb.Spec.PipelineName,
			Labels: map[string]string{
				constants.LabelPipeline:     rb.Spec.PipelineName,
				constants.LabelRepo:         repoLabel,
				constants.LabelManagedBy:    constants.ManagedByPlatformController,
				constants.LabelOrganization: rb.Spec.AphexOrg,
			},
		},
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, namespace, func(obj client.Object) interface{} {
		ns, ok := obj.(*corev1.Namespace)
		if !ok {
			return nil
		}
		return ns.Labels
	})

	if err != nil {
		return fmt.Errorf("failed to provision namespace: %w", err)
	}

	r.Log.V(1).Info("Namespace provisioning result",
		"namespace", rb.Spec.PipelineName,
		"result", result.String())

	return r.provisionPipelineContext(ctx, rb)
}

func (r *RepoBindingReconciler) provisionPipelineContext(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pipeline-context",
			Namespace: rb.Spec.PipelineName,
			Labels: map[string]string{
				constants.LabelManagedBy:    constants.ManagedByPlatformController,
				constants.LabelOrganization: rb.Spec.AphexOrg,
			},
		},
		Data: map[string]string{
			"REPO_URL":      fmt.Sprintf("https://github.com/%s/%s", rb.Spec.RepoOrg, rb.Spec.RepoName),
			"REPO_ORG":      rb.Spec.RepoOrg,
			"REPO_NAME":     rb.Spec.RepoName,
			"APHEX_ORG":     rb.Spec.AphexOrg,
			"PIPELINE_NAME": rb.Spec.PipelineName,
			"ORG_NAMESPACE": fmt.Sprintf("org-%s", rb.Spec.AphexOrg),
		},
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	_, err := helper.CreateOrUpdate(ctx, configMap, func(obj client.Object) interface{} {
		cm, ok := obj.(*corev1.ConfigMap)
		if !ok {
			return nil
		}
		return cm.Data
	})
	if err != nil {
		return fmt.Errorf("failed to provision pipeline context: %w", err)
	}
	return nil
}

// provisionPipeline creates the Tekton Pipeline resource from the spec using idempotent provisioning.
func (r *RepoBindingReconciler) provisionPipeline(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during pipeline provisioning: %w", ctx.Err())
	default:
	}

	// Validate YAML before parsing
	yamlValidator := validators.NewYAMLValidator()
	unstructuredPipeline, err := yamlValidator.ValidateAndDecode([]byte(rb.Spec.PipelineSpec), validators.TektonPipelineGVK)
	if err != nil {
		r.Log.Error(err, "Pipeline YAML validation failed")
		return fmt.Errorf("pipeline YAML validation failed: %w", err)
	}

	// Decode YAML using Kubernetes decoder (handles JSON tags properly)
	pipeline := &tektonv1.Pipeline{}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(rb.Spec.PipelineSpec)), 4096)
	if err := decoder.Decode(pipeline); err != nil {
		r.Log.Error(err, "Failed to parse pipeline YAML")
		return fmt.Errorf("failed to parse pipeline YAML: %w", err)
	}

	r.Log.V(1).Info("Parsed and validated pipeline",
		"originalName", pipeline.Name,
		"kind", pipeline.Kind,
		"apiVersion", pipeline.APIVersion,
		"validatedGVK", unstructuredPipeline.GetObjectKind().GroupVersionKind().String())

	// Set name and namespace from RepoBinding (override whatever is in the YAML)
	pipeline.Name = rb.Spec.PipelineName
	pipeline.Namespace = rb.Spec.PipelineName
	pipeline.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   constants.TektonPipelineGroup,
		Version: constants.TektonPipelineVersion,
		Kind:    "Pipeline",
	})

	// Add labels for tracking
	if pipeline.Labels == nil {
		pipeline.Labels = make(map[string]string)
	}
	pipeline.Labels[constants.LabelPipeline] = rb.Spec.PipelineName
	pipeline.Labels[constants.LabelManagedBy] = constants.ManagedByPlatformController

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, pipeline, func(obj client.Object) interface{} {
		p, ok := obj.(*tektonv1.Pipeline)
		if !ok {
			return nil
		}
		return p.Spec
	})

	if err != nil {
		return fmt.Errorf("failed to provision Pipeline: %w", err)
	}

	r.Log.V(1).Info("Pipeline provisioning result",
		"name", rb.Spec.PipelineName,
		"namespace", rb.Spec.PipelineName,
		"result", result.String())

	return nil
}

// provisionServiceAccount creates or updates the tenant service account using idempotent provisioning.
func (r *RepoBindingReconciler) provisionServiceAccount(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during service account provisioning: %w", ctx.Err())
	default:
	}

	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PipelineRunnerServiceAccount,
			Namespace: rb.Spec.PipelineName,
			Labels: map[string]string{
				constants.LabelPipeline:  rb.Spec.PipelineName,
				constants.LabelManagedBy: constants.ManagedByPlatformController,
			},
		},
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateIfNotExists(ctx, serviceAccount)

	if err != nil {
		return fmt.Errorf("failed to provision service account: %w", err)
	}

	r.Log.V(1).Info("Service account provisioning result",
		"namespace", rb.Spec.PipelineName,
		"name", constants.PipelineRunnerServiceAccount,
		"result", result.String())

	return nil
}

// provisionRBAC creates or updates the pipeline RBAC (Role, RoleBinding, ClusterRole, ClusterRoleBinding)
func (r *RepoBindingReconciler) provisionRBAC(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during RBAC provisioning: %w", ctx.Err())
	default:
	}

	// Use standard profile for all pipelines
	profile := "standard"

	// Create namespace-scoped Role
	if err := r.provisionRole(ctx, rb, profile); err != nil {
		return fmt.Errorf("failed to provision role: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled after role provisioning: %w", ctx.Err())
	default:
	}

	// Create namespace-scoped RoleBinding
	if err := r.provisionRoleBinding(ctx, rb); err != nil {
		return fmt.Errorf("failed to provision rolebinding: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled after rolebinding provisioning: %w", ctx.Err())
	default:
	}

	// Create cluster-scoped ClusterRole for Tekton Triggers resources
	if err := r.provisionClusterRole(ctx, rb); err != nil {
		return fmt.Errorf("failed to provision cluster role: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled after cluster role provisioning: %w", ctx.Err())
	default:
	}

	// Create cluster-scoped ClusterRoleBinding
	if err := r.provisionClusterRoleBinding(ctx, rb); err != nil {
		return fmt.Errorf("failed to provision cluster rolebinding: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled after cluster rolebinding provisioning: %w", ctx.Err())
	default:
	}

	// Create ArgoCD RoleBinding in argocd namespace
	if err := r.provisionArgoCDRoleBinding(ctx, rb); err != nil {
		return fmt.Errorf("failed to provision argocd rolebinding: %w", err)
	}

	// Check context cancellation between steps
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled after argocd rolebinding provisioning: %w", ctx.Err())
	default:
	}

	// Create ArgoCD AppProject for pipeline isolation
	if err := r.provisionArgoCDAppProject(ctx, rb); err != nil {
		return fmt.Errorf("failed to provision argocd appproject: %w", err)
	}

	return nil
}

// provisionRole creates or updates the pipeline Role using idempotent provisioning.
func (r *RepoBindingReconciler) provisionRole(ctx context.Context, rb *platformv1alpha1.RepoBinding, profile string) error {
	role := r.buildRole(rb.Spec.PipelineName, profile)
	role.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "Role",
	})

	if r.rbacValidator != nil {
		if err := r.rbacValidator.ValidateRole(ctx, role); err != nil {
			return fmt.Errorf("RBAC validation failed: %w", err)
		}
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, role, func(obj client.Object) interface{} {
		roleObj, ok := obj.(*rbacv1.Role)
		if !ok {
			return nil
		}
		return roleObj.Rules
	})

	if err != nil {
		return fmt.Errorf("failed to provision role: %w", err)
	}

	r.Log.V(1).Info("Role provisioning result",
		"namespace", rb.Spec.PipelineName,
		"profile", profile,
		"result", result.String())

	return nil
}

// provisionRoleBinding creates or updates the pipeline RoleBinding using idempotent provisioning.
func (r *RepoBindingReconciler) provisionRoleBinding(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PipelineRunnerRoleName,
			Namespace: rb.Spec.PipelineName,
			Labels: map[string]string{
				constants.LabelPipeline:  rb.Spec.PipelineName,
				constants.LabelManagedBy: constants.ManagedByPlatformController,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     constants.PipelineRunnerRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      constants.PipelineRunnerServiceAccount,
				Namespace: rb.Spec.PipelineName,
			},
		},
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateIfNotExists(ctx, roleBinding)

	if err != nil {
		return fmt.Errorf("failed to provision rolebinding: %w", err)
	}

	r.Log.V(1).Info("RoleBinding provisioning result",
		"namespace", rb.Spec.PipelineName,
		"result", result.String())

	return nil
}

// provisionClusterRole creates or updates the pipeline ClusterRole for cluster-scoped Tekton Triggers resources.
func (r *RepoBindingReconciler) provisionClusterRole(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	clusterRoleName := fmt.Sprintf("pipeline-runner-%s", rb.Spec.PipelineName)

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"triggers.tekton.dev"},
				Resources: []string{"clusterinterceptors", "clustertriggerbindings"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	// Set cluster-scoped owner labels for tracking
	provisioners.SetClusterScopedOwnerLabels(clusterRole, rb.Name, rb.Namespace, "RepoBinding")
	clusterRole.Labels[constants.LabelPipeline] = rb.Spec.PipelineName

	if r.rbacValidator != nil {
		if err := r.rbacValidator.ValidateClusterRole(ctx, clusterRole); err != nil {
			return fmt.Errorf("RBAC validation failed: %w", err)
		}
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, clusterRole, func(obj client.Object) interface{} {
		cr, ok := obj.(*rbacv1.ClusterRole)
		if !ok {
			return nil
		}
		return cr.Rules
	})

	if err != nil {
		return fmt.Errorf("failed to provision ClusterRole: %w", err)
	}

	r.Log.V(1).Info("ClusterRole provisioning result",
		"clusterRole", clusterRoleName,
		"result", result.String())

	return nil
}

// provisionClusterRoleBinding creates or updates the pipeline ClusterRoleBinding.
func (r *RepoBindingReconciler) provisionClusterRoleBinding(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	clusterRoleName := fmt.Sprintf("pipeline-runner-%s", rb.Spec.PipelineName)
	clusterRoleBindingName := fmt.Sprintf("pipeline-runner-%s", rb.Spec.PipelineName)

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      constants.PipelineRunnerServiceAccount,
				Namespace: rb.Spec.PipelineName,
			},
		},
	}

	// Set cluster-scoped owner labels for tracking
	provisioners.SetClusterScopedOwnerLabels(clusterRoleBinding, rb.Name, rb.Namespace, "RepoBinding")
	clusterRoleBinding.Labels[constants.LabelPipeline] = rb.Spec.PipelineName

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, clusterRoleBinding, func(obj client.Object) interface{} {
		crb, ok := obj.(*rbacv1.ClusterRoleBinding)
		if !ok {
			return nil
		}
		return struct {
			RoleRef  rbacv1.RoleRef
			Subjects []rbacv1.Subject
		}{crb.RoleRef, crb.Subjects}
	})

	if err != nil {
		return fmt.Errorf("failed to provision ClusterRoleBinding: %w", err)
	}

	r.Log.V(1).Info("ClusterRoleBinding provisioning result",
		"clusterRoleBinding", clusterRoleBindingName,
		"result", result.String())

	return nil
}

// provisionArgoCDRoleBinding creates a RoleBinding in the argocd namespace
// granting the pipeline's service account permission to manage ArgoCD Applications.
func (r *RepoBindingReconciler) provisionArgoCDRoleBinding(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	roleBindingName := fmt.Sprintf("%s-argocd-access", rb.Spec.PipelineName)

	roleBinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: constants.ArgoCDNamespace,
			Labels: map[string]string{
				constants.LabelPipeline:  rb.Spec.PipelineName,
				constants.LabelManagedBy: constants.ManagedByPlatformController,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.ArgoCDApplicationDeployerRole,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      constants.PipelineRunnerServiceAccount,
				Namespace: rb.Spec.PipelineName,
			},
		},
	}

	helper := provisioners.NewIdempotentHelper(r.Client, r.Log)
	result, err := helper.CreateOrUpdate(ctx, roleBinding, func(obj client.Object) interface{} {
		rbObj, ok := obj.(*rbacv1.RoleBinding)
		if !ok {
			return nil
		}
		return rbObj.Subjects
	})

	if err != nil {
		return fmt.Errorf("failed to provision ArgoCD RoleBinding: %w", err)
	}

	r.Log.V(1).Info("ArgoCD RoleBinding provisioning result",
		"name", roleBindingName,
		"pipeline", rb.Spec.PipelineName,
		"result", result.String())

	return nil
}

// provisionArgoCDAppProject creates an AppProject scoped to the pipeline's allowed destinations.
//
// NOTE: We use unstructured here instead of importing github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1
// because ArgoCD's API types have deep transitive dependencies on k8s.io/kubernetes which causes
// version conflicts with our k8s.io/api version. The unstructured approach avoids this dependency
// hell while still providing type-safe interaction with the ArgoCD CRD.
func (r *RepoBindingReconciler) provisionArgoCDAppProject(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	projectName := rb.Spec.PipelineName

	appProject := &unstructured.Unstructured{}
	appProject.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   constants.ArgoCDGroup,
		Version: constants.ArgoCDVersion,
		Kind:    constants.ArgoCDProjectKind,
	})
	appProject.SetName(projectName)
	appProject.SetNamespace(constants.ArgoCDNamespace)
	appProject.SetLabels(map[string]string{
		constants.LabelPipeline:  rb.Spec.PipelineName,
		constants.LabelManagedBy: constants.ManagedByPlatformController,
	})

	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)

	spec := map[string]interface{}{
		"description": fmt.Sprintf("Project for %s pipeline", rb.Spec.PipelineName),
		"destinations": []interface{}{
			map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": rb.Spec.PipelineName,
			},
			map[string]interface{}{
				"server":    "https://kubernetes.default.svc",
				"namespace": orgNamespace,
			},
		},
		"sourceRepos": []interface{}{
			fmt.Sprintf("https://github.com/%s/%s", rb.Spec.RepoOrg, rb.Spec.RepoName),
			fmt.Sprintf("https://github.com/%s/%s.git", rb.Spec.RepoOrg, rb.Spec.RepoName),
		},
		"sourceNamespaces": []interface{}{
			rb.Spec.PipelineName,
		},
		"clusterResourceWhitelist": []interface{}{},
		"namespaceResourceWhitelist": []interface{}{
			map[string]interface{}{"group": "*", "kind": "*"},
		},
	}
	appProject.Object["spec"] = spec

	existingProject := &unstructured.Unstructured{}
	existingProject.SetGroupVersionKind(appProject.GroupVersionKind())
	err := r.Get(ctx, client.ObjectKey{Name: projectName, Namespace: constants.ArgoCDNamespace}, existingProject)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.V(1).Info("Creating ArgoCD AppProject",
				"name", projectName,
				"pipeline", rb.Spec.PipelineName)
			if err := r.Create(ctx, appProject); err != nil {
				return fmt.Errorf("failed to create ArgoCD AppProject: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to get ArgoCD AppProject: %w", err)
	}

	r.Log.V(1).Info("Updating ArgoCD AppProject",
		"name", projectName,
		"pipeline", rb.Spec.PipelineName)
	existingProject.Object["spec"] = spec
	existingProject.SetLabels(appProject.GetLabels())
	if err := r.Update(ctx, existingProject); err != nil {
		return fmt.Errorf("failed to update ArgoCD AppProject: %w", err)
	}

	return nil
}

// buildRole constructs a Role based on the permission profile
func (r *RepoBindingReconciler) buildRole(namespace, profile string) *rbacv1.Role {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PipelineRunnerRoleName,
			Namespace: namespace,
			Labels: map[string]string{
				constants.LabelPipeline:  namespace,
				constants.LabelManagedBy: constants.ManagedByPlatformController,
			},
		},
	}

	// Standard rules (common to both profiles)
	standardRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "pods/log"},
			Verbs:     []string{"get", "list", "create", "update", "delete", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"configmaps", "secrets"},
			Verbs:     []string{"get", "list", "create", "update", "delete"},
		},
		{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"pipelineruns", "taskruns"},
			Verbs:     []string{"get", "list", "create", "watch"},
		},
		{
			APIGroups: []string{"triggers.tekton.dev"},
			Resources: []string{"eventlisteners", "triggers", "triggerbindings", "triggertemplates", "interceptors"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"persistentvolumeclaims"},
			Verbs:     []string{"get", "list"},
		},
	}

	if profile == "elevated" {
		// Add elevated rules
		elevatedRules := []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list", "create", "update", "delete"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"get", "list", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list", "watch"},
			},
		}
		// Keep all standard rules except the PVC rule (which gets replaced), then add elevated rules
		role.Rules = append([]rbacv1.PolicyRule{}, standardRules[:4]...)
		role.Rules = append(role.Rules, elevatedRules...)
	} else {
		role.Rules = standardRules
	}

	return role
}

// updateEventListenerNamespaces adds the pipeline namespace to the EventListener's namespace selector.
func (r *RepoBindingReconciler) updateEventListenerNamespaces(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during EventListener update: %w", ctx.Err())
	default:
	}

	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)
	eventListener := &triggersv1beta1.EventListener{}

	if err := r.Get(ctx, client.ObjectKey{Name: constants.GitHubListenerName, Namespace: orgNamespace}, eventListener); err != nil {
		return fmt.Errorf("failed to get EventListener: %w", err)
	}

	for _, ns := range eventListener.Spec.NamespaceSelector.MatchNames {
		if ns == rb.Spec.PipelineName {
			return nil
		}
	}

	eventListener.Spec.NamespaceSelector.MatchNames = append(
		eventListener.Spec.NamespaceSelector.MatchNames,
		rb.Spec.PipelineName,
	)

	if err := r.Update(ctx, eventListener); err != nil {
		return fmt.Errorf("failed to update EventListener: %w", err)
	}

	r.Log.V(1).Info("Updated EventListener namespace selector",
		"eventListener", constants.GitHubListenerName,
		"namespace", orgNamespace,
		"addedNamespace", rb.Spec.PipelineName)

	return nil
}

// provisionTriggerTemplate creates or updates the TriggerTemplate for pipeline execution.
func (r *RepoBindingReconciler) provisionTriggerTemplate(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during TriggerTemplate provisioning: %w", ctx.Err())
	default:
	}

	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)

	// Get template from catalog
	template, err := r.TemplateCatalog.Get(rb.Spec.TemplateRef)
	if err != nil {
		return fmt.Errorf("failed to get template from catalog: %w", err)
	}

	// Convert to TriggerTemplate and materialize in org namespace
	triggerTemplate := template.ToTriggerTemplate(orgNamespace, rb.Spec.AphexOrg)
	triggerTemplate.SetName(fmt.Sprintf("%s-trigger-template", rb.Spec.PipelineName))
	triggerTemplate.Labels[constants.LabelPipeline] = rb.Spec.PipelineName

	// Apply the TriggerTemplate
	if err := r.Patch(ctx, triggerTemplate, client.Apply, client.ForceOwnership, client.FieldOwner("platform-controller")); err != nil {
		return fmt.Errorf("failed to apply TriggerTemplate: %w", err)
	}

	r.Log.V(1).Info("Materialized template from catalog",
		"template", rb.Spec.TemplateRef,
		"namespace", orgNamespace,
		"name", triggerTemplate.Name)

	return nil
}

// provisionTrigger creates or updates the Trigger for pipeline execution.
func (r *RepoBindingReconciler) provisionTrigger(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled during Trigger provisioning: %w", ctx.Err())
	default:
	}

	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)

	// Find the pipeline's namespace
	pipelineNamespace, err := r.findPipelineNamespace(ctx, rb.Spec.PipelineName)
	if err != nil {
		return fmt.Errorf("failed to find pipeline %q: %w", rb.Spec.PipelineName, err)
	}

	repoFullName := fmt.Sprintf("%s/%s", rb.Spec.RepoOrg, rb.Spec.RepoName)

	trigger := &triggersv1beta1.Trigger{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "triggers.tekton.dev/v1beta1",
			Kind:       "Trigger",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-trigger", rb.Spec.PipelineName),
			Namespace: orgNamespace,
			Labels: map[string]string{
				constants.LabelPipeline:     rb.Spec.PipelineName,
				constants.LabelManagedBy:    constants.ManagedByPlatformController,
				constants.LabelOrganization: rb.Spec.AphexOrg,
			},
		},
		Spec: triggersv1beta1.TriggerSpec{
			Interceptors: []*triggersv1beta1.TriggerInterceptor{
				{
					Ref: triggersv1beta1.InterceptorRef{
						Name: "cel",
						Kind: triggersv1beta1.ClusterInterceptorKind,
					},
					Params: []triggersv1beta1.InterceptorParams{
						{
							Name:  "filter",
							Value: apiextensionsv1.JSON{Raw: []byte(fmt.Sprintf("%q", fmt.Sprintf("body.repository.full_name == '%s'", repoFullName)))},
						},
					},
				},
			},
			Bindings: []*triggersv1beta1.TriggerSpecBinding{
				{Ref: constants.GitHubPushBindingName},
				{Name: "pipeline-name", Value: stringPtr(rb.Spec.PipelineName)},
				{Name: "pipeline-namespace", Value: stringPtr(pipelineNamespace)},
				{Name: "org-name", Value: stringPtr(rb.Spec.AphexOrg)},
				{Name: "repo-full-name", Value: stringPtr(repoFullName)},
				{Name: "event-type", Value: stringPtr("$(header.X-Github-Event)")},
				{Name: "event-id", Value: stringPtr("$(header.X-Github-Delivery)")},
				{Name: "triggered-at", Value: stringPtr("$(body.repository.pushed_at)")},
			},
			Template: triggersv1beta1.TriggerSpecTemplate{
				Ref: stringPtr(fmt.Sprintf("%s-trigger-template", rb.Spec.PipelineName)),
			},
		},
	}

	if err := r.Patch(ctx, trigger, client.Apply, client.ForceOwnership, client.FieldOwner("platform-controller")); err != nil {
		return fmt.Errorf("failed to apply Trigger: %w", err)
	}

	r.Log.V(1).Info("Provisioned trigger with CEL filter",
		"trigger", trigger.Name,
		"template", *trigger.Spec.Template.Ref,
		"pipeline", rb.Spec.PipelineName,
		"repoFilter", repoFullName)

	return nil
}

// updateRepoBindingStatusWithWebhookInfo updates the RepoBinding status with webhook configuration details
func (r *RepoBindingReconciler) updateRepoBindingStatusWithWebhookInfo(ctx context.Context, rb *platformv1alpha1.RepoBinding) error {
	orgNamespace := fmt.Sprintf("org-%s", rb.Spec.AphexOrg)
	orgSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: constants.WebhookSecretName, Namespace: orgNamespace}, orgSecret)
	if err != nil {
		return fmt.Errorf("failed to get webhook secret from organization namespace %s: %w", orgNamespace, err)
	}

	pipelineSecret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Name: constants.WebhookSecretName, Namespace: rb.Spec.PipelineName}, pipelineSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			pipelineSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.WebhookSecretName,
					Namespace: rb.Spec.PipelineName,
					Labels: map[string]string{
						constants.LabelManagedBy:    constants.ManagedByPlatformController,
						constants.LabelOrganization: rb.Spec.AphexOrg,
					},
				},
				Type: orgSecret.Type,
				Data: orgSecret.Data,
			}

			if err := r.Create(ctx, pipelineSecret); err != nil {
				return fmt.Errorf("failed to copy webhook secret to pipeline namespace: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check webhook secret in pipeline namespace: %w", err)
		}
	}

	webhookSecret, ok := pipelineSecret.Data["secret"]
	if !ok {
		return fmt.Errorf("webhook secret missing 'secret' key")
	}

	rb.Status.WebhookURL = fmt.Sprintf("https://%s.%s", rb.Spec.RepoOrg, constants.WebhookDomain)
	rb.Status.WebhookSecret = string(webhookSecret)

	// Build configuration instructions message
	instructions := fmt.Sprintf(`Registration successful!

Next steps - Configure GitHub webhook:
1. Go to: https://github.com/%s/%s/settings/hooks/new
2. Payload URL: %s
3. Content type: application/json
4. Secret: %s
5. Events: Push events, Pull request events
6. Active: ✓
7. Click "Add webhook"`,
		rb.Spec.RepoOrg,
		rb.Spec.RepoName,
		rb.Status.WebhookURL,
		rb.Status.WebhookSecret)

	rb.Status.Message = instructions

	r.Log.Info("Updated RepoBinding status with webhook information",
		"tenant", rb.Spec.PipelineName,
		"webhookURL", rb.Status.WebhookURL)

	return nil
}

// findPipelineNamespace searches for a pipeline by name across all accessible namespaces
func (r *RepoBindingReconciler) findPipelineNamespace(ctx context.Context, pipelineName string) (string, error) {
	// List all namespaces
	namespaceList := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaceList); err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Search each namespace for the pipeline
	for i := range namespaceList.Items {
		namespace := namespaceList.Items[i].Name

		// Try to get the pipeline in this namespace using typed client
		pipeline := &tektonv1.Pipeline{}
		err := r.Get(ctx, client.ObjectKey{
			Name:      pipelineName,
			Namespace: namespace,
		}, pipeline)

		if err == nil {
			// Found it!
			return namespace, nil
		}
	}

	return "", fmt.Errorf("pipeline %q not found in any accessible namespace", pipelineName)
}
