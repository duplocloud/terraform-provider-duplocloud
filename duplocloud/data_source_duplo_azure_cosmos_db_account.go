package duplocloud

import (
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAzureCosmosDBAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db_account` manages a azure cosmos db account in Duplo.",

		Read: dataSourceAzureCosmosDBAccountRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the host will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The name for availability set",
				Type:        schema.TypeString,
				Required:    true,
			},
			"kind": {
				Description: "Indicates the type of database account. This can only be set at database account creation. \n Allowed Account Kind : GlobalDocumentDB, MongoDB, Parse",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"locations": {
				Description: "An array that contains the georeplication locations enabled for the Cosmos DB account.",
				Type:        schema.TypeString,
				Computed:    true,
				//	Elem:        &schema.Schema{Type: schema.TypeMap},
			},
			"type": {
				Description: "Specifies the  Cosmos DB account type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consistency_policy": {
				Description: "",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_staleness_prefix": {
							Description: "Max number of stale requests tolerated. Accepted range for this values 1 to 2147483647",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"max_interval_in_seconds": {
							Description: "Max amount of time staleness (in seconds) is tolerated",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"default_consistency_level": {
							Description: "Specify the default consistency level and configuration settings of the Cosmos DB account. Possible values include: 'Eventual', 'Session', 'BoundedStaleness','Strong', 'ConsistentPrefix'",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"capabilities": {
				Description: "Name of the Cosmos DB capability", //' Current values also include 'EnableTable', 'EnableGremlin',
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name of the Cosmos DB capability, for example, 'EnableServerless'.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"capacity_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_policy": {
				Description: "Backup policy for cosmos db account",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_interval": {
							Description: "Backup interval in minutes",
							Computed:    true,
							Type:        schema.TypeInt,
						},

						"backup_retention_interval": {
							Description: "Backup retention interval in hours",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"backup_storage_redundancy": {
							Description: "Backup storage redundancy type. Valid values are Geo, Local, Zone. Defaults to Geo.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"type": {
							Description: "Valid values Periodic, Continuous",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"continuous_mode_tier": {
							Description: "The continuous mode tier for the Cosmos DB account. This is only applicable if the backup policy type is Continuous.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"disable_key_based_metadata_write_access": {
				Description: "Disable write operations on metadata resources (databases, containers, throughput) via account keys",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"enable_free_tier": {
				Description: "Flag to indicate whether Free Tier is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"public_network_access": {
				Description: "Flag to indicate whether to enable/disable public network access.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAzureCosmosDBAccountRead(d *schema.ResourceData, m interface{}) error {
	tenantId := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDBAccount(tenantId, name)
	if err != nil && err.Status() != 404 {
		return fmt.Errorf("Error fetching cosmos db account %s details for tenantId %s", name, tenantId)
	}
	if rp == nil {
		log.Printf("[DEBUG] Cosmos DB account %s not found for tenantId %s, removing from state", name, tenantId)
		d.SetId("")
		return nil
	}
	id := fmt.Sprintf("%s/cosmosdb/account/%s", tenantId, name)
	d.SetId(id)

	d.Set("tenant_id", tenantId)
	flattenAzureCosmosDBAccount(d, *rp)
	return nil

}
