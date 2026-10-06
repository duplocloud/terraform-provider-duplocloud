package duplocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// DUPLO-44696.  A duplocloud_k8_secret that omits secret_data never converges: the apply
// sends an empty SecretData, Read writes the literal "{}" into state, and the next plan
// compares that against the "" an absent attribute produces.  Deliberately named apart
// from the fixtures in resource_duplo_k8_secret_test.go so the two files can coexist.

const driftSecretID = "v2/subscriptions/a1b2c3d4-0000-4000-8000-000000000001/K8SecretApiV2/mysecret"

// driftObject builds a value of the resource's own type, defaulting unnamed attributes to
// null - which is what Terraform sends for an attribute the configuration omits.
func driftObject(attrs map[string]cty.Value) cty.Value {
	vals := map[string]cty.Value{}
	for name, ty := range resourceK8Secret().CoreConfigSchema().ImpliedType().AttributeTypes() {
		if v, ok := attrs[name]; ok {
			vals[name] = v
		} else {
			vals[name] = cty.NullVal(ty)
		}
	}
	return cty.ObjectVal(vals)
}

func driftState(secretData cty.Value) cty.Value {
	return driftObject(map[string]cty.Value{
		"id":          cty.StringVal(driftSecretID),
		"tenant_id":   cty.StringVal("a1b2c3d4-0000-4000-8000-000000000001"),
		"secret_name": cty.StringVal("mysecret"),
		"secret_type": cty.StringVal("Opaque"),
		"secret_data": secretData,
	})
}

func driftConfig(secretData cty.Value) cty.Value {
	return driftObject(map[string]cty.Value{
		"tenant_id":   cty.StringVal("a1b2c3d4-0000-4000-8000-000000000001"),
		"secret_name": cty.StringVal("mysecret"),
		"secret_type": cty.StringVal("Opaque"),
		"secret_data": secretData,
	})
}

// driftPlan drives the real plan path, so DiffSuppressFunc is exercised as Terraform
// calls it rather than in isolation.
func driftPlan(t *testing.T, state, config cty.Value) *terraform.InstanceDiff {
	t.Helper()
	r := resourceK8Secret()

	is, err := r.ShimInstanceStateFromValue(state)
	assert.NoError(t, err)
	is.ID = driftSecretID
	is.RawState = state
	is.RawConfig = config
	is.RawPlan = config

	diff, err := r.SimpleDiff(context.Background(), is, terraform.NewResourceConfigShimmed(config, r.CoreConfigSchema()), nil)
	assert.NoError(t, err)
	return diff
}

// --- the bug ---------------------------------------------------------------

