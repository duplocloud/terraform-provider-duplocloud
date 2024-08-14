package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantCleanUpTimersSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the cleanup timers will be created in.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"expiry_time": {
			Description: "The expiry time of the tenant.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"pause_time": {
			Description: "The time to pause the tenant.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func dataSourceTenantCleanUpTimers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantCleanUpTimersRead,

		Schema: tenantCleanUpTimersSchema(),
	}
}

func dataSourceTenantCleanUpTimersRead(d *schema.ResourceData, m interface{}) error {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] dataSourceTenantCleanUpTimersRead(%s): start", tenantId)

	c := m.(*duplosdk.Client)
	tenantCleanUpTimers, err := c.GetTenantCleanUpTimers(tenantId)
	if err != nil {
		return fmt.Errorf("unable to retrieve tenant clean up timers for '%s': %s", tenantId, err)
	}

	if tenantCleanUpTimers == nil {
		d.SetId("") // object missing
		return nil
	}

	d.SetId(tenantId)
	_ = d.Set("expiry_time", tenantCleanUpTimers.ExpiryTime)
	_ = d.Set("pause_time", tenantCleanUpTimers.PauseTime)

	log.Printf("[TRACE] dataSourceTenantCleanUpTimersRead(%s): end", tenantId)
	return nil
}
