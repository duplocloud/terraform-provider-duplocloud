package duplocloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"math/big"
	"regexp"
	"strconv"
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
		CustomizeDiff: customizeK8SecretDiff,
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
	configured, _ := d.Get("secret_annotations").(map[string]interface{})
	flattenK8sSecret(d, rp, !manage)
	d.Set("secret_annotations", k8sSecretAnnotationsForState(configured, rp.SecretAnnotations))
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

const k8sSecretSkipEncodingAnnotation = "duplocloud.net/skip-encoding"

// k8sSecretAnnotationsForState reconciles the skip-encoding annotation with the
// configuration.  The backend rewrites it as bool.ToString() on every update, which adds
// "False" to a secret that never had it and turns a configured "true" into "True".  An
// unconfigured "false" is dropped, since it is what the backend backfills and means the
// same as no annotation; any other unconfigured value stays, so removing an enabled
// skip-encoding still plans.  The configured spelling is kept when the two agree ignoring
// case, which is how the backend parses it.  Neither changes the secret.
func k8sSecretAnnotationsForState(configured map[string]interface{}, backend map[string]string) map[string]string {
	annotations := make(map[string]string, len(backend))
	for k, v := range backend {
		annotations[k] = v
	}
	want, ok := configured[k8sSecretSkipEncodingAnnotation].(string)
	if !ok {
		if strings.EqualFold(annotations[k8sSecretSkipEncodingAnnotation], "false") {
			delete(annotations, k8sSecretSkipEncodingAnnotation)
		}
	} else if got, ok := annotations[k8sSecretSkipEncodingAnnotation]; ok && strings.EqualFold(got, want) {
		annotations[k8sSecretSkipEncodingAnnotation] = want
	}
	return annotations
}

// suppressK8sSecretSkipEncodingDrift clears a secret_annotations change that only Read's
// reconciliation would have hidden.  Read judges ownership of the annotation by d.Get,
// which during a refresh is prior state, so state written before that reconciliation
// existed still holds the backend's value and keeps planning its removal.  The raw
// configuration is authoritative here, so the same reconciliation is applied against it.
func suppressK8sSecretSkipEncodingDrift(diff *schema.ResourceDiff) error {
	if !diff.HasChange("secret_annotations") {
		return nil
	}
	config := diff.GetRawConfig()
	if config.IsNull() || !config.IsKnown() || !config.Type().IsObjectType() || !config.Type().HasAttribute("secret_annotations") {
		return nil
	}
	v := config.GetAttr("secret_annotations")
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil
	}
	configured := map[string]interface{}{}
	want := map[string]string{}
	for k, e := range v.AsValueMap() {
		if e.IsNull() {
			continue
		}
		configured[k] = e.AsString()
		want[k] = e.AsString()
	}
	o, _ := diff.GetChange("secret_annotations")
	backend := map[string]string{}
	for k, e := range o.(map[string]interface{}) {
		backend[k], _ = e.(string)
	}
	if maps.Equal(k8sSecretAnnotationsForState(configured, backend), want) {
		return diff.Clear("secret_annotations")
	}
	return nil
}

func customizeK8SecretDiff(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	if err := suppressK8sSecretSkipEncodingDrift(diff); err != nil {
		return err
	}
	return validateK8sSecretDataManagement(ctx, diff, m)
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

	manage := manageK8sSecretData(d)

	// The mirror of the re-enable guard below.  CustomizeDiff defers while the configured
	// ownership is unknown, so a value that only resolves to false at apply arrives here
	// with secret_data still set.  Skipping it silently would discard the configured
	// secret and still report a successful apply.
	if !manage && ctyAttrIsSet(d.GetRawConfig(), "secret_data") {
		return nil, errors.New(errDataWhileUnmanaged)
	}

	// The data must be decoded as JSON.  It is skipped entirely when Terraform does not
	// own the contents of the secret - the caller supplies whatever is already there.
	if manage {
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
			err := decodeSecretJSON(data, &duplo.SecretData)
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
// errDataWhileUnmanaged is raised at plan time by CustomizeDiff and again on the write
// path, which is the only one that always has a known ownership value to judge.
const errDataWhileUnmanaged = "secret_data must not be set when manage_secret_data is false: in that mode " +
	"Terraform tracks only the existence of the secret, and its contents are left to whoever owns them"

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
		// Report a difference rather than suppressing one, so a value neither side can
		// parse shows up in the plan instead of being silently held equal.
		log.Printf("[TRACE] secretDataDiff(%s): %s", k, err)
		return false
	}
	return state
}

