package duplocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
)

// Expanders

func flattenPodTemplateSpec(t corev1.PodTemplateSpec, d *schema.ResourceData, meta interface{}, prefix ...string) ([]interface{}, error) {
	template := make(map[string]interface{})

	jobPrefix := "spec.0.template.0." //schema tree from job
	if len(prefix) > 0 {
		jobPrefix = prefix[0]
	}
	//spec.0.job_template.0.spec.0.template.0.spec
	template["metadata"] = flattenMetadata(t.ObjectMeta, d, meta, jobPrefix)
	var podSpec interface{}
	if len(prefix) > 0 {
		podSpec = d.Get(prefix[0] + "spec")
	} else {
		podSpec = d.Get(jobPrefix + "spec")
	}
	spec, err := flattenPodSpec(t.Spec, podSpec)
	if err != nil {
		return []interface{}{template}, err
	}
	template["spec"] = spec

	return []interface{}{template}, nil
}
