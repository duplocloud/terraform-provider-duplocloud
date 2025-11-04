package duplocloud

import (
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAzureCosmosDBDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db_database` manages a azure cosmos db database in Duplo.",

		Read: dataSourceAzureCosmosDBDatabaseRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant in which the Cosmos DB database will be created for specified account",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The name for Cosmsos DB database.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"account_name": {
				Description: "Indicates on which account database will be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Specifies the Cosmos DB account type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace": {
				Description: "The Azure resource provider namespace for Cosmos DB. Defaults to 'Microsoft.DocumentDB'. This value is typically not changed and identifies the resource type within Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAzureCosmosDBDatabaseRead(d *schema.ResourceData, m interface{}) error {
	tenantId := d.Get("tenant_id").(string)
	accountName := d.Get("account_name").(string)
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDB(tenantId, accountName, name)
	if err != nil && err.Status() != 404 {
		return fmt.Errorf("Error fetching cosmos db database %s from account %s details for tenantId %s :%s", name, accountName, tenantId, err.Error())
	}
	if rp == nil {
		log.Printf("[DEBUG] dataSourceAzureCosmosDBDatabaseRead: Cosmos DB database %s from account %s for tenantId %s not found, removing from state", name, accountName, tenantId)
		d.SetId("")
		return nil
	}
	id := fmt.Sprintf("%s/cosmosdb/%s/database/%s", tenantId, accountName, name)
	d.SetId(id)
	d.Set("tenant_id", tenantId)
	d.Set("account_name", accountName)
	flattenAzureCosmosDB(d, *rp)
	return nil
}
