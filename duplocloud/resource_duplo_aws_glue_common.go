package duplocloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// glueBodyJSONDescription is reused across resources for the body_json attribute.
const glueBodyJSONDescription = "JSON-encoded body of the AWS Glue API request. " +
	"Typed fields (e.g. `name`, `role`, `type`, `database_name`, `registry_name`) are merged in automatically — do not duplicate them here. " +
	"Use `jsonencode({...})` for readability. Omit the attribute entirely when all configurable fields are typed (it defaults to `{}`); do not set it to an empty string. " +
	"Drift on AWS-computed fields (CreationTime, CatalogId, etc.) is suppressed."

// glueCommonTenantIDSchema returns the standard tenant_id schema entry.
func glueCommonTenantIDSchema() *schema.Schema {
	return &schema.Schema{
		Description:  "The GUID of the tenant in which the Glue resource is provisioned.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.IsUUID,
	}
}

// glueBodyJSONSchema returns the standard body_json schema entry.
func glueBodyJSONSchema() *schema.Schema {
	return &schema.Schema{
		Description:      glueBodyJSONDescription,
		Type:             schema.TypeString,
		Optional:         true,
		Computed:         true,
		DiffSuppressFunc: glueBodyJSONDiffSuppress,
		ValidateFunc:     validation.StringIsJSON,
	}
}

// glueFullnameSchema is the standard `fullname` computed attribute. AWS Glue
// knows resources by their tenant-prefixed name (e.g.
// `duploservices-<tenant>-<short>`). Cross-resource references in body_json
// that point at AWS — `JobName`, `WorkflowName`, etc. — must use this value,
// not the short `name`.
func glueFullnameSchema() *schema.Schema {
	return &schema.Schema{
		Description: "The full AWS-known name of the resource, tenant-prefixed where applicable. " +
			"Nested resources whose names are not prefixed (e.g. tables) report the same value as `name`. " +
			"Reference this from another resource's `body_json` when AWS needs the prefixed form (e.g. trigger `JobName`).",
		Type:     schema.TypeString,
		Computed: true,
	}
}

// glueBodyJSONDiffSuppress suppresses a diff when the user's configured JSON
// is a deep subset of the JSON stored in state. AWS returns many computed
// fields (CreationTime, CatalogId, ResourceArn, etc.) that we keep in state
// but do not want to churn the plan over.
//
// Empty input is treated as `{}` so users can leave body_json unspecified
// when all configurable fields are typed elsewhere (e.g. minimal Databases).
func glueBodyJSONDiffSuppress(_, old, new string, _ *schema.ResourceData) bool {
	if old == new {
		return true
	}
	oldObj, ok := parseGlueBody(old)
	if !ok {
		return false
	}
	newObj, ok := parseGlueBody(new)
	if !ok {
		return false
	}
	return isJSONSubset(newObj, oldObj)
}

// parseGlueBody parses a body_json string into a value. The empty string maps
// to an empty object; malformed JSON returns ok=false.
func parseGlueBody(s string) (interface{}, bool) {
	if s == "" {
		return map[string]interface{}{}, true
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, false
	}
	return v, true
}

// isJSONSubset reports whether every leaf in `sub` is present and equal in
// `sup`. Maps recurse key-wise; arrays recurse element-wise (same length).
// Scalars require deep equality.
func isJSONSubset(sub, sup interface{}) bool {
	switch s := sub.(type) {
	case map[string]interface{}:
		m, ok := sup.(map[string]interface{})
		if !ok {
			return false
		}
		for k, v := range s {
			sv, ok := m[k]
			if !ok {
				return false
			}
			if !isJSONSubset(v, sv) {
				return false
			}
		}
		return true
	case []interface{}:
		l, ok := sup.([]interface{})
		if !ok || len(l) != len(s) {
			return false
		}
		for i, v := range s {
			if !isJSONSubset(v, l[i]) {
				return false
			}
		}
		return true
	case nil:
		return sup == nil
	default:
		return reflect.DeepEqual(sub, sup)
	}
}

