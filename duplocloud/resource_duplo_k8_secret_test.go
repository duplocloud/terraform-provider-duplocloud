package duplocloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

const k8sSecretTestID = "v2/subscriptions/3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31/K8SecretApiV2/env-var-os-global"

// fakeRawAccessor stands in for *schema.ResourceData / *schema.ResourceDiff, which are
// impractical to build with the raw config and raw state populated independently.
type fakeRawAccessor struct {
	config cty.Value
	state  cty.Value
}

func (f fakeRawAccessor) GetRawConfig() cty.Value { return f.config }
func (f fakeRawAccessor) GetRawState() cty.Value  { return f.state }

// k8sSecretType is the object type the SDK derives from the resource schema.  Building
// fixtures from it, rather than by hand, keeps them honest as the schema grows.
func k8sSecretType() cty.Type {
	return resourceK8Secret().CoreConfigSchema().ImpliedType()
}

// k8sSecretObject builds a value of that type, defaulting every attribute not named here
// to null - which is what Terraform sends for an attribute the configuration omits.
func k8sSecretObject(attrs map[string]cty.Value) cty.Value {
	vals := map[string]cty.Value{}
	for name, ty := range k8sSecretType().AttributeTypes() {
		if v, ok := attrs[name]; ok {
			vals[name] = v
		} else {
			vals[name] = cty.NullVal(ty)
		}
	}
	return cty.ObjectVal(vals)
}

// k8sSecretRaw is the minimal object the ownership helpers actually read.
func k8sSecretRaw(manageSecretData cty.Value) cty.Value {
	return k8sSecretObject(map[string]cty.Value{
		"secret_name":        cty.StringVal("env-var-os-global"),
		"manage_secret_data": manageSecretData,
	})
}

// k8sSecretRawNoAttr is state written by a provider version that predates the attribute.
// It is deliberately hand-built: the whole point is that the attribute is absent from the
// object type, which a fixture derived from the current schema can never reproduce.
func k8sSecretRawNoAttr() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"secret_name": cty.StringVal("env-var-os-global"),
	})
}

