package duplocloud

import (
	"regexp"
	"time"

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
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
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
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"principal_id": {
						Description: "",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"tenant_id": {
						Description: "",
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
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"arm_resource_id": {
									Description:  "The user identity dictionary key references will be ARM resource ids in the form: '/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{identityName}'.",
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\\.ManagedIdentity/userAssignedIdentities/[^/]+$`), "invalid arm_resource_id"),
								},
								"principal_id": {
									Description: "",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"client_id": {
									Description: "",
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
					"exposed_heafers": {
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
								"collection_name": {
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
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"total_throughput_limit": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		"connector_offer": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"key_vault_key_uri": {
			Description: "",
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
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
		"network_acl_bypass": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"database_account_offer_type": {
			Description: "",
			Optional:    true,
			Type:        schema.TypeString,
		},
	}
}

func resourceAzureCosmosDB() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_availability_set` manages logical groupings of VMs that enhance reliability by placing VMs in different fault domains to minimize correlated failures, offering improved VM-to-VM latency and high availability, with no extra cost beyond the VM instances themselves, though they may still be affected by shared infrastructure failures.",

		ReadContext:   resourceAzureAvailabilitySetRead,
		CreateContext: resourceAzureAvailabilitySetCreate,
		//	UpdateContext: resourceAzureAvailabilitySetUpdate,
		DeleteContext: resourceAzureAvailabilitySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAzureAvailablitySetSchema(),
		CustomizeDiff: validateAvailabilitySetAttribute,
	}
}