// parseSecretData decodes a secret_data value into the map it represents.  An attribute
// the configuration omits reaches here as "", while a secret with no keys is stored as
// "{}" - the two describe the same secret, so treating "" as an empty object is what lets
// an omitted secret_data settle instead of planning the same change on every run.
func parseSecretData(s string) (map[string]interface{}, error) {
	obj := map[string]interface{}{}
	if s == "" {
		return obj, nil
	}
	if err := decodeSecretJSON(s, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// decodeSecretJSON decodes s the way json.Unmarshal would, but keeps numbers as their
// digit strings.  Decoding into float64 merges integers above 2^53, so an edit from
// 9007199254740992 to 9007199254740993 would compare equal and never plan.
func decodeSecretJSON(s string, v interface{}) error {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return errors.New("unexpected data after the JSON value")
	}
	return nil
}

func secretDataCompare(old, new string) (bool, error) {
	obj1, err := parseSecretData(old)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 1: %v", err)
	}

	obj2, err := parseSecretData(new)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 2: %v", err)
	}
	if len(obj1) != len(obj2) {
		return false, nil
	}
	for k, v := range obj2 {
		if v1, ok := obj1[k]; !ok || !secretValueEqual(v1, v) {
			return false, nil
		}
	}
	return true, nil
}

// secretValueEqual compares one value held in state against the configured one.  The
// backend decodes any value that looks like a JSON object or array before returning it,
// so a dockerconfigjson secret comes back as an object while the configuration holds the
// string it was written from.  Such a string is decoded before comparing, but only into an
// object or array - a scalar is never decoded by the backend, so 1 and "1" stay distinct.
func secretValueEqual(state, config interface{}) bool {
	if _, ok := state.(string); ok {
		if n, ok := config.(json.Number); ok {
			return state == newtonsoftNumberString(n)
		}
		return state == fmt.Sprintf("%v", config)
	}
	if s, ok := config.(string); ok {
		var decoded interface{}
		if err := decodeSecretJSON(s, &decoded); err != nil {
			return false
		}
		switch decoded.(type) {
		case map[string]interface{}, []interface{}:
			config = decoded
		}
	}
	return secretJSONEqual(state, config)
}

// newtonsoftNumberString returns the text the backend stores for a configured scalar
// number.  The backend (net48, Newtonsoft 13) parses an integer literal as a long or
// BigInteger and writes its digits back, and any other number as a double, which it writes
// with .NET Framework's "R" format plus a ".0" when that has no decimal point or exponent.
// So 1.50 is stored as 1.5 and 1e3 as 1000.0, while 1 and 1.0 stay different strings.
func newtonsoftNumberString(n json.Number) string {
	s := string(n)
	if !strings.ContainsAny(s, ".eE") {
		if i, ok := new(big.Int).SetString(s, 10); ok {
			return i.String()
		}
		return s
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s
	}
	if f == 0 {
		// .NET Framework prints negative zero as "0".
		return "0.0"
	}
	text := netFrameworkGeneral(f, 15)
	if back, err := strconv.ParseFloat(text, 64); err != nil || back != f {
		text = netFrameworkGeneral(f, 17)
	}
	if strings.ContainsAny(text, ".E") {
		return text
	}
	return text + ".0"
}

// netFrameworkGeneral formats f the way .NET's "G" format does at the given precision:
// trailing zeros dropped, fixed-point while the decimal exponent is above -5 and below the
// precision, and otherwise scientific with a signed exponent of at least two digits.
func netFrameworkGeneral(f float64, precision int) string {
	e := strconv.FormatFloat(f, 'e', precision-1, 64)
	sign := ""
	if e[0] == '-' {
		sign, e = "-", e[1:]
	}
	mantissa, expText, _ := strings.Cut(e, "e")
	exp, _ := strconv.Atoi(expText)
	digits := strings.TrimRight(strings.Replace(mantissa, ".", "", 1), "0")
	if digits == "" {
		digits = "0"
	}

	if exp > -5 && exp < precision {
		if exp < 0 {
			return sign + "0." + strings.Repeat("0", -exp-1) + digits
		}
		if len(digits) <= exp+1 {
			return sign + digits + strings.Repeat("0", exp+1-len(digits))
		}
		return sign + digits[:exp+1] + "." + digits[exp+1:]
	}

	out := digits[:1]
	if len(digits) > 1 {
		out += "." + digits[1:]
	}
	expSign := "+"
	if exp < 0 {
		expSign, exp = "-", -exp
	}
	return fmt.Sprintf("%s%sE%s%02d", sign, out, expSign, exp)
}

// secretJSONEqual compares two decoded JSON values, matching numbers by value rather than
// spelling.  The backend round-trips a decoded value through Newtonsoft, which can rewrite
// 1e3 as 1000.0 or 1.50 as 1.5, so comparing the digit strings would plan on every run.
// big.Rat holds each number exactly, so integers above 2^53 still stay distinct.
func secretJSONEqual(a, b interface{}) bool {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			w, ok := bv[k]
			if !ok || !secretJSONEqual(v, w) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !secretJSONEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case json.Number:
		bv, ok := b.(json.Number)
		if !ok {
			return false
		}
		x, okA := new(big.Rat).SetString(string(av))
		y, okB := new(big.Rat).SetString(string(bv))
		return okA && okB && x.Cmp(y) == 0
	default:
		return a == b
	}
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
	msg := errDataWhileUnmanaged
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
