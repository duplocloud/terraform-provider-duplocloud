package duplocloud

import (
	"strconv"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	batchv1 "k8s.io/api/batch/v1"
)

func flattenK8sJob(d *schema.ResourceData, duplo *duplosdk.DuploK8sJob, meta interface{}) error {
	d.Set("tenant_id", duplo.TenantId)
	d.Set("metadata", flattenMetadata(duplo.Metadata, d, meta))
	jobSpec, err := flattenJobV1Spec(duplo.Spec, d, meta)
	if err != nil {
		return err
	}
	d.Set("spec", jobSpec)

	return nil
}

func flattenJobV1Spec(in batchv1.JobSpec, d *schema.ResourceData, meta interface{}, prefix ...string) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.ActiveDeadlineSeconds != nil {
		att["active_deadline_seconds"] = *in.ActiveDeadlineSeconds
	}

	if in.BackoffLimit != nil {
		att["backoff_limit"] = *in.BackoffLimit
	}

	if in.Completions != nil {
		att["completions"] = *in.Completions
	}

	if in.CompletionMode != nil {
		att["completion_mode"] = string(*in.CompletionMode)
	}

	if in.ManualSelector != nil {
		att["manual_selector"] = *in.ManualSelector
	}

	if in.Parallelism != nil {
		att["parallelism"] = *in.Parallelism
	}

	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}
	// Remove server-generated labels
	labels := in.Template.ObjectMeta.Labels

	delete(labels, "controller-uid")

	delete(labels, "job-name")

	podSpec, err := flattenPodTemplateSpec(in.Template, d, meta, prefix...)
	if err != nil {
		return nil, err
	}
	att["template"] = podSpec

	if in.TTLSecondsAfterFinished != nil {
		att["ttl_seconds_after_finished"] = strconv.Itoa(int(*in.TTLSecondsAfterFinished))
	}

	return []interface{}{att}, nil
}

func expandJobV1Spec(j []interface{}) (batchv1.JobSpec, error) {
	obj := batchv1.JobSpec{}

	if len(j) == 0 || j[0] == nil {
		return obj, nil
	}

	in := j[0].(map[string]interface{})

	if v, ok := in["active_deadline_seconds"].(int); ok && v > 0 {
		obj.ActiveDeadlineSeconds = ptrToInt64(int64(v))
	}

	if v, ok := in["backoff_limit"].(int); ok && v >= 0 {
		obj.BackoffLimit = ptrToInt32(int32(v))
	}

	if v, ok := in["completions"].(int); ok && v > 0 {
		obj.Completions = ptrToInt32(int32(v))
	}

	if v, ok := in["completion_mode"].(string); ok && v != "" {
		m := batchv1.CompletionMode(v)
		obj.CompletionMode = &m
	}

	if v, ok := in["manual_selector"]; ok {
		obj.ManualSelector = ptrToBool(v.(bool))
	}

	if v, ok := in["parallelism"].(int); ok && v >= 0 {
		obj.Parallelism = ptrToInt32(int32(v))
	}

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		obj.Selector = expandLabelSelector(v)
	}

	template, err := expandPodTemplate(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = *template

	if v, ok := in["ttl_seconds_after_finished"].(string); ok && v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return obj, err
		}
		obj.TTLSecondsAfterFinished = ptrToInt32(int32(i))
	}

	return obj, nil
}
