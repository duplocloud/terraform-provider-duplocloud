package duplocloud

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func expandStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}

func expandMetadata(in []interface{}) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}
	if len(in) == 0 || in[0] == nil {
		return meta
	}

	m := in[0].(map[string]interface{})

	if v, ok := m["annotations"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Annotations = expandStringMap(m["annotations"].(map[string]interface{}))
	}

	if v, ok := m["labels"].(map[string]interface{}); ok && len(v) > 0 {
		meta.Labels = expandStringMap(m["labels"].(map[string]interface{}))
	}

	if v, ok := m["generate_name"]; ok {
		meta.GenerateName = v.(string)
	}
	if v, ok := m["name"]; ok {
		meta.Name = v.(string)
	}
	if v, ok := m["namespace"]; ok {
		meta.Namespace = v.(string)
	}

	return meta
}

func flattenMetadata(meta metav1.ObjectMeta, d *schema.ResourceData, providerMetadata interface{}, metaPrefix ...string) []interface{} {
	m := make(map[string]interface{})
	prefix := ""
	if len(metaPrefix) > 0 {
		prefix = metaPrefix[0]
	}

	m["annotations"] = d.Get(prefix + "metadata.0.annotations").(map[string]interface{})

	if meta.GenerateName != "" {
		m["generate_name"] = meta.GenerateName
	}

	configLabels := d.Get(prefix + "metadata.0.labels").(map[string]interface{})
	labels := removeInternalKeys(meta.Labels, configLabels)
	// we can pass a second argument to removeKeys to ignore a set of keys
	m["labels"] = removeKeys(labels, configLabels, nil)
	m["name"] = meta.Name
	m["resource_version"] = meta.ResourceVersion
	m["uid"] = fmt.Sprintf("%v", meta.UID)
	m["generation"] = meta.Generation

	if meta.Namespace != "" {
		m["namespace"] = meta.Namespace
	}

	return []interface{}{m}
}

func removeInternalKeys(m map[string]string, d map[string]interface{}) map[string]string {
	for k := range m {
		if isInternalKey(k) && !isKeyInMap(k, d) {
			delete(m, k)
		}
	}
	return m
}

// removeKeys removes given Kubernetes metadata(annotations and labels) keys.
// In that case, they won't be available in the TF state file and will be ignored during apply/plan operations.
func removeKeys(m map[string]string, d map[string]interface{}, ignoreKubernetesMetadataKeys []string) map[string]string {
	for k := range m {
		if ignoreKey(k, ignoreKubernetesMetadataKeys) && !isKeyInMap(k, d) {
			delete(m, k)
		}
	}
	return m
}

func isKeyInMap(key string, d map[string]interface{}) bool {
	if d == nil {
		return false
	}
	for k := range d {
		if k == key {
			return true
		}
	}
	return false
}

func isInternalKey(annotationKey string) bool {
	u, err := url.Parse("//" + annotationKey)
	if err != nil {
		return false
	}

	// allow user specified application specific keys
	if u.Hostname() == "app.kubernetes.io" {
		return false
	}

	// allow AWS load balancer configuration annotations
	if u.Hostname() == "service.beta.kubernetes.io" {
		return false
	}

	// internal *.kubernetes.io keys
	if strings.HasSuffix(u.Hostname(), "kubernetes.io") {
		return true
	}

	// Specific to DaemonSet annotations, generated & controlled by the server.
	if strings.Contains(annotationKey, "deprecated.daemonset.template.generation") {
		return true
	}
	return false
}

// ignoreKey reports whether the Kubernetes metadata(annotations and labels) key contains
// any match of the regular expression pattern from the expressions slice.
func ignoreKey(key string, expressions []string) bool {
	for _, e := range expressions {
		if ok, _ := regexp.MatchString(e, key); ok {
			return true
		}
	}

	return false
}

func ptrToString(s string) *string {
	return &s
}

// nolint
func ptrToBool(b bool) *bool {
	return &b
}

// nolint
func ptrToInt32(i int32) *int32 {
	return &i
}

// nolint
func ptrToInt64(i int64) *int64 {
	return &i
}

func sliceOfString(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = s.(string)
	}
	return result
}

