package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	glueSchemaTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "SchemaName", identity: true},
		{tfKey: "data_format", jsonKey: "DataFormat"},
		{tfKey: "compatibility", jsonKey: "Compatibility"},
	}
	// GetSchema does not echo SchemaDefinition (that lives on the schema
	// version, fetched separately). Preserve the user-supplied value across
	// reads so it doesn't churn the plan.
	glueSchemaWrap = glueWrap{preserveOnRead: []string{"SchemaDefinition"}}
)

func resourceDuploAwsGlueSchema() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_schema` manages an AWS Glue Schema under a registry in a DuploCloud tenant. " +
			"The provider prefixes `registry_name` with `duploservices-<tenant>-` when calling the backend, so reference the short registry name. " +
			"The schema name itself is also prepended with the tenant prefix server-side.",
		CreateContext: resourceDuploAwsGlueSchemaCreate,
		ReadContext:   resourceDuploAwsGlueSchemaRead,
		UpdateContext: resourceDuploAwsGlueSchemaUpdate,
		DeleteContext: resourceDuploAwsGlueSchemaDelete,
		Importer:      &schema.ResourceImporter{StateContext: resourceDuploAwsGlueSchemaImport},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"registry_name": {
				Description: "The short name of the parent Glue schema registry (without the `duploservices-<tenant>-` prefix).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The short name of the Glue schema. The backend sends it as `SchemaName`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"data_format": {
				Description:  "The data format of the schema. One of `AVRO`, `JSON`, `PROTOBUF`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"AVRO", "JSON", "PROTOBUF"}, false),
			},
			"compatibility": {
				Description: "The compatibility mode of the schema (e.g. `NONE`, `BACKWARD`, `FORWARD`, `FULL`).",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueSchemaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	regShort := d.Get("registry_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueSchemaCreate(%s, %s, %s): start", tenantID, regShort, name)

	c := m.(*duplosdk.Client)
	regPrefixed, err := gluePrefixedName(c, tenantID, regShort)
	if err != nil {
		return diag.FromErr(err)
	}
	body, _, err := glueBuildRequestBody(d, glueSchemaTypedFields, glueSchemaWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, cerr := c.AwsGlueSchemaCreate(tenantID, regPrefixed, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s/%s", tenantID, regShort, name))
	return resourceDuploAwsGlueSchemaRead(ctx, d, m)
}

func resourceDuploAwsGlueSchemaRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, regShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueSchemaRead(%s, %s, %s): start", tenantID, regShort, name)

	c := m.(*duplosdk.Client)
	regPrefixed, err := gluePrefixedName(c, tenantID, regShort)
	if err != nil {
		return diag.FromErr(err)
	}
	rp, cerr := c.AwsGlueSchemaGet(tenantID, regPrefixed, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("registry_name", regShort)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueSchemaTypedFields, glueSchemaWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueSchemaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, regShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueSchemaUpdate(%s, %s, %s): start", tenantID, regShort, name)

	c := m.(*duplosdk.Client)
	regPrefixed, err := gluePrefixedName(c, tenantID, regShort)
	if err != nil {
		return diag.FromErr(err)
	}
	namePrefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	// AWS UpdateSchema accepts only Compatibility, Description, and a
	// SchemaVersionNumber checkpoint. It rejects (with "Cannot update to the
	// same compatibility setting") when Compatibility matches the current
	// value, so only send it on actual change. Everything else in body_json
	// (SchemaDefinition, SchemaCheckpoint, LatestSchemaVersion, timestamps)
	// is read-only state that must not be sent back.
	body := duplosdk.GlueResource{"SchemaName": namePrefixed}
	if d.HasChange("compatibility") {
		if compat := d.Get("compatibility").(string); compat != "" {
			body["Compatibility"] = compat
			body["SchemaVersionNumber"] = duplosdk.GlueResource{"LatestVersion": true}
		}
	}
	if d.HasChange("body_json") {
		oldRaw, newRaw := d.GetChange("body_json")
		newBody, parseErr := glueParseBodyJSON(newRaw.(string))
		if parseErr != nil {
			return diag.FromErr(fmt.Errorf("body_json: %w", parseErr))
		}
		if desc, ok := newBody["Description"]; ok {
			body["Description"] = desc
		} else if oldBody, oldErr := glueParseBodyJSON(oldRaw.(string)); oldErr == nil {
			// Description was previously set in body_json and is now removed —
			// send "" so AWS clears it. Without this, the removal silently
			// no-ops because the diff suppression hides the field drop.
			if _, hadDesc := oldBody["Description"]; hadDesc {
				body["Description"] = ""
			}
		}
	}
	// No-op: nothing changed that AWS UpdateSchema can act on. Skip the call.
	if len(body) == 1 {
		return resourceDuploAwsGlueSchemaRead(ctx, d, m)
	}
	if _, cerr := c.AwsGlueSchemaUpdate(tenantID, regPrefixed, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueSchemaRead(ctx, d, m)
}

// resourceDuploAwsGlueSchemaImport pre-populates body_json with the schema's
// SchemaDefinition before the framework calls Read. This avoids a one-time
// post-import drift, because GetSchema does not echo SchemaDefinition (it
// lives on the schema version, which is fetched separately). preserveOnRead
// then carries the definition through future refreshes.
func resourceDuploAwsGlueSchemaImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	tenantID, regShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return nil, err
	}
	c := m.(*duplosdk.Client)
	regPrefixed, err := gluePrefixedName(c, tenantID, regShort)
	if err != nil {
		return nil, err
	}
	versions, cerr := c.AwsGlueSchemaVersionList(tenantID, regPrefixed, name)
	if cerr != nil {
		log.Printf("[WARN] resourceDuploAwsGlueSchemaImport(%s, %s, %s): version list failed: %s", tenantID, regShort, name, cerr.Error())
		return []*schema.ResourceData{d}, nil
	}
	latestID, latestVer := "", 0.0
	for _, v := range versions {
		id, _ := v["SchemaVersionId"].(string)
		num, _ := v["VersionNumber"].(float64)
		if id == "" {
			continue
		}
		if num > latestVer {
			latestVer = num
			latestID = id
		}
	}
	if latestID == "" {
		return []*schema.ResourceData{d}, nil
	}
	ver, cerr := c.AwsGlueSchemaVersionGet(tenantID, regPrefixed, name, latestID)
	if cerr != nil {
		log.Printf("[WARN] resourceDuploAwsGlueSchemaImport(%s, %s, %s): version get failed: %s", tenantID, regShort, name, cerr.Error())
		return []*schema.ResourceData{d}, nil
	}
	def, ok := ver["SchemaDefinition"].(string)
	if !ok || def == "" {
		return []*schema.ResourceData{d}, nil
	}
	body := duplosdk.GlueResource{"SchemaDefinition": def}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	d.Set("body_json", string(b))
	return []*schema.ResourceData{d}, nil
}

func resourceDuploAwsGlueSchemaDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, regShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueSchemaDelete(%s, %s, %s): start", tenantID, regShort, name)

	c := m.(*duplosdk.Client)
	regPrefixed, err := gluePrefixedName(c, tenantID, regShort)
	if err != nil {
		return diag.FromErr(err)
	}
	if cerr := c.AwsGlueSchemaDelete(tenantID, regPrefixed, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
