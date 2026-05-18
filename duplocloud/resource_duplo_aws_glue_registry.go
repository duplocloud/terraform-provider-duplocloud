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
	glueRegistryTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "RegistryName", identity: true},
	}
	glueRegistryWrap = glueWrap{} // flat request and response
)

func resourceDuploAwsGlueRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_registry` manages an AWS Glue Schema Registry in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the registry name automatically.",
		CreateContext: resourceDuploAwsGlueRegistryCreate,
		ReadContext:   resourceDuploAwsGlueRegistryRead,
		UpdateContext: resourceDuploAwsGlueRegistryUpdate,
		DeleteContext: resourceDuploAwsGlueRegistryDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue schema registry. The backend prepends `duploservices-<tenant>-` and sends it as `RegistryName`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueRegistryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueRegistryCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueRegistryTypedFields, glueRegistryWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueRegistryCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueRegistryRead(ctx, d, m)
}

func resourceDuploAwsGlueRegistryRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueRegistryRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueRegistryGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueRegistryTypedFields, glueRegistryWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueRegistryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueRegistryUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueRegistryTypedFields, glueRegistryWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	prefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	glueOverrideIdentity(body, glueRegistryTypedFields, glueRegistryWrap, prefixed)
	if _, cerr := c.AwsGlueRegistryUpdate(tenantID, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueRegistryRead(ctx, d, m)
}

func resourceDuploAwsGlueRegistryDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueRegistryDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueRegistryDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
