package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source listing secrets
func dataSourceTenantAwsKmsKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsKmsKeyRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func dataSourceTenantAwsKmsKeyRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsKmsKeyRead ******** start")

	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	var kmsKey *duplosdk.DuploAwsKmsKey
	var err error

	// Allow getting information from a specific key by name, or just the default key.
	kmsKeyName, ok := d.GetOk("key_name")
	if ok && kmsKeyName != nil && kmsKeyName.(string) != "" {
		kmsKey, err = c.TenantGetKmsKeyByName(tenantID, kmsKeyName.(string))
	} else {
		kmsKey, err = c.TenantGetTenantKmsKey(tenantID)
	}
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, kmsKey.KeyID))
	d.Set("key_id", kmsKey.KeyID)
	d.Set("key_name", kmsKey.KeyName)
	d.Set("key_arn", kmsKey.KeyArn)

	log.Printf("[TRACE] dataSourceTenantAwsKmsKeyRead ******** end")
	return nil
}