// nullObject is what the SDK returns when Terraform sent no config (a refresh) or no
// prior state (a create).
func nullObject() cty.Value {
	return cty.NullVal(k8sSecretType())
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

// The case that actually occurs in practice: core's UpgradeResourceState leaves the new
// attribute present-but-null rather than absent.
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

func TestCtyAttrIsSet(t *testing.T) {
	set := k8sSecretObject(map[string]cty.Value{"secret_data": cty.StringVal(`{"foo":"bar"}`)})

	assert.True(t, ctyAttrIsSet(set, "secret_data"))
	assert.False(t, ctyAttrIsSet(k8sSecretObject(nil), "secret_data"))
	assert.False(t, ctyAttrIsSet(nullObject(), "secret_data"))
	assert.False(t, ctyAttrIsSet(k8sSecretRawNoAttr(), "secret_data"))
	assert.NotPanics(t, func() { assert.False(t, ctyAttrIsSet(cty.NilVal, "secret_data")) })
}

// k8sSecretPlan drives the real plan path - CustomizeDiff and DiffSuppressFunc both -
// rather than calling them with a hand-rolled accessor.  That is what pins the assumption
// the whole design rests on: that the SDK wires raw config and raw state where this code
// looks for them.  schemaMap.Diff copies them off the InstanceState, so that is where
// they have to be set.
func k8sSecretPlan(t *testing.T, state, config cty.Value) (*terraform.InstanceDiff, error) {
	t.Helper()
	r := resourceK8Secret()

	is := &terraform.InstanceState{}
	if !state.IsNull() {
		shimmed, err := r.ShimInstanceStateFromValue(state)
		assert.NoError(t, err)
		is = shimmed
		is.ID = k8sSecretTestID
	}
	is.RawState = state
	is.RawConfig = config
	is.RawPlan = config

	return r.SimpleDiff(context.Background(), is, terraform.NewResourceConfigShimmed(config, r.CoreConfigSchema()), nil)
}

// k8sSecretState is a prior state for a secret that exists.
func k8sSecretState(manageSecretData, secretData cty.Value) cty.Value {
	return k8sSecretObject(map[string]cty.Value{
		"id":                 cty.StringVal(k8sSecretTestID),
		"tenant_id":          cty.StringVal("3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31"),
		"secret_name":        cty.StringVal("env-var-os-global"),
		"secret_type":        cty.StringVal("Opaque"),
		"manage_secret_data": manageSecretData,
		"secret_data":        secretData,
	})
}

// k8sSecretConfig is a configuration for the same secret.
func k8sSecretConfig(manageSecretData, secretData cty.Value) cty.Value {
	return k8sSecretObject(map[string]cty.Value{
		"tenant_id":          cty.StringVal("3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31"),
		"secret_name":        cty.StringVal("env-var-os-global"),
		"secret_type":        cty.StringVal("Opaque"),
		"manage_secret_data": manageSecretData,
		"secret_data":        secretData,
	})
}

var (
	nullString = cty.NullVal(cty.String)
	nullBool   = cty.NullVal(cty.Bool)
	maskedData = cty.StringVal(`{"DB_PASSWORD":"**********"}`)
	realData   = cty.StringVal(`{"DB_PASSWORD":"hunter2"}`)
)

func TestValidateK8sSecretDataManagement_DataWithUnmanagedConfigErrors(t *testing.T) {
	_, err := k8sSecretPlan(t, nullObject(), k8sSecretConfig(cty.False, realData))

	assert.ErrorContains(t, err, "secret_data must not be set when manage_secret_data is false")
}

// The `false` came from prior state, so the message has to say so - the configuration in
// front of the user contains no manage_secret_data at all.
func TestValidateK8sSecretDataManagement_UnmanagedFromStateNamesTheSource(t *testing.T) {
	_, err := k8sSecretPlan(t, k8sSecretState(cty.False, maskedData), k8sSecretConfig(nullBool, realData))

	assert.ErrorContains(t, err, "manage_secret_data is false in state")
	assert.ErrorContains(t, err, "set manage_secret_data = true explicitly")
}

// Re-enabling management with no data to write would post an empty secret over values
// Terraform has only ever seen masked.
func TestValidateK8sSecretDataManagement_ReEnablingWithoutDataErrors(t *testing.T) {
	_, err := k8sSecretPlan(t, k8sSecretState(cty.False, maskedData), k8sSecretConfig(cty.True, nullString))

	assert.ErrorContains(t, err, "secret_data is required when manage_secret_data is set back to true")
}

func TestValidateK8sSecretDataManagement_ReEnablingWithDataIsAccepted(t *testing.T) {
	_, err := k8sSecretPlan(t, k8sSecretState(cty.False, maskedData), k8sSecretConfig(cty.True, realData))

	assert.NoError(t, err)
}

func TestValidateK8sSecretDataManagement_UnmanagedWithoutDataIsAccepted(t *testing.T) {
	_, err := k8sSecretPlan(t, k8sSecretState(cty.False, maskedData), k8sSecretConfig(cty.False, nullString))

	assert.NoError(t, err)
}

// Emptying a managed secret on purpose stays legal - the new guard must not reach it.
func TestValidateK8sSecretDataManagement_ManagedSecretDroppingDataIsAccepted(t *testing.T) {
	_, err := k8sSecretPlan(t, k8sSecretState(cty.True, realData), k8sSecretConfig(nullBool, nullString))

	assert.NoError(t, err)
}

func TestSecretDataDiff_SuppressedWhenUnmanaged(t *testing.T) {
	diff, err := k8sSecretPlan(t, k8sSecretState(cty.False, maskedData), k8sSecretConfig(cty.False, nullString))

	assert.NoError(t, err)
	assert.NotContains(t, diff.Attributes, "secret_data")
}

func TestSecretDataDiff_NotSuppressedWhenManaged(t *testing.T) {
	rotated := cty.StringVal(`{"DB_PASSWORD":"hunter3"}`)
	diff, err := k8sSecretPlan(t, k8sSecretState(cty.True, realData), k8sSecretConfig(cty.True, rotated))

	assert.NoError(t, err)
	assert.Contains(t, diff.Attributes, "secret_data")
}

// k8sSecretData builds the ResourceData that Create, Update and Read are handed.
func k8sSecretData(t *testing.T, state, config cty.Value) *schema.ResourceData {
	t.Helper()
	r := resourceK8Secret()

	is, err := r.ShimInstanceStateFromValue(state)
	assert.NoError(t, err)
	is.ID = k8sSecretTestID
	is.RawState = state
	is.RawConfig = config

	return r.Data(is)
}

func TestExpandK8sSecret_SkipsDataWhenUnmanaged(t *testing.T) {
	state := k8sSecretState(cty.False, maskedData)
	d := k8sSecretData(t, state, k8sSecretConfig(cty.False, nullString))

	rq, err := expandK8sSecret(d)

	assert.NoError(t, err)
	// Never the mask, and never nil: SecretData has no omitempty, so a nil map would go
	// out as JSON null instead of the empty object every other path sends.
	assert.NotNil(t, rq.SecretData)
	assert.Empty(t, rq.SecretData)
}

func TestExpandK8sSecret_SendsDataWhenManaged(t *testing.T) {
	state := k8sSecretState(cty.True, realData)
	d := k8sSecretData(t, state, k8sSecretConfig(cty.True, realData))

	rq, err := expandK8sSecret(d)

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"DB_PASSWORD": "hunter2"}, rq.SecretData)
}

func TestFlattenK8sSecret_MasksTheValues(t *testing.T) {
	d := k8sSecretData(t, k8sSecretState(cty.False, maskedData), nullObject())
	duplo := &duplosdk.DuploK8sSecret{
		TenantID:   "3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31",
		SecretName: "env-var-os-global",
		SecretType: "Opaque",
		SecretData: map[string]interface{}{"DB_PASSWORD": "hunter2"},
	}

	flattenK8sSecret(d, duplo, true)

	assert.Equal(t, `{"DB_PASSWORD":"**********"}`, d.Get("secret_data"))
	assert.NotContains(t, d.Get("secret_data"), "hunter2")
}

