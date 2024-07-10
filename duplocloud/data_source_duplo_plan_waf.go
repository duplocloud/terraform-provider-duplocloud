package duplocloud

import (
	"context"
	"log"
	"strconv"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planWafSchema(single bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"waf_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"waf_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"dashboard_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
	return result
}

func dataSourcePlanWafs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plans` retrieves a list of plans from Duplo.",

		ReadContext: dataSourcePlanWafsRead,
		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan for waf.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: planWafSchema(false),
				},
			},
		},
	}
}

func dataSourcePlanWaf() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_waf` retrieves details of a plan in Duplo.",

		ReadContext: dataSourcePlanWafRead,
		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan for waf.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"waf_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"waf_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourcePlanWafsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourcePlanWafsRead(): start")
	planId := d.Get("plan_id").(string)
	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.PlanWAFGetList(planId)
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	data := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		plan := map[string]interface{}{
			//			"plan_id":       planId,
			"waf_name":      duplo.WebAclName,
			"waf_arn":       duplo.WebAclId,
			"dashboard_url": duplo.DashboardUrl,
		}

		data = append(data, plan)
	}

	if err := d.Set("data", data); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] dataSourcePlanWafsRead(): end")
	return nil
}

// READ/SEARCH resources
func dataSourcePlanWafRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanWafRead(%s): start", planID)
	name := d.Get("waf_name").(string)
	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan configs.
	duplo, err := c.PlanWAFGet(planID, name)
	if err != nil {
		return diag.Errorf("failed to retrieve plan waf for '%s': %s", planID, err)
	}
	d.Set("waf_name", duplo.WebAclName)
	d.Set("waf_arn", duplo.WebAclId)
	d.Set("dashboard_url", duplo.DashboardUrl)
	log.Printf("[TRACE] dataSourcePlanWafRead(%s): end", planID)
	return nil
}