func flattenResourceList(l api.ResourceList) map[string]string {
	m := make(map[string]string)
	for k, v := range l {
		m[string(k)] = v.String()
	}
	return m
}

func newStringSet(f schema.SchemaSetFunc, in []string) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return schema.NewSet(f, out)
}
func newInt64Set(f schema.SchemaSetFunc, in []int64) *schema.Set {
	var out = make([]interface{}, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return schema.NewSet(f, out)
}

func flattenLocalObjectReferenceArray(in []api.LocalObjectReference) []interface{} {
	att := []interface{}{}
	for _, v := range in {
		m := map[string]interface{}{
			"name": v.Name,
		}
		att = append(att, m)
	}
	return att
}

func flattenNodeSelectorRequirementList(in []api.NodeSelectorRequirement) []map[string]interface{} {
	att := make([]map[string]interface{}, len(in))
	for i, v := range in {
		m := map[string]interface{}{}
		m["key"] = v.Key
		m["values"] = newStringSet(schema.HashString, v.Values)
		m["operator"] = string(v.Operator)
		att[i] = m
	}
	return att
}

func flattenNodeSelectorTerm(in api.NodeSelectorTerm) []interface{} {
	att := make(map[string]interface{})
	if len(in.MatchExpressions) > 0 {
		att["match_expressions"] = flattenNodeSelectorRequirementList(in.MatchExpressions)
	}
	if len(in.MatchFields) > 0 {
		att["match_fields"] = flattenNodeSelectorRequirementList(in.MatchFields)
	}
	return []interface{}{att}
}

func flattenNodeSelectorTerms(in []api.NodeSelectorTerm) []interface{} {
	att := make([]interface{}, len(in))
	for i, n := range in {
		att[i] = flattenNodeSelectorTerm(n)[0]
	}
	return att
}

func expandNodeSelectorTerm(l []interface{}) *api.NodeSelectorTerm {
	if len(l) == 0 || l[0] == nil {
		return &api.NodeSelectorTerm{}
	}
	in := l[0].(map[string]interface{})
	obj := api.NodeSelectorTerm{}
	if v, ok := in["match_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchExpressions = expandNodeSelectorRequirementList(v)
	}
	if v, ok := in["match_fields"].([]interface{}); ok && len(v) > 0 {
		obj.MatchFields = expandNodeSelectorRequirementList(v)
	}
	return &obj
}

func expandNodeSelectorTerms(l []interface{}) []api.NodeSelectorTerm {
	if len(l) == 0 || l[0] == nil {
		return []api.NodeSelectorTerm{}
	}
	obj := make([]api.NodeSelectorTerm, len(l))
	for i, n := range l {
		obj[i] = *expandNodeSelectorTerm([]interface{}{n})
	}
	return obj
}

func expandNodeSelectorRequirementList(in []interface{}) []api.NodeSelectorRequirement {
	att := []api.NodeSelectorRequirement{}
	if len(in) < 1 {
		return att
	}
	att = make([]api.NodeSelectorRequirement, len(in))
	for i, c := range in {
		p := c.(map[string]interface{})
		att[i].Key = p["key"].(string)
		att[i].Operator = api.NodeSelectorOperator(p["operator"].(string))
		att[i].Values = expandStringSlice(p["values"].(*schema.Set).List())
	}
	return att
}

func expandStringSlice(s []interface{}) []string {
	result := make([]string, len(s))
	for k, v := range s {
		// Handle the Terraform parser bug which turns empty strings in lists to nil.
		if v == nil {
			result[k] = ""
		} else {
			result[k] = v.(string)
		}
	}
	return result
}

func expandMapToResourceList(m map[string]interface{}) (*api.ResourceList, error) {
	out := make(api.ResourceList)
	for stringKey, origValue := range m {
		key := api.ResourceName(stringKey)
		var value resource.Quantity

		if v, ok := origValue.(int); ok {
			q := resource.NewQuantity(int64(v), resource.DecimalExponent)
			value = *q
		} else if v, ok := origValue.(string); ok {
			var err error
			value, err = resource.ParseQuantity(v)
			if err != nil {
				return &out, err
			}
		} else {
			return &out, fmt.Errorf("Unexpected value type: %#v", origValue)
		}

		out[key] = value
	}
	return &out, nil
}
