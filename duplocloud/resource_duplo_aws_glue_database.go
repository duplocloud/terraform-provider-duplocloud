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
	glueDatabaseTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
	}
	glueDatabaseWrap = glueWrap{request: "DatabaseInput", response: "Database"}
)

func resourceDuploAwsGlueDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_database` manages an AWS Glue Data Catalog database in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the database name automatically.",
		CreateContext: resourceDuploAwsGlueDatabaseCreate,
		ReadContext:   resourceDuploAwsGlueDatabaseRead,
		UpdateContext: resourceDuploAwsGlueDatabaseUpdate,
		DeleteContext: resourceDuploAwsGlueDatabaseDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue database. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueDatabaseCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueDatabaseTypedFields, glueDatabaseWrap)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueDatabaseCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueDatabaseRead(ctx, d, m)
}

func resourceDuploAwsGlueDatabaseRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueDatabaseGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueDatabaseTypedFields, glueDatabaseWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueDatabaseUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueDatabaseTypedFields, glueDatabaseWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	prefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	glueOverrideIdentity(body, glueDatabaseTypedFields, glueDatabaseWrap, prefixed)

	if _, cerr := c.AwsGlueDatabaseUpdate(tenantID, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueDatabaseRead(ctx, d, m)
}

func resourceDuploAwsGlueDatabaseDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueDatabaseDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