func TestFlattenK8sSecret_StoresTheValuesWhenManaged(t *testing.T) {
	d := k8sSecretData(t, k8sSecretState(cty.True, realData), nullObject())
	duplo := &duplosdk.DuploK8sSecret{
		TenantID:   "3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31",
		SecretName: "env-var-os-global",
		SecretType: "Opaque",
		SecretData: map[string]interface{}{"DB_PASSWORD": "hunter2"},
	}

	flattenK8sSecret(d, duplo, false)

	assert.Equal(t, `{"DB_PASSWORD":"hunter2"}`, d.Get("secret_data"))
}

// The importer has no configuration behind it, so it has to seed the attribute or the
// follow-up read would land in tracking-only mode and mask a secret Terraform owns.
func TestK8SecretImporter_SeedsManagedData(t *testing.T) {
	r := resourceK8Secret()
	d := r.Data(&terraform.InstanceState{ID: k8sSecretTestID})

	results, err := r.Importer.StateContext(context.Background(), d, nil)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, true, results[0].Get("manage_secret_data"))
}

// k8sSecretServer serves GetAllK8Secrets and captures whatever gets posted back to
// CreateOrUpdateK8Secret, so a write can be inspected as it goes out on the wire.
func k8sSecretServer(t *testing.T, list string, posted *string) (*httptest.Server, *duplosdk.Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "CreateOrUpdateK8Secret") {
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			*posted = string(body)
			// K8SecretCreateOrUpdate passes a nil response type, so postAPI wants no body.
			return
		}
		fmt.Fprint(w, list)
	}))
	c, err := duplosdk.NewClient(srv.URL, "fake-token")
	if err != nil {
		t.Fatalf("NewClient: %s", err)
	}
	return srv, c
}

const k8sSecretTenant = "3c9bd2e4-1f5a-4d7b-8a62-1e9f0c5b7d31"

// The critical case: applying an unmanaged block against a secret that already exists.
// Create and Update are the same whole-object replace, and a create has no prior state to
// warn it the secret is there, so without the read-before-write guard this posts an empty
// SecretData over an installer-created secret and every service depending on it loses its
// environment.
func TestCarryForwardSecretData_CreateKeepsExistingData(t *testing.T) {
	var posted string
	srv, c := k8sSecretServer(t, `[{"SecretName":"env-var-os-global","SecretType":"Opaque",`+
		`"SecretData":{"DB_PASSWORD":"hunter2"}}]`, &posted)
	defer srv.Close()

	rq := &duplosdk.DuploK8sSecret{SecretName: "env-var-os-global", SecretType: "Opaque",
		SecretData: map[string]interface{}{}}

	assert.NoError(t, carryForwardSecretData(c, k8sSecretTenant, "env-var-os-global", rq, true))
	assert.Equal(t, map[string]interface{}{"DB_PASSWORD": "hunter2"}, rq.SecretData)
}

// A secret that is not there yet is the ordinary create, not a failure.
func TestCarryForwardSecretData_CreateToleratesMissingSecret(t *testing.T) {
	var posted string
	srv, c := k8sSecretServer(t, `[]`, &posted)
	defer srv.Close()

	rq := &duplosdk.DuploK8sSecret{SecretName: "env-var-os-global",
		SecretData: map[string]interface{}{}}

	assert.NoError(t, carryForwardSecretData(c, k8sSecretTenant, "env-var-os-global", rq, true))
	assert.NotNil(t, rq.SecretData)
}

// The same 404 has to fail an update: there is a resource in state, so a secret that has
// gone missing is a real problem and silently posting an empty one would compound it.
func TestCarryForwardSecretData_UpdateRejectsMissingSecret(t *testing.T) {
	var posted string
	srv, c := k8sSecretServer(t, `[]`, &posted)
	defer srv.Close()

	rq := &duplosdk.DuploK8sSecret{SecretName: "env-var-os-global",
		SecretData: map[string]interface{}{}}

	err := carryForwardSecretData(c, k8sSecretTenant, "env-var-os-global", rq, false)

	assert.ErrorContains(t, err, "env-var-os-global")
}

// SecretData is not omitempty, so a nil map marshals as JSON null rather than the empty
// object every other path sends.  A backend response carrying no SecretData must not turn
// into a null on the one path whose entire job is to leave the data alone.
func TestCarryForwardSecretData_NeverSendsJSONNull(t *testing.T) {
	var posted string
	srv, c := k8sSecretServer(t, `[{"SecretName":"env-var-os-global","SecretType":"Opaque"}]`, &posted)
	defer srv.Close()

	rq := &duplosdk.DuploK8sSecret{SecretName: "env-var-os-global", SecretType: "Opaque",
		SecretData: map[string]interface{}{}}

	assert.NoError(t, carryForwardSecretData(c, k8sSecretTenant, "env-var-os-global", rq, true))
	assert.NotNil(t, rq.SecretData)

	assert.NoError(t, c.K8SecretCreate(k8sSecretTenant, rq))
	assert.Contains(t, posted, `"SecretData":{}`)
	assert.NotContains(t, posted, `"SecretData":null`)
}
