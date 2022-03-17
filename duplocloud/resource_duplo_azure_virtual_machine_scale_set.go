package duplocloud

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureVirtualMachineScaleSetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure virtual machine scale set will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description:  "Specifies the name of the virtual machine scale set resource.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"is_minion": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent pool that this VM is added to.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Default:     0,
		},
		"allocation_tags": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"zones": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
		"identity": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description:      "Specifies the identity type to be assigned to the scale set. Allowable values are `SystemAssigned` and `UserAssigned`.",
						Type:             schema.TypeString,
						Required:         true,
						DiffSuppressFunc: CaseDifference,
						ValidateFunc: validation.StringInSlice([]string{
							"SystemAssigned",
							"UserAssigned",
						}, false),
					},
					"identity_ids": {
						Description: "Specifies a list of user managed identity ids to be assigned to the VMSS. Required if `type` is `UserAssigned`.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"principal_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"sku": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description:  "Specifies the size of virtual machines in a scale set.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"tier": {
						Description:      "Specifies the tier of virtual machines in a scale set. Possible values, `standard` or `basic`.",
						Type:             schema.TypeString,
						DiffSuppressFunc: CaseDifference,
						Optional:         true,
						Computed:         true,
					},

					"capacity": {
						Description:  "Specifies the number of virtual machines in the scale set.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(0),
					},
				},
			},
		},
		"license_type": {
			Description: "Specifies the Windows OS license type. If supplied, the only allowed values are `Windows_Client` and `Windows_Server`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Windows_Client",
				"Windows_Server",
			}, false),
		},

		"upgrade_policy_mode": {
			Description: "Specifies the mode of an upgrade to virtual machines in the scale set. Possible values, `Rolling`, `Manual`, or `Automatic`. When choosing Rolling, you will need to set a health probe.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Automatic",
				"Manual",
				"Rolling",
			}, false),
		},

		"health_probe_id": {
			Description: "Specifies the identifier for the load balancer health probe. Required when using `Rolling` as your `upgrade_policy_mode`",
			Type:        schema.TypeString,
			Optional:    true,
		},

		"automatic_os_upgrade": {
			Description: "Automatic OS patches can be applied by Azure to your scaleset. This is particularly useful when `upgrade_policy_mode` is set to `Rolling`.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"rolling_upgrade_policy": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"max_batch_instance_percent": {
						Description:  "The maximum percent of total virtual machine instances that will be upgraded simultaneously by the rolling upgrade.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      20,
						ValidateFunc: validation.IntBetween(5, 100),
					},

					"max_unhealthy_instance_percent": {
						Description:  "The maximum percentage of the total virtual machine instances in the scale set that can be simultaneously unhealthy.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      20,
						ValidateFunc: validation.IntBetween(5, 100),
					},

					"max_unhealthy_upgraded_instance_percent": {
						Description:  "The maximum percentage of upgraded virtual machine instances that can be found to be in an unhealthy state.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      20,
						ValidateFunc: validation.IntBetween(5, 100),
					},

					"pause_time_between_batches": {
						Description: "The wait time between completing the update for all virtual machines in one batch and starting the next batch.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "PT0S",
					},
				},
			},
			DiffSuppressFunc: azureRmVirtualMachineScaleSetSuppressRollingUpgradePolicyDiff,
		},
		"overprovision": {
			Description: "Specifies whether the virtual machine scale set should be overprovisioned.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},

		"single_placement_group": {
			Description: "Specifies whether the scale set is limited to a single placement group with a maximum size of 100 virtual machines. If set to false, managed disks must be used.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			ForceNew:    true,
		},

		"priority": {
			Description: "Specifies the priority for the Virtual Machines in the Scale Set.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Low",
				"Regular",
			}, false),
			Default: "Regular",
		},

		"eviction_policy": {
			Description: "Specifies the eviction policy for Virtual Machines in this Scale Set.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Deallocate",
				"Delete",
			}, false),
		},
		"os_profile": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"computer_name_prefix": {
						Description: "Specifies the computer name prefix for all of the virtual machines in the scale set.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},

					"admin_username": {
						Description:  "Specifies the administrator account name to use for all the instances of virtual machines in the scale set.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"admin_password": {
						Description:  "Specifies the administrator password to use for all the instances of virtual machines in a scale set.",
						Type:         schema.TypeString,
						Optional:     true,
						Sensitive:    true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"custom_data": {
						Description: "Specifies custom data to supply to the machine. On linux-based systems, this can be used as a cloud-init script. On other systems, this will be copied as a file on disk.",
						Type:        schema.TypeString,
						Optional:    true,
						//StateFunc:        userDataStateFunc,
						//DiffSuppressFunc: userDataDiffSuppressFunc,
					},
				},
			},
		},
		"os_profile_secrets": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source_vault_id": {
						Description: "Specifies the key vault to use.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"vault_certificates": {
						Description: "A collection of Vault Certificates as documented below.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"certificate_url": {
									Description: "It is the Base64 encoding of a JSON Object that which is encoded in UTF-8 of which the contents need to be `data`, `dataType` and `password`.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"certificate_store": {
									Description: "Specifies the certificate store on the Virtual Machine where the certificate should be added to.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"os_profile_windows_config": {
			Type:     schema.TypeSet,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"provision_vm_agent": {
						Description: "Indicates whether virtual machine agent should be provisioned on the virtual machines in the scale set.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
					"enable_automatic_upgrades": {
						Description: "Indicates whether virtual machines in the scale set are enabled for automatic updates.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
					"winrm": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"protocol": {
									Description: "Specifies the protocol of listener.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"certificate_url": {
									Description: "Specifies URL of the certificate with which new Virtual Machines is provisioned.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"additional_unattend_config": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"pass": {
									Description: "Specifies the name of the pass that the content applies to. The only allowable value is `oobeSystem`.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"component": {
									Description: "Specifies the name of the component to configure with the added content. The only allowable value is `Microsoft-Windows-Shell-Setup`.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"setting_name": {
									Description: "Specifies the name of the setting to which the content applies. Possible values are: `FirstLogonCommands` and `AutoLogon`.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"content": {
									Description: "Specifies the base-64 encoded XML formatted content that is added to the unattend.xml file for the specified path and component.",
									Type:        schema.TypeString,
									Required:    true,
									Sensitive:   true,
								},
							},
						},
					},
				},
			},
			Set: resourceVirtualMachineScaleSetOsProfileWindowsConfigHash,
		},
		"os_profile_linux_config": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"disable_password_authentication": {
						Description: "Specifies whether password authentication should be disabled.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						ForceNew:    true,
					},
					"ssh_keys": {
						Description: "Specifies a collection of `path` and `key_data` to be placed on the virtual machine.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:     schema.TypeString,
									Required: true,
								},
								"key_data": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
			Set: resourceVirtualMachineScaleSetOsProfileLinuxConfigHash,
		},

		"network_profile": {
			Type:     schema.TypeSet,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description:  "Specifies the name of the network interface configuration.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},

					"primary": {
						Description: "Indicates whether network interfaces created from the network interface configuration will be the primary NIC of the VM.",
						Type:        schema.TypeBool,
						Required:    true,
					},

					"accelerated_networking": {
						Description: "Specifies whether to enable accelerated networking or not.",
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
					},

					"ip_forwarding": {
						Description: "Whether IP forwarding is enabled on this NIC.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},

					"network_security_group_id": {
						Description: "Specifies the identifier for the network security group.",
						Type:        schema.TypeString,
						Optional:    true,
						//ValidateFunc: azure.ValidateResourceID,
					},

					"dns_settings": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"dns_servers": {
									Description: "Specifies an array of dns servers.",
									Type:        schema.TypeList,
									Required:    true,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},

					"ip_configuration": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description:  "Specifies name of the IP configuration.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringIsNotEmpty,
								},

								"subnet_id": {
									Description: "Specifies the identifier of the subnet.",
									Type:        schema.TypeString,
									Required:    true,
									//ValidateFunc: azure.ValidateResourceID,
								},

								"application_gateway_backend_address_pool_ids": {
									Description: " Specifies an array of references to backend address pools of application gateways.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Set:         schema.HashString,
								},

								"application_security_group_ids": {
									Description: "Specifies up to 20 application security group IDs.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
										//ValidateFunc: azure.ValidateResourceID,
									},
									Set:      schema.HashString,
									MaxItems: 20,
								},

								"load_balancer_backend_address_pool_ids": {
									Description: "Specifies an array of references to backend address pools of load balancers.",
									Type:        schema.TypeSet,
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Set:         schema.HashString,
								},

								"load_balancer_inbound_nat_rules_ids": {
									Description: "Specifies an array of references to inbound NAT pools for load balancers.",
									Type:        schema.TypeSet,
									Optional:    true,
									Computed:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Set:         schema.HashString,
								},

								"primary": {
									Description: "Specifies if this ip_configuration is the primary one.",
									Type:        schema.TypeBool,
									Required:    true,
								},

								"public_ip_address_configuration": {
									Description: "Describes a virtual machines scale set IP Configuration's PublicIPAddress configuration.",
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Description: "The name of the public ip address configuration.",
												Type:        schema.TypeString,
												Required:    true,
											},

											"idle_timeout": {
												Description:  "The idle timeout in minutes. This value must be between 4 and 30.",
												Type:         schema.TypeInt,
												Required:     true,
												ValidateFunc: validation.IntBetween(4, 32),
											},

											"domain_name_label": {
												Description: "The domain name label for the dns settings.",
												Type:        schema.TypeString,
												Required:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Set: resourceVirtualMachineScaleSetNetworkConfigurationHash,
		},
		"boot_diagnostics": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Description: "Whether to enable boot diagnostics for the virtual machine.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},

					"storage_uri": {
						Description: "Blob endpoint for the storage account to hold the virtual machine's diagnostic files.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"storage_profile_os_disk": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Specifies the disk name. Must be specified when using unmanaged disk ('managed_disk_type' property not set).",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"image": {
						Description: "Specifies the blob uri for user image. A virtual machine scale set creates an os disk in the same container as the user image.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"vhd_containers": {
						Description: "Specifies the vhd uri. Cannot be used when `image` or `managed_disk_type` is specified.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Set:         schema.HashString,
					},

					"managed_disk_type": {
						Description: "Specifies the type of managed disk to create. Value you must be either `Standard_LRS`, `StandardSSD_LRS` or `Premium_LRS`. Cannot be used when `vhd_containers` or `image` is specified.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"Standard_LRS",
							"StandardSSD_LRS",
							"Premium_LRS",
						}, false),
					},

					"caching": {
						Description: "Specifies the caching requirements. Possible values include: `None` (default), `ReadOnly`, `ReadWrite`.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},

					"os_type": {
						Description: "Specifies the operating system Type, valid values are `windows`, `linux`.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"create_option": {
						Description: "Specifies how the virtual machine should be created. The only possible option is `FromImage`.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
			Set: resourceVirtualMachineScaleSetStorageProfileOsDiskHash,
		},

		"storage_profile_data_disk": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"lun": {
						Description: "Specifies the Logical Unit Number of the disk in each virtual machine in the scale set.",
						Type:        schema.TypeInt,
						Required:    true,
					},

					"create_option": {
						Description: "Specifies how the data disk should be created. The only possible options are `FromImage` and `Empty`.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"caching": {
						Description: "Specifies the caching requirements. Possible values include: `None` (default), `ReadOnly`, `ReadWrite`.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},

					"disk_size_gb": {
						Description: "Specifies the size of the disk in GB. This element is required when creating an empty disk.",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},

					"managed_disk_type": {
						Description: "Specifies the type of managed disk to create. Value must be either `Standard_LRS`, `StandardSSD_LRS` or `Premium_LRS`.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"Standard_LRS",
							"StandardSSD_LRS",
							"Premium_LRS",
						}, false),
					},
				},
			},
		},

		"storage_profile_image_reference": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Description: "Specifies the ID of the (custom) image to use to create the virtual machine scale set.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"publisher": {
						Description: "Specifies the publisher of the image used to create the virtual machines.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"offer": {
						Description: "Specifies the offer of the image used to create the virtual machines.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"sku": {
						Description: "Specifies the SKU of the image used to create the virtual machines.",
						Type:        schema.TypeString,
						Optional:    true,
					},

					"version": {
						Description: "Specifies the version of the image used to create the virtual machines.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
			Set: resourceVirtualMachineScaleSetStorageProfileImageReferenceHash,
		},
		"plan": {
			Type:     schema.TypeSet,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Specifies the name of the image from the marketplace.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"publisher": {
						Description: "Specifies the publisher of the image.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"product": {
						Description: "Specifies the product of the image from the marketplace.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"extension": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Specifies the name of the extension.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"publisher": {
						Description: "The publisher of the extension, available publishers can be found by using the Azure CLI..",
						Type:        schema.TypeString,
						Required:    true,
					},

					"type": {
						Description: "The type of extension, available types for a publisher can be found using the Azure CLI.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"type_handler_version": {
						Description: "Specifies the version of the extension to use, available versions can be found using the Azure CLI.",
						Type:        schema.TypeString,
						Required:    true,
					},

					"auto_upgrade_minor_version": {
						Description: "Specifies whether or not to use the latest minor version available.",
						Type:        schema.TypeBool,
						Optional:    true,
					},

					"provision_after_extensions": {
						Description: "Specifies a dependency array of extensions required to be executed before, the array stores the name of each extension.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						Set: schema.HashString,
					},

					"settings": {
						Description:  "The settings passed to the extension, these are specified as a JSON object in a string.",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringIsJSON,
						//DiffSuppressFunc: schema.SuppressJsonDiff,
					},

					"protected_settings": {
						Description:  "The protected_settings passed to the extension, like settings, these are specified as a JSON object in a string.",
						Type:         schema.TypeString,
						Optional:     true,
						Sensitive:    true,
						ValidateFunc: validation.StringIsJSON,
						//DiffSuppressFunc: schema.SuppressJsonDiff,
					},
				},
			},
			Set: resourceVirtualMachineScaleSetExtensionHash,
		},
		"proximity_placement_group_id": {
			Description:      "The ID of the Proximity Placement Group to which this Virtual Machine should be assigned. ",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: CaseDifference,
		},
	}
}

