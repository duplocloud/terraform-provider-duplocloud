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
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The endpoint used to connect to the CosmosDB account.",
			},
			"read_endpoints": {
				Type:        schema.TypeList,
				Description: "The list of read endpoints for the CosmosDB account.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"write_endpoints": {
				Type:        schema.TypeList,
				Description: "The list of write endpoints for the CosmosDB account.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"primary_master_key": {
				Type:        schema.TypeString,
				Description: "The primary key for the CosmosDB account.",
				Computed:    true,
				Sensitive:   true, // Sensitive information
			},
			"secondary_master_key": {
				Type:        schema.TypeString,
				Description: "The secondary key for the CosmosDB account.",
				Computed:    true,
				Sensitive:   true, // Sensitive information
			},
			"primary_readonly_master_key": {
				Type:        schema.TypeString,
				Description: "The primary readonly key for the CosmosDB account.",
				Computed:    true,
				Sensitive:   true, // Sensitive information
			},
			"secondary_readonly_master_key": {
				Type:        schema.TypeString,
				Description: "The secondary readonly key for the CosmosDB account.",
				Computed:    true,
				Sensitive:   true, // Sensitive information
			},
			"primary_sql_connection_string": {
				Type:        schema.TypeString,
				Description: "The primary SQL connection string for the CosmosDB account.",
				Computed:    true,
			},
			"secondary_sql_connection_string": {
				Type:        schema.TypeString,
				Description: "The secondary SQL connection string for the CosmosDB account.",
				Computed:    true,
			},
			"primary_readonly_sql_connection_string": {
				Type:        schema.TypeString,
				Description: "The primary readonly SQL connection string for the CosmosDB account.",
				Computed:    true,
			},
			"secondary_readonly_sql_connection_string": {
				Type:        schema.TypeString,
				Description: "The secondary readonly SQL connection string for the CosmosDB account.",
				Computed:    true,
			},
			"primary_mongo_connection_string": {
				Type:        schema.TypeString,
				Description: "The primary MongoDB connection string for the CosmosDB account.",
				Computed:    true,
			},
			"secondary_mongo_connection_string": {
				Type:        schema.TypeString,
				Description: "The secondary MongoDB connection string for the CosmosDB account.",
				Computed:    true,
			},
			"primary_readonly_mongo_connection_string": {
				Type:        schema.TypeString,
				Description: "The primary readonly MongoDB connection string for the CosmosDB account.",
				Computed:    true,
			},
			"secondary_readonly_mongo_connection_string": {
				Type:        schema.TypeString,
				Description: "The secondary readonly MongoDB connection string for the CosmosDB account.",
				Computed:    true,
			},
			"geo_location": {
				Type:        schema.TypeList,
				Description: "Specifies a geo_location resource, used to define where data should be replicated with the failover_priority 0 specifying the primary location",
				Computed:    true,

				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location_name": {
							Description: "The name of the Azure region to host replicated data",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"failover_priority": {
							Description: "The failover priority of the region. A failover priority of 0 indicates a write region. The maximum value for a failover priority = (total number of regions - 1). Failover priority values must be unique for each of the regions in which the database account exists. Changing this causes the location to be re-provisioned and cannot be changed for the location with failover priority 0",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"is_zone_redundant": {
							Description: "Should zone redundancy be enabled for this region?",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
			"is_virtual_network_filter_enabled": {
				Description: "Enables virtual network filtering for this Cosmos DB account.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"virtual_network_rule": {
				Description: "A list of virtual network rules for the Cosmos DB account. This is used to define which subnets are allowed to access this CosmosDB account",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Description: "The ID of the subnet to allow access to this CosmosDB account. This should be in the format /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"ignore_missing_vnet_service_endpoint": {
							Description: "If set to true, the specified subnet will be added as a virtual network rule even if its CosmosDB service endpoint is not active",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
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
