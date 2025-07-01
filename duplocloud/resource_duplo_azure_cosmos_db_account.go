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
			Description:  "Indicates the type of database account. This can only be set at database account creation. \n Allowed Account Kind : GlobalDocumentDB, MongoDB, Parse",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"GlobalDocumentDB", "MongoDB", "Parse"}, false),
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
			Description: "Specify the consistency policy for the Cosmos DB account. This is only applicable for GlobalDocumentDB accounts.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_staleness_prefix": {
						Description:  "Max number of stale requests tolerated. Accepted range for this values 1 to 2147483647",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(1, 2147483647),
					},
					"max_interval_in_seconds": {
						Description: "Max amount of time staleness (in seconds) is tolerated",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
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
			Description: "Backup policy for cosmos db account",
			Type:        schema.TypeList,
			Computed:    true,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"backup_interval": {
						Description: "Backup interval in minutes",
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
						Description:  "Valid values Periodic, Continuous",
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
	}
}

func resourceAzureCosmosDBAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db` manages cosmos db resource for azure",

		ReadContext:   resourceAzureCosmosDBAccountRead,
		CreateContext: resourceAzureCosmosDBAccountCreate,
		UpdateContext: resourceAzureCosmosDBAccountUpdate,
		DeleteContext: resourceAzureCosmosDBAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
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
	if rp == nil {
		log.Printf("[DEBUG] Cosmos DB account %s not found for tenantId %s, removing from state", idParts[3], idParts[0])
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", idParts[0])
	flattenAzureCosmosDBAccount(d, *rp)
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
		return diag.Errorf("Error deleting cosmos db account %s for tenantId %s : %s", idParts[3], idParts[0], err.Error())
	}
	return nil
}
func expandAzureCosmosDBAccount(d *schema.ResourceData) duplosdk.DuploAzureCosmosDBAccount {
	obj := duplosdk.DuploAzureCosmosDBAccount{}
	obj.Name = d.Get("name").(string)
	obj.Kind = d.Get("kind").(string)
	obj.AccountType = d.Get("type").(string)
	obj.Locations = []map[string]interface{}{}
	obj.ConsistencyPolicy = expandConsistencyPolicy(d.Get("consistency_policy").([]interface{}))
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

func flattenAzureCosmosDBAccount(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDBAccount) {
	d.Set("name", rp.Name)
	d.Set("kind", rp.Kind)
	d.Set("disable_key_based_metadata_write_access", rp.DisableKeyBasedMetadataWriteAccess)
	d.Set("enable_free_tier", rp.IsFreeTierEnabled)
	d.Set("public_network_access", rp.PublicNetworkAccess)
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
}

func flattenBackupPolicy(bp duplosdk.DuploAzureCosmosDBAccount) []interface{} {
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
	return nil
}