func resourceAzureVirtualMachineScaleSet() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_virtual_machine_scale_set` manages an azure virtual machine scale set in Duplo.",

		ReadContext:   resourceAzureVirtualMachineScaleSetRead,
		CreateContext: resourceAzureVirtualMachineScaleSetCreate,
		UpdateContext: resourceAzureVirtualMachineScaleSetUpdate,
		DeleteContext: resourceAzureVirtualMachineScaleSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureVirtualMachineScaleSetSchema(),
	}
}

func resourceAzureVirtualMachineScaleSetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineScaleSetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureVirtualMachineScaleSetGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure virtual machine scale set %s : %s", tenantID, name, clientErr)
	}

	d.Set("name", name)
	d.Set("tenant_id", tenantID)

	flattenAzureVirtualMachineScaleSet(d, duplo)

	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureVirtualMachineScaleSetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq, err := expandAzureVirtualMachineScaleSet(d)
	if err != nil {
		return diag.Errorf("Error expanding virtual machine scale set (%s, %s) : %s", tenantID, name, err)
	}
	clientErr := c.AzureVirtualMachineScaleSetCreate(tenantID, rq)
	if clientErr != nil {
		return diag.Errorf("Error creating tenant %s azure virtual machine scale set '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure virtual machine scale set", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureVirtualMachineScaleSetGet(tenantID, name)
	})

	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureVirtualMachineScaleSetRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureVirtualMachineScaleSetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureVirtualMachineScaleSetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineScaleSetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureVirtualMachineScaleSetDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure virtual machine scale set '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure virtual machine scale set", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureVirtualMachineScaleSetGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVirtualMachineScaleSetDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureVirtualMachineScaleSet(d *schema.ResourceData) (*duplosdk.DuploAzureVirtualMachineScaleSet, error) {

	zones := expandZones(d.Get("zones").([]interface{}))
	vmScaleSet := &duplosdk.DuploAzureVirtualMachineScaleSet{
		NameEx:               d.Get("name").(string),
		IsMinion:             d.Get("is_minion").(bool),
		AllocationTags:       d.Get("allocation_tags").(string),
		AgentPlatform:        d.Get("agent_platform").(int),
		Overprovision:        d.Get("overprovision").(bool),
		SinglePlacementGroup: d.Get("single_placement_group").(bool),
		Zones:                zones,
	}
	vmScaleSet.Sku = expandVirtualMachineScaleSetSku(d)

	upgradePolicy := d.Get("upgrade_policy_mode").(string)
	automaticOsUpgrade := d.Get("automatic_os_upgrade").(bool)
	priority := d.Get("priority").(string)
	evictionPolicy := d.Get("eviction_policy").(string)

	UpgradePolicy := &duplosdk.DuploAzureVirtualMachineScaleSetUpgradePolicy{
		Mode: upgradePolicy,
		AutomaticOSUpgradePolicy: &duplosdk.DuploAzureVirtualMachineScaleSetAutomaticOSUpgradePolicy{
			EnableAutomaticOSUpgrade: automaticOsUpgrade,
		},
		RollingUpgradePolicy: expandAzureRmRollingUpgradePolicy(d),
	}
	vmScaleSet.UpgradePolicy = UpgradePolicy

	storageProfile := duplosdk.DuploVirtualMachineScaleSetStorageProfile{}
	osDisk, err := expandAzureRMVirtualMachineScaleSetsStorageProfileOsDisk(d)
	if err != nil {
		return nil, err
	}
	storageProfile.OsDisk = osDisk

	if _, ok := d.GetOk("storage_profile_data_disk"); ok {
		storageProfile.DataDisks = expandAzureRMVirtualMachineScaleSetsStorageProfileDataDisk(d)
	}

	if _, ok := d.GetOk("storage_profile_image_reference"); ok {
		imageRef, err2 := expandAzureRmVirtualMachineScaleSetStorageProfileImageReference(d)
		if err2 != nil {
			return nil, err
		}
		storageProfile.ImageReference = imageRef
	}

	osProfile := expandAzureRMVirtualMachineScaleSetsOsProfile(d)
	if err != nil {
		return nil, err
	}

	extensions, err := expandAzureRMVirtualMachineScaleSetExtensions(d)
	if err != nil {
		return nil, err
	}

	virtualMachineProfile := &duplosdk.DuploAzureScaleSetVirtualMachineProfile{
		NetworkProfile:   expandAzureRmVirtualMachineScaleSetNetworkProfile(d),
		StorageProfile:   &storageProfile,
		OsProfile:        osProfile,
		ExtensionProfile: extensions,
		Priority:         priority,
		EvictionPolicy:   evictionPolicy,
	}

	if v, ok := d.GetOk("license_type"); ok {
		virtualMachineProfile.LicenseType = v.(string)
	}

	if _, ok := d.GetOk("boot_diagnostics"); ok {
		diagnosticProfile := expandAzureRMVirtualMachineScaleSetsDiagnosticProfile(d)
		virtualMachineProfile.DiagnosticsProfile = &diagnosticProfile
	}

	if v, ok := d.GetOk("health_probe_id"); ok {
		virtualMachineProfile.NetworkProfile.HealthProbe = &duplosdk.DuploApiEntityReference{
			Id: v.(string),
		}
	}
	vmScaleSet.VirtualMachineProfile = virtualMachineProfile
	if v, ok := d.GetOk("proximity_placement_group_id"); ok {
		vmScaleSet.ProximityPlacementGroup = &duplosdk.DuploSubResource{
			Id: v.(string),
		}
	}

	if _, ok := d.GetOk("identity"); ok {
		vmScaleSet.Identity = expandAzureRmVirtualMachineScaleSetIdentity(d)
	}

	if _, ok := d.GetOk("plan"); ok {
		vmScaleSet.Plan = expandAzureRmVirtualMachineScaleSetPlan(d)
	}

	return vmScaleSet, nil
}

func parseAzureVirtualMachineScaleSetIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureVirtualMachineScaleSet(d *schema.ResourceData, duplo *duplosdk.DuploAzureVirtualMachineScaleSet) {
	// d.Set("tenant_id", tenantID)
	// d.Set("Schema_attr", duplo.name)
	// d.Set("Schema_attr", duplo.type)
	// d.Set("Schema_attr", duplo.properties.upgradePolicy)
	// d.Set("Schema_attr", duplo.properties.doNotRunExtensionsOnOverprovisionedVMs)
	// d.Set("Schema_attr", duplo.id)
	// d.Set("Schema_attr", duplo.properties.overprovision)
	// d.Set("Schema_attr", duplo.properties.uniqueId)
	// d.Set("Schema_attr", duplo.properties.singlePlacementGroup)
	// d.Set("Schema_attr", duplo.location)
	// d.Set("Schema_attr", duplo.tags)
	// d.Set("Schema_attr", duplo.sku)
	// d.Set("Schema_attr", duplo.properties.virtualMachineProfile)
	// d.Set("Schema_attr", duplo.properties.provisioningState)

}

// When upgrade_policy_mode is not Rolling, we will just ignore rolling_upgrade_policy (returns true).
func azureRmVirtualMachineScaleSetSuppressRollingUpgradePolicyDiff(k, _, new string, d *schema.ResourceData) bool {
	if k == "rolling_upgrade_policy.#" && new == "0" {
		return strings.ToLower(d.Get("upgrade_policy_mode").(string)) != "rolling"
	}
	return false
}

func resourceVirtualMachineScaleSetOsProfileWindowsConfigHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		if v, ok := m["provision_vm_agent"]; ok {
			buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
		}
		if v, ok := m["enable_automatic_upgrades"]; ok {
			buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
		}
	}

	return schema.HashString(buf.String())
}

func resourceVirtualMachineScaleSetOsProfileLinuxConfigHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%t-", m["disable_password_authentication"].(bool)))

		if sshKeys, ok := m["ssh_keys"].([]interface{}); ok {
			for _, item := range sshKeys {
				k := item.(map[string]interface{})
				if path, ok := k["path"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", path.(string)))
				}
				if data, ok := k["key_data"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", data.(string)))
				}
			}
		}
	}

	return schema.HashString(buf.String())
}

func resourceVirtualMachineScaleSetNetworkConfigurationHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
		buf.WriteString(fmt.Sprintf("%t-", m["primary"].(bool)))

		if v, ok := m["accelerated_networking"]; ok {
			buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
		}
		if v, ok := m["ip_forwarding"]; ok {
			buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
		}
		if v, ok := m["network_security_group_id"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["dns_settings"].(map[string]interface{}); ok {
			if k, ok := v["dns_servers"]; ok {
				buf.WriteString(fmt.Sprintf("%s-", k))
			}
		}
		if ipConfig, ok := m["ip_configuration"].([]interface{}); ok {
			for _, it := range ipConfig {
				config := it.(map[string]interface{})
				if name, ok := config["name"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", name.(string)))
				}
				if subnetid, ok := config["subnet_id"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", subnetid.(string)))
				}
				if appPoolId, ok := config["application_gateway_backend_address_pool_ids"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", appPoolId.(*schema.Set).List()))
				}
				if appSecGroup, ok := config["application_security_group_ids"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", appSecGroup.(*schema.Set).List()))
				}
				if lbPoolIds, ok := config["load_balancer_backend_address_pool_ids"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", lbPoolIds.(*schema.Set).List()))
				}
				if lbInNatRules, ok := config["load_balancer_inbound_nat_rules_ids"]; ok {
					buf.WriteString(fmt.Sprintf("%s-", lbInNatRules.(*schema.Set).List()))
				}
				if primary, ok := config["primary"]; ok {
					buf.WriteString(fmt.Sprintf("%t-", primary.(bool)))
				}
				if publicIPConfig, ok := config["public_ip_address_configuration"].([]interface{}); ok {
					for _, publicIPIt := range publicIPConfig {
						publicip := publicIPIt.(map[string]interface{})
						if publicIPConfigName, ok := publicip["name"]; ok {
							buf.WriteString(fmt.Sprintf("%s-", publicIPConfigName.(string)))
						}
						if idle_timeout, ok := publicip["idle_timeout"]; ok {
							buf.WriteString(fmt.Sprintf("%d-", idle_timeout.(int)))
						}
						if dnsLabel, ok := publicip["domain_name_label"]; ok {
							buf.WriteString(fmt.Sprintf("%s-", dnsLabel.(string)))
						}
					}
				}
			}
		}
	}

	return schema.HashString(buf.String())
}

func resourceVirtualMachineScaleSetStorageProfileOsDiskHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))

		if v, ok := m["vhd_containers"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(*schema.Set).List()))
		}
	}

	return schema.HashString(buf.String())
}

func resourceVirtualMachineScaleSetStorageProfileImageReferenceHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		if v, ok := m["publisher"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["offer"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["sku"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["version"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["id"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
	}

	return schema.HashString(buf.String())
}

func resourceVirtualMachineScaleSetExtensionHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", m["publisher"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", m["type_handler_version"].(string)))

		if v, ok := m["auto_upgrade_minor_version"]; ok {
			buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
		}

		if v, ok := m["provision_after_extensions"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", v.(*schema.Set).List()))
		}

		// we need to ensure the whitespace is consistent
		settings := m["settings"].(string)
		if settings != "" {
			expandedSettings, err := structure.ExpandJsonFromString(settings)
			if err == nil {
				serializedSettings, err := structure.FlattenJsonToString(expandedSettings)
				if err == nil {
					buf.WriteString(fmt.Sprintf("%s-", serializedSettings))
				}
			}
		}
	}

	return schema.HashString(buf.String())
}

func expandVirtualMachineScaleSetSku(d *schema.ResourceData) *duplosdk.DuploAzureVirtualMachineScaleSetSku {
	skuConfig := d.Get("sku").([]interface{})
	config := skuConfig[0].(map[string]interface{})

	sku := &duplosdk.DuploAzureVirtualMachineScaleSetSku{
		Name:     config["name"].(string),
		Capacity: config["capacity"].(int),
	}

	if tier, ok := config["tier"].(string); ok && tier != "" {
		sku.Tier = tier
	}

	return sku
}

func expandAzureRMVirtualMachineScaleSetsStorageProfileOsDisk(d *schema.ResourceData) (*duplosdk.DuploVirtualMachineScaleSetOSDisk, error) {
	osDiskConfigs := d.Get("storage_profile_os_disk").(*schema.Set).List()

	osDiskConfig := osDiskConfigs[0].(map[string]interface{})
	name := osDiskConfig["name"].(string)
	image := osDiskConfig["image"].(string)
	vhd_containers := osDiskConfig["vhd_containers"].(*schema.Set).List()
	caching := osDiskConfig["caching"].(string)
	osType := osDiskConfig["os_type"].(string)
	createOption := osDiskConfig["create_option"].(string)
	managedDiskType := osDiskConfig["managed_disk_type"].(string)

	if managedDiskType == "" && name == "" {
		return nil, fmt.Errorf("[ERROR] `name` must be set in `storage_profile_os_disk` for unmanaged disk")
	}

	osDisk := &duplosdk.DuploVirtualMachineScaleSetOSDisk{
		Name:         name,
		Caching:      caching,
		OsType:       osType,
		CreateOption: createOption,
	}

	if image != "" {
		osDisk.Image = &duplosdk.DuploVirtualHardDisk{
			Uri: image,
		}
	}

	if len(vhd_containers) > 0 {
		var vhdContainers []string
		for _, v := range vhd_containers {
			str := v.(string)
			vhdContainers = append(vhdContainers, str)
		}
		osDisk.VhdContainers = vhdContainers
	}

	managedDisk := &duplosdk.DuploVirtualMachineScaleSetManagedDiskParameters{}

	if managedDiskType != "" {
		if name != "" {
			return nil, fmt.Errorf("[ERROR] Conflict between `name` and `managed_disk_type` on `storage_profile_os_disk` (please remove name or set it to blank)")
		}

		managedDisk.StorageAccountType = managedDiskType
		osDisk.ManagedDisk = managedDisk
	}

	// BEGIN: code to be removed after GH-13016 is merged
	if image != "" && managedDiskType != "" {
		return nil, fmt.Errorf("[ERROR] Conflict between `image` and `managed_disk_type` on `storage_profile_os_disk` (only one or the other can be used)")
	}

	if len(vhd_containers) > 0 && managedDiskType != "" {
		return nil, fmt.Errorf("[ERROR] Conflict between `vhd_containers` and `managed_disk_type` on `storage_profile_os_disk` (only one or the other can be used)")
	}
	// END: code to be removed after GH-13016 is merged

	return osDisk, nil
}

func expandAzureRMVirtualMachineScaleSetsStorageProfileDataDisk(d *schema.ResourceData) *[]duplosdk.DuploVirtualMachineScaleSetDataDisk {
	disks := d.Get("storage_profile_data_disk").([]interface{})
	dataDisks := make([]duplosdk.DuploVirtualMachineScaleSetDataDisk, 0, len(disks))
	for _, diskConfig := range disks {
		config := diskConfig.(map[string]interface{})

		createOption := config["create_option"].(string)
		managedDiskType := config["managed_disk_type"].(string)
		lun := config["lun"].(int)

		dataDisk := duplosdk.DuploVirtualMachineScaleSetDataDisk{
			Lun:          lun,
			CreateOption: createOption,
		}

		managedDiskVMSS := &duplosdk.DuploVirtualMachineScaleSetManagedDiskParameters{}

		if len(managedDiskType) > 0 {
			managedDiskVMSS.StorageAccountType = managedDiskType
		} else {
			managedDiskVMSS.StorageAccountType = "Standard_LRS"
		}

		// assume that data disks in VMSS can only be Managed Disks
		dataDisk.ManagedDisk = managedDiskVMSS
		if v := config["caching"].(string); v != "" {
			dataDisk.Caching = v
		}

		if v := config["disk_size_gb"]; v != nil {
			diskSize := config["disk_size_gb"].(int)
			dataDisk.DiskSizeGB = diskSize
		}

		dataDisks = append(dataDisks, dataDisk)
	}

	return &dataDisks
}

func expandAzureRmVirtualMachineScaleSetStorageProfileImageReference(d *schema.ResourceData) (*duplosdk.DuploStorageProfileImageReference, error) {
	storageImageRefs := d.Get("storage_profile_image_reference").(*schema.Set).List()

	storageImageRef := storageImageRefs[0].(map[string]interface{})

	imageID := storageImageRef["id"].(string)
	publisher := storageImageRef["publisher"].(string)

	imageReference := duplosdk.DuploStorageProfileImageReference{}

	if imageID != "" && publisher != "" {
		return nil, fmt.Errorf("[ERROR] Conflict between `id` and `publisher` (only one or the other can be used)")
	}

	if imageID != "" {
		imageReference.Id = storageImageRef["id"].(string)
	} else {
		offer := storageImageRef["offer"].(string)
		sku := storageImageRef["sku"].(string)
		version := storageImageRef["version"].(string)

		imageReference.Publisher = publisher
		imageReference.Offer = offer
		imageReference.Sku = sku
		imageReference.Version = version
	}

	return &imageReference, nil
}

func expandAzureRMVirtualMachineScaleSetsOsProfile(d *schema.ResourceData) *duplosdk.DuploVirtualMachineScaleSetOSProfile {
	osProfileConfigs := d.Get("os_profile").([]interface{})

	osProfileConfig := osProfileConfigs[0].(map[string]interface{})
	namePrefix := osProfileConfig["computer_name_prefix"].(string)
	username := osProfileConfig["admin_username"].(string)
	password := osProfileConfig["admin_password"].(string)
	customData := osProfileConfig["custom_data"].(string)

	osProfile := &duplosdk.DuploVirtualMachineScaleSetOSProfile{
		ComputerNamePrefix: namePrefix,
		AdminUsername:      username,
	}

	if password != "" {
		osProfile.AdminPassword = password
	}

	if customData != "" {
		customData = Base64EncodeIfNot(customData)
		osProfile.CustomData = customData
	}

	if _, ok := d.GetOk("os_profile_secrets"); ok {
		secrets := expandAzureRmVirtualMachineScaleSetOsProfileSecrets(d)
		if secrets != nil {
			osProfile.Secrets = secrets
		}
	}

	if _, ok := d.GetOk("os_profile_linux_config"); ok {
		osProfile.LinuxConfiguration = expandAzureRmVirtualMachineScaleSetOsProfileLinuxConfig(d)
	}

	if _, ok := d.GetOk("os_profile_windows_config"); ok {
		winConfig := expandAzureRmVirtualMachineScaleSetOsProfileWindowsConfig(d)
		if winConfig != nil {
			osProfile.WindowsConfiguration = winConfig
		}
	}

	return osProfile
}

func expandAzureRmVirtualMachineScaleSetOsProfileLinuxConfig(d *schema.ResourceData) *duplosdk.DuploOSProfileLinuxConfiguration {
	osProfilesLinuxConfig := d.Get("os_profile_linux_config").(*schema.Set).List()

	linuxConfig := osProfilesLinuxConfig[0].(map[string]interface{})
	disablePasswordAuth := linuxConfig["disable_password_authentication"].(bool)

	linuxKeys := linuxConfig["ssh_keys"].([]interface{})
	sshPublicKeys := make([]duplosdk.DuploSshPublicKey, 0, len(linuxKeys))
	for _, key := range linuxKeys {
		if key == nil {
			continue
		}
		sshKey := key.(map[string]interface{})
		path := sshKey["path"].(string)
		keyData := sshKey["key_data"].(string)

		sshPublicKey := duplosdk.DuploSshPublicKey{
			Path:    path,
			KeyData: keyData,
		}

		sshPublicKeys = append(sshPublicKeys, sshPublicKey)
	}

	config := &duplosdk.DuploOSProfileLinuxConfiguration{
		DisablePasswordAuthentication: disablePasswordAuth,
		Ssh: &duplosdk.DuploOSProfileSshConfiguration{
			PublicKeys: &sshPublicKeys,
		},
	}

	return config
}

func expandAzureRmVirtualMachineScaleSetOsProfileWindowsConfig(d *schema.ResourceData) *duplosdk.DuploOSProfileWindowsConfiguration {
	osProfilesWindowsConfig := d.Get("os_profile_windows_config").(*schema.Set).List()

	osProfileConfig := osProfilesWindowsConfig[0].(map[string]interface{})
	config := &duplosdk.DuploOSProfileWindowsConfiguration{}

	if v := osProfileConfig["provision_vm_agent"]; v != nil {
		provision := v.(bool)
		config.ProvisionVMAgent = provision
	}

	if v := osProfileConfig["enable_automatic_upgrades"]; v != nil {
		update := v.(bool)
		config.EnableAutomaticUpdates = update
	}

	if v := osProfileConfig["winrm"]; v != nil {
		winRm := v.([]interface{})
		if len(winRm) > 0 {
			winRmListeners := make([]duplosdk.DuploOSProfileWinRMListener, 0, len(winRm))
			for _, winRmConfig := range winRm {
				config := winRmConfig.(map[string]interface{})

				protocol := config["protocol"].(string)
				winRmListener := duplosdk.DuploOSProfileWinRMListener{
					Protocol: protocol,
				}
				if v := config["certificate_url"].(string); v != "" {
					winRmListener.CertificateUrl = v
				}

				winRmListeners = append(winRmListeners, winRmListener)
			}
			config.WinRM = &duplosdk.DuploOSProfileWinRMConfiguration{
				Listeners: &winRmListeners,
			}
		}
	}
	if v := osProfileConfig["additional_unattend_config"]; v != nil {
		additionalConfig := v.([]interface{})
		if len(additionalConfig) > 0 {
			additionalConfigContent := make([]duplosdk.DuploWinConfigAdditionalUnattendContent, 0, len(additionalConfig))
			for _, addConfig := range additionalConfig {
				config := addConfig.(map[string]interface{})
				pass := config["pass"].(string)
				component := config["component"].(string)
				settingName := config["setting_name"].(string)
				content := config["content"].(string)

				addContent := duplosdk.DuploWinConfigAdditionalUnattendContent{
					PassName:      pass,
					ComponentName: component,
					SettingName:   settingName,
				}

				if content != "" {
					addContent.Content = content
				}

				additionalConfigContent = append(additionalConfigContent, addContent)
			}
			config.AdditionalUnattendContent = &additionalConfigContent
		}
	}
	return config
}

func expandAzureRmVirtualMachineScaleSetOsProfileSecrets(d *schema.ResourceData) *[]duplosdk.DuploVaultSecretGroup {
	secretsConfig := d.Get("os_profile_secrets").(*schema.Set).List()
	secrets := make([]duplosdk.DuploVaultSecretGroup, 0, len(secretsConfig))

	for _, secretConfig := range secretsConfig {
		config := secretConfig.(map[string]interface{})
		sourceVaultId := config["source_vault_id"].(string)

		vaultSecretGroup := duplosdk.DuploVaultSecretGroup{
			SourceVault: &duplosdk.DuploSubResource{
				Id: sourceVaultId,
			},
		}

		if v := config["vault_certificates"]; v != nil {
			certsConfig := v.([]interface{})
			certs := make([]duplosdk.DuploVaultCertificate, 0, len(certsConfig))
			for _, certConfig := range certsConfig {
				config := certConfig.(map[string]interface{})

				certUrl := config["certificate_url"].(string)
				cert := duplosdk.DuploVaultCertificate{
					CertificateUrl: certUrl,
				}
				if v := config["certificate_store"].(string); v != "" {
					cert.CertificateStore = v
				}

				certs = append(certs, cert)
			}
			vaultSecretGroup.VaultCertificates = &certs
		}

		secrets = append(secrets, vaultSecretGroup)
	}

	return &secrets
}

func expandAzureRMVirtualMachineScaleSetExtensions(d *schema.ResourceData) (*duplosdk.DuploVirtualMachineScaleSetExtensionProfile, error) {
	extensions := d.Get("extension").(*schema.Set).List()
	resources := make([]duplosdk.DuploVirtualMachineScaleSetExtension, 0, len(extensions))
	for _, e := range extensions {
		config := e.(map[string]interface{})
		name := config["name"].(string)
		publisher := config["publisher"].(string)
		t := config["type"].(string)
		version := config["type_handler_version"].(string)

		extension := duplosdk.DuploVirtualMachineScaleSetExtension{
			Name:               name,
			Publisher:          publisher,
			Type:               t,
			TypeHandlerVersion: version,
		}

		if u := config["auto_upgrade_minor_version"]; u != nil {
			upgrade := u.(bool)
			extension.AutoUpgradeMinorVersion = upgrade
		}

		if a := config["provision_after_extensions"]; a != nil {
			provision_after_extensions := config["provision_after_extensions"].(*schema.Set).List()
			if len(provision_after_extensions) > 0 {
				var provisionAfterExtensions []string
				for _, a := range provision_after_extensions {
					str := a.(string)
					provisionAfterExtensions = append(provisionAfterExtensions, str)
				}
				extension.ProvisionAfterExtensions = provisionAfterExtensions
			}
		}

		if s := config["settings"].(string); s != "" {
			settings, err := structure.ExpandJsonFromString(s)
			if err != nil {
				return nil, fmt.Errorf("unable to parse settings: %+v", err)
			}
			extension.Settings = settings
		}

		if s := config["protected_settings"].(string); s != "" {
			protectedSettings, err := structure.ExpandJsonFromString(s)
			if err != nil {
				return nil, fmt.Errorf("unable to parse protected_settings: %+v", err)
			}
			extension.ProtectedSettings = protectedSettings
		}

		resources = append(resources, extension)
	}

	return &duplosdk.DuploVirtualMachineScaleSetExtensionProfile{
		Extensions: &resources,
	}, nil
}

func expandAzureRmRollingUpgradePolicy(d *schema.ResourceData) *duplosdk.DuploAzureVirtualMachineScaleSetRollingUpgradePolicy {
	if config, ok := d.GetOk("rolling_upgrade_policy.0"); ok {
		policy := config.(map[string]interface{})
		return &duplosdk.DuploAzureVirtualMachineScaleSetRollingUpgradePolicy{
			MaxBatchInstancePercent:             policy["max_batch_instance_percent"].(int),
			MaxUnhealthyInstancePercent:         policy["max_unhealthy_instance_percent"].(int),
			MaxUnhealthyUpgradedInstancePercent: policy["max_unhealthy_upgraded_instance_percent"].(int),
			PauseTimeBetweenBatches:             policy["pause_time_between_batches"].(string),
		}
	}
	return nil
}

func expandAzureRmVirtualMachineScaleSetNetworkProfile(d *schema.ResourceData) *duplosdk.DuploVirtualMachineScaleSetNetworkProfile {
	scaleSetNetworkProfileConfigs := d.Get("network_profile").(*schema.Set).List()
	networkProfileConfig := make([]duplosdk.DuploVirtualMachineScaleSetNetworkConfiguration, 0, len(scaleSetNetworkProfileConfigs))

	for _, npProfileConfig := range scaleSetNetworkProfileConfigs {
		config := npProfileConfig.(map[string]interface{})

		name := config["name"].(string)
		primary := config["primary"].(bool)
		acceleratedNetworking := config["accelerated_networking"].(bool)
		ipForwarding := config["ip_forwarding"].(bool)

		dnsSettingsConfigs := config["dns_settings"].([]interface{})
		dnsSettings := duplosdk.DuploVirtualMachineScaleSetNetworkConfigurationDnsSettings{}
		for _, dnsSettingsConfig := range dnsSettingsConfigs {
			dns_settings := dnsSettingsConfig.(map[string]interface{})

			if v := dns_settings["dns_servers"]; v != nil {
				dns_servers := dns_settings["dns_servers"].([]interface{})
				if len(dns_servers) > 0 {
					var dnsServers []string
					for _, v := range dns_servers {
						str := v.(string)
						dnsServers = append(dnsServers, str)
					}
					dnsSettings.DnsServers = dnsServers
				}
			}
		}
		ipConfigurationConfigs := config["ip_configuration"].([]interface{})
		ipConfigurations := make([]duplosdk.DuploVirtualMachineScaleSetIPConfiguration, 0, len(ipConfigurationConfigs))
		for _, ipConfigConfig := range ipConfigurationConfigs {
			ipconfig := ipConfigConfig.(map[string]interface{})
			name := ipconfig["name"].(string)
			primary := ipconfig["primary"].(bool)
			subnetId := ipconfig["subnet_id"].(string)

			ipConfiguration := duplosdk.DuploVirtualMachineScaleSetIPConfiguration{
				Name: name,
				Subnet: &duplosdk.DuploApiEntityReference{
					Id: subnetId,
				},
			}

			ipConfiguration.Primary = primary

			if v := ipconfig["application_gateway_backend_address_pool_ids"]; v != nil {
				pools := v.(*schema.Set).List()
				resources := make([]duplosdk.DuploSubResource, 0, len(pools))
				for _, p := range pools {
					id := p.(string)
					resources = append(resources, duplosdk.DuploSubResource{
						Id: id,
					})
				}
				ipConfiguration.ApplicationGatewayBackendAddressPools = &resources
			}

			if v := ipconfig["application_security_group_ids"]; v != nil {
				asgs := v.(*schema.Set).List()
				resources := make([]duplosdk.DuploSubResource, 0, len(asgs))
				for _, p := range asgs {
					id := p.(string)
					resources = append(resources, duplosdk.DuploSubResource{
						Id: id,
					})
				}
				ipConfiguration.ApplicationSecurityGroups = &resources
			}

			if v := ipconfig["load_balancer_backend_address_pool_ids"]; v != nil {
				pools := v.(*schema.Set).List()
				resources := make([]duplosdk.DuploSubResource, 0, len(pools))
				for _, p := range pools {
					id := p.(string)
					resources = append(resources, duplosdk.DuploSubResource{
						Id: id,
					})
				}
				ipConfiguration.LoadBalancerBackendAddressPools = &resources
			}

			if v := ipconfig["load_balancer_inbound_nat_rules_ids"]; v != nil {
				rules := v.(*schema.Set).List()
				rulesResources := make([]duplosdk.DuploSubResource, 0, len(rules))
				for _, m := range rules {
					id := m.(string)
					rulesResources = append(rulesResources, duplosdk.DuploSubResource{
						Id: id,
					})
				}
				ipConfiguration.LoadBalancerInboundNatPools = &rulesResources
			}

			if v := ipconfig["public_ip_address_configuration"]; v != nil {
				publicIpConfigs := v.([]interface{})
				for _, publicIpConfigConfig := range publicIpConfigs {
					publicIpConfig := publicIpConfigConfig.(map[string]interface{})

					domainNameLabel := publicIpConfig["domain_name_label"].(string)
					dnsSettings := duplosdk.DuploVirtualMachineScaleSetPublicIPAddressConfigurationDnsSettings{
						DomainNameLabel: domainNameLabel,
					}
					publicIPConfigName := publicIpConfig["name"].(string)
					idleTimeout := publicIpConfig["idle_timeout"].(int)
					config := duplosdk.DuploVirtualMachineScaleSetPublicIPAddressConfiguration{
						Name:                 publicIPConfigName,
						DnsSettings:          &dnsSettings,
						IdleTimeoutInMinutes: idleTimeout,
					}

					ipConfiguration.PublicIPAddressConfiguration = &config
				}
			}

			ipConfigurations = append(ipConfigurations, ipConfiguration)
		}

		nProfile := duplosdk.DuploVirtualMachineScaleSetNetworkConfiguration{
			Name:                        name,
			Primary:                     primary,
			IpConfigurations:            &ipConfigurations,
			EnableAcceleratedNetworking: acceleratedNetworking,
			EnableIPForwarding:          ipForwarding,
			DnsSettings:                 &dnsSettings,
		}

		if v := config["network_security_group_id"].(string); v != "" {
			networkSecurityGroupId := duplosdk.DuploSubResource{
				Id: v,
			}
			nProfile.NetworkSecurityGroup = &networkSecurityGroupId
		}

		networkProfileConfig = append(networkProfileConfig, nProfile)
	}

	return &duplosdk.DuploVirtualMachineScaleSetNetworkProfile{
		NetworkInterfaceConfigurations: &networkProfileConfig,
	}
}

func expandAzureRMVirtualMachineScaleSetsDiagnosticProfile(d *schema.ResourceData) duplosdk.DuploDiagnosticsProfile {
	bootDiagnosticConfigs := d.Get("boot_diagnostics").([]interface{})
	bootDiagnosticConfig := bootDiagnosticConfigs[0].(map[string]interface{})

	enabled := bootDiagnosticConfig["enabled"].(bool)
	storageURI := bootDiagnosticConfig["storage_uri"].(string)

	bootDiagnostic := &duplosdk.DuploBootDiagnostics{
		Enabled:    enabled,
		StorageUri: storageURI,
	}

	diagnosticsProfile := duplosdk.DuploDiagnosticsProfile{
		BootDiagnostics: bootDiagnostic,
	}

	return diagnosticsProfile
}

func expandZones(v []interface{}) []string {
	zones := make([]string, 0)
	for _, zone := range v {
		zones = append(zones, zone.(string))
	}
	if len(zones) > 0 {
		return zones
	} else {
		return nil
	}
}

func expandAzureRmVirtualMachineScaleSetIdentity(d *schema.ResourceData) *duplosdk.DuploAzureVirtualMachineScaleSetIdentity {
	v := d.Get("identity")
	identities := v.([]interface{})
	identity := identities[0].(map[string]interface{})
	identityType := identity["type"].(string)

	identityIds := make(map[string]*duplosdk.DuploAzureVirtualMachineScaleSetIdentityValue)
	for _, id := range identity["identity_ids"].([]interface{}) {
		identityIds[id.(string)] = &duplosdk.DuploAzureVirtualMachineScaleSetIdentityValue{}
	}

	vmssIdentity := duplosdk.DuploAzureVirtualMachineScaleSetIdentity{
		Type: identityType,
	}

	if vmssIdentity.Type == "UserAssigned" {
		vmssIdentity.UserAssignedIdentities = identityIds
	}

	return &vmssIdentity
}

func expandAzureRmVirtualMachineScaleSetPlan(d *schema.ResourceData) *duplosdk.DuploAzureVirtualMachineScaleSetPlan {
	planConfigs := d.Get("plan").(*schema.Set).List()

	planConfig := planConfigs[0].(map[string]interface{})

	publisher := planConfig["publisher"].(string)
	name := planConfig["name"].(string)
	product := planConfig["product"].(string)

	return &duplosdk.DuploAzureVirtualMachineScaleSetPlan{
		Publisher: publisher,
		Name:      name,
		Product:   product,
	}
}
