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
	glueWorkflowTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
	}
	glueWorkflowWrap = glueWrap{response: "Workflow"}
)

func resourceDuploAwsGlueWorkflow() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_workflow` manages an AWS Glue Workflow in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the workflow name automatically.",
		CreateContext: resourceDuploAwsGlueWorkflowCreate,
		ReadContext:   resourceDuploAwsGlueWorkflowRead,
		UpdateContext: resourceDuploAwsGlueWorkflowUpdate,
		DeleteContext: resourceDuploAwsGlueWorkflowDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue workflow. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueWorkflowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueWorkflowCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueWorkflowTypedFields, glueWorkflowWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueWorkflowCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueWorkflowRead(ctx, d, m)
}

func resourceDuploAwsGlueWorkflowRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueWorkflowRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueWorkflowGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueWorkflowTypedFields, glueWorkflowWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueWorkflowUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueWorkflowUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueWorkflowTypedFields, glueWorkflowWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	prefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	glueOverrideIdentity(body, glueWorkflowTypedFields, glueWorkflowWrap, prefixed)
	if _, cerr := c.AwsGlueWorkflowUpdate(tenantID, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueWorkflowRead(ctx, d, m)
}

func resourceDuploAwsGlueWorkflowDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueWorkflowDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueWorkflowDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
