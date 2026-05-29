package duplocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	glueTableTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
	}
	glueTableWrap = glueWrap{request: "TableInput", response: "Table"}
)

func resourceDuploAwsGlueTable() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_table` manages an AWS Glue Data Catalog table under a database in a DuploCloud tenant. " +
			"The provider prefixes `database_name` with `duploservices-<tenant>-` when calling the backend, so reference the short database name.",
		CreateContext: resourceDuploAwsGlueTableCreate,
		ReadContext:   resourceDuploAwsGlueTableRead,
		UpdateContext: resourceDuploAwsGlueTableUpdate,
		DeleteContext: resourceDuploAwsGlueTableDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"database_name": {
				Description: "The short name of the parent Glue database (without the `duploservices-<tenant>-` prefix).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the Glue table. Table names are not prefixed.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	dbShort := d.Get("database_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueTableCreate(%s, %s, %s): start", tenantID, dbShort, name)

	c := m.(*duplosdk.Client)
	dbPrefixed, err := gluePrefixedName(c, tenantID, dbShort)
	if err != nil {
		return diag.FromErr(err)
	}
	body, _, err := glueBuildRequestBody(d, glueTableTypedFields, glueTableWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, cerr := c.AwsGlueTableCreate(tenantID, dbPrefixed, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s/%s", tenantID, dbShort, name))
	return resourceDuploAwsGlueTableRead(ctx, d, m)
}

func resourceDuploAwsGlueTableRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, dbShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTableRead(%s, %s, %s): start", tenantID, dbShort, name)

	c := m.(*duplosdk.Client)
	dbPrefixed, err := gluePrefixedName(c, tenantID, dbShort)
	if err != nil {
		return diag.FromErr(err)
	}
	rp, cerr := c.AwsGlueTableGet(tenantID, dbPrefixed, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("database_name", dbShort)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueTableTypedFields, glueTableWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueTableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, dbShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTableUpdate(%s, %s, %s): start", tenantID, dbShort, name)

	c := m.(*duplosdk.Client)
	dbPrefixed, err := gluePrefixedName(c, tenantID, dbShort)
	if err != nil {
		return diag.FromErr(err)
	}
	body, _, err := glueBuildRequestBody(d, glueTableTypedFields, glueTableWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, cerr := c.AwsGlueTableUpdate(tenantID, dbPrefixed, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueTableRead(ctx, d, m)
}

func resourceDuploAwsGlueTableDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, dbShort, name, err := glueThreePartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTableDelete(%s, %s, %s): start", tenantID, dbShort, name)

	c := m.(*duplosdk.Client)
	dbPrefixed, err := gluePrefixedName(c, tenantID, dbShort)
	if err != nil {
		return diag.FromErr(err)
	}
	if cerr := c.AwsGlueTableDelete(tenantID, dbPrefixed, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
