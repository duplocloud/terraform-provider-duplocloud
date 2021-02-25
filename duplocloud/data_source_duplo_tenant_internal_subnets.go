package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source listing secrets
func dataSourceTenantInternalSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantSecretsRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

/// READ resource
func dataSourceTenantInternalSubnetsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantInternalSubnetsRead ******** start")

	// List the secrets from Duplo.
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	subnetIDs, err := c.TenantGetInternalSubnets(tenantID)
	if err != nil {
		return fmt.Errorf("Failed to list subnets: %s", err)
	}
	d.SetId(tenantID)
	d.Set("subnet_ids", subnetIDs)

	log.Printf("[TRACE] dataSourceTenantInternalSubnetsRead ******** end")
	return nil
}
