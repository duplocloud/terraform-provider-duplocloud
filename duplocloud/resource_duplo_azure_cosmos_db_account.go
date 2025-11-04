package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureCosmosDBAccountchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the host will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description:  "The name for availability set",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9._-]{0,78}[a-zA-Z0-9_]$"), `The length must be between 1 and 80 characters. The first character must be a letter or number. The last character must be a letter, number, or underscore. The remaining characters must be letters, numbers, periods, underscores, or dashes.`),
		},
		"kind": {
			Description: `Indicates the type of database account. This can only be set at database account creation.
			Allowed Account Kind : GlobalDocumentDB.
			Future support MongoDB, Parse`,
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"GlobalDocumentDB"}, false), //, "MongoDB", "Parse"}, false),
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
			Default:     "Microsoft.DocumentDB/databaseAccounts",
			Optional:    true,
		},
		"consistency_policy": {
			Description: "Specify the consistency policy for the Cosmos DB account.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_staleness_prefix": {
						Description:  "When used with the 'Bounded Staleness' consistency level, this value represents the number of stale requests tolerated. The accepted range for this value is 10 – 2147483647. Defaults to 100. Required when 'consistency_level' is set to 'BoundedStaleness'",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(1, 2147483647),
					},
					"max_interval_in_seconds": {
						Description:  "When used with the 'Bounded Staleness' consistency level, this value represents the time amount of staleness (in seconds) tolerated. The accepted range for this value is 5 - 86400 (1 day). Required when consistency_level is set to BoundedStaleness.",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(5, 86400),
					},
					"default_consistency_level": {
						Description:  "Specify the default consistency level and configuration settings of the Cosmos DB account. Possible values include: 'Eventual', 'Session', 'BoundedStaleness','Strong', 'ConsistentPrefix'",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"Eventual", "Session", "BoundedStaleness", "Strong", "ConsistentPrefix"}, false),
					},
				},
			},
		},
		"capabilities": {
			Description: "Name of the Cosmos DB capability", //' Current values also include 'EnableTable', 'EnableGremlin',
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description:  "Name of the Cosmos DB capability, for example, 'EnableServerless'.",
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{ /*"EnableCassandra", "EnableTable", "EnableGremlin",*/ "EnableServerless"}, false),
						Required:     true,
					},
				},
			},
		},
		"capacity_mode": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"backup_policy": {
			Description: `Backup policy for cosmos db account. 
			> ⚠️ **Note:**: 
			> You can only configure backup_interval, backup_retention_interval and backup_storage_redundancy when the type field is set to periodic`,
			Type:     schema.TypeList,
			Computed: true,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"backup_interval": {
						Description: "Backup interval in minutes. Can be configured when type is set to Periodic",
						Optional:    true,
						Type:        schema.TypeInt,
					},

					"backup_retention_interval": {
						Description: "Backup retention interval in hours",
						Optional:    true,
						Type:        schema.TypeInt,
					},
					"backup_storage_redundancy": {
						Description:  "Backup storage redundancy type. Valid values are Geo, Local, Zone. Defaults to Geo.",
						Optional:     true,
						Type:         schema.TypeString,
						Computed:     true,
						ValidateFunc: validation.StringInSlice([]string{"Geo", "Local", "Zone"}, false),
					},
					"type": {
						Description: `The type of backup. Possible values are Periodic and Continuous
						> ⚠️ **Note:**: 
						> Update from Periodic to Continuous type is allowed. To change from Periodic to Continuous resource need to be recreated`,
						Optional:     true,
						Type:         schema.TypeString,
						Default:      "Periodic",
						ValidateFunc: validation.StringInSlice([]string{"Periodic", "Continuous"}, false),
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
			Optional:    true,
			Default:     false,
		},
		"enable_free_tier": {
			Description: "Flag to indicate whether Free Tier is enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true,
		},
		"public_network_access": {
			Description:  "Flag to indicate whether to enable/disable public network access.",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "Enabled",
			ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
		},

		// ... (rest of the schema unchanged)
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
			Required:    true,
			ForceNew:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"location_name": {
						Description: "The name of the Azure region to host replicated data",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"failover_priority": {
						Description: "The failover priority of the region. A failover priority of 0 indicates a write region. The maximum value for a failover priority = (total number of regions - 1). Failover priority values must be unique for each of the regions in which the database account exists. Changing this causes the location to be re-provisioned and cannot be changed for the location with failover priority 0",
						Type:        schema.TypeInt,
						Required:    true,
					},
					"is_zone_redundant": {
						Description: "Should zone redundancy be enabled for this region?",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						ForceNew:    true,
					},
				},
			},
		},
		"is_virtual_network_filter_enabled": {
			Description: "Enables virtual network filtering for this Cosmos DB account.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"virtual_network_rule": {
			Description: "A list of virtual network rules for the Cosmos DB account. This is used to define which subnets are allowed to access this CosmosDB account",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// Suppress diff if both old and new have the same set of subnet_ids, regardless of order
				oldRaw, newRaw := d.GetChange("virtual_network_rule")
				oldSlice, okOld := oldRaw.([]interface{})
				newSlice, okNew := newRaw.([]interface{})
				if !okOld || !okNew {
					return false
				}
				if len(oldSlice) != len(newSlice) {
					return false
				}
				if len(oldSlice) == 0 {
					return true
				}
				// Collect subnet_ids from both lists
				oldMap := make(map[string]bool)
				for _, v := range oldSlice {
					if m, ok := v.(map[string]interface{}); ok {
						if id, ok := m["subnet_id"].(string); ok {
							oldMap[id] = true
						}
					}
				}
				newMap := make(map[string]bool)
				for _, v := range newSlice {
					if m, ok := v.(map[string]interface{}); ok {
						if id, ok := m["subnet_id"].(string); ok {
							newMap[id] = true
						}
					}
				}
				if len(oldMap) != len(newMap) {
					return false
				}
				for id := range oldMap {
					if !newMap[id] {
						return false
					}
				}
				return true
			},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"subnet_id": {
						Description: "The ID of the subnet to allow access to this CosmosDB account. This should be in the format /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}",
						Type:        schema.TypeString,
						Required:    true,
					},
					"ignore_missing_vnet_service_endpoint": {
						Description: "If set to true, the specified subnet will be added as a virtual network rule even if its CosmosDB service endpoint is not active",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
				},
			},
		},
	}
}

func resourceAzureCosmosDBAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db_account` manages cosmos db account resource for azure",

		ReadContext:   resourceAzureCosmosDBAccountRead,
		CreateContext: resourceAzureCosmosDBAccountCreate,
		UpdateContext: resourceAzureCosmosDBAccountUpdate,
		DeleteContext: resourceAzureCosmosDBAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema:        duploAzureCosmosDBAccountchema(),
		CustomizeDiff: validateCosmosDBAccountParameters,
	}
}

func resourceAzureCosmosDBAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDBAccount(idParts[0], idParts[3])
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db account %s details for tenantId %s", idParts[3], idParts[0])
	}
	if err != nil && err.Status() == 404 {
		log.Printf("[DEBUG] Cosmos DB account %s not found for tenantId %s, removing from state", idParts[3], idParts[0])
		d.SetId("")
		return nil
	}
	if rp == nil {
		log.Printf("[DEBUG] Cosmos DB account %s not found for tenantId %s, removing from state", idParts[3], idParts[0])
		d.SetId("")
		return nil
	}

	d.Set("tenant_id", idParts[0])
	flattenAzureCosmosDBAccount(d, *rp)
	rps, err := c.GetCosmosDBAccountConnectionStringList(idParts[0], idParts[3])
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db account %s details for tenantId %s", idParts[3], idParts[0])
	}
	if len(rps) > 0 {
		flattenConnectionStrings(d, rps)
	}
	ak, err := c.GetCosmosDBAccountKeys(idParts[0], idParts[3])
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db account %s keys for tenantId %s : %s", idParts[3], idParts[0], err.Error())
	}
	if ak != nil {
		flattenAccountKey(d, ak)
	}
	return nil
}
func resourceAzureCosmosDBAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, name := idParts[0], idParts[3]
	rq := expandAzureCosmosDBAccount(d)
	c := m.(*duplosdk.Client)
	cerr := c.UpdateCosmosDBAccount(tenantId, name, rq)
	if cerr != nil {
		return diag.Errorf("Error updating cosmos db account %s for tenantId %s : %s", tenantId, name, cerr.Error())
	}
	d.SetId(id)
	err := cosmosDBAccountWaitUntilReady(ctx, c, idParts[0], rq.Name, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("Error waiting for cosmos db account %s to be ready for tenantId %s : %s", rq.Name, tenantId, err.Error())
	}

	diag := resourceAzureCosmosDBAccountRead(ctx, d, m)
	if diag != nil {
		return diag
	}

	return nil
}
func resourceAzureCosmosDBAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("resourceAzureCosmosDBCreate started for %s", tenantId)
	rq := expandAzureCosmosDBAccount(d)
	c := m.(*duplosdk.Client)
	cerr := c.CreateCosmosDBAccount(tenantId, rq)
	if cerr != nil {
		return diag.Errorf("Error creating cosmos db for tenantId %s : %s", tenantId, cerr.Error())
	}
	id := fmt.Sprintf("%s/cosmosdb/account/%s", tenantId, rq.Name)
	d.SetId(id)
	err := cosmosDBAccountWaitUntilReady(ctx, c, tenantId, rq.Name, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("Error waiting for cosmos db account %s to be ready for tenantId %s : %s", rq.Name, tenantId, err.Error())
	}
	diag := resourceAzureCosmosDBAccountRead(ctx, d, m)
	if diag != nil {
		return diag
	}
	log.Printf("resourceAzureCosmosDBCreate end for %s", tenantId)

	return nil
}

func resourceAzureCosmosDBAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	c := m.(*duplosdk.Client)
	err := c.DeleteCosmosDBAccount(idParts[0], idParts[3])
	if err != nil {
		if err.Status() == 404 {
			log.Printf("[DEBUG] Cosmos DB account %s not found for tenantId %s, removing from state", idParts[3], idParts[0])
			return nil
		}
		return diag.Errorf("Error deleting cosmos db account %s for tenantId %s : %s", idParts[3], idParts[0], err.Error())
	}
	derr := cosomosDbAccountWaitUntilDelete(ctx, c, idParts[0], idParts[3], d.Timeout("delete"))
	if derr != nil {
		return diag.Errorf("Error waiting for cosmos db account %s to be deleted for tenantId %s : %s", idParts[3], idParts[0], derr.Error())
	}
	return nil
}
func expandAzureCosmosDBAccount(d *schema.ResourceData) duplosdk.DuploAzureCosmosDBAccount {
	obj := duplosdk.DuploAzureCosmosDBAccount{}
	obj.Name = d.Get("name").(string)
	obj.Kind = d.Get("kind").(string)
	obj.AccountType = d.Get("type").(string)
	obj.Locations = []map[string]interface{}{}

	if v, ok := d.GetOk("geo_location"); ok {
		obj.GeoLocations = expandGeoLocations(v.([]interface{}))
	}
	if v, ok := d.GetOk("consistency_policy"); ok {
		obj.ConsistencyPolicy = expandConsistencyPolicy(v.([]interface{}))
	} else {
		obj.ConsistencyPolicy = nil
	}
	obj.Capabilities = expandCapablities(d.Get("capabilities").([]interface{}))
	obj.BackupIntervalInMinutes, obj.BackupRetentionIntervalInHours, obj.BackupPolicyType, obj.BackupStorageRedundancy = expandBackupPolicy(d.Get("backup_policy").([]interface{}))
	if obj.Capabilities != nil && len(*obj.Capabilities) > 0 && (*obj.Capabilities)[0].Name == "EnableServerless" {
		obj.CapacityMode = "Serverless"
	} else {
		obj.CapacityMode = "Provisioned"
	}
	obj.DisableKeyBasedMetadataWriteAccess = d.Get("disable_key_based_metadata_write_access").(bool)
	obj.IsFreeTierEnabled = d.Get("enable_free_tier").(bool)
	obj.PublicNetworkAccess = d.Get("public_network_access").(string)
	obj.IsVirtualNetworkFilterEnabled = d.Get("is_virtual_network_filter_enabled").(bool)
	obj.VirtualNetworkRulesRequest = expandVirtualNetworkRules(d.Get("virtual_network_rule").([]interface{}))
	return obj
}
func expandBackupPolicy(inf []interface{}) (int, int, string, string) {
	var backupInterval, backupRetentionInterval int
	var backupType, backupStorageRedundancy string
	if len(inf) > 0 {
		m := inf[0].(map[string]interface{})
		backupInterval = m["backup_interval"].(int)
		backupRetentionInterval = m["backup_retention_interval"].(int)
		backupType = m["type"].(string)
		backupStorageRedundancy = m["backup_storage_redundancy"].(string)
	}
	return backupInterval, backupRetentionInterval, backupType, backupStorageRedundancy
}

func flattenAzureCosmosDBAccount(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDBAccountResponse) {
	d.Set("name", rp.Name)
	d.Set("kind", rp.Kind)
	d.Set("disable_key_based_metadata_write_access", rp.DisableKeyBasedMetadataWriteAccess)
	d.Set("enable_free_tier", rp.IsFreeTierEnabled)
	d.Set("public_network_access", rp.PublicNetworkAccess)
	d.Set("endpoint", rp.DocumentEndpoint)
	if rp.Capabilities != nil && len(*rp.Capabilities) > 0 {
		d.Set("capabilities", flattenCapablities(*rp.Capabilities))
	}
	if rp.ConsistencyPolicy != nil {
		d.Set("consistency_policy", flattenConsistencyPolicy(*rp.ConsistencyPolicy))

	}
	d.Set("capacity_mode", rp.CapacityMode)
	d.Set("backup_policy", flattenBackupPolicy(rp))
	if rp.Locations != nil {
		lb, _ := json.Marshal(rp.Locations)
		d.Set("locations", string(lb))
	}
	if rp.ResourceType != nil {
		d.Set("type", rp.ResourceType.Namespace+"/"+rp.ResourceType.Type)
	}
	if len(rp.WriteLocations) > 0 {
		d.Set("write_endpoints", flattenLocationEndpoints(rp.WriteLocations))
	}
	if len(rp.ReadLocations) > 0 {
		d.Set("read_endpoints", flattenLocationEndpoints(rp.ReadLocations))
	}
	if len(rp.GeoLocationsResponse) > 0 {
		d.Set("geo_location", flattenGeoLocations(rp.GeoLocationsResponse))
	}
	if len(rp.VirtualNetworkRules) > 0 {
		d.Set("virtual_network_rule", flattenVirtualNetworkRules(rp.VirtualNetworkRules))
	}
}

func flattenBackupPolicy(bp duplosdk.DuploAzureCosmosDBAccountResponse) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["backup_interval"] = bp.BackupIntervalInMinutes
	m["backup_retention_interval"] = bp.BackupRetentionIntervalInHours
	m["backup_storage_redundancy"] = bp.BackupStorageRedundancy
	m["type"] = bp.BackupPolicyType
	if bp.ContinuousModeTier != "" {
		m["continuous_mode_tier"] = bp.ContinuousModeTier
	}
	obj = append(obj, m)
	return obj

}
func expandCapablities(inf []interface{}) *[]duplosdk.DuploAzureCosmosDBCapability {
	obj := []duplosdk.DuploAzureCosmosDBCapability{}
	for _, i := range inf {
		mp := i.(map[string]interface{})
		o := duplosdk.DuploAzureCosmosDBCapability{
			Name: mp["name"].(string),
		}
		obj = append(obj, o)
	}
	return &obj
}

func flattenCapablities(cap []duplosdk.DuploAzureCosmosDBCapability) []interface{} {
	mp := map[string]interface{}{}
	for _, i := range cap {
		mp["name"] = i.Name
	}
	return []interface{}{mp}
}

func expandConsistencyPolicy(inf []interface{}) *duplosdk.DuploAzureCosmosDBConsistencyPolicy {
	obj := duplosdk.DuploAzureCosmosDBConsistencyPolicy{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		if v, ok := m["default_consistency_level"]; ok {
			obj.DefaultConsistencyLevel = v.(string)
		}
		obj.MaxIntervalInSeconds = m["max_interval_in_seconds"].(int)
		obj.MaxStalenessPrefix = m["max_staleness_prefix"].(int)

	}
	return &obj
}

func flattenConsistencyPolicy(cons duplosdk.DuploAzureCosmosDBConsistencyPolicy) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["default_consistency_level"] = duplosdk.ConsistencyLevelMap[cons.DefaultConsistencyLevel.(float64)]
	m["max_interval_in_seconds"] = cons.MaxIntervalInSeconds
	m["max_staleness_prefix"] = cons.MaxStalenessPrefix
	obj = append(obj, m)

	return obj
}

/*
//func expandAzureCosmosDB(d *schema.ResourceData) duplosdk.DuploAzureCosmosDBRequest {
//	obj := duplosdk.DuploAzureCosmosDBRequest{}
//	obj.Name = d.Get("name").(string)
//	obj.Kind = d.Get("kind").(string)
//	obj.Identity = expandIdentity(d.Get("identity").([]interface{}))
//	prop := duplosdk.DuploAzureCosmosDBProperties{}
//	prop.Locations = expandStringSlice(d.Get("locations").([]interface{}))
//	prop.IpRules = expandStringSlice(d.Get("ip_rules").([]interface{}))
//	prop.IsVirtualNetworkFilterEnabled = d.Get("is_virtual_network_filter_enabled").(bool)
//	prop.EnableAutomaticFailover = d.Get("enable_automatic_failover").(bool)
//	prop.EnableMultipleWriteLocations = d.Get("enable_multiple_write_locations").(bool)
//	prop.EnableCassandraConnector = d.Get("enable_cassandra_connector").(bool)
//	prop.DisableKeyBasedMetadataWriteAccess = d.Get("disable_key_based_metadata_write_access").(bool)
//	prop.DisableLocalAuth = d.Get("disable_local_auth").(bool)
//	prop.EnableFreeTier = d.Get("enable_free_tier").(bool)
//	prop.EnableAnalyticalStorage = d.Get("enable_analytical_storage").(bool)
//	prop.ConnectorOffer = d.Get("connector_offer").(string)
//	prop.KeyVaultKeyUri = d.Get("key_vault_key_uri").(string)
//	prop.DefaultIdentity = d.Get("default_identity").(string)
//	prop.PublicNetworkAccess = d.Get("public_network_access").(string)
//	prop.CreateMode = d.Get("create_mode").(string)
//	prop.NetworkAclBypass = d.Get("network_acl_bypass").(string)
//	prop.DatabaseAccountOfferType = d.Get("database_account_offer_type").(string)
//	prop.Capabilities = expandCapablities(d.Get("capablities").([]interface{}))
//	prop.ConsistencyPolicy = expandConsistencyPolicy(d.Get("consistency_policy").([]interface{}))
//	prop.VirtualNetworkRules = expandVirtualNetworkRules(d.Get("virtual_network_rules").([]interface{}))
//	prop.ApiProperties = expandApiProperties(d.Get("api_properties").([]interface{}))
//	prop.AnalyticalStorageConfiguration = expandAnalyticalStorageConfiguration(d.Get("analytical_storage_configuration").([]interface{}))
//	prop.BackupPolicy = expandBackupPolicy(d.Get("backup_policy").([]interface{}))
//	prop.Cors = expandCors(d.Get("cors").([]interface{}))
//	prop.RestoreParameters = expandRestoreParams(d.Get("restore_parameters").([]interface{}))
//	prop.Capacity = expandCapacity(d.Get("capacity").([]interface{}))
//	prop.NetworkAclBypassResourceIds = expandStringSlice(d.Get("network_acl_bypass_resource_ids").([]interface{}))
//	obj.Properties = &prop
//	obj.Location = d.Get("location").(string)
//	return obj
//}

//	func flattenAzureCosmosDB(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDBResponse) {
//		d.Set("name", rp.Name)
//		d.Set("kind", rp.Kind)
//		d.Set("location", rp.Location)
//
//		if rp.Identity != nil {
//			d.Set("identity", flattenIdentity(*rp.Identity))
//		}
//		if len(rp.Locations) > 0 {
//			d.Set("location", flattenStringList(rp.Locations))
//		}
//		if len(rp.IpRules) > 0 {
//			d.Set("ip_rules", flattenStringList(rp.IpRules))
//		}
//		d.Set("is_virtual_network_filter_enabled", rp.IsVirtualNetworkFilterEnabled)
//		d.Set("enable_automatic_failover", rp.EnableAutomaticFailover)
//		d.Set("enable_multiple_write_locations", rp.EnableMultipleWriteLocations)
//		d.Set("enable_cassandra_connector", rp.EnableCassandraConnector)
//		d.Set("disable_key_based_metadata_write_access", rp.DisableKeyBasedMetadataWriteAccess)
//		d.Set("disable_local_auth", rp.DisableLocalAuth)
//		d.Set("enable_free_tier", rp.EnableFreeTier)
//		d.Set("connector_offer", rp.ConnectorOffer)
//		d.Set("key_vault_key_uri", rp.KeyVaultKeyUri)
//		d.Set("default_identity", rp.DefaultIdentity)
//		d.Set("public_network_access", rp.PublicNetworkAccess)
//		d.Set("create_mode", rp.CreateMode)
//		d.Set("network_acl_bypass", rp.NetworkAclBypass)
//		d.Set("database_account_offer_type", rp.DatabaseAccountOfferType)
//		if len(*rp.Capabilities) > 0 {
//			d.Set("capabilities", flattenCapablities(*rp.Capabilities))
//		}
//		if rp.ConsistencyPolicy != nil {
//			d.Set("consistency_policy", flattenConsistencyPolicy(*rp.ConsistencyPolicy))
//
//		}
//		if len(*rp.VirtualNetworkRules) > 0 {
//			d.Set("virtual_network_rules", flattenConsistencyPolicy(*rp.ConsistencyPolicy))
//		}
//		if rp.ApiProperties != nil {
//			d.Set("api_properties", flattenApiProperties(*rp.ApiProperties))
//		}
//		if rp.AnalyticalStorageConfiguration != nil {
//			d.Set("analytical_storage_configuration", flattenAnalyticalStorageConfiguration(*rp.AnalyticalStorageConfiguration))
//		}
//		if rp.BackupPolicy != nil {
//			d.Set("backup_policy", flattenBackupPolicy(rp.BackupPolicy))
//		}
//		if rp.Cors != nil {
//			d.Set("cors", flattenCors(*rp.Cors))
//		}
//		if rp.RestoreParameters != nil {
//			d.Set("restore_parameters", flattenRestoreParams(*rp.RestoreParameters))
//		}
//		if rp.Capacity != nil {
//			d.Set("capacity", flattenCapacity(*rp.Capacity))
//		}
//	}
func expandIdentity(identity []interface{}) *duplosdk.DuploAzureCosmosDBManagedServiceIdentity {
	o := duplosdk.DuploAzureCosmosDBManagedServiceIdentity{}
	for _, i := range identity {
		m := i.(map[string]interface{})
		o.PrincipalId = m["principal_id"].(string)
		o.ResourceIdentityType = m["resource_identity_type"].(string)
		rsrcId := m["arm_resource_id"]
		if rsrcId != nil {
			um := make(map[string]duplosdk.DuploAzureCosmosDBManagedServiceIdentityUserAssignedIdentities)
			um[rsrcId.(string)] = duplosdk.DuploAzureCosmosDBManagedServiceIdentityUserAssignedIdentities{}
			o.UserAssignedIdentities = um
		}
	}
	return &o
}

func flattenIdentity(identity duplosdk.DuploAzureCosmosDBManagedServiceIdentity) map[string]interface{} {
	m := make(map[string]interface{})
	m["principal_id"] = identity.PrincipalId
	m["resource_identity_type"] = identity.ResourceIdentityType
	for key, val := range identity.UserAssignedIdentities {
		m1 := make(map[string]interface{})
		m1["principal_id"] = val.PrincipalId
		m1["client_id"] = val.ClientId
		m1["arm_resource_id"] = key
		m["user_assigned_identities"] = m1
	}
	return m
}


func expandVirtualNetworkRules(inf []interface{}) *[]duplosdk.DuploAzureCosmosDBVirtualNetworkRule {
	obj := []duplosdk.DuploAzureCosmosDBVirtualNetworkRule{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		o := duplosdk.DuploAzureCosmosDBVirtualNetworkRule{
			Id:                               m["id"].(string),
			IgnoreMissingVNetServiceEndpoint: m["ignore_missing_vnet_service_endpoint"].(bool),
		}
		obj = append(obj, o)
	}
	return &obj
}

func flattenVirtualNetworkRules(nr []duplosdk.DuploAzureCosmosDBVirtualNetworkRule) []interface{} {
	obj := []interface{}{}
	for _, i := range nr {
		m := make(map[string]interface{})
		m["id"] = i.Id
		m["ignore_missing_vnet_service_endpoint"] = i.IgnoreMissingVNetServiceEndpoint
		obj = append(obj, m)

	}
	return obj
}

func expandApiProperties(inf []interface{}) *duplosdk.DuploAzureCosmosDBApiProperties {
	obj := duplosdk.DuploAzureCosmosDBApiProperties{}
	for _, i := range inf {
		obj.ServerVersion = i.(string)
	}
	return &obj
}

func flattenApiProperties(p duplosdk.DuploAzureCosmosDBApiProperties) []interface{} {
	obj := []interface{}{}
	m := map[string]interface{}{
		"server_version": p.ServerVersion,
	}
	obj = append(obj, m)
	return obj
}

func expandAnalyticalStorageConfiguration(inf []interface{}) *duplosdk.DuploAzureCosmosDBAnalyticalStorageConfiguration {
	obj := duplosdk.DuploAzureCosmosDBAnalyticalStorageConfiguration{}
	for _, i := range inf {
		obj.SchemaType = i.(string)
	}
	return &obj
}

func flattenAnalyticalStorageConfiguration(asc duplosdk.DuploAzureCosmosDBAnalyticalStorageConfiguration) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["schema_type"] = asc.SchemaType
	obj = append(obj, m)
	return obj
}

//func expandBackupPolicy(inf []interface{}) *duplosdk.DuploAzureCosmosDBBackupPolicy {
//	obj := duplosdk.DuploAzureCosmosDBBackupPolicy{}
//	if len(inf) > 0 {
//		m := inf[0].(map[string]interface{})
//		minf := m["migration_state"].([]interface{})
//		for _, i := range minf {
//			mi := i.(map[string]interface{})
//			o := duplosdk.DuploAzureCosmosDBBackupPolicyMigrationState{
//				TargetType: mi["target_type"].(string),
//				StartTime:  mi["start_time"].(string),
//				Status:     mi["status"].(string),
//			}
//			obj.BackupPolicyMigrationState = o
//
//		}
//	}
//	return &obj
//}

//func flattenBackupPolicy(bp *duplosdk.DuploAzureCosmosDBBackupPolicy) []interface{} {
//	obj := []interface{}{}
//	m := make(map[string]interface{})
//	m1 := make(map[string]interface{})
//	m1["target_type"] = bp.BackupPolicyMigrationState.TargetType
//	m1["start_time"] = bp.BackupPolicyMigrationState.StartTime
//	m1["status"] = bp.BackupPolicyMigrationState.Status
//
//	m["migration_state"] = m1
//	obj = append(obj, m)
//	return obj
//}

func expandCors(inf []interface{}) *duplosdk.DuploAzureCosmosDBCorsPolicy {
	obj := duplosdk.DuploAzureCosmosDBCorsPolicy{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		obj.AllowedHeaders = m["allowed_headers"].(string)
		obj.AllowedMethods = m["allowed_methods"].(string)
		obj.AllowedOrigins = m["allowed_origins"].(string)
		obj.ExposedHeaders = m["exposed_headers"].(string)
		obj.MaxAgeInSeconds = m["max_age_in_seconds"].(float64)
	}
	return &obj
}

func flattenCors(c duplosdk.DuploAzureCosmosDBCorsPolicy) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["allowed_headers"] = c.AllowedHeaders
	m["allowed_methods"] = c.AllowedMethods
	m["allowed_origins"] = c.AllowedOrigins
	m["exposed_headers"] = c.ExposedHeaders
	m["max_age_in_seconds"] = c.MaxAgeInSeconds
	obj = append(obj, m)
	return obj
}
func expandRestoreParams(inf []interface{}) *duplosdk.DuploAzureCosmosDBRestoreParameters {
	obj := duplosdk.DuploAzureCosmosDBRestoreParameters{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		obj.RestoreMode = m["restore_mode"].(string)
		obj.TablesToRestore = expandStringSlice(m["tables_to_restore"].([]interface{}))
		obj.DatabasesToRestore = duplosdk.DuploAzureDatabaseRestoreResource{
			DatabaseName:    m["database_to_restore.0.database_name"].(string),
			CollectionNames: expandStringSlice(m["database_to_restore.0.collection_names"].([]interface{})),
		}
	}
	return &obj
}

func flattenRestoreParams(rp duplosdk.DuploAzureCosmosDBRestoreParameters) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["restore_mode"] = rp.RestoreMode
	m["tables_to_restore"] = toInterfaceSlice(rp.TablesToRestore)
	m["database_to_restore.0.database_name"] = rp.DatabasesToRestore.DatabaseName
	m["database_to_restore.0.collection_names"] = toInterfaceSlice(rp.DatabasesToRestore.CollectionNames)

	return obj
}

func expandCapacity(inf []interface{}) *duplosdk.DuploAzureCosmosDBCapacity {
	obj := duplosdk.DuploAzureCosmosDBCapacity{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		obj.TotalThroughputLimit = m["total_throughput_limit"].(int)
	}
	return &obj
}

func flattenCapacity(c duplosdk.DuploAzureCosmosDBCapacity) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["total_throughput_limit"] = c.TotalThroughputLimit

	return obj
}

*/

func cosmosDBAccountWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.GetCosmosDBAccount(tenantID, name)
			//			log.Printf("[TRACE] Dynamodb status is (%s).", rp.TableStatus.Value)
			status := "pending"
			if err == nil {
				if rp.ProvisioningState == "Succeeded" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] cosmosDBAccountWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func validateCosmosDBAccountParameters(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Prevent backup_interval and backup_retention_interval if backup_policy.type is "Continuous"
	if d.HasChange("backup_policy") {
		// Get old and new values of backup_policy
		oldRaw, newRaw := d.GetChange("backup_policy")

		oldList, okOld := oldRaw.([]interface{})
		newList, okNew := newRaw.([]interface{})

		if okOld && okNew && len(oldList) > 0 && len(newList) > 0 {
			oldMap, ok1 := oldList[0].(map[string]interface{})
			newMap, ok2 := newList[0].(map[string]interface{})

			if ok1 && ok2 {
				oldType := ""
				newType := ""

				if v, ok := oldMap["type"]; ok && v != nil {
					oldType = strings.ToLower(v.(string))
				}
				if v, ok := newMap["type"]; ok && v != nil {
					newType = strings.ToLower(v.(string))
				}

				// Trigger force recreation only from continuous -> periodic
				if oldType == "continuous" && newType == "periodic" {
					return fmt.Errorf("updating resource from Continuous to Perioduic not allowed. Resource need to be recreated")
				}
			}
		}

		// Additional validation logic
		backupPolicies := d.Get("backup_policy").([]interface{})
		if len(backupPolicies) > 0 {
			bp := backupPolicies[0].(map[string]interface{})
			bpType := ""
			if v, ok := bp["type"]; ok && v != nil {
				bpType = strings.ToLower(v.(string))
			}
			if bpType == "continuous" {
				if bp["backup_interval"] != nil && bp["backup_interval"].(int) > 0 {
					return fmt.Errorf("backup_interval cannot be set when backup_policy.type is 'Continuous'")
				}
				if bp["backup_retention_interval"] != nil && bp["backup_retention_interval"].(int) > 0 {
					return fmt.Errorf("backup_retention_interval cannot be set when backup_policy.type is 'Continuous'")
				}
				if bp["backup_storage_redundancy"] != nil && bp["backup_storage_redundancy"].(string) != "" {
					return fmt.Errorf("backup_storage_redundancy cannot be set manually when backup_policy.type is 'Continuous'")
				}
			}
		}
	}

	update := false
	if d.HasChange("consistency_policy") {
		consistencyPolicies := d.Get("consistency_policy").([]interface{})
		if len(consistencyPolicies) > 0 {
			cp := consistencyPolicies[0].(map[string]interface{})
			if cp["default_consistency_level"] != nil && cp["default_consistency_level"].(string) != "BoundedStaleness" {
				if cp["max_staleness_prefix"] != nil && cp["max_staleness_prefix"].(int) != 100 {
					cp["max_staleness_prefix"] = 100
					update = true
				}
				if cp["max_interval_in_seconds"] != nil && cp["max_interval_in_seconds"].(int) != 5 {
					cp["max_interval_in_seconds"] = 5
					update = true

				}
			}
			if update {
				err := d.SetNew("consistency_policy", []interface{}{cp})
				return err
			}
		}
	}

	oldRaw, newRaw := d.GetChange("geo_location")

	oldList, okOld := oldRaw.([]interface{})
	newList, okNew := newRaw.([]interface{})
	if !okOld || !okNew {
		return nil
	}

	// Convert old and new lists into map[location]geo
	type geoConfig struct {
		FailoverPriority int
		Index            int
	}

	oldMap := map[string]geoConfig{}
	newMap := map[string]geoConfig{}

	for i, item := range oldList {
		if item == nil {
			continue
		}
		m := item.(map[string]interface{})
		location := strings.ToLower(m["location_name"].(string))
		oldMap[location] = geoConfig{
			FailoverPriority: m["failover_priority"].(int),
			Index:            i,
		}
	}

	for i, item := range newList {
		if item == nil {
			continue
		}
		m := item.(map[string]interface{})
		location := strings.ToLower(m["location_name"].(string))
		newMap[location] = geoConfig{
			FailoverPriority: m["failover_priority"].(int),
			Index:            i,
		}
	}

	// Loop through all locations present in both old and new maps
	for location, oldGeo := range oldMap {
		newGeo, exists := newMap[location]
		if !exists {
			// Region removed – handled elsewhere (ForceNew on geo_location list)
			continue
		}

		// Check if failover_priority changed
		if oldGeo.FailoverPriority != newGeo.FailoverPriority {
			if oldGeo.FailoverPriority == 0 || newGeo.FailoverPriority == 0 {
				// Changing write region's priority directly is forbidden
				return fmt.Errorf(
					"cannot change failover_priority for location %q with priority 0 (write region must be changed via failover)",
					location,
				)
			}

			// Allow change, but force re-creation of that geo_location block
			attrPath := fmt.Sprintf("geo_location.%d.failover_priority", newGeo.Index)
			if err := d.ForceNew(attrPath); err != nil {
				return fmt.Errorf("failed to mark geo_location.%s for ForceNew: %w", location, err)
			}
		}
	}

	return nil
}

func flattenConnectionStrings(d *schema.ResourceData, cs []duplosdk.DuploAzureCosmosDBAccountConnectionString) {
	for _, conn := range cs {
		kind := strings.ToLower(conn.KeyKind)
		ktype := strings.ToLower(conn.KeyType)
		if kind == "primary" && ktype == "sql" {
			d.Set("primary_sql_connection_string", conn.ConnectionString)
		}
		if kind == "secondary" && ktype == "sql" {
			d.Set("secondary_sql_connection_string", conn.ConnectionString)
		}
		if kind == "primaryreadonly" && ktype == "sql" {
			d.Set("primary_readonly_sql_connection_string", conn.ConnectionString)
		}
		if kind == "secondaryreadonly" && ktype == "sql" {
			d.Set("secondary_readonly_sql_connection_string", conn.ConnectionString)
		}
		if kind == "primary" && ktype == "mongo" {
			d.Set("primary_mongo_connection_string", conn.ConnectionString)
		}
		if kind == "secondary" && ktype == "mongo" {
			d.Set("secondary_mongo_connection_string", conn.ConnectionString)
		}
		if kind == "primaryreadonly" && ktype == "mongo" {
			d.Set("primary_readonly_mongo_connection_string", conn.ConnectionString)
		}
		if kind == "secondaryreadonly" && ktype == "mongo" {
			d.Set("secondary_readonly_mongo_connection_string", conn.ConnectionString)
		}
	}

}

func flattenAccountKey(d *schema.ResourceData, ak *duplosdk.DuploAzureCosmosDBAccountKeys) {
	d.Set("primary_master_key", ak.PrimaryMasterKey)
	d.Set("secondary_master_key", ak.SecondaryMasterKey)
	d.Set("primary_readonly_master_key", ak.PrimaryReadonlyMasterKey)
	d.Set("secondary_readonly_master_key", ak.SecondaryReadonlyMasterKey)
}

func flattenLocationEndpoints(locations []duplosdk.DuploAzureCosmosDBAccountLocation) []interface{} {
	obj := []interface{}{}
	for _, loc := range locations {
		if loc.DocumentEndpoint != "" {
			obj = append(obj, loc.DocumentEndpoint)
		}
	}
	return obj
}

func expandGeoLocations(inf []interface{}) []duplosdk.DuploAzureCosmosDBAccountLocationRequest {
	objs := []duplosdk.DuploAzureCosmosDBAccountLocationRequest{}
	for _, i := range inf {
		obj := duplosdk.DuploAzureCosmosDBAccountLocationRequest{}
		m := i.(map[string]interface{})
		if v, ok := m["location_name"]; ok {
			obj.LocationName = v.(string)
		}
		if v, ok := m["failover_priority"]; ok {
			obj.FailoverPriority = v.(int)
		}
		if v, ok := m["is_zone_redundant"]; ok {
			obj.IsZoneRedundant = v.(bool)
		}
		objs = append(objs, obj)
	}
	return objs
}

func flattenGeoLocations(locs []duplosdk.DuploAzureCosmosDBAccountLocation) []interface{} {
	obj := []interface{}{}
	for _, loc := range locs {
		m := make(map[string]interface{})
		m["location_name"] = loc.LocationName.Name
		m["failover_priority"] = loc.FailoverPriority
		m["is_zone_redundant"] = loc.IsZoneRedundant
		obj = append(obj, m)
	}
	return obj
}

func cosomosDbAccountWaitUntilDelete(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"deleting"},
		Target:       []string{"deleted"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			status := "deleting"
			rp, err := c.GetCosmosDBAccount(tenantID, name)
			if err != nil && err.Status() != 404 {
				return rp, "", err
			}
			if rp == nil || (err != nil && err.Status() == 404) {
				status = "deleted"
			}
			return rp, status, nil
		},
	}
	log.Printf("[DEBUG] awsElasticSearchDomainWaitUntilDeleted (%s/%s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func expandVirtualNetworkRules(inf []interface{}) []duplosdk.DuploAzureCosmosDBVirtualNetworkRuleRequest {
	obj := []duplosdk.DuploAzureCosmosDBVirtualNetworkRuleRequest{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		o := duplosdk.DuploAzureCosmosDBVirtualNetworkRuleRequest{
			Id: struct {
				ResourceId string `json:"resourceId"`
			}{
				ResourceId: m["subnet_id"].(string),
			},
			IgnoreMissingVNetServiceEndpoint: m["ignore_missing_vnet_service_endpoint"].(bool),
		}
		obj = append(obj, o)
	}
	return obj
}

func flattenVirtualNetworkRules(nr []duplosdk.DuploAzureCosmosDBVirtualNetworkRule) []interface{} {
	obj := []interface{}{}
	for _, i := range nr {
		m := make(map[string]interface{})
		m["subnet_id"] = "/subscriptions/" + i.Id.SubscriptionId + "/resourceGroups/" + i.Id.ResourceGroupName + "/providers/Microsoft.Network/virtualNetworks/" + i.Id.Parent.InfraName + "/subnets/" + i.Id.SubnetName
		m["ignore_missing_vnet_service_endpoint"] = i.IgnoreMissingVNetServiceEndpoint
		obj = append(obj, m)

	}
	return obj
}
