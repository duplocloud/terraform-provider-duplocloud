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
	spec, err := flattenPodSpec(t.Spec)
	if err != nil {
		return []interface{}{template}, err
	}
	template["spec"] = spec

	return []interface{}{template}, nil
}

//func patchUpdateStrategyRollingUpdate(keyPrefix, pathPrefix string, d *schema.ResourceData) (PatchOperations, error) {
//	ops := PatchOperations{}
//	if d.HasChange(keyPrefix + "partition") {
//		log.Printf("[TRACE] StatefulSet.Spec.UpdateStrategy.RollingUpdate.Partition has changes")
//		if p, ok := d.Get(keyPrefix + "partition").(int); ok {
//			ops = append(ops, &ReplaceOperation{
//				Path:  pathPrefix + "partition",
//				Value: p,
//			})
//		}
//	}
//	return ops, nil
//}
