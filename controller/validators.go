package controller

import (
	"fmt"
	"regexp"
	"strings"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
)

var (
	// Approved GitHub organizations
	approvedOrgs = []string{
		"bdchatham",
	}

	// Privileged namespace names that cannot be used for tenants
	privilegedNamespaces = []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default",
		"pipeline-system",
		"pipeline-catalog",
		"auth-system",
		"tekton-pipelines",
		"tekton-pipelines-resolvers",
	}

	// Valid namespace pattern
	namespacePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

	// Valid pipeline name pattern
	pipelineNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateRepoBinding validates a RepoBinding spec
func ValidateRepoBinding(rb *platformv1alpha1.RepoBinding, catalog *TemplateCatalog) error {
	// Validate repository organization
	if err := validateOrganization(rb.Spec.RepoOrg); err != nil {
		return err
	}

	// Validate namespace name pattern
	if err := validateNamespacePattern(rb.Spec.PipelineName); err != nil {
		return err
	}

	// Reject privileged namespace names
	if err := validateNotPrivilegedNamespace(rb.Spec.PipelineName); err != nil {
		return err
	}

	// Validate template reference
	if err := validateTemplateRef(rb.Spec.TemplateRef, catalog); err != nil {
		return err
	}

	// Validate pipeline name
	return validatePipelineName(rb.Spec.PipelineName)
}

// validateOrganization checks if the repository organization is in the approved list
func validateOrganization(org string) error {
	for _, approvedOrg := range approvedOrgs {
		if org == approvedOrg {
			return nil
		}
	}
	return &ValidationError{
		Field:   "repoOrg",
		Message: fmt.Sprintf("Repository organization '%s' not in approved list", org),
	}
}

// validateNamespacePattern checks if the namespace name matches the required pattern
func validateNamespacePattern(name string) error {
	if !namespacePattern.MatchString(name) {
		return &ValidationError{
			Field:   "pipelineName",
			Message: "Namespace name must match pattern ^[a-z0-9-]+$",
		}
	}
	return nil
}

// validateNotPrivilegedNamespace checks if the namespace name is not a privileged namespace
func validateNotPrivilegedNamespace(name string) error {
	for _, privileged := range privilegedNamespaces {
		if name == privileged || strings.HasPrefix(name, privileged+"-") {
			return &ValidationError{
				Field:   "pipelineName",
				Message: fmt.Sprintf("Cannot create namespace with privileged name '%s'", name),
			}
		}
	}
	return nil
}

// validateTemplateRef checks if the template exists in the catalog
func validateTemplateRef(templateRef string, catalog *TemplateCatalog) error {
	if templateRef == "" {
		return &ValidationError{
			Field:   "templateRef",
			Message: "Template reference is required",
		}
	}

	if !catalog.Exists(templateRef) {
		return &ValidationError{
			Field:   "templateRef",
			Message: fmt.Sprintf("Template %q not found in catalog (available: %v)", templateRef, catalog.List()),
		}
	}

	return nil
}

// validatePipelineName checks if the pipeline name matches the required pattern
func validatePipelineName(name string) error {
	if name == "" {
		return &ValidationError{
			Field:   "pipelineName",
			Message: "Pipeline name is required",
		}
	}

	if !pipelineNamePattern.MatchString(name) {
		return &ValidationError{
			Field:   "pipelineName",
			Message: "Pipeline name must match pattern ^[a-z0-9-]+$",
		}
	}
	return nil
}
