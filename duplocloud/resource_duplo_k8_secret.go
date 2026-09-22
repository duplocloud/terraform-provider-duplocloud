package duplocloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the secret will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"secret_name": {
			Description:  "The name of the secret.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: ValidateDnsSubdomainRFC1123(),
		},
		"secret_type": {
			Description: "The type of the secret.  Usually `\"Opaque\"`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"client_secret_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_data": {
			Description: "A JSON encoded string representing the secret metadata. " +
				"You can use the `jsonencode()` function to convert map or object data, if needed. You can use the `jsondecode()` function to read data.",
			Type:             schema.TypeString,
			Optional:         true,
			Sensitive:        true,
			ValidateFunc:     ValidateJSONObjectString,
			DiffSuppressFunc: secretDataDiff,
		},
		"manage_secret_data": {
			Description: "Whether Terraform manages the contents of the secret.  Defaults to `true`.\n\n" +
				"When `true`, `secret_data` is sent to Duplo on create and update, and the secret's values are stored " +
				"in Terraform state.\n\n" +
				"Set this to `false` to track a secret whose values are owned outside of Terraform - for example, one " +
				"created by the Duplo installer and rotated out of band.  Terraform then manages only the existence of " +
				"the secret: `secret_data` must be omitted from the configuration, its values are masked in state, and " +
				"creates and updates carry the existing data forward instead of overwriting it.\n\n" +
				"Two things this mode does not protect against.  `tenant_id`, `secret_name` and `secret_type` force a " +
				"new resource, and a replacement destroys and recreates the secret, its data included - so an import " +
				"that guesses `secret_type` wrong still loses the contents.  And `terraform import` has no " +
				"configuration behind it, so importing writes the secret's values to state once before this attribute " +
				"can be set to `false`; rotate the secret afterwards, or scrub the state history, if that matters.\n\n" +
				"This attribute keeps its last applied value when it is removed from the configuration, so set it back " +
				"to `true` explicitly to resume managing the contents.",
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"secret_annotations": {
			Description: "Annotations for the secret.\n\n**Note: : To skip encoding of an already encoded value string of a k8's secrete add `duplocloud.net/skip-encoding: \"true\"`",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"secret_labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the secret",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// SCHEMA for resource crud
func resourceK8Secret() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_secret` manages a kubernetes secret in a Duplo tenant.",

		ReadContext:   resourceK8SecretRead,
		CreateContext: resourceK8SecretCreate,
		UpdateContext: resourceK8SecretUpdate,
		DeleteContext: resourceK8SecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// Imported state has no configuration behind it, so seed the historical
				// behavior explicitly.  Without this the attribute comes back null and an
				// import of a secret that Terraform is meant to manage would silently land
				// in tracking-only mode.
				if err := d.Set("manage_secret_data", true); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},
		CustomizeDiff: validateK8sSecretDataManagement,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sSecretSchema(),
	}
}

// / READ resource
func resourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, cerr := c.K8SecretGet(tenantId, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceK8SecretRead(%s, %s): object not found", tenantId, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s secret %s : %s", tenantId, name, cerr)
	}

	manage := manageK8sSecretData(d)
	flattenK8sSecret(d, rp, !manage)
	d.Set("manage_secret_data", manage)

	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): end", tenantId, name)
	return nil
}

// / CREATE resource
func resourceK8SecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	name := d.Get("secret_name").(string)

	log.Printf("[TRACE] resourceK8SecretCreate(%s, %s): start", tenantId, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)

	// The secret may well exist already - tracking one the installer created is the point
	// of this mode - and there is no state to tell us so, because this is a create.  Carry
	// whatever is there forward rather than posting an empty map over it.
	if !manageK8sSecretData(d) {
		if err := carryForwardSecretData(c, tenantId, name, rq, true); err != nil {
			return diag.FromErr(err)
		}
	}

	cerr := c.K8SecretCreate(tenantId, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantId, name))

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8SecretCreate(%s, %s): end", tenantId, name)
	return diags
}

// / UPDATE resource
func resourceK8SecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretUpdate(%s, %s): start", tenantId, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)

	// When Terraform does not own the contents of the secret, carry the existing data
	// forward.  CreateOrUpdateK8Secret replaces the whole object, so an update driven by
	// a label or annotation change would otherwise wipe values managed outside Terraform.
	if !manageK8sSecretData(d) {
		if err := carryForwardSecretData(c, tenantId, name, rq, false); err != nil {
			return diag.FromErr(err)
		}
	}

	cerr := c.K8SecretUpdate(tenantId, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8SecretUpdate(%s, %s): end", tenantId, name)
	return diags
}

// / DELETE resource
func resourceK8SecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, cerr := c.K8SecretGet(tenantId, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): object not found", tenantId, name)
			return nil
		}
		return diag.FromErr(cerr)
	}
	if rp != nil && rp.SecretName != "" {
		cerr := c.K8SecretDelete(tenantId, name)
		if cerr != nil {
			return diag.FromErr(cerr)
		}
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): end", tenantId, name)
	return nil
}

func parseK8sSecretIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sSecret(d *schema.ResourceData, duplo *duplosdk.DuploK8sSecret, maskSecretData bool) {
	// First, set the simple fields.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("secret_name", duplo.SecretName)
	d.Set("secret_type", duplo.SecretType)
	d.Set("secret_version", duplo.SecretVersion)
	if maskSecretData {
		for key := range duplo.SecretData {
			duplo.SecretData[key] = "**********"
		}
	}
	// Next, set the JSON encoded strings.
	toJsonStringState("secret_data", duplo.SecretData, d)

	// Finally, set the map
	d.Set("secret_annotations", duplo.SecretAnnotations)
	filter := map[string]struct{}{
		"app":        {},
		"owner":      {},
		"tenantid":   {},
		"tenantname": {},
	}
	m := d.Get("secret_labels")
	if m != nil {
		for k := range m.(map[string]interface{}) {
			delete(filter, k)
		}
	}
	op := make(map[string]interface{})

	if duplo.SecretLabels != nil {
		for k, v := range duplo.SecretLabels {
			if _, ok := filter[k]; !ok {
				op[k] = v
			}
		}
	}
	d.Set("secret_labels", op)
	log.Printf("[TRACE] flattenK8sSecret(%s, %s): flattened %d key(s)", duplo.TenantID, duplo.SecretName, len(duplo.SecretData))

}

func expandK8sSecret(d *schema.ResourceData) (*duplosdk.DuploK8sSecret, error) {
	duplo := duplosdk.DuploK8sSecret{
		SecretName: d.Get("secret_name").(string),
		SecretType: d.Get("secret_type").(string),
	}

	// The annotations must be converted to a map of strings.
	if v, ok := d.GetOk("secret_annotations"); ok && !isInterfaceNil(v) {
		duplo.SecretAnnotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.SecretAnnotations[key] = value.(string)
		}
	}

	if v, ok := d.GetOk("secret_labels"); ok && !isInterfaceNil(v) {
		duplo.SecretLabels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			if !isStringValid(regexp.MustCompile("^(([A-Za-z0-9][-/_.A-Za-z0-9]*)?[A-Za-z0-9])?$"), key) {
				return nil, secretLabelValidationError(duplo.SecretName, key)
			}
			v := value.(string)
			if !isStringValid(regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"), v) {
				return nil, secretLabelValidationError(duplo.SecretName, v)
			}
			duplo.SecretLabels[key] = v
		}
	}

	// The data must be decoded as JSON.  It is skipped entirely when Terraform does not
	// own the contents of the secret - the caller supplies whatever is already there.
	if manageK8sSecretData(d) {
		data := d.Get("secret_data").(string)
		// Management has just been turned back on with nothing to write.  CustomizeDiff
		// catches this at plan time, but only when the configured value was known then, so
		// the check is repeated here where it always is.
		if data == "" {
			if wasManaged, ok := ctyBoolAttr(d.GetRawState(), "manage_secret_data"); ok && !wasManaged {
				return nil, errors.New(errReEnableWithoutData)
			}
		}
		if data != "" {
			err := json.Unmarshal([]byte(data), &duplo.SecretData)
			if err != nil {
				return nil, err
			}
		}
		d.Set("client_secret_version", hashForData(data))
	}
	if duplo.SecretData == nil {
		duplo.SecretData = map[string]interface{}{}
	}

	return &duplo, nil
}

// errReEnableWithoutData is raised in two places: at plan time by CustomizeDiff, and
// again on the write path, which is the only one that always has a known value to judge.
const errReEnableWithoutData = "secret_data is required when manage_secret_data is set back to true: the value " +
	"held in state is masked, so applying would overwrite the secret with an empty one - supply the secret's " +
	"contents, or use `secret_data = jsonencode({})` to empty it on purpose"

// carryForwardSecretData points the request at the data the backend already holds, so that
// a write made while Terraform does not own the contents leaves them alone.
// CreateOrUpdateK8Secret backs both K8SecretCreate and K8SecretUpdate and replaces the
// whole object, so anything left out of the request is lost - including the values of a
// secret that was created by the installer and has never been in Terraform's state.
//
// On create, a secret that does not exist yet is the ordinary case, so a 404 is not an
// error.  The empty map expandK8sSecret supplies is kept whenever the backend reports no
// data of its own, because SecretData is not omitempty and a nil map marshals as null.
func carryForwardSecretData(c *duplosdk.Client, tenantID, name string, rq *duplosdk.DuploK8sSecret, creating bool) error {
	current, cerr := c.K8SecretGet(tenantID, name)
	if cerr != nil {
		if creating && cerr.Status() == 404 {
			return nil
		}
		verb := "update"
		if creating {
			verb = "create"
		}
		return fmt.Errorf("Unable to retrieve tenant %s k8s secret %s for %s: %s", tenantID, name, verb, cerr)
	}
	if current != nil && current.SecretData != nil {
		rq.SecretData = current.SecretData
	}
	return nil
}

func secretLabelValidationError(name, value string) error {
	return fmt.Errorf(`Secret '%s' is invalid: metadata.labels: Invalid value: '%s,': a valid label must 
	be an empty string or consist of alphanumeric characters, '-', '', or '.', and must start and end with an alphanumeric 
	character (e.g. 'MyValue', or 'my_value', or '12345',
	 regex used for validation is '((A-Za-z0-9][-A-Za-z0-9.]*)?[A-Za-z0-9])?').`, name, value)
}

func secretDataDiff(k, old, new string, d *schema.ResourceData) bool {
	// Ownership is not settled yet, so show the change rather than hiding it on the
	// strength of a prior state that may not survive the apply.
	if ctyAttrIsUnknown(d.GetRawConfig(), "manage_secret_data") {
		return false
	}
	// Terraform does not own the contents of the secret, so the masked values held in
	// state never need to be reconciled against the configuration.
	if !manageK8sSecretData(d) {
		return true
	}
	state, err := secretDataCompare(old, new)
	if err != nil {
		log.Printf("TRACE secretDataCompare : %s", err.Error())
		return state
	}
	return state
}
func secretDataCompare(old, new string) (bool, error) {
	var obj1, obj2 map[string]interface{}

	// Unmarshal the first JSON string into a map
	if err := json.Unmarshal([]byte(old), &obj1); err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 1: %v", err)
	}

	// Unmarshal the second JSON string into a map
	if err := json.Unmarshal([]byte(new), &obj2); err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 2: %v", err)
	}
	if len(obj1) != len(obj2) {
		return false, nil
	}
	for k, v := range obj2 {
		if v1, ok := obj1[k]; !ok {
			return false, nil
		} else {
			s := fmt.Sprintf("%v", v)
			if v1 != s {
				return false, nil
			}
		}
	}
	return true, nil
}

// k8sSecretRawAccessor is satisfied by both *schema.ResourceData and *schema.ResourceDiff.
type k8sSecretRawAccessor interface {
	GetRawConfig() cty.Value
	GetRawState() cty.Value
}

// manageK8sSecretData reports whether Terraform owns the contents of the secret.
//
// manage_secret_data is Optional+Computed rather than defaulted on purpose.  State
// written before the attribute existed carries no value for it, and a Default would make
// every such resource plan a change on the next refresh - which, for anyone already
// tracking a secret they do not own, would push an empty SecretData and wipe it.  So the
// raw configuration is consulted first, then the raw prior state, and an absent value
// means the historical "Terraform manages the data" behavior.
func manageK8sSecretData(d k8sSecretRawAccessor) bool {
	if v, ok := ctyBoolAttr(d.GetRawConfig(), "manage_secret_data"); ok {
		return v
	}
	if v, ok := ctyBoolAttr(d.GetRawState(), "manage_secret_data"); ok {
		return v
	}
	return true
}

// ctyBoolAttr reads a boolean attribute out of a cty object, reporting whether it was
// present and usable.  Raw config is null during a refresh and raw state is null during a
// create, so every access has to tolerate an absent object.
func ctyBoolAttr(obj cty.Value, name string) (bool, bool) {
	if obj.IsNull() || !obj.IsKnown() || !obj.Type().IsObjectType() || !obj.Type().HasAttribute(name) {
		return false, false
	}
	v := obj.GetAttr(name)
	if v.IsNull() || !v.IsKnown() {
		return false, false
	}
	return v.True(), true
}

// validateK8sSecretDataManagement rejects configurations whose secret data and declared
// ownership of it disagree, rather than letting the apply quietly write the wrong thing.
func validateK8sSecretDataManagement(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	config, state := diff.GetRawConfig(), diff.GetRawState()
	hasSecretData := ctyAttrIsSet(config, "secret_data")

	// An unknown value says nothing about ownership yet, and manageK8sSecretData falls back
	// to prior state for it - which is precisely what is about to change.  Judging it now
	// would reject a configuration that is about to become valid, or wave through a
	// re-enable that is about to empty the secret.  Both are settled on the write path.
	if ctyAttrIsUnknown(config, "manage_secret_data") {
		return nil
	}

	if manageK8sSecretData(diff) {
		// Management is being turned back on with nothing to write.  The value sitting in
		// state is a mask, so the plan renders as an ordinary rotation while the apply would
		// post an empty secret over data Terraform has never seen.  Reaching here with a
		// `false` in state means the configuration said `true` explicitly, since an absent
		// value falls back to state.
		if wasManaged, ok := ctyBoolAttr(state, "manage_secret_data"); ok && !wasManaged && !hasSecretData {
			return errors.New(errReEnableWithoutData)
		}
		return nil
	}

	if !hasSecretData {
		return nil
	}
	msg := "secret_data must not be set when manage_secret_data is false: in that mode Terraform tracks only " +
		"the existence of the secret, and its contents are left to whoever owns them"
	if _, inConfig := ctyBoolAttr(config, "manage_secret_data"); !inConfig {
		// The `false` came from prior state, so the configuration in front of the user does
		// not mention manage_secret_data at all and the message above reads as a provider bug.
		msg += ". manage_secret_data is false in state; it is Optional and Computed, so removing it from the " +
			"configuration keeps the last applied value - set manage_secret_data = true explicitly to resume " +
			"managing the contents"
	}
	return errors.New(msg)
}

// ctyAttrIsUnknown reports whether an attribute is present but not yet computed.  The
// ownership helpers cannot read such a value, and treating it as absent is wrong in a way
// that matters, so the plan-time callers ask about it separately.
func ctyAttrIsUnknown(obj cty.Value, name string) bool {
	if obj.IsNull() || !obj.IsKnown() || !obj.Type().IsObjectType() || !obj.Type().HasAttribute(name) {
		return false
	}
	v := obj.GetAttr(name)
	return !v.IsNull() && !v.IsKnown()
}

// ctyAttrIsSet reports whether an attribute is present in a cty object and carries a
// value.  Raw config is null during a refresh and raw state is null during a create, so
// every access has to tolerate an absent object.
func ctyAttrIsSet(obj cty.Value, name string) bool {
	if obj.IsNull() || !obj.IsKnown() || !obj.Type().IsObjectType() || !obj.Type().HasAttribute(name) {
		return false
	}
	return !obj.GetAttr(name).IsNull()
}
