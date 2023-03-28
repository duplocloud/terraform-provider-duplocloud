package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceTenantAwsRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsRegionRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ resource
func dataSourceTenantAwsRegionRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsRegionRead ******** start")

	// Get the region from Duplo.
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	awsRegion, err := c.TenantGetAwsRegion(tenantID)
	d.SetId(tenantID)
	if err != nil {
		return fmt.Errorf("failed to read AWS region from tenant '%s': %s", tenantID, err)
	}

	// Set the Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("aws_region", awsRegion)

	log.Printf("[TRACE] dataSourceTenantAwsRegionRead ******** end")
	return nil
}
