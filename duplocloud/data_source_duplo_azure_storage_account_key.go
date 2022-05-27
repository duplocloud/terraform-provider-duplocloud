package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAzureStorageAccountKey() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storage_account_key` retrieves a azure storage account key in Duplo.",

		Read: dataSourceAzureStorageAccountKeyRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"storage_account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAzureStorageAccountKeyRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceAzureStorageAccountKeyRead ******** start")

	tenantID := d.Get("tenant_id").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	c := m.(*duplosdk.Client)

	keyValue, err := c.StorageAccountGetKey(tenantID, storageAccountName)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", tenantID, storageAccountName))

	d.Set("key_name", keyValue.Key)
	d.Set("key_value", keyValue.Value)

	log.Printf("[TRACE] dataSourceAzureStorageAccountKeyRead ******** end")
	return nil
}
