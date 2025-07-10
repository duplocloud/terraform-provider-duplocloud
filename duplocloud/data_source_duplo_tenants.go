package duplocloud

import (
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source listing tenants
func dataSourceTenants() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantsRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: tenantSchemaComputed(true),
				},
			},
		},
	}
}

// READ resource
func dataSourceTenantsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantsRead ******** start")

	// List the tenants from Duplo.
	c := m.(*duplosdk.Client)
	planID := ""
	var duploTenants []duplosdk.DuploTenant
	if v, ok := d.GetOk("plan_id"); ok && v != nil {
		planID = v.(string)
	}
	duploTenants, err := c.ListTenantsForUserByPlan(planID)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %s", err)
	}

	// Set the Terraform resource data
	tenants := make([]map[string]interface{}, 0, len(duploTenants))
	for _, duploTenant := range duploTenants {
		tenant := map[string]interface{}{
			"id":          duploTenant.TenantID,
			"name":        duploTenant.AccountName,
			"plan_id":     duploTenant.PlanID,
			"infra_owner": duploTenant.InfraOwner,
			"tags":        keyValueToState("tags", duploTenant.Tags),
		}
		if duploTenant.TenantPolicy != nil {
			tenant["policy"] = []map[string]interface{}{{
				"allow_volume_mapping": true,
				"block_external_ep":    true,
			}}
		}
		tenants = append(tenants, tenant)
	}

	d.SetId("user-tenants")
	d.Set("tenants", tenants)

	log.Printf("[TRACE] dataSourceTenantsRead ******** end")
	return nil
}
