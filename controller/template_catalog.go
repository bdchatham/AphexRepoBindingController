package controller

import (
	"fmt"
)

// TemplateCatalog manages platform-provided dispatcher templates
type TemplateCatalog struct {
	templates map[string]*DispatcherTemplate
}

// NewTemplateCatalog creates a new template catalog with all available templates
func NewTemplateCatalog() *TemplateCatalog {
	return &TemplateCatalog{
		templates: map[string]*DispatcherTemplate{
			"run-pipeline-v1": NewRunPipelineV1(),
		},
	}
}

// Get retrieves a template by name
func (c *TemplateCatalog) Get(name string) (*DispatcherTemplate, error) {
	template, ok := c.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %q not found in catalog (available: %v)", name, c.List())
	}
	return template, nil
}

// List returns all available template names
func (c *TemplateCatalog) List() []string {
	names := make([]string, 0, len(c.templates))
	for name := range c.templates {
		names = append(names, name)
	}
	return names
}

// Exists checks if a template exists in the catalog
func (c *TemplateCatalog) Exists(name string) bool {
	_, ok := c.templates[name]
	return ok
}
