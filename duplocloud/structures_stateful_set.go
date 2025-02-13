package duplocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	corev1 "k8s.io/api/core/v1"
)

// Expanders

func flattenPodTemplateSpec(t corev1.PodTemplateSpec, d *schema.ResourceData, meta interface{}, prefix ...string) ([]interface{}, error) {
	template := make(map[string]interface{})

	metaPrefix := "spec.0.template.0."
	if len(prefix) > 0 {
		metaPrefix = prefix[0]
	}
	template["metadata"] = flattenMetadata(t.ObjectMeta, d, meta, metaPrefix)
	spec, err := flattenPodSpec(t.Spec, d)
	if err != nil {
		return []interface{}{template}, err
	}
	template["spec"] = spec

	return []interface{}{template}, nil
}
