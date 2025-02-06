package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePlanKMSList() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms` manages the list of kms avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan kms for a specific plan.",

		ReadContext: dataSourcePlanKMSReadList,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan to configure.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_name": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourcePlanKMSReadList(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)

	log.Printf("[TRACE] dataSourcePlanKMSRead(%s): start", planID)
	c := m.(*duplosdk.Client)

	duplo, err := c.PlanKMSGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}
	count := 0
	if duplo != nil {
		count = len(*duplo)
	} else {

		d.SetId("")
		return nil

	}
	data := make([]map[string]interface{}, 0, count)
	// Set the simple fields first.
	for _, v := range *duplo {
		val := map[string]interface{}{
			"kms_name": v.KeyName,
			"kms_id":   v.KeyId,
			"kms_arn":  v.KeyArn,
		}
		data = append(data, val)
	}
	d.Set("data", data)
	// Build a list of current state, to replace the user-supplied settings.
	d.SetId(planID)

	log.Printf("[TRACE] dataSourcePlanKMSRead(%s): end", planID)
	return nil
}
