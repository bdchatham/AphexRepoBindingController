// Package webhooks provides validating admission webhooks for platform CRDs.
// These webhooks reject invalid resources at admission time, preventing bad
// configurations from entering the cluster.
package webhooks

import (
	"fmt"
	"regexp"
	"strings"
)

// Common validation patterns
var (
	// namespacePattern validates Kubernetes namespace names
	namespacePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

	// githubOrgPattern validates GitHub organization names
	githubOrgPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)

	// githubRepoPattern validates GitHub repository names
	githubRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

	// privilegedNamespaces that cannot be used for tenant namespaces
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
		"argocd",
		"platform-system",
	}
)

// WebhookValidationError represents a validation failure with field context.
type WebhookValidationError struct {
	Field   string
	Message string
}

func (e *WebhookValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// newValidationError creates a new WebhookValidationError.
func newValidationError(field, message string) *WebhookValidationError {
	return &WebhookValidationError{Field: field, Message: message}
}

// isPrivilegedNamespace checks if a namespace name is privileged.
func isPrivilegedNamespace(name string) bool {
	for _, ns := range privilegedNamespaces {
		if name == ns || strings.HasPrefix(name, ns+"-") {
			return true
		}
	}
	return false
}
