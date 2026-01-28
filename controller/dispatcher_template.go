package controller

import (
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DispatcherTemplate represents a thin dispatcher template that converts
// webhook events into PipelineRuns
type DispatcherTemplate struct {
	Name    string
	Version string
	Spec    triggersv1beta1.TriggerTemplateSpec
}

// NewRunPipelineV1 creates the run-pipeline-v1 dispatcher template
func NewRunPipelineV1() *DispatcherTemplate {
	return &DispatcherTemplate{
		Name:    "run-pipeline-v1",
		Version: "v1",
		Spec: triggersv1beta1.TriggerTemplateSpec{
			Params: []triggersv1beta1.ParamSpec{
				{Name: "git-url", Description: "Repository clone URL"},
				{Name: "git-revision", Description: "Git commit SHA or branch"},
				{Name: "repo-full-name", Description: "Repository name (org/repo)"},
				{Name: "event-type", Description: "Webhook event type"},
				{Name: "event-id", Description: "Unique event identifier"},
				{Name: "pipeline-name", Description: "Name of the pipeline to run"},
				{Name: "pipeline-namespace", Description: "Namespace where pipeline exists"},
				{Name: "org-name", Description: "Organization name", Default: stringPtr("")},
				{Name: "triggered-at", Description: "Webhook timestamp", Default: stringPtr("")},
			},
			ResourceTemplates: []triggersv1beta1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: buildPipelineRunTemplate(),
					},
				},
			},
		},
	}
}

// ToTriggerTemplate converts a DispatcherTemplate to a Tekton TriggerTemplate
func (t *DispatcherTemplate) ToTriggerTemplate(namespace, orgName string) *triggersv1beta1.TriggerTemplate {
	return &triggersv1beta1.TriggerTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "triggers.tekton.dev/v1beta1",
			Kind:       "TriggerTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: namespace,
			Labels: map[string]string{
				"platform.aphex/template-version": t.Version,
				"platform.aphex/managed-by":       "platform",
				"platform.aphex/organization":     orgName,
			},
		},
		Spec: t.Spec,
	}
}

// buildPipelineRunTemplate creates the PipelineRun resource template
func buildPipelineRunTemplate() []byte {
	return []byte(`{
			"apiVersion": "tekton.dev/v1",
			"kind": "PipelineRun",
			"metadata": {
				"generateName": "$(tt.params.pipeline-name)-",
				"namespace": "$(tt.params.pipeline-namespace)",
				"labels": {
					"platform.aphex/triggered": "true",
					"platform.aphex/event-type": "$(tt.params.event-type)",
					"platform.aphex/event-id": "$(tt.params.event-id)"
				}
			},
			"spec": {
				"pipelineRef": {
					"resolver": "cluster",
					"params": [
						{"name": "kind", "value": "pipeline"},
						{"name": "name", "value": "$(tt.params.pipeline-name)"},
						{"name": "namespace", "value": "$(tt.params.pipeline-namespace)"}
					]
				},
				"params": [
					{"name": "git-url", "value": "$(tt.params.git-url)"},
					{"name": "git-revision", "value": "$(tt.params.git-revision)"},
					{"name": "repo-full-name", "value": ["$(tt.params.repo-full-name)"]},
					{"name": "event-type", "value": "$(tt.params.event-type)"},
					{"name": "event-id", "value": "$(tt.params.event-id)"},
					{"name": "triggered-at", "value": "$(tt.params.triggered-at)"},
					{"name": "org-name", "value": "$(tt.params.org-name)"}
				],
				"workspaces": [
					{
						"name": "source",
						"volumeClaimTemplate": {
							"spec": {
								"accessModes": ["ReadWriteOnce"],
								"resources": {
									"requests": {
										"storage": "1Gi"
									}
								}
							}
						}
					}
				],
				"taskRunTemplate": {
					"serviceAccountName": "pipeline-runner"
				},
				"timeouts": {
					"pipeline": "1h"
				}
			}
		}`)
}
