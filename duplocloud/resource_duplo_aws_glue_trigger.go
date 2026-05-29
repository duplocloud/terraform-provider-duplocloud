package duplocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	glueTriggerTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
		{tfKey: "type", jsonKey: "Type"},
	}
	// StartOnCreation is a create-only field that GetTrigger never echoes back;
	// preserve it from prior state so it doesn't surface as perpetual drift.
	glueTriggerWrap = glueWrap{response: "Trigger", preserveOnRead: []string{"StartOnCreation"}}
)

func resourceDuploAwsGlueTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_trigger` manages an AWS Glue Trigger in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the trigger name automatically.",
		CreateContext: resourceDuploAwsGlueTriggerCreate,
		ReadContext:   resourceDuploAwsGlueTriggerRead,
		UpdateContext: resourceDuploAwsGlueTriggerUpdate,
		DeleteContext: resourceDuploAwsGlueTriggerDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue trigger. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description:  "The trigger type. One of `SCHEDULED`, `CONDITIONAL`, `ON_DEMAND`, `EVENT`. Cannot be changed after creation.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"SCHEDULED", "CONDITIONAL", "ON_DEMAND", "EVENT"}, false),
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueTriggerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueTriggerCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueTriggerTypedFields, glueTriggerWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueTriggerCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueTriggerRead(ctx, d, m)
}

func resourceDuploAwsGlueTriggerRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTriggerRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueTriggerGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueTriggerTypedFields, glueTriggerWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueTriggerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTriggerUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueTriggerTypedFields, glueTriggerWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	// AWS UpdateTrigger takes {Name, TriggerUpdate{...}} — TriggerUpdate has
	// no Name or Type field. Strip both and wrap so the backend can forward
	// to AWS without "triggerUpdate must not be null".
	delete(body, "Name")
	delete(body, "Type")
	wrapped := duplosdk.GlueResource{"TriggerUpdate": body}
	if _, cerr := c.AwsGlueTriggerUpdate(tenantID, name, wrapped); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueTriggerRead(ctx, d, m)
}

func resourceDuploAwsGlueTriggerDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueTriggerDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueTriggerDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
