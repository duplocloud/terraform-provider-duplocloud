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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		"capacity": {
			Description: "Specifies the [size of the Virtual Machine](https://docs.microsoft.com/azure/virtual-machines/sizes-general). See also [Azure VM Naming Conventions](https://docs.microsoft.com/azure/virtual-machines/vm-naming-conventions).",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			ForceNew:    true,
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
			ForceNew:    true,
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
			ForceNew:    true,
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
			Computed:    true,
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
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"minion_tags": {
			Description: "A map of tags to assign to the resource. Example - `AllocationTags` can be passed as tag key with any value.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			ForceNew:    true, // relaunch instance
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
		"wait_until_ready": {
			Description: "Whether or not to wait until azure virtual machine to be ready, after creation.",
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
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureNativeHostGet(tenantID, name)
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
	d.Set("tenant_id", tenantID)
	flattenAzureVirtualMachine(d, duplo)
	minion, _ := c.GetMinionForHost(tenantID, name)
	if minion != nil {
		log.Printf("[TRACE] Minion found for host (%s).", name)
		d.Set("is_minion", true)
		d.Set("agent_platform", minion.AgentPlatform)
	} else {
		log.Printf("[TRACE] Minion not found for host (%s).", name)
		d.Set("is_minion", false)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("friendly_name").(string)
	log.Printf("[TRACE] resourceAzureVirtualMachineCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureVirtualMachine(d)
	err = c.AzureNativeHostCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure virtual machine '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure virtual machine", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureNativeHostGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the virtual machine to be ready.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = virtualMachineWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureVirtualMachineRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVirtualMachineCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAzureVirtualMachineCreate(ctx, d, m)
}

func resourceAzureVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureVirtualMachineIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVirtualMachineDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureNativeHostDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure virtual machine '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure virtual machine", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureNativeHostGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVirtualMachineDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureVirtualMachine(d *schema.ResourceData) *duplosdk.DuploNativeHost {
	diskSizeKV, usernameKV, passwordKV := duplosdk.DuploKeyStringValue{}, duplosdk.DuploKeyStringValue{},
		duplosdk.DuploKeyStringValue{}

	if v, ok := d.GetOk("disk_size_gb"); ok {
		diskSizeKV = duplosdk.DuploKeyStringValue{
			Key:   "OsDiskSize",
			Value: strconv.Itoa(v.(int)),
		}
	}
	if v, ok := d.GetOk("admin_username"); ok && v != nil && v.(string) != "" {
		usernameKV = duplosdk.DuploKeyStringValue{
			Key:   "Username",
			Value: v.(string),
		}
	}
	if v, ok := d.GetOk("admin_password"); ok && v != nil && v.(string) != "" {
		passwordKV = duplosdk.DuploKeyStringValue{
			Key:   "Password",
			Value: v.(string),
		}
	}
	joinDomainKV := duplosdk.DuploKeyStringValue{
		Key:   "JoinDomain",
		Value: strconv.FormatBool(d.Get("join_domain").(bool)),
	}
	logAnalyticKV := duplosdk.DuploKeyStringValue{
		Key:   "JoinLogAnalytics",
		Value: strconv.FormatBool(d.Get("enable_log_analytics").(bool)),
	}
	return &duplosdk.DuploNativeHost{
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
		MetaData: &[]duplosdk.DuploKeyStringValue{
			diskSizeKV, usernameKV, passwordKV, joinDomainKV, logAnalyticKV,
		},
		TagsEx:     keyValueFromState("tags", d),
		MinionTags: keyValueFromState("minion_tags", d),
		Volumes:    expandAzureNativeHostVolumes("volume", d),
		NetworkInterfaces: &[]duplosdk.DuploNativeHostNetworkInterface{
			{
				SubnetID: d.Get("subnet_id").(string),
			},
		},
	}
}

func expandAzureNativeHostVolumes(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostVolume {
	result := []duplosdk.DuploNativeHostVolume{}

	if rawlist, ok := d.GetOk(key); ok && rawlist != nil && len(rawlist.([]interface{})) > 0 {
		volumes := rawlist.([]interface{})

		result = make([]duplosdk.DuploNativeHostVolume, 0, len(volumes))
		for _, raw := range volumes {
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
	stateConf := &resource.StateChangeConf{
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
