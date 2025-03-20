package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureCosmosDBchema() map[string]*schema.Schema {
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
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"location": {
			Description: "Specifies the primary write region for Cosmos DB.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"ip_rules": {
			Description: "List of IpRules.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"is_virtual_network_filter_enabled": {
			Description: "Flag to indicate whether to enable/disable Virtual Network ACL rules.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_automatic_failover": {
			Description: "Enables automatic failover of the write region in the rare event that the region is unavailable due to an outage. Automatic failover will result in a new write region for the account and is chosen based on the failover priorities configured for the account.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_multiple_write_locations": {
			Description: "Enables the account to write in multiple locations.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_cassandra_connector": {
			Description: "Enables the cassandra connector on the Cosmos DB C* account",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"disable_key_based_metadata_write_access": {
			Description: "Disable write operations on metadata resources (databases, containers, throughput) via account keys",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"disable_local_auth": {
			Description: "Opt-out of local authentication and ensure only MSI and AAD can be used exclusively for authentication.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_free_tier": {
			Description: "Flag to indicate whether Free Tier is enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_analytical_storage": {
			Description: "Flag to indicate whether to enable storage analytics.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"identity": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"principal_id": {
						Description: "System assigned principal id",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"tenant_id": {
						Description: "System assigned tenant id",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"resource_identity_type": {
						Description:  "Flag to indicate whether to enable storage analytics.",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"SystemAssigned", "UserAssigned", "SystemAssignedUserAssigned", "None"}, false),
					},
					"user_assigned_identities": {
						Description: "List of user identities associated with resource",
						Optional:    true,
						Type:        schema.TypeList,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"arm_resource_id": {
									Description:  "The user identity dictionary key references will be ARM resource ids in the form: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{identityName}'.",
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\\.ManagedIdentity/userAssignedIdentities/[^/]+$`), "invalid arm_resource_id"),
								},
								"principal_id": {
									Description: "User assigned principal id",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"client_id": {
									Description: "User assigned client id",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
		"consistency_policy": {
			Description: "",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_staleness_prefix": {
						Description:  "Max number of stale requests tolerated. Accepted range for this values 1 to 2147483647",
						Type:         schema.TypeFloat,
						Optional:     true,
						ValidateFunc: validation.IntBetween(1, 2147483647),
					},
					"max_interval_in_seconds": {
						Description: "Max amount of time staleness (in seconds) is tolerated",
						Type:        schema.TypeInt,
						Optional:    true,
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
		"capablities": {
			Description: "Name of the Cosmos DB capability, for example, 'EnableCassandra' Current values also include 'EnableTable' and 'EnableGremlin'",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"EnableCassandra", "EnableTable", "EnableGremlin"}, false),
			},
		},
		"virtual_network_rules": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Description:  "resource ID of a subnet, for example: /subscriptions/{subscriptionId}/resourceGroups/{groupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}",
						Optional:     true,
						Type:         schema.TypeString,
						ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\\.Network/virtualNetworks/[^/]+/subnets/[^/]+$`), "invalid virtual_network_rules.id"),
					},
					"ignore_missing_vnet_service_endpoint": {
						Description: "create firewall rule before the virtual network has vnet service endpoint enabled",
						Optional:    true,
						Default:     false,
						Type:        schema.TypeBool,
					},
				},
			},
		},
		"api_properties": {
			Type:     schema.TypeList,
			Computed: true,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"server_version": {
						Description:  "Server version of an a MongoDB account",
						Optional:     true,
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"3.2", "3.6", "4.0", "4.2"}, false),
					},
				},
			},
		},

		"analytical_storage_configuration": {
			Description: "Specify analytical storage specific properties",
			Type:        schema.TypeList,
			Computed:    true,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"schema_type": {
						Description:  "Valid values WellDefined, FullFidelity",
						Optional:     true,
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"WellDefined", "FullFidelity"}, false),
					},
				},
			},
		},

		"backup_policy": {
			Description: "Specify analytical storage specific properties",
			Type:        schema.TypeList,
			Computed:    true,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"migration_state": {
						Description: "Specify analytical storage specific properties",
						Type:        schema.TypeList,
						Computed:    true,
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"status": {
									Description:  "Valid values WellDefined, FullFidelity",
									Optional:     true,
									Type:         schema.TypeString,
									ValidateFunc: validation.StringInSlice([]string{"WellDefined", "FullFidelity"}, false),
								},
								"target_type": {
									Description:  "Valid values WellDefined, FullFidelity",
									Optional:     true,
									Type:         schema.TypeString,
									ValidateFunc: validation.StringInSlice([]string{"WellDefined", "FullFidelity"}, false),
								},
								"start_time": {
									Description:  "Valid values WellDefined, FullFidelity",
									Optional:     true,
									Type:         schema.TypeString,
									ValidateFunc: validation.StringInSlice([]string{"WellDefined", "FullFidelity"}, false),
								},
							},
						},
					},
				},
			},
		},

		"cors": {
			Description: "Specify analytical storage specific properties",
			Type:        schema.TypeList,
			Computed:    true,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allowed_origins": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"allowed_methods": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"allowed_headers": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"exposed_headers": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"max_age_in_seconds": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeFloat,
					},
				},
			},
		},

		"restore_parameters": {
			Description: "Specify analytical storage specific properties",
			Type:        schema.TypeList,
			Computed:    true,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"databases_to_restore": {
						Description: "Specify analytical storage specific properties",
						Type:        schema.TypeList,
						Computed:    true,
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"database_name": {
									Description:  "Valid values WellDefined, FullFidelity",
									Optional:     true,
									Type:         schema.TypeString,
									ValidateFunc: validation.StringInSlice([]string{"WellDefined", "FullFidelity"}, false),
								},
								"collection_names": {
									Description: "Valid values WellDefined, FullFidelity",
									Optional:    true,
									Type:        schema.TypeList,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"restore_mode": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeFloat,
					},
					"tables_to_restore": {
						Description: "Valid values WellDefined, FullFidelity",
						Optional:    true,
						Type:        schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},

		"capacity": {
			Description: "Specifiy capacity property for the account",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"total_throughput_limit": {
						Description:  "Total throughput limit imposed on the Account.A totalThroughputLimit of 2000 imposes a strict limit of max throughput that can be provisioned on that account to be 2000. A totalThroughputLimit of -1 indicates no limits on provisioning of throughput.",
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntAtMost(2000),
					},
				},
			},
		},
		"connector_offer": {
			Description: "Specify cassandra connector offer type for the Cosmos DB database C*",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"key_vault_key_uri": {
			Description: "The URI of the key vault",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"default_identity": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"public_network_access": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"create_mode": {
			Description:  "Indicate the mode of account creation. Possible values include: 'Default', 'Restore'",
			Optional:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"Default", "Restore"}, false),
		},
		"network_acl_bypass": {
			Description:  "Indicates what services are allowed to bypass firewall checks. Possible values include: 'None', 'AzureServices'",
			Optional:     true,
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"None", "AzureServices"}, false),
		},
		"database_account_offer_type": {
			Description: "",
			Required:    true,
			Type:        schema.TypeString,
		},
		"network_acl_bypass_resource_ids": {
			Description: "Resource Ids for Network Acl Bypass for the Cosmos DB account.",
			Optional:    true,
			Type:        schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func resourceAzureCosmosDB() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db` manages cosmos db resource for azure",

		ReadContext:   resourceAzureCosmosDBRead,
		CreateContext: resourceAzureCosmosDBCreate,
		UpdateContext: resourceAzureCosmosDBUpdate,
		DeleteContext: resourceAzureCosmosDBDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureCosmosDBchema(),
	}
}

func resourceAzureCosmosDBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDB(idParts[0], idParts[2])
	if err != nil {
		return diag.Errorf("Error fetching cosmos db account %s details for tenantId %s", idParts[2], idParts[0])
	}
	flattenAzureCosmosDB(d, *rp)
	return nil
}
func resourceAzureCosmosDBUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
func resourceAzureCosmosDBCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("resourceAzureCosmosDBCreate started for %s", tenantId)
	rq := expandAzureCosmosDB(d)
	c := m.(*duplosdk.Client)
	err := c.CreateCosmosDB(tenantId, rq)
	if err != nil {
		return diag.Errorf("Error creating cosmos db for tenantId %s : %s", tenantId, err.Error())
	}
	id := fmt.Sprintf("%s/cosmosdb/%s", tenantId, rq.Name)
	d.SetId(id)
	diag := resourceAzureCosmosDBRead(ctx, d, m)
	if diag != nil {
		return diag
	}
	log.Printf("resourceAzureCosmosDBCreate end for %s", tenantId)

	return nil
}

func resourceAzureCosmosDBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandAzureCosmosDB(d *schema.ResourceData) duplosdk.DuploAzureCosmosDBRequest {
	obj := duplosdk.DuploAzureCosmosDBRequest{}
	obj.Name = d.Get("name").(string)
	obj.Kind = d.Get("kind").(string)
	obj.Identity = expandIdentity(d.Get("identity").([]interface{}))
	prop := duplosdk.DuploAzureCosmosDBProperties{}
	prop.Locations = expandStringSlice(d.Get("locations").([]interface{}))
	prop.IpRules = expandStringSlice(d.Get("ip_rules").([]interface{}))
	prop.IsVirtualNetworkFilterEnabled = d.Get("is_virtual_network_filter_enabled").(bool)
	prop.EnableAutomaticFailover = d.Get("enable_automatic_failover").(bool)
	prop.EnableMultipleWriteLocations = d.Get("enable_multiple_write_locations").(bool)
	prop.EnableCassandraConnector = d.Get("enable_cassandra_connector").(bool)
	prop.DisableKeyBasedMetadataWriteAccess = d.Get("disable_key_based_metadata_write_access").(bool)
	prop.DisableLocalAuth = d.Get("disable_local_auth").(bool)
	prop.EnableFreeTier = d.Get("enable_free_tier").(bool)
	prop.EnableAnalyticalStorage = d.Get("enable_analytical_storage").(bool)
	prop.ConnectorOffer = d.Get("connector_offer").(string)
	prop.KeyVaultKeyUri = d.Get("key_vault_key_uri").(string)
	prop.DefaultIdentity = d.Get("default_identity").(string)
	prop.PublicNetworkAccess = d.Get("public_network_access").(string)
	prop.CreateMode = d.Get("create_mode").(string)
	prop.NetworkAclBypass = d.Get("network_acl_bypass").(string)
	prop.DatabaseAccountOfferType = d.Get("database_account_offer_type").(string)
	prop.Capabilities = expandCapablities(d.Get("capablities").([]interface{}))
	prop.ConsistencyPolicy = expandConsistencyPolicy(d.Get("consistency_policy").([]interface{}))
	prop.VirtualNetworkRules = expandVirtualNetworkRules(d.Get("virtual_network_rules").([]interface{}))
	prop.ApiProperties = expandApiProperties(d.Get("api_properties").([]interface{}))
	prop.AnalyticalStorageConfiguration = expandAnalyticalStorageConfiguration(d.Get("analytical_storage_configuration").([]interface{}))
	prop.BackupPolicy = expandBackupPolicy(d.Get("backup_policy").([]interface{}))
	prop.Cors = expandCors(d.Get("cors").([]interface{}))
	prop.RestoreParameters = expandRestoreParams(d.Get("restore_parameters").([]interface{}))
	prop.Capacity = expandCapacity(d.Get("capacity").([]interface{}))
	prop.NetworkAclBypassResourceIds = expandStringSlice(d.Get("network_acl_bypass_resource_ids").([]interface{}))
	obj.Properties = &prop
	obj.Location = d.Get("location").(string)
	return obj
}

func flattenAzureCosmosDB(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDBResponse) {
	d.Set("name", rp.Name)
	d.Set("kind", rp.Kind)
	d.Set("location", rp.Location)

	if rp.Identity != nil {
		d.Set("identity", flattenIdentity(*rp.Identity))
	}
	if len(rp.Locations) > 0 {
		d.Set("location", flattenStringList(rp.Locations))
	}
	if len(rp.IpRules) > 0 {
		d.Set("ip_rules", flattenStringList(rp.IpRules))
	}
	d.Set("is_virtual_network_filter_enabled", rp.IsVirtualNetworkFilterEnabled)
	d.Set("enable_automatic_failover", rp.EnableAutomaticFailover)
	d.Set("enable_multiple_write_locations", rp.EnableMultipleWriteLocations)
	d.Set("enable_cassandra_connector", rp.EnableCassandraConnector)
	d.Set("disable_key_based_metadata_write_access", rp.DisableKeyBasedMetadataWriteAccess)
	d.Set("disable_local_auth", rp.DisableLocalAuth)
	d.Set("enable_free_tier", rp.EnableFreeTier)
	d.Set("connector_offer", rp.ConnectorOffer)
	d.Set("key_vault_key_uri", rp.KeyVaultKeyUri)
	d.Set("default_identity", rp.DefaultIdentity)
	d.Set("public_network_access", rp.PublicNetworkAccess)
	d.Set("create_mode", rp.CreateMode)
	d.Set("network_acl_bypass", rp.NetworkAclBypass)
	d.Set("database_account_offer_type", rp.DatabaseAccountOfferType)
	if len(*rp.Capabilities) > 0 {
		d.Set("capabilities", flattenCapablities(*rp.Capabilities))
	}
	if rp.ConsistencyPolicy != nil {
		d.Set("consistency_policy", flattenConsistencyPolicy(*rp.ConsistencyPolicy))

	}
	if len(*rp.VirtualNetworkRules) > 0 {
		d.Set("virtual_network_rules", flattenConsistencyPolicy(*rp.ConsistencyPolicy))
	}
	if rp.ApiProperties != nil {
		d.Set("api_properties", flattenApiProperties(*rp.ApiProperties))
	}
	if rp.AnalyticalStorageConfiguration != nil {
		d.Set("analytical_storage_configuration", flattenAnalyticalStorageConfiguration(*rp.AnalyticalStorageConfiguration))
	}
	if rp.BackupPolicy != nil {
		d.Set("backup_policy", flattenBackupPolicy(rp.BackupPolicy))
	}
	if rp.Cors != nil {
		d.Set("cors", flattenCors(*rp.Cors))
	}
	if rp.RestoreParameters != nil {
		d.Set("restore_parameters", flattenRestoreParams(*rp.RestoreParameters))
	}
	if rp.Capacity != nil {
		d.Set("capacity", flattenCapacity(*rp.Capacity))
	}
}
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

func expandCapablities(inf []interface{}) *[]duplosdk.DuploAzureCosmosDBCapability {
	obj := []duplosdk.DuploAzureCosmosDBCapability{}
	for _, i := range inf {
		o := duplosdk.DuploAzureCosmosDBCapability{
			Name: i.(string),
		}
		obj = append(obj, o)
	}
	return &obj
}

func flattenCapablities(cap []duplosdk.DuploAzureCosmosDBCapability) []interface{} {
	s := []string{}
	for _, i := range cap {
		s = append(s, i.Name)
	}
	return flattenStringList(s)
}

func expandConsistencyPolicy(inf []interface{}) *duplosdk.DuploAzureCosmosDBConsistencyPolicy {
	obj := duplosdk.DuploAzureCosmosDBConsistencyPolicy{}
	for _, i := range inf {
		m := i.(map[string]interface{})
		obj.DefaultConsistencyLevel = m["default_consistency_level"].(string)
		obj.MaxIntervalInSeconds = m["max_interval_in_seconds"].(int)
		obj.MaxStalenessPrefix = m["max_staleness_prefix"].(float64)
	}
	return &obj
}

func flattenConsistencyPolicy(cons duplosdk.DuploAzureCosmosDBConsistencyPolicy) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m["default_consistency_level"] = cons.DefaultConsistencyLevel
	m["max_interval_in_seconds"] = cons.MaxIntervalInSeconds
	m["max_staleness_prefix"] = cons.MaxStalenessPrefix
	obj = append(obj, m)

	return obj
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

func expandBackupPolicy(inf []interface{}) *duplosdk.DuploAzureCosmosDBBackupPolicy {
	obj := duplosdk.DuploAzureCosmosDBBackupPolicy{}
	if len(inf) > 0 {
		m := inf[0].(map[string]interface{})
		minf := m["migration_state"].([]interface{})
		for _, i := range minf {
			mi := i.(map[string]interface{})
			o := duplosdk.DuploAzureCosmosDBBackupPolicyMigrationState{
				TargetType: mi["target_type"].(string),
				StartTime:  mi["start_time"].(string),
				Status:     mi["status"].(string),
			}
			obj.BackupPolicyMigrationState = o

		}
	}
	return &obj
}

func flattenBackupPolicy(bp *duplosdk.DuploAzureCosmosDBBackupPolicy) []interface{} {
	obj := []interface{}{}
	m := make(map[string]interface{})
	m1 := make(map[string]interface{})
	m1["target_type"] = bp.BackupPolicyMigrationState.TargetType
	m1["start_time"] = bp.BackupPolicyMigrationState.StartTime
	m1["status"] = bp.BackupPolicyMigrationState.Status

	m["migration_state"] = m1
	obj = append(obj, m)
	return obj
}

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
