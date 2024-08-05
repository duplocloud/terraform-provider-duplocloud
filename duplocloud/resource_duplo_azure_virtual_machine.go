package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureVirtualMachineSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the host will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"friendly_name": {
			Description:      "The short name of the host.",
			Type:             schema.TypeString,
			Optional:         false,
			Required:         true,
			ForceNew:         true, // relaunch instance
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"fullname": {
			Description: "The full name of the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"capacity": {
			Description: "Specifies the [size of the Virtual Machine](https://docs.microsoft.com/azure/virtual-machines/sizes-general). See also [Azure VM Naming Conventions](https://docs.microsoft.com/azure/virtual-machines/vm-naming-conventions).",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
		},
		"instance_id": {
			Description: "The Azure Virtual Machine ID of the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"is_minion": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"join_domain": {
			Description: "Join a Windows Server virtual machine to an Azure Active Directory Domain Services.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"ad_domain_type": {
			Description: "Specify domain service provided by Microsoft Azure for managing identities and access in the cloud. Valid values are `aadjoin` or `addsjoin`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"aadjoin",
				"addsjoin",
			}, false),
		},
		"timezone": {
			Description: "Specifies the time zone of the virtual machine, [the possible values are defined here](https://jackstromberg.com/2017/01/list-of-time-zones-consumed-by-azure/).",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"enable_log_analytics": {
			Description: "Enable log analytics on virtual machine.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"image_id": {
			Description: "The Image ID to use to create virtual machine. Provide id as semicolon separated string with sequence of sku, publisher and offer. For example, 16.04-LTS;Canonical;UbuntuServe",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent pool that this host is added to.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
		},
		"subnet_id": {
			Description: "Subnet ID which should be associated with the Virtual Machine.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"admin_username": {
			Description: "Specifies the name of the local administrator account.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"admin_password": {
			Description: "The password associated with the local administrator account.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"base64_user_data": {
			Description: "Base64 encoded user data to associated with the host.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},
		"allocated_public_ip": {
			Description: "Whether or not to allocate a public IP.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"encrypt_disk": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true,
		},
		"disk_size_gb": {
			Description: "Specifies the size of the OS Disk in gigabytes",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     128,
		},
		"os_disk_type": {
			Description: "Specifies the type of managed disk to create. Possible values are either `Standard_LRS`, `StandardSSD_LRS`, `Premium_LRS`, `PremiumV2_LRS`, `Premium_ZRS`, `StandardSSD_ZRS` or `UltraSSD_LRS`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Standard_LRS",
				"StandardSSD_LRS",
				"Premium_LRS",
				"PremiumV2_LRS",
				"StandardSSD_ZRS",
				"UltraSSD_LRS",
			}, true),
		},
		"status": {
			Description: "The current status of the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"user_account": {
			Description: "The name of the tenant that the host will be created in.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Type:             schema.TypeList,
			Optional:         true,
			Computed:         true,
			Elem:             KeyValueSchema(),
			DiffSuppressFunc: suppressAzureManagedTags,
		},
		"minion_tags": {
			Description: "A map of tags to assign to the resource. Example - `AllocationTags` can be passed as tag key with any value.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"volume": {
			Type:     schema.TypeList,
			Optional: true,
			ForceNew: true, // relaunch instance
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"iops": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"name": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"size": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"volume_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"volume_type": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"disk_control_type": {
			Description: "disk control types refer to the different levels of management and performance control provided for disks attached to virtual machines (VMs)",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"NVMe",
				"SCSI",
			}, true),
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until azure virtual machine to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"enable_encrypt_at_host": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"security_type": {
			Description: `Specify "Standard" or "Trusted Launch" security type. Defaults to "Standard".
			Use Trusted Launch for the security of "Generation 2" virtual machines (VMs). [Supported Sizes](https://learn.microsoft.com/en-us/azure/virtual-machines/trusted-launch#virtual-machines-sizes)
			`,
			Type:     schema.TypeString,
			Optional: true,
			Default:  "Standard",
			ValidateFunc: validation.StringInSlice([]string{
				"Standard",
				"TrustedLaunch",
			}, false),
		},
		"enable_security_boot": {
			Description: "Specify to enable Secure Boot for your VM. Used with security_type=TrustedLaunch",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"enable_vtpm": {
			Description: "Specify to enable virtual Trusted Platform Module (vTPM) for Azure VM. Used with security_type=TrustedLaunch",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceAzureVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_virtual_machine` manages an Azure virtual machine in Duplo.",

		ReadContext:   resourceAzureVirtualMachineRead,
		CreateContext: resourceAzureVirtualMachineCreate,
		UpdateContext: resourceAzureVirtualMachineUpdate,
		DeleteContext: resourceAzureVirtualMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureVirtualMachineSchema(),
	}
}

func resourceAzureVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var fullname string
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineRead(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)

	if c.IsAzureCustomPrefixesEnabled() {
		fullname = c.AddPrefixSuffixFromResourceName(name, "vm", false)
	} else {
		fullname = name
	}

	duplo, clientErr := c.AzureNativeHostGet(tenantID, fullname)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure virtual machine %s : %s", tenantID, name, clientErr)
	}
	d.Set("friendly_name", name)
	d.Set("fullname", fullname)
	d.Set("tenant_id", tenantID)
	flattenAzureVirtualMachine(d, duplo)
	minion, _ := c.GetMinionForHost(tenantID, fullname)
	if minion != nil {
		log.Printf("[TRACE] Minion found for host (%s).", fullname)
		d.Set("is_minion", true)
		d.Set("agent_platform", minion.AgentPlatform)
	} else {
		log.Printf("[TRACE] Minion not found for host (%s).", fullname)
		d.Set("is_minion", false)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineRead(%s, %s): end", tenantID, fullname)
	return nil
}

func resourceAzureVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	var fullname string

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("friendly_name").(string)
	log.Printf("[TRACE] resourceAzureVirtualMachineCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	if c.IsAzureCustomPrefixesEnabled() {
		fullname = c.AddPrefixSuffixFromResourceName(name, "vm", false)
	} else {
		fullname = name
	}

	rq := expandAzureVirtualMachine(d)
	err = c.AzureNativeHostCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure virtual machine '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure virtual machine", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureNativeHostGet(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the virtual machine to be ready.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = virtualMachineWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureVirtualMachineRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVirtualMachineCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("friendly_name").(string)
	fullname := d.Get("fullname").(string)
	log.Printf("[TRACE] resourceAzureVirtualMachineUpdate(%s, %s): start", tenantID, fullname)
	c := m.(*duplosdk.Client)

	if d.HasChange("capacity") {
		clientErr := c.UpdateAzureVirtualMachineSize(tenantID, &duplosdk.UpdateAzureVirtualMachineSizeReq{
			Capacity:     d.Get("capacity").(string),
			FriendlyName: fullname,
		})
		if clientErr != nil {
			return diag.Errorf("Error updating tenant %s azure virtual machine capacity '%s': %s", tenantID, name, err)
		}
		time.Sleep(time.Duration(40) * time.Second)
		err = virtualMachineWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if needsAzureVMUpdate(d) {
		rq := expandAzureVirtualMachine(d)
		rq.FriendlyName = fullname
		err = c.AzureNativeHostCreate(rq)
		if err != nil {
			return diag.Errorf("Error creating tenant %s azure virtual machine '%s': %s", tenantID, name, err)
		}

		id := fmt.Sprintf("%s/%s", tenantID, name)
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure virtual machine", id, func() (interface{}, duplosdk.ClientError) {
			return c.AzureNativeHostGet(tenantID, fullname)
		})
		if diags != nil {
			return diags
		}
		d.SetId(id)

		//By default, wait until the virtual machine to be ready.
		if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
			err = virtualMachineWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		diags = resourceAzureVirtualMachineRead(ctx, d, m)
		log.Printf("[TRACE] resourceAzureVirtualMachineUpdate(%s, %s): end", tenantID, name)
		return diags
	}
	return nil
}

func resourceAzureVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineIdParts(id)
	fullname := d.Get("fullname").(string)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineDelete(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureNativeHostDelete(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure virtual machine '%s': %s", tenantID, fullname, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure virtual machine", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureNativeHostGet(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVirtualMachineDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureVirtualMachine(d *schema.ResourceData) *duplosdk.DuploNativeHost {
	metadata := []duplosdk.DuploKeyStringValue{}
	if v, ok := d.GetOk("disk_size_gb"); ok {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "OsDiskSize",
			Value: strconv.Itoa(v.(int)),
		})
	}
	if v, ok := d.GetOk("os_disk_type"); ok && v != nil && v.(string) != "" {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "OsDiskType",
			Value: v.(string),
		})
	}
	if v, ok := d.GetOk("admin_username"); ok && v != nil && v.(string) != "" {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "Username",
			Value: v.(string),
		})
	}
	if v, ok := d.GetOk("admin_password"); ok && v != nil && v.(string) != "" {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "Password",
			Value: v.(string),
		})
	}
	if v, ok := d.GetOk("ad_domain_type"); ok && v != nil && v.(string) != "" {
		if v.(string) == "aadjoin" {
			metadata = append(metadata, duplosdk.DuploKeyStringValue{
				Key:   "JoinAADDomain",
				Value: "true",
			})
		} else {
			metadata = append(metadata, duplosdk.DuploKeyStringValue{
				Key:   "JoinDomain",
				Value: "true",
			})
		}

	}
	if v, ok := d.GetOk("join_domain"); ok && v != nil {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "JoinDomain",
			Value: strconv.FormatBool(d.Get("join_domain").(bool)),
		})
	}
	if v, ok := d.GetOk("enable_log_analytics"); ok && v != nil {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "JoinLogAnalytics",
			Value: strconv.FormatBool(d.Get("enable_log_analytics").(bool)),
		})
	}
	if v, ok := d.GetOk("timezone"); ok && v != nil && v.(string) != "" {
		metadata = append(metadata, duplosdk.DuploKeyStringValue{
			Key:   "Timezone",
			Value: v.(string),
		})
	}
	data := &duplosdk.DuploNativeHost{
		TenantID:          d.Get("tenant_id").(string),
		InstanceID:        d.Get("instance_id").(string),
		FriendlyName:      d.Get("friendly_name").(string),
		Capacity:          d.Get("capacity").(string),
		IsMinion:          d.Get("is_minion").(bool),
		ImageID:           d.Get("image_id").(string),
		Base64UserData:    d.Get("base64_user_data").(string),
		AgentPlatform:     d.Get("agent_platform").(int),
		AllocatedPublicIP: d.Get("allocated_public_ip").(bool),
		Cloud:             2, // For Azure
		EncryptDisk:       d.Get("encrypt_disk").(bool),
		MetaData:          &metadata,
		TagsEx:            keyValueFromState("tags", d),
		MinionTags:        keyValueFromState("minion_tags", d),
		Volumes:           expandAzureNativeHostVolumes("volume", d),
		NetworkInterfaces: &[]duplosdk.DuploNativeHostNetworkInterface{
			{
				SubnetID: d.Get("subnet_id").(string),
			},
		},
		SecurityType:    d.Get("security_type").(string),
		IsEncryptAtHost: d.Get("enable_encrypt_at_host").(bool),
		DiskControlType: d.Get("disk_control_type").(string),
	}
	if data.SecurityType == "TrustedLaunch" {
		data.IsSecureBoot = d.Get("enable_security_boot").(bool)
		data.IsvTPM = d.Get("enable_vtpm").(bool)
	}
	return data
}

func expandAzureNativeHostVolumes(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostVolume {
	result := []duplosdk.DuploNativeHostVolume{}

	if rawlist, ok := d.GetOk(key); ok && rawlist != nil && len(rawlist.([]interface{})) > 0 {
		volumes := rawlist.([]interface{})

		for _, raw := range volumes {
			result = make([]duplosdk.DuploNativeHostVolume, 0, len(volumes))
			volume := raw.(map[string]interface{})

			duplo := duplosdk.DuploNativeHostVolume{}
			if v, ok := volume["iops"]; ok && v != nil && v.(int) > 0 {
				duplo.Iops = v.(int)
			}
			if v, ok := volume["name"]; ok && v != nil && v.(string) != "" {
				duplo.Name = v.(string)
			}
			if v, ok := volume["size"]; ok && v != nil && v.(int) > 0 {
				duplo.Size = v.(int)
			}
			if v, ok := volume["volume_id"]; ok && v != nil && v.(string) != "" {
				duplo.VolumeID = v.(string)
			}
			if v, ok := volume["volume_type"]; ok && v != nil && v.(string) != "" {
				duplo.VolumeType = v.(string)
			}
			result = append(result, duplo)
		}
	}

	return &result
}
func parseAzureVirtualMachineIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureVirtualMachine(d *schema.ResourceData, duplo *duplosdk.DuploNativeHost) {
	d.Set("instance_id", duplo.InstanceID)
	d.Set("capacity", duplo.Capacity)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("status", duplo.Status)
	d.Set("user_account", duplo.UserAccount)
	d.Set("tags", flattenTags(duplo.TagsEx))
	d.Set("security_type", duplo.SecurityType)
	if duplo.SecurityType == "TrustedLaunch" {
		d.Set("enable_security_boot", duplo.IsSecureBoot)
		d.Set("enable_vtpm", duplo.IsvTPM)
	}
	d.Set("enable_encrypt_at_host", duplo.IsEncryptAtHost)
	d.Set("disk_control_type", duplo.DiskControlType)
}

func flattenTags(tags *[]duplosdk.DuploKeyStringValue) []interface{} {
	if tags != nil {
		managedTags := DuploManagedAzureTags()
		output := []interface{}{}
		for _, duploObject := range *tags {
			if Contains(managedTags, duploObject.Key) {
				continue
			}

			jo := make(map[string]interface{})
			jo["key"] = duploObject.Key
			jo["value"] = duploObject.Value
			output = append(output, jo)
		}
		return output
	}

	return make([]interface{}, 0)
}

func virtualMachineWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AzureNativeHostGet(tenantID, name)
			log.Printf("[TRACE] Virtual machine provisioning state is (%s).", rp.Status)
			status := "pending"
			if err == nil {
				if rp.Status == "VM running" {
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
	log.Printf("[DEBUG] virtualMachineWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func needsAzureVMUpdate(d *schema.ResourceData) bool {
	return d.HasChange("join_domain") ||
		d.HasChange("ad_domain_type") ||
		d.HasChange("enable_log_analytics") ||
		d.HasChange("timezone") ||
		d.HasChange("disk_size_gb") ||
		d.HasChange("os_disk_type") ||
		d.HasChange("volume")
}
