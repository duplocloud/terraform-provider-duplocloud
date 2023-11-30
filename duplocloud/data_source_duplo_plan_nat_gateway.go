package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planNgwSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"state": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"vpc_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"subnet_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"addresses": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allocation_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"network_interface_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"private_ip": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"public_ip": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

func planNgwDataSourceSchema() map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"nat_gateways": {
			Description: "The list of NAT gateways for this plan.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: planNgwSchema(),
			},
		},
	}

	// Always require the plan ID.
	result["plan_id"] = &schema.Schema{
		Description: "The plan ID",
		Type:        schema.TypeString,
		Required:    true,
	}

	return result
}

func dataSourcePlanNgws() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_nat_gateways` retrieves a list of NAT gateways for a given plan.",

		ReadContext: dataSourcePlanNgwsRead,
		Schema:      planNgwDataSourceSchema(),
	}
}

func dataSourcePlanNgwsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanNgwsRead(%s): start", planID)

	c := m.(*duplosdk.Client)
	all, err := c.PlanNgwGetList(planID)
	if err != nil {
		return diag.Errorf("failed to read NAT gateway for plan: %s", err)
	}
	d.Set("nat_gateways", flattenPlanNgws(all))

	d.SetId(planID)

	log.Printf("[TRACE] dataSourcePlanNgwsRead(%s): end", planID)
	return nil
}

func flattenPlanNgws(list *[]duplosdk.DuploPlanNgw) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, ngw := range *list {
		result = append(result, map[string]interface{}{
			"id":        ngw.NatGatewayId,
			"state":     ngw.State.Value,
			"vpc_id":    ngw.VpcId,
			"subnet_id": ngw.SubnetId,
			"addresses": flattenPlanNgwAddresses(ngw.NatGatewayAddresses),
			"tags":      keyValueToState("tags", ngw.Tags),
		})
	}
	return result
}

func flattenPlanNgwAddresses(duplo *[]duplosdk.DuploPlanNgwAddress) []map[string]interface{} {
	if duplo == nil {
		return []map[string]interface{}{}
	}

	list := []map[string]interface{}{}
	for _, item := range *duplo {
		list = append(list, map[string]interface{}{
			"allocation_id":        item.AllocationId,
			"network_interface_id": item.NetworkInterfaceId,
			"private_ip":           item.PrivateIP,
			"public_ip":            item.PublicIP,
		})
	}
	return list
}