// What Read leaves behind after an apply that sent an empty SecretData, against a
// configuration that never mentioned secret_data.  These are the same secret.
func TestSecretDataCompare_EmptyStringMatchesEmptyObject(t *testing.T) {
	equal, err := secretDataCompare("{}", "")

	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestSecretDataCompare_EmptyStringMatchesEmptyString(t *testing.T) {
	equal, err := secretDataCompare("", "")

	assert.NoError(t, err)
	assert.True(t, equal)
}

// The plan-level proof: omitting secret_data must settle instead of planning the same
// change on every run.
func TestSecretDataDiff_OmittedDataAgainstAnEmptySecretConverges(t *testing.T) {
	diff := driftPlan(t, driftState(cty.StringVal("{}")), driftConfig(cty.NullVal(cty.String)))

	assert.NotContains(t, diff.Attributes, "secret_data")
}

// --- what must keep diffing ------------------------------------------------

// Removing secret_data from a populated secret is a real change: the user is emptying it.
func TestSecretDataCompare_EmptyStringDoesNotMatchPopulated(t *testing.T) {
	equal, err := secretDataCompare(`{"DB_PASSWORD":"hunter2"}`, "")

	assert.NoError(t, err)
	assert.False(t, equal)
}

func TestSecretDataDiff_EmptyingAPopulatedSecretStillDiffs(t *testing.T) {
	diff := driftPlan(t, driftState(cty.StringVal(`{"DB_PASSWORD":"hunter2"}`)), driftConfig(cty.NullVal(cty.String)))

	assert.Contains(t, diff.Attributes, "secret_data")
}

// Adding data to a secret that was empty is a change in the other direction.
func TestSecretDataCompare_PopulatingAnEmptySecretDiffs(t *testing.T) {
	equal, err := secretDataCompare("", `{"DB_PASSWORD":"hunter2"}`)

	assert.NoError(t, err)
	assert.False(t, equal)
}

// Ordinary rotations and no-ops are unaffected.
func TestSecretDataCompare_UnchangedAndRotatedValues(t *testing.T) {
	same, err := secretDataCompare(`{"a":"1","b":"2"}`, `{"b":"2","a":"1"}`)
	assert.NoError(t, err)
	assert.True(t, same, "key order must not matter")

	rotated, err := secretDataCompare(`{"a":"1"}`, `{"a":"2"}`)
	assert.NoError(t, err)
	assert.False(t, rotated)
}

// Malformed JSON is still an error, and still surfaces as a difference rather than being
// quietly suppressed.
func TestSecretDataCompare_MalformedJSONIsStillAnError(t *testing.T) {
	equal, err := secretDataCompare("{", "{}")

	assert.Error(t, err)
	assert.False(t, equal)
}

// A value in state that cannot be parsed must surface as a difference, not be held equal.
// Driven through the plan so it pins secretDataDiff's error branch and not just the
// comparison underneath it.
func TestSecretDataDiff_UnparseableStateStillDiffs(t *testing.T) {
	diff := driftPlan(t, driftState(cty.StringVal("{")), driftConfig(cty.NullVal(cty.String)))

	assert.Contains(t, diff.Attributes, "secret_data")
}

// --- DUPLO-44790 -----------------------------------------------------------

// The backend decodes any secret value that looks like a JSON object or array before
// returning it, so state holds an object where the configuration holds the string it
// came from.  dockerconfigjson and dockercfg secrets always look like that.
const (
	dockerConfigJSONState  = `{".dockerconfigjson":{"auths":{"example.io":{"auth":"dTpw"}}}}`
	dockerConfigJSONConfig = `{".dockerconfigjson":"{\"auths\":{\"example.io\":{\"auth\":\"dTpw\"}}}"}`
)

func TestSecretDataCompare_DecodedObjectMatchesItsJSONString(t *testing.T) {
	equal, err := secretDataCompare(dockerConfigJSONState, dockerConfigJSONConfig)
	assert.NoError(t, err)
	assert.True(t, equal, "dockerconfigjson")

	equal, err = secretDataCompare(`{".dockercfg":{"example.io":{"auth":"dTpw"}}}`, `{".dockercfg":"{\"example.io\":{\"auth\":\"dTpw\"}}"}`)
	assert.NoError(t, err)
	assert.True(t, equal, "dockercfg")

	equal, err = secretDataCompare(`{"hosts":["a","b"]}`, `{"hosts":"[\"a\",\"b\"]"}`)
	assert.NoError(t, err)
	assert.True(t, equal, "array")
}

func TestSecretDataCompare_DecodedObjectStillDiffsOnRealChange(t *testing.T) {
	equal, err := secretDataCompare(dockerConfigJSONState, `{".dockerconfigjson":"{\"auths\":{\"example.io\":{\"auth\":\"bmV3\"}}}"}`)
	assert.NoError(t, err)
	assert.False(t, equal, "rotated credential")

	equal, err = secretDataCompare(dockerConfigJSONState, `{".dockerconfigjson":"not json"}`)
	assert.NoError(t, err)
	assert.False(t, equal, "non-JSON string against an object")
}

// Only objects and arrays are decoded by the backend, so a scalar that parses as JSON
// is still compared as the string it is.
func TestSecretDataCompare_JSONScalarStringIsNotDecoded(t *testing.T) {
	for _, tc := range []struct{ state, config string }{
		{`{"x":1}`, `{"x":"1"}`},
		{`{"x":true}`, `{"x":"true"}`},
		{`{"x":null}`, `{"x":"null"}`},
	} {
		equal, err := secretDataCompare(tc.state, tc.config)
		assert.NoError(t, err)
		assert.False(t, equal, tc.config)
	}
}

// The object form in the configuration compares against the object in state as well.
func TestSecretDataCompare_ObjectInConfigMatchesObjectInState(t *testing.T) {
	equal, err := secretDataCompare(dockerConfigJSONState, dockerConfigJSONState)
	assert.NoError(t, err)
	assert.True(t, equal)
}

func TestSecretDataDiff_DockerConfigJSONConverges(t *testing.T) {
	diff := driftPlan(t, driftState(cty.StringVal(dockerConfigJSONState)), driftConfig(cty.StringVal(dockerConfigJSONConfig)))

	assert.NotContains(t, diff.Attributes, "secret_data")
}

// Integers above 2^53 do not survive a float64, so both sides are decoded with their
// digits intact and an edit between two such integers still plans.
func TestSecretDataCompare_LargeIntegersStayDistinct(t *testing.T) {
	state := `{"k":{"n":9007199254740992}}`

	equal, err := secretDataCompare(state, `{"k":"{\"n\":9007199254740993}"}`)
	assert.NoError(t, err)
	assert.False(t, equal, "changed integer")

	equal, err = secretDataCompare(state, `{"k":"{\"n\":9007199254740992}"}`)
	assert.NoError(t, err)
	assert.True(t, equal, "same integer")

	equal, err = secretDataCompare(`{"n":9007199254740992}`, `{"n":9007199254740993}`)
	assert.NoError(t, err)
	assert.False(t, equal, "top-level integer")
}

// Trailing data after the JSON value is still rejected, as json.Unmarshal did.
func TestSecretDataCompare_TrailingDataIsNotJSON(t *testing.T) {
	_, err := secretDataCompare(`{"k":"v"} junk`, `{"k":"v"}`)
	assert.Error(t, err)

	equal, err := secretDataCompare(`{"k":{"a":1}}`, `{"k":"{\"a\":1} junk"}`)
	assert.NoError(t, err)
	assert.False(t, equal, "config string with trailing data is compared as a string")
}
