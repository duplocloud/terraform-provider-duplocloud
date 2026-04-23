package duplocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func expandDaemonSetSpec(s []interface{}) (appsv1.DaemonSetSpec, error) {
	obj := appsv1.DaemonSetSpec{}
	if len(s) == 0 || s[0] == nil {
		return obj, nil
	}
	in := s[0].(map[string]interface{})

	if v, ok := in["min_ready_seconds"].(int); ok {
		obj.MinReadySeconds = int32(v)
	}

	if v, ok := in["revision_history_limit"].(int); ok {
		obj.RevisionHistoryLimit = ptrToInt32(int32(v))
	}

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		obj.Selector = expandLabelSelector(v)
	}

	template, err := expandPodTemplate(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = *template

	if v, ok := in["update_strategy"].([]interface{}); ok && len(v) > 0 {
		obj.UpdateStrategy = expandDaemonSetUpdateStrategy(v)
	}

	return obj, nil
}

func expandDaemonSetUpdateStrategy(s []interface{}) appsv1.DaemonSetUpdateStrategy {
	obj := appsv1.DaemonSetUpdateStrategy{}
	if len(s) == 0 || s[0] == nil {
		return obj
	}
	in := s[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && v != "" {
		obj.Type = appsv1.DaemonSetUpdateStrategyType(v)
	}

	if v, ok := in["rolling_update"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		ru := v[0].(map[string]interface{})
		rollingUpdate := &appsv1.RollingUpdateDaemonSet{}
		if mu, ok := ru["max_unavailable"].(string); ok && mu != "" {
			parsed := intstr.Parse(mu)
			rollingUpdate.MaxUnavailable = &parsed
		}
		if ms, ok := ru["max_surge"].(string); ok && ms != "" {
			parsed := intstr.Parse(ms)
			rollingUpdate.MaxSurge = &parsed
		}
		obj.RollingUpdate = rollingUpdate
	}

	return obj
}

func flattenDaemonSetSpec(in appsv1.DaemonSetSpec, d *schema.ResourceData, meta interface{}) ([]interface{}, error) {
	att := make(map[string]interface{})

	att["min_ready_seconds"] = in.MinReadySeconds

	if in.RevisionHistoryLimit != nil {
		att["revision_history_limit"] = *in.RevisionHistoryLimit
	}

	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}

	podSpec, err := flattenPodTemplateSpec(in.Template, d, meta, "spec.0.template.0.")
	if err != nil {
		return nil, err
	}
	att["template"] = podSpec

	att["update_strategy"] = flattenDaemonSetUpdateStrategy(in.UpdateStrategy)

	return []interface{}{att}, nil
}

func flattenDaemonSetUpdateStrategy(in appsv1.DaemonSetUpdateStrategy) []interface{} {
	att := make(map[string]interface{})

	att["type"] = string(in.Type)

	if in.RollingUpdate != nil {
		// Always emit both fields so state matches schema defaults when the backend
		// omits zero-value fields (Kubernetes drops MaxSurge=0 on round-trip).
		ru := map[string]interface{}{
			"max_unavailable": "1",
			"max_surge":       "0",
		}
		if in.RollingUpdate.MaxUnavailable != nil {
			ru["max_unavailable"] = in.RollingUpdate.MaxUnavailable.String()
		}
		if in.RollingUpdate.MaxSurge != nil {
			ru["max_surge"] = in.RollingUpdate.MaxSurge.String()
		}
		att["rolling_update"] = []interface{}{ru}
	}

	return []interface{}{att}
}
