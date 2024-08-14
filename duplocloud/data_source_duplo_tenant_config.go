package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for secrets
func tenantConfigSchemaComputed() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"metadata": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// Data source retrieving a secret
func dataSourceTenantConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantConfigRead,

		Schema: tenantConfigSchemaComputed(),
	}
}

// READ resource
func dataSourceTenantConfigRead(d *schema.ResourceData, m interface{}) error {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] dataSourceTenantConfigRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetConfig(tenantID)
	if err != nil {
		return fmt.Errorf("unable to retrieve tenant config for '%s': %s", tenantID, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the fields
	d.SetId(duplo.TenantID)
	_ = d.Set("metadata", keyValueToState("metadata", duplo.Metadata))

	log.Printf("[TRACE] dataSourceTenantConfigRead(%s): end", tenantID)
	return nil
}
