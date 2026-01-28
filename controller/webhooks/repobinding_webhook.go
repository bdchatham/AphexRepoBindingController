package webhooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
)

// RepoBindingValidator validates RepoBinding resources at admission time.
// It implements webhook.CustomValidator interface.
type RepoBindingValidator struct {
	logger       logr.Logger
	approvedOrgs []string
}

// NewRepoBindingValidator creates a new RepoBindingValidator.
func NewRepoBindingValidator(logger logr.Logger, approvedOrgs []string) *RepoBindingValidator {
	return &RepoBindingValidator{
		logger:       logger.WithName("RepoBindingValidator"),
		approvedOrgs: approvedOrgs,
	}
}

// SetupWebhookWithManager registers the webhook with the manager.
func (v *RepoBindingValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&platformv1alpha1.RepoBinding{}).
		WithValidator(v).
		Complete()
}

// ValidateCreate validates a RepoBinding on creation.
func (v *RepoBindingValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	rb, ok := obj.(*platformv1alpha1.RepoBinding)
	if !ok {
		return nil, fmt.Errorf("expected RepoBinding, got %T", obj)
	}

	v.logger.V(1).Info("Validating RepoBinding creation",
		"name", rb.Name,
		"namespace", rb.Namespace,
	)

	if err := v.validateSpec(rb); err != nil {
		v.logger.Info("RepoBinding validation failed",
			"name", rb.Name,
			"error", err.Error(),
		)
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate validates a RepoBinding on update.
func (v *RepoBindingValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	rb, ok := newObj.(*platformv1alpha1.RepoBinding)
	if !ok {
		return nil, fmt.Errorf("expected RepoBinding, got %T", newObj)
	}

	v.logger.V(1).Info("Validating RepoBinding update",
		"name", rb.Name,
		"namespace", rb.Namespace,
	)

	if err := v.validateSpec(rb); err != nil {
		v.logger.Info("RepoBinding validation failed",
			"name", rb.Name,
			"error", err.Error(),
		)
		return nil, err
	}

	// Validate immutable fields
	oldRB, ok := oldObj.(*platformv1alpha1.RepoBinding)
	if !ok {
		return nil, fmt.Errorf("expected RepoBinding, got %T", oldObj)
	}

	if err := v.validateImmutableFields(oldRB, rb); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateDelete validates a RepoBinding on deletion.
func (v *RepoBindingValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validateSpec validates the RepoBinding spec fields.
func (v *RepoBindingValidator) validateSpec(rb *platformv1alpha1.RepoBinding) error {
	// Validate AphexOrg (required)
	if rb.Spec.AphexOrg == "" {
		return newValidationError("spec.aphexOrg", "aphexOrg is required")
	}
	if !namespacePattern.MatchString(rb.Spec.AphexOrg) {
		return newValidationError("spec.aphexOrg",
			"aphexOrg must be a valid Kubernetes name (lowercase alphanumeric and hyphens)")
	}

	// Validate RepoOrg (required)
	if rb.Spec.RepoOrg == "" {
		return newValidationError("spec.repoOrg", "repoOrg is required")
	}
	if !githubOrgPattern.MatchString(rb.Spec.RepoOrg) {
		return newValidationError("spec.repoOrg",
			"repoOrg must be a valid GitHub organization name")
	}

	// Validate RepoOrg is in approved list
	if err := v.validateApprovedOrg(rb.Spec.RepoOrg); err != nil {
		return err
	}

	// Validate RepoName (required)
	if rb.Spec.RepoName == "" {
		return newValidationError("spec.repoName", "repoName is required")
	}
	if !githubRepoPattern.MatchString(rb.Spec.RepoName) {
		return newValidationError("spec.repoName",
			"repoName must be a valid GitHub repository name")
	}

	// Validate PipelineName (required)
	if rb.Spec.PipelineName == "" {
		return newValidationError("spec.pipelineName", "pipelineName is required")
	}
	if !namespacePattern.MatchString(rb.Spec.PipelineName) {
		return newValidationError("spec.pipelineName",
			"pipelineName must be a valid Kubernetes name (lowercase alphanumeric and hyphens)")
	}

	// Validate PipelineName is not a privileged namespace
	if isPrivilegedNamespace(rb.Spec.PipelineName) {
		return newValidationError("spec.pipelineName",
			fmt.Sprintf("pipelineName %q is a privileged namespace and cannot be used", rb.Spec.PipelineName))
	}

	// Validate TemplateRef (required)
	if rb.Spec.TemplateRef == "" {
		return newValidationError("spec.templateRef", "templateRef is required")
	}

	// Validate PipelineSpec (required)
	if rb.Spec.PipelineSpec == "" {
		return newValidationError("spec.pipelineSpec", "pipelineSpec is required")
	}

	// Validate PipelineSpec is valid YAML with expected structure
	return v.validatePipelineSpec(rb.Spec.PipelineSpec)
}

// validateApprovedOrg checks if the organization is in the approved list.
func (v *RepoBindingValidator) validateApprovedOrg(org string) error {
	for _, approved := range v.approvedOrgs {
		if org == approved {
			return nil
		}
	}
	return newValidationError("spec.repoOrg",
		fmt.Sprintf("organization %q is not in the approved list", org))
}

// validatePipelineSpec validates the pipeline spec YAML structure.
func (v *RepoBindingValidator) validatePipelineSpec(spec string) error {
	// Basic validation - check for required YAML structure markers
	if !strings.Contains(spec, "apiVersion:") {
		return newValidationError("spec.pipelineSpec", "pipelineSpec must contain apiVersion")
	}
	if !strings.Contains(spec, "kind:") {
		return newValidationError("spec.pipelineSpec", "pipelineSpec must contain kind")
	}
	if !strings.Contains(spec, "Pipeline") {
		return newValidationError("spec.pipelineSpec", "pipelineSpec must define a Pipeline resource")
	}

	return nil
}

// validateImmutableFields checks that immutable fields have not changed.
func (v *RepoBindingValidator) validateImmutableFields(oldRB, newRB *platformv1alpha1.RepoBinding) error {
	// PipelineName is immutable as it determines the namespace
	if oldRB.Spec.PipelineName != newRB.Spec.PipelineName {
		return newValidationError("spec.pipelineName",
			"pipelineName is immutable and cannot be changed after creation")
	}

	// AphexOrg is immutable as it determines organization binding
	if oldRB.Spec.AphexOrg != newRB.Spec.AphexOrg {
		return newValidationError("spec.aphexOrg",
			"aphexOrg is immutable and cannot be changed after creation")
	}

	return nil
}
