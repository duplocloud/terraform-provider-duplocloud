package duplocloud

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

// fakeRawAccessor stands in for *schema.ResourceData / *schema.ResourceDiff, which are
// impractical to build with the raw config and raw state populated independently.
type fakeRawAccessor struct {
	config cty.Value
	state  cty.Value
}

func (f fakeRawAccessor) GetRawConfig() cty.Value { return f.config }
func (f fakeRawAccessor) GetRawState() cty.Value  { return f.state }

// k8sSecretRaw builds the cty object shape the SDK hands back for this resource.
func k8sSecretRaw(manageSecretData cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"secret_name":        cty.StringVal("env-var-os-global"),
		"manage_secret_data": manageSecretData,
	})
}

// k8sSecretRawNoAttr is state written by a provider version that predates the attribute.
func k8sSecretRawNoAttr() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"secret_name": cty.StringVal("env-var-os-global"),
	})
}

// nullObject is what the SDK returns when Terraform sent no config (a refresh) or no
// prior state (a create).
func nullObject() cty.Value {
	return cty.NullVal(cty.Object(map[string]cty.Type{
		"secret_name":        cty.String,
		"manage_secret_data": cty.Bool,
	}))
}

func TestManageK8sSecretData_ConfigExplicitFalse_NotManaged(t *testing.T) {
	d := fakeRawAccessor{config: k8sSecretRaw(cty.False), state: nullObject()}

	assert.False(t, manageK8sSecretData(d))
}

func TestManageK8sSecretData_ConfigExplicitTrue_Managed(t *testing.T) {
	d := fakeRawAccessor{config: k8sSecretRaw(cty.True), state: k8sSecretRaw(cty.False)}

	// Configuration wins over prior state, so opting back in takes effect immediately.
	assert.True(t, manageK8sSecretData(d))
}

func TestManageK8sSecretData_CreateWithAttributeOmitted_Managed(t *testing.T) {
	d := fakeRawAccessor{config: k8sSecretRaw(cty.NullVal(cty.Bool)), state: nullObject()}

	assert.True(t, manageK8sSecretData(d))
}

func TestManageK8sSecretData_RefreshFallsBackToState(t *testing.T) {
	// A refresh carries no configuration at all.
	d := fakeRawAccessor{config: nullObject(), state: k8sSecretRaw(cty.False)}

	assert.False(t, manageK8sSecretData(d))
}

// The upgrade-safety case: state written before manage_secret_data existed must keep the
// historical behavior, or the first refresh after upgrading would mask the data and plan
// a change against every existing secret.
func TestManageK8sSecretData_LegacyStateWithoutAttribute_Managed(t *testing.T) {
	d := fakeRawAccessor{config: nullObject(), state: k8sSecretRawNoAttr()}

	assert.True(t, manageK8sSecretData(d))
}

func TestManageK8sSecretData_LegacyStateNullAttribute_Managed(t *testing.T) {
	d := fakeRawAccessor{config: nullObject(), state: k8sSecretRaw(cty.NullVal(cty.Bool))}

	assert.True(t, manageK8sSecretData(d))
}

func TestManageK8sSecretData_UnknownConfigFallsBackToState(t *testing.T) {
	d := fakeRawAccessor{config: k8sSecretRaw(cty.UnknownVal(cty.Bool)), state: k8sSecretRaw(cty.False)}

	assert.False(t, manageK8sSecretData(d))
}

// ResourceDiff.GetRawConfig returns the diff's RawConfig verbatim, which can be an
// uninitialised cty.Value.  Reaching into it must not panic.
func TestManageK8sSecretData_NilValuesAreTolerated(t *testing.T) {
	d := fakeRawAccessor{config: cty.NilVal, state: cty.NilVal}

	assert.NotPanics(t, func() {
		assert.True(t, manageK8sSecretData(d))
	})
}

func TestCtyBoolAttr_NonObjectIsNotUsable(t *testing.T) {
	assert.NotPanics(t, func() {
		_, ok := ctyBoolAttr(cty.StringVal("not an object"), "manage_secret_data")
		assert.False(t, ok)
	})
}

func TestCtyBoolAttr_MissingAttributeIsNotUsable(t *testing.T) {
	_, ok := ctyBoolAttr(k8sSecretRawNoAttr(), "manage_secret_data")

	assert.False(t, ok)
}

func TestCtyBoolAttr_PresentValueIsReported(t *testing.T) {
	v, ok := ctyBoolAttr(k8sSecretRaw(cty.True), "manage_secret_data")

	assert.True(t, ok)
	assert.True(t, v)
}
