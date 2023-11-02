package duplocloud

import corev1 "k8s.io/api/core/v1"

func expandPodTemplate(l []interface{}) (*corev1.PodTemplateSpec, error) {
	obj := &corev1.PodTemplateSpec{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})

	obj.ObjectMeta = expandMetadata(in["metadata"].([]interface{}))

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		podSpec, err := expandPodSpec(in["spec"].([]interface{}))
		if err != nil {
			return obj, err
		}
		obj.Spec = *podSpec
	}
	return obj, nil
}
