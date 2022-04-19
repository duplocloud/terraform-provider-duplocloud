package duplocloud

import (
	"context"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SCHEMA for resource data/search
func dataSourceDuploEcsServices() *schema.Resource {
	itemSchema := ecsServiceComputedSchema()

	makeSchemaComputed(itemSchema["tenant_id"])
	makeSchemaComputed(itemSchema["name"])

	return &schema.Resource{
		ReadContext: dataSourceDuploEcsServicesRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: itemSchema,
				},
			},
		},
	}
}

func dataSourceDuploEcsServicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServicesRead: start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	list, err := c.EcsServiceList(tenantID)
	if err != nil {
		return diag.Errorf("Unable to list tenant %s services: %s", tenantID, err)
	}
	d.SetId(tenantID)

	// Build the list for TF state
	itemCount := 0
	if list != nil {
		itemCount = len(*list)
	}
	services := make([]map[string]interface{}, 0, itemCount)
	if itemCount > 0 {
		for _, duplo := range *list {

			// Get the base information from the replication controller.
			service := map[string]interface{}{
				"name":                              duplo.Name,
				"tenant_id":                         duplo.TenantID,
				"task_definition":                   duplo.TaskDefinition,
				"replicas":                          duplo.Replicas,
				"health_check_grace_period_seconds": duplo.HealthCheckGracePeriodSeconds,
				"old_task_definition_buffer_size":   duplo.OldTaskDefinitionBufferSize,
				"is_target_group_only":              duplo.IsTargetGroupOnly,
				"dns_prfx":                          duplo.DNSPrfx,
			}

			// Get the load balancer
			loadBalancers, err := flattenDuploEcsServiceLbs(&duplo, c)
			if err != nil {
				return diag.FromErr(err)
			}
			d.Set("load_balancer", loadBalancers)

			services = append(services, service)
		}
	}
	if err := d.Set("services", services); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceDuploServicesRead: end")
	return nil
}
