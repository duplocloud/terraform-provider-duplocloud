package duplocloud

import (
	"context"
	"strconv"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planSchema(single bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"cloud": {
			Description: "The numerical index of the cloud provider for this plan" +
				"Will be one of:\n\n" +
				"   - `0` : AWS\n" +
				"   - `2` : Azure\n",
			Type:     schema.TypeInt,
			Computed: true,
		},
		"region": {
			Description: "The cloud provider region.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"vpc_id": {
			Description: "The VPC or VNet ID.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_subnet_ids": {
			Description: "The private subnets for the VPC or VNet.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"public_subnet_ids": {
			Description: "The public subnets for the VPC or VNet.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}

	if single {
		result["plan_id"] = &schema.Schema{
			Description: "The plan ID",
			Type:        schema.TypeString,
			Required:    true,
		}
	} else {
		result["plan_id"] = &schema.Schema{
			Description: "The plan ID",
			Type:        schema.TypeString,
			Computed:    true,
		}
	}
	return result
}

func dataSourcePlans() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plans` retrieves a list of infrastructures in Duplo.",

		ReadContext: dataSourcePlansRead,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: planSchema(false),
				},
			},
		},
	}
}

func dataSourcePlan() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan` retrieves details of a plan in Duplo.",

		ReadContext: dataSourcePlanRead,
		Schema:      planSchema(true),
	}
}

func dataSourcePlansRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourcePlansRead(): start")

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.PlanGetList()
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	data := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		plan := map[string]interface{}{
			"plan_id": duplo.Name,
		}
		flattenPlanCloudConfig(plan, &duplo)

		data = append(data, plan)
	}

	if err := d.Set("data", data); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] dataSourcePlansRead(): end")
	return nil
}

/// READ/SEARCH resources
func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourcePlanRead(): start")

	c := m.(*duplosdk.Client)
	var name string

	// Look up by tenant ID, if requested.
	if v, ok := d.GetOk("tenant_id"); ok && v != nil && v.(string) != "" {
		tenant, err := c.TenantGet(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		name = tenant.PlanID // plan ID always matches infra name
	} else {
		name = d.Get("infra_name").(string)
	}

	// Get the object from Duplo, detecting a missing object
	missing, err := infrastructureRead(c, d, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure '%s': %s", name, err)
	}
	if missing {
		return diag.Errorf("No infrastructure named '%s' was found", name)
	}
	d.SetId("-")

	log.Printf("[TRACE] dataSourcePlanRead(): end")
	return nil
}

func flattenPlanCloudConfig(plan map[string]interface{}, duplo *duplosdk.DuploPlan) {
	if duplo.AwsConfig != nil && len(duplo.AwsConfig) > 0 {
		plan["vpc_id"] = duplo.AwsConfig["VpcId"]
		plan["region"] = duplo.AwsConfig["AwsRegion"]
		if duplo.AwsConfig["AwsElbSubnet"] != nil {
			plan["public_subnet_ids"] = strings.Split(duplo.AwsConfig["AwsElbSubnet"].(string), ";")
		}
		if duplo.AwsConfig["AwsInternalElbSubnet"] != nil {
			plan["private_subnet_ids"] = strings.Split(duplo.AwsConfig["AwsInternalElbSubnet"].(string), ";")
		}
	}
}