// glueUnwrapDuploStringValues recursively replaces single-key {"Value": "X"}
// objects with the bare string. The DuploCloud backend serializes C# enum
// fields with this wrapper (DuploStringValue); AWS itself does not use this
// shape in Glue payloads, so unwrapping is safe and makes BE responses
// match the user's flat string form.
func glueUnwrapDuploStringValues(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		if len(x) == 1 {
			if s, ok := x["Value"].(string); ok {
				return s
			}
		}
		for k, val := range x {
			x[k] = glueUnwrapDuploStringValues(val)
		}
		return x
	case []interface{}:
		for i, val := range x {
			x[i] = glueUnwrapDuploStringValues(val)
		}
		return x
	default:
		return v
	}
}

// glueParseBodyJSON parses the user-supplied body_json into a map.
// Empty string yields an empty map (allowed when all fields are typed).
func glueParseBodyJSON(s string) (duplosdk.GlueResource, error) {
	if s == "" {
		return duplosdk.GlueResource{}, nil
	}
	out := duplosdk.GlueResource{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// glueMarshalBodyJSON marshals a Glue resource map back to a JSON string after
// stripping the typed top-level keys.
func glueMarshalBodyJSON(r duplosdk.GlueResource, stripKeys []string) (string, error) {
	if r == nil {
		return "", nil
	}
	out := make(duplosdk.GlueResource, len(r))
	for k, v := range r {
		out[k] = v
	}
	for _, k := range stripKeys {
		delete(out, k)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// glueResourceString reads a top-level string field from a Glue response.
func glueResourceString(r duplosdk.GlueResource, key string) string {
	if r == nil {
		return ""
	}
	if v, ok := r[key].(string); ok {
		return v
	}
	return ""
}

// glueTypedField describes how a typed Terraform attribute maps onto a key in
// the Glue API JSON body. The provider merges the value into the request body
// on write and projects it back out of the response on read; the key is also
// stripped from the body_json stored in state.
//
// If identity is true the field value comes from the ID at read time (e.g.
// the short "name" attribute) and must not be overwritten from the response —
// the backend returns the tenant-prefixed name there which would diverge from
// what the user wrote and from the ID.
type glueTypedField struct {
	tfKey    string // terraform schema attribute name (e.g. "name", "role")
	jsonKey  string // AWS Glue request/response key (e.g. "Name", "Role", "RegistryName")
	identity bool   // do not d.Set this from the response
}

// glueWrap describes the AWS-shape wrappers around Glue create/update bodies
// and Get responses. Several Glue APIs wrap the actual fields in an "*Input"
// envelope on the request side and an "*" envelope on the response side
// (e.g. DatabaseInput / Database, ConnectionInput / Connection, TableInput /
// Table). Crawlers/Jobs/Triggers/Workflows are unwrapped on request but
// wrapped on read. Registries and Schemas are flat on both.
type glueWrap struct {
	request  string // empty or e.g. "DatabaseInput"
	response string // empty or e.g. "Database"
	// preserveOnRead lists body_json keys that should be copied from prior
	// state to the freshly-read body_json. Use this for create-only fields
	// that the GET endpoint does not echo back (e.g. Glue SchemaDefinition,
	// which is only retrievable via a separate GetSchemaVersion call).
	preserveOnRead []string
}

// glueTwoPartID parses an ID of the form "tenantID/name".
func glueTwoPartID(id string) (tenantID, name string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form tenantID/name, got %q", id)
	}
	return parts[0], parts[1], nil
}

// glueThreePartID parses an ID of the form "tenantID/parentName/name".
func glueThreePartID(id string) (tenantID, parent, name string, err error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("expected ID in the form tenantID/parent/name, got %q", id)
	}
	return parts[0], parts[1], parts[2], nil
}

// glueBuildRequestBody parses body_json and merges the typed Terraform
// attributes into it, returning the final request body plus the list of typed
// keys to strip when reflecting state back. If wrap.request is non-empty the
// inner object is wrapped under that key (e.g. {"DatabaseInput": {...}}).
func glueBuildRequestBody(d *schema.ResourceData, typed []glueTypedField, wrap glueWrap) (duplosdk.GlueResource, []string, error) {
	inner, err := glueParseBodyJSON(d.Get("body_json").(string))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid body_json: %w", err)
	}
	stripKeys := make([]string, 0, len(typed))
	for _, f := range typed {
		if v, ok := d.Get(f.tfKey).(string); ok && v != "" {
			inner[f.jsonKey] = v
		}
		stripKeys = append(stripKeys, f.jsonKey)
	}
	if wrap.request != "" {
		return duplosdk.GlueResource{wrap.request: inner}, stripKeys, nil
	}
	return inner, stripKeys, nil
}

// glueRoleARNSuppressDiff treats an IAM role ARN and the bare role name as
// equivalent. The DuploCloud backend may store and return either form;
// users typically write the ARN. Compare on the trailing `/<name>` segment.
func glueRoleARNSuppressDiff(_, old, new string, _ *schema.ResourceData) bool {
	return glueRoleShortName(old) == glueRoleShortName(new)
}

// glueRoleShortName returns the role name component of an ARN, or the input
// unchanged if there is no `/`.
func glueRoleShortName(s string) string {
	if i := strings.LastIndex(s, "/"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// gluePrefixedName returns the backend-side prefixed name for a user-facing
// short name. The backend expects parent resource names (database, registry)
// in their prefixed form when used in URL paths.
func gluePrefixedName(c *duplosdk.Client, tenantID, shortName string) (string, error) {
	prefix, err := c.GetDuploServicesPrefix(tenantID, "")
	if err != nil {
		return "", err
	}
	return prefix + "-" + shortName, nil
}

// glueOverrideIdentity rewrites every identity-marked field in the request
// body with the given value, navigating into the request wrap key if set.
// Used on Update for resources whose AWS API requires the Input.Name to
// match the URL Name parameter — the BE prefixes the URL form but not the
// inner Input, so the provider supplies the prefixed value directly.
func glueOverrideIdentity(body duplosdk.GlueResource, typed []glueTypedField, wrap glueWrap, value string) {
	inner := body
	if wrap.request != "" {
		if v, ok := body[wrap.request].(duplosdk.GlueResource); ok {
			inner = v
		} else if v, ok := body[wrap.request].(map[string]interface{}); ok {
			inner = v
		}
	}
	for _, f := range typed {
		if f.identity {
			inner[f.jsonKey] = value
		}
	}
}

// glueApplyResponse writes typed fields and body_json back to Terraform state
// after a successful read.
//
//   - wrap.response unwraps the response (e.g. {"Database": {...}} -> {...}).
//   - DuploStringValue ({"Value": "X"}) objects are flattened to bare strings.
//   - typed fields are projected to their TF attributes (skipping identity).
//   - wrap.preserveOnRead pulls forward create-only fields from prior state
//     that the GET endpoint does not echo back.
func glueApplyResponse(d *schema.ResourceData, rp duplosdk.GlueResource, typed []glueTypedField, wrap glueWrap) error {
	if wrap.response != "" {
		if inner, ok := rp[wrap.response].(map[string]interface{}); ok {
			rp = inner
		}
	}
	if normalized, ok := glueUnwrapDuploStringValues(rp).(map[string]interface{}); ok {
		rp = normalized
	}
	stripKeys := make([]string, 0, len(typed))
	for _, f := range typed {
		stripKeys = append(stripKeys, f.jsonKey)
		if f.identity {
			d.Set("fullname", glueResourceString(rp, f.jsonKey))
			continue
		}
		d.Set(f.tfKey, glueResourceString(rp, f.jsonKey))
	}
	if len(wrap.preserveOnRead) > 0 {
		var prior map[string]interface{}
		if priorJSON := d.Get("body_json").(string); priorJSON != "" {
			_ = json.Unmarshal([]byte(priorJSON), &prior)
		}
		for _, k := range wrap.preserveOnRead {
			if _, present := rp[k]; present {
				continue
			}
			if v, ok := prior[k]; ok {
				rp[k] = v
			}
		}
	}
	body, err := glueMarshalBodyJSON(rp, stripKeys)
	if err != nil {
		return err
	}
	d.Set("body_json", body)
	return nil
}
