package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func nativeHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"instance_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"user_account": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"friendly_name": {
			Type:             schema.TypeString,
			Optional:         false,
			Required:         true,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"capacity": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, // relaunch instnace
		},
		"zone": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true, // relaunch instance
			Default:  0,
		},
		"is_minion": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"image_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true, // relaunch instance
		},
		"base64_user_data": {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"is_ebs_optimized": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"allocated_public_ip": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
			ForceNew: true, // relaunch instance
		},
		"encrypt_disk": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"identity_role": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"private_ip_address": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"metadata": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"tags": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"minion_tags": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"volumes": {
			Type:             schema.TypeSet,
			Optional:         true,
			ForceNew:         true, // relaunch instance
			DiffSuppressFunc: diffSuppressFuncIgnore,
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
		"network_interfaces": {
			Type:             schema.TypeSet,
			Optional:         true,
			ForceNew:         true, // relaunch instance
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"network_interface_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"subnet_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"device_index": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"associate_public_ip": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"groups": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"metadata": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     KeyValueSchema(),
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceAwsHost() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceAwsHostRead,
		CreateContext: resourceAwsHostCreate,
		UpdateContext: resourceAwsHostUpdate,
		DeleteContext: resourceAwsHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: nativeHostSchema(),
	}
}

func resourceAwsHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceAwsHostRead(%s): start", id)
	tenantID, instanceID, err := nativeHostIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.NativeHostGet(tenantID, instanceID)
	if err != nil {
		return diag.Errorf("Unable to retrieve AWS host '%s': %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Apply the data
	nativeHostToState(d, duplo)

	log.Printf("[TRACE] resourceAwsHostRead(%s): end", id)
	return nil
}

func resourceAwsHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Build a request.
	rq := expandNativeHost(d)
	log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): start", rq.TenantID, rq.FriendlyName)

	// Create the host in Duplo.
	c := m.(*duplosdk.Client)
	rp, err := c.NativeHostCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating AWS host '%s': %s", rq.FriendlyName, err)
	}
	if rq.InstanceID == "" {
		return diag.Errorf("Error creating AWS host '%s': no instance ID was received", rq.FriendlyName)
	}

	// Wait up to 60 seconds for Duplo to be able to return the host details.
	id := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", rp.TenantID, rp.InstanceID)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "AWS host", id, func() (interface{}, error) {
		return c.NativeHostGet(rp.TenantID, rp.InstanceID)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Then, wait until the host is completely ready.
	err = nativeHostWaitUntilReady(c, rp.TenantID, rp.InstanceID, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	// Read the host from the backend again.
	diags = resourceAwsHostRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): end", rq.TenantID, rq.FriendlyName)
	return diags
}

/// UPDATE resource
func resourceAwsHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Build a request.
	rq := expandNativeHost(d)
	log.Printf("[TRACE] resourceAwsHostUpdate(%s, %s): start", rq.TenantID, rq.InstanceID)

	// Update the host in Duplo.
	c := m.(*duplosdk.Client)
	_, err := c.NativeHostUpdate(rq)
	if err != nil {
		return diag.Errorf("Error creating AWS host '%s': %s", rq.FriendlyName, err)
	}

	// Read the host from the backend again.
	diags := resourceAwsHostRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): end", rq.TenantID, rq.FriendlyName)
	return diags
}

/// DELETE resource
func resourceAwsHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceAwsHostDelete(%s): start", id)
	tenantID, instanceID, err := nativeHostIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete the host from Duplo
	c := m.(*duplosdk.Client)
	err = c.NativeHostDelete(tenantID, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for the host to be missing
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "AWS host", id, func() (interface{}, error) {
		return c.NativeHostGet(tenantID, instanceID)
	})

	log.Printf("[TRACE] resourceAwsHostDelete(%s): end", id)
	return diags
}

func expandNativeHost(d *schema.ResourceData) *duplosdk.DuploNativeHost {
	return &duplosdk.DuploNativeHost{
		TenantID:          d.Get("tenant_id").(string),
		UserAccount:       d.Get("user_account").(string),
		FriendlyName:      d.Get("friendly_name").(string),
		Capacity:          d.Get("capacity").(string),
		Zone:              d.Get("zone").(int),
		IsMinion:          d.Get("is_minion").(bool),
		ImageID:           d.Get("image_id").(string),
		Base64UserData:    d.Get("base64_user_data").(string),
		AgentPlatform:     d.Get("agent_platform").(int),
		IsEbsOptimized:    d.Get("is_ebs_optimized").(bool),
		AllocatedPublicIP: d.Get("allocated_public_ip").(bool),
		Cloud:             d.Get("cloud").(int),
		EncryptDisk:       d.Get("encrypt_disk").(bool),
		MetaData:          duplosdk.KeyValueFromState("metadata", d),
		Tags:              duplosdk.KeyValueFromState("tag", d),
		MinionTags:        duplosdk.KeyValueFromState("minion_tags", d),
		Volumes:           expandNativeHostVolumes("volumes", d),
		NetworkInterfaces: expandNativeHostNetworkInterfaces("network_interfaces", d),
	}
}

