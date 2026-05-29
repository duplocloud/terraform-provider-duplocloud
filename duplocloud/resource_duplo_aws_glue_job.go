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
	glueJobTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
		{tfKey: "role", jsonKey: "Role"},
	}
	glueJobWrap = glueWrap{response: "Job"}
)

func resourceDuploAwsGlueJob() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_job` manages an AWS Glue Job in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the job name automatically. " +
			"The IAM role must have a trust relationship with `glue.amazonaws.com` and permissions for the script's data sources.",
		CreateContext: resourceDuploAwsGlueJobCreate,
		ReadContext:   resourceDuploAwsGlueJobRead,
		UpdateContext: resourceDuploAwsGlueJobUpdate,
		DeleteContext: resourceDuploAwsGlueJobDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue job. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"role": {
				Description:      "The ARN (or short name) of the IAM role the job assumes.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: glueRoleARNSuppressDiff,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueJobCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueJobTypedFields, glueJobWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueJobCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueJobRead(ctx, d, m)
}

func resourceDuploAwsGlueJobRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueJobRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueJobGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueJobTypedFields, glueJobWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueJobUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueJobTypedFields, glueJobWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	// AWS UpdateJob takes {JobName, JobUpdate{...}} — JobUpdate has no Name
	// field and the shape differs from CreateJob's flat form. Strip Name and
	// wrap the rest in JobUpdate so the backend can forward it to AWS.
	delete(body, "Name")
	wrapped := duplosdk.GlueResource{"JobUpdate": body}
	if _, cerr := c.AwsGlueJobUpdate(tenantID, name, wrapped); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueJobRead(ctx, d, m)
}

func resourceDuploAwsGlueJobDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueJobDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueJobDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
