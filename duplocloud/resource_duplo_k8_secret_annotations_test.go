package duplocloud

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// DUPLO-44791.  The backend writes duplocloud.net/skip-encoding onto a secret the first
// time it is updated, so a configuration that sets secret_annotations explicitly would
// plan the key's removal on every run.

func TestK8sSecretAnnotationsForState_DropsSkipEncodingTheConfigDoesNotMention(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}

	assert.Equal(t, map[string]string{}, k8sSecretAnnotationsForState(map[string]interface{}{}, backend), "empty map")
	assert.Equal(t, map[string]string{}, k8sSecretAnnotationsForState(nil, backend), "omitted")

	backend = map[string]string{"n": "1", "duplocloud.net/skip-encoding": "False"}
	assert.Equal(t, map[string]string{"n": "1"}, k8sSecretAnnotationsForState(map[string]interface{}{"n": "1"}, backend), "other keys kept")
}

func TestK8sSecretAnnotationsForState_KeepsSkipEncodingTheConfigSets(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "true"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, backend, k8sSecretAnnotationsForState(configured, backend))
}

func TestK8sSecretAnnotationsForState_LeavesTheBackendMapAlone(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}

	k8sSecretAnnotationsForState(nil, backend)

	assert.Contains(t, backend, "duplocloud.net/skip-encoding")
}

// The backend rewrites the annotation as bool.ToString() on every update, so a configured
// "true" comes back as "True".  It parses the value case-insensitively, so the two agree.
func TestK8sSecretAnnotationsForState_KeepsTheConfiguredSpellingOfSkipEncoding(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "True"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, map[string]string{"duplocloud.net/skip-encoding": "true"}, k8sSecretAnnotationsForState(configured, backend))
}

func TestK8sSecretAnnotationsForState_ReportsADifferentSkipEncodingValue(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, backend, k8sSecretAnnotationsForState(configured, backend))
}

// State written before Read reconciled the annotation still holds the backend's value.
// Read cannot drop it then, since it sees prior state rather than configuration, so the
// plan has to settle it against the raw configuration instead.

func annotationsPlan(t *testing.T, state, config cty.Value) *terraform.InstanceDiff {
	t.Helper()
	data := cty.StringVal(`{"k":"v"}`)
	return driftPlan(t,
		driftObjectWith(driftState(data), "secret_annotations", state),
		driftObjectWith(driftConfig(data), "secret_annotations", config))
}

func assertNoAnnotationsDiff(t *testing.T, diff *terraform.InstanceDiff) {
	t.Helper()
	if diff == nil {
		return
	}
	for k := range diff.Attributes {
		assert.NotContains(t, k, "secret_annotations")
	}
}

func driftObjectWith(obj cty.Value, name string, v cty.Value) cty.Value {
	vals := obj.AsValueMap()
	vals[name] = v
	return cty.ObjectVal(vals)
}

func TestK8sSecretAnnotationsPlan_StaleBackendSkipEncodingConverges(t *testing.T) {
	state := cty.MapVal(map[string]cty.Value{"n": cty.StringVal("1"), "duplocloud.net/skip-encoding": cty.StringVal("False")})
	config := cty.MapVal(map[string]cty.Value{"n": cty.StringVal("1")})

	diff := annotationsPlan(t, state, config)

	assertNoAnnotationsDiff(t, diff)
}

func TestK8sSecretAnnotationsPlan_StaleSkipEncodingSpellingConverges(t *testing.T) {
	state := cty.MapVal(map[string]cty.Value{"duplocloud.net/skip-encoding": cty.StringVal("True")})
	config := cty.MapVal(map[string]cty.Value{"duplocloud.net/skip-encoding": cty.StringVal("true")})

	diff := annotationsPlan(t, state, config)

	assertNoAnnotationsDiff(t, diff)
}

func TestK8sSecretAnnotationsPlan_RealAnnotationChangeStillPlans(t *testing.T) {
	state := cty.MapVal(map[string]cty.Value{"n": cty.StringVal("1"), "duplocloud.net/skip-encoding": cty.StringVal("False")})
	config := cty.MapVal(map[string]cty.Value{"n": cty.StringVal("2")})

	diff := annotationsPlan(t, state, config)

	assert.Contains(t, diff.Attributes, "secret_annotations.n")
}

func TestK8sSecretAnnotationsPlan_ConfiguredSkipEncodingValueChangeStillPlans(t *testing.T) {
	state := cty.MapVal(map[string]cty.Value{"duplocloud.net/skip-encoding": cty.StringVal("False")})
	config := cty.MapVal(map[string]cty.Value{"duplocloud.net/skip-encoding": cty.StringVal("true")})

	diff := annotationsPlan(t, state, config)

	assert.Contains(t, diff.Attributes, "secret_annotations.duplocloud.net/skip-encoding")
}
