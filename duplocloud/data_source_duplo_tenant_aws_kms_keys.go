package duplocloud

import (
	"encoding/json"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source listing secrets
func dataSourceTenantAwsKmsKeys() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsKmsKeysRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"selectable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// READ resource
func dataSourceTenantAwsKmsKeysRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsKmsKeysRead ******** 1 start")

	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	// Get all keys from duplo
	allKeys, err := c.TenantGetAllKmsKeys(tenantID)
	if err != nil {
		return err
	}

	// Allow the user to omit filter and include all keys (even ones with missing IDs)
	selectable, ok := d.GetOk("selectable")
	noFiltering := (ok && selectable != nil && !selectable.(bool))

	// Build a list of selected keys
	selectedKeys := make([]map[string]interface{}, 0, len(allKeys))
	for _, key := range allKeys {
		if noFiltering || key.KeyID != "" {
			selectedKeys = append(selectedKeys, map[string]interface{}{
				"key_id":   key.KeyID,
				"key_name": key.KeyName,
				"key_arn":  key.KeyArn,
			})
		}
	}
	d.SetId(tenantID)

	// Apply the result
	dumpKeys, _ := json.Marshal(selectedKeys)
	log.Printf("[TRACE] dataSourceTenantAwsKmsKeysRead ******** 2 dump: %s", dumpKeys)
	d.Set("keys", selectedKeys)

	log.Printf("[TRACE] dataSourceTenantAwsKmsKeysRead ******** 3 end")
	return nil
}
