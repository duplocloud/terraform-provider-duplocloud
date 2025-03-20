package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planWafV2DataSourceSchema(single bool) map[string]*schema.Schema {

	// Create a fully computed schema.
	wafs_schema := planWafSchemaV2()
	for k := range wafs_schema {
		wafs_schema[k].Required = false
		wafs_schema[k].Computed = true
	}

	// For a single waf, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = wafs_schema
		result["name"].Computed = false
		result["name"].Required = true

		// For a list of wafs, move the list under the result key.
	} else {
		result = map[string]*schema.Schema{
			"wafs": {
				Description: "The list of wafs for this plan.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: wafs_schema,
				},
			},
		}

	}

	// Always require the plan ID.
	result["plan_id"] = &schema.Schema{
		Description: "The plan ID",
		Type:        schema.TypeString,
		Required:    true,
	}

	return result
}

func dataSourcePlanWafsV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_wafs` retrieves a list of wafs for a given plan.",

		ReadContext: dataSourcePlanWafsReadV2,
		Schema:      planWafV2DataSourceSchema(false),
	}
}

func dataSourcePlanWafV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_waf` retrieves details of a specific waf for a given plan.",

		ReadContext: dataSourcePlanWafReadV2,
		Schema:      planWafV2DataSourceSchema(true),
	}
}

func dataSourcePlanWafsReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanWafsRead(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanWafs(c, planID)
	if diags != nil {
		return diags
	}
	// Populate the results from the list.
	d.Set("wafs", flattenPlanWafs(all))

	d.SetId(planID + "/waf")

	log.Printf("[TRACE] dataSourcePlanWafsRead(%s): end", planID)
	return nil
}

// READ/SEARCH resources
func dataSourcePlanWafReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourcePlanWafRead(%s, %s): start", planID, name)

	// Get the plan certificate from Duplo.
	c := m.(*duplosdk.Client)
	duplo, diag := getPlanWaf(c, planID, name)
	if diag != nil {
		return diag
	}
	d.SetId(planID + "/waf/" + duplo.WebAclName)
	d.Set("name", duplo.WebAclName)
	d.Set("arn", duplo.WebAclId)
	d.Set("dashboard_url", duplo.DashboardUrl)
	log.Printf("[TRACE] dataSourcePlanWafRead(): end")
	return nil
}

func getPlanWaf(c *duplosdk.Client, planID, name string) (*duplosdk.DuploPlanWAF, diag.Diagnostics) {
	rsp, err := c.PlanWAFGet(planID, name)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan certificate for '%s/%s': %s", planID, name, err)
	}

	// If it failed, try the fallback method.
	if rsp == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		if plan.WafInfos != nil {
			for _, v := range *plan.WafInfos {
				if v.WebAclName == name {
					rsp = &v
				}
			}
		}
	}

	if rsp == nil {
		return nil, diag.Errorf("failed to retrieve plan waf for '%s/%s': %s", planID, name, err)
	}
	return rsp, nil
}

func getPlanWafs(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanWAF, diag.Diagnostics) {
	resp, err := c.PlanWAFGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan wafs for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if resp == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		resp = plan.WafInfos
	}

	return resp, nil
}

func planWafSchemaV2() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the waf  issued",
		},
		"arn": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ARN of the waf",
		},
		"dashboard_url": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The dashboard url associated to waf",
		},
	}
}
