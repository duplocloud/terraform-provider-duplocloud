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
	glueConnectionTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
	}
	glueConnectionWrap = glueWrap{request: "ConnectionInput", response: "Connection"}
)

func resourceDuploAwsGlueConnection() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_connection` manages an AWS Glue Connection in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the connection name automatically.",
		CreateContext: resourceDuploAwsGlueConnectionCreate,
		ReadContext:   resourceDuploAwsGlueConnectionRead,
		UpdateContext: resourceDuploAwsGlueConnectionUpdate,
		DeleteContext: resourceDuploAwsGlueConnectionDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue connection. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueConnectionCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueConnectionTypedFields, glueConnectionWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueConnectionCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueConnectionRead(ctx, d, m)
}

func resourceDuploAwsGlueConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueConnectionRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueConnectionGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueConnectionTypedFields, glueConnectionWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueConnectionUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueConnectionTypedFields, glueConnectionWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	prefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	glueOverrideIdentity(body, glueConnectionTypedFields, glueConnectionWrap, prefixed)
	if _, cerr := c.AwsGlueConnectionUpdate(tenantID, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueConnectionRead(ctx, d, m)
}

func resourceDuploAwsGlueConnectionDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueConnectionDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueConnectionDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