func expandNativeHostVolumes(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostVolume {
	var result []duplosdk.DuploNativeHostVolume

	if rawlist, ok := d.GetOk(key); ok && rawlist != nil && len(rawlist.([]interface{})) > 0 {
		volumes := rawlist.([]interface{})

		log.Printf("[TRACE] expandNativeHostVolumes ********: found %s", key)

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

func expandNativeHostNetworkInterfaces(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostNetworkInterface {
	var result []duplosdk.DuploNativeHostNetworkInterface

	if rawlist, ok := d.GetOk(key); ok && rawlist != nil && len(rawlist.([]interface{})) > 0 {
		nics := rawlist.([]interface{})

		log.Printf("[TRACE] expandNativeHostNetworkInterface ********: found %s", key)

		result = make([]duplosdk.DuploNativeHostNetworkInterface, 0, len(nics))
		for _, raw := range nics {
			nic := raw.(map[string]interface{})

			duplo := duplosdk.DuploNativeHostNetworkInterface{
				AssociatePublicIP: nic["associate_public_ip"].(bool),
				MetaData:          duplosdk.KeyValueFromMap("metadata", nic),
			}

			if v, ok := nic["subnet_id"]; ok && v != nil && v.(string) != "" {
				duplo.SubnetID = v.(string)
			}
			if v, ok := nic["network_interface_id"]; ok && v != nil && v.(string) != "" {
				duplo.NetworkInterfaceID = v.(string)
			}
			if v, ok := nic["device_index"]; ok && v != nil && v.(int) > 0 {
				duplo.DeviceIndex = v.(int)
			}
			if v, ok := nic["groups"]; ok && v != nil {
				duplo.Groups, _ = getStringArray(nic, "groups")
			}

			result = append(result, duplo)
		}
	}

	return &result
}

func nativeHostToState(d *schema.ResourceData, duplo *duplosdk.DuploNativeHost) {
	d.Set("instance_id", duplo.InstanceID)
	d.Set("user_account", duplo.UserAccount)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("friendly_name", duplo.FriendlyName)
	d.Set("capacity", duplo.Capacity)
	d.Set("zone", duplo.Zone)
	d.Set("is_minion", duplo.IsMinion)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("is_ebs_optimized", duplo.IsEbsOptimized)
	d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	d.Set("cloud", duplo.Cloud)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("status", duplo.Status)
	d.Set("identity_role", duplo.IdentityRole)
	d.Set("private_ip_address", duplo.PrivateIPAddress)
	d.Set("metadata", duplosdk.KeyValueToState("metadata", duplo.MetaData))
	d.Set("tags", duplosdk.KeyValueToState("tags", duplo.Tags))
	d.Set("minion_tags", duplosdk.KeyValueToState("minion_tags", duplo.MinionTags))
	d.Set("volumes", flattenNativeHostVolumes(duplo.Volumes))
	d.Set("network_interfaces", flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces))
}

func flattenNativeHostVolumes(duplo *[]duplosdk.DuploNativeHostVolume) []map[string]interface{} {
	if duplo == nil {
		return []map[string]interface{}{}
	}

	list := make([]map[string]interface{}, len(*duplo), len(*duplo))
	for _, item := range *duplo {
		list = append(list, map[string]interface{}{
			"iops":        item.Iops,
			"name":        item.Name,
			"size":        item.Size,
			"volume_id":   item.VolumeID,
			"volume_type": item.VolumeType,
		})
	}

	return list
}

func flattenNativeHostNetworkInterfaces(duplo *[]duplosdk.DuploNativeHostNetworkInterface) []map[string]interface{} {
	if duplo == nil {
		return []map[string]interface{}{}
	}

	list := make([]map[string]interface{}, len(*duplo), len(*duplo))
	for _, item := range *duplo {
		nic := map[string]interface{}{
			"associate_public_ip": item.AssociatePublicIP,
			"metadata":            duplosdk.KeyValueToState("metadata", item.MetaData),
		}

		if item.NetworkInterfaceID != "" {
			nic["network_interface_id"] = item.NetworkInterfaceID
		}
		if item.SubnetID != "" {
			nic["subnet_id"] = item.SubnetID
		}
		if item.Groups != nil {
			nic["groups"] = item.Groups
		}
		if item.DeviceIndex > 0 {
			nic["device_index"] = item.DeviceIndex
		}

		list = append(list, nic)
	}

	return list
}

func nativeHostIdParts(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) < 5 {
		return "", "", fmt.Errorf("Invalid resource ID: %s", id)
	}
	return idParts[2], idParts[4], nil
}

// NativeHostWaitForCreation waits for creation of an AWS Host by the Duplo API
func nativeHostWaitUntilReady(c *duplosdk.Client, tenantID, instanceID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.NativeHostGet(tenantID, instanceID)
			status := "pending"
			if err == nil && rp.Status == "running" {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] duploNativeHostWaitUntilReady(%s, %s)", tenantID, instanceID)
	_, err := stateConf.WaitForState()
	return err
}
