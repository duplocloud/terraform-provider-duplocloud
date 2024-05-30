package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePlanKMS() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms` manages the list of kms avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan kms for a specific plan.",

		ReadContext: dataSourcePlanKMSRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan to configure.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"kms_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePlanKMSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	name := d.Get("kms_name").(string)
	log.Printf("[TRACE] dataSourcePlanKMSRead(%s): start", planID)
	c := m.(*duplosdk.Client)

	duplo, err := c.PlanGetKMSKey(planID, name)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}
	// Set the simple fields first.
	d.Set("kms_id", duplo.KeyId)
	d.Set("kms_name", duplo.KeyName)
	d.Set("kms_arn", duplo.KeyArn)
	// Build a list of current state, to replace the user-supplied settings.
	d.SetId(planID + "/kms/" + duplo.KeyName)

	log.Printf("[TRACE] dataSourcePlanKMSRead(%s): end", planID)
	return nil
}
