package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func nativeHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the host will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"user_account": {
			Description:      "The name of the tenant that the host will be created in.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"friendly_name": {
			Description:      "The short name of the host.",
			Type:             schema.TypeString,
			Optional:         false,
			Required:         true,
			ForceNew:         true, // relaunch instance
			DiffSuppressFunc: diffSuppressIfSame,
		},
		"instance_id": {
			Description: "The AWS EC2 instance ID of the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"capacity": {
			Description: "The AWS EC2 instance type.",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			ForceNew:    true, // relaunch instnace
		},
		"zone": {
			Description: "The availability zone to launch the host in, expressed as a number and starting at 0.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true, // relaunch instance
			Default:     0,
		},
		"is_minion": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true, // relaunch instance
			Default:  true,
		},
		"image_id": {
			Description: "The AMI ID to use.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // relaunch instance
		},
		"base64_user_data": {
			Description: "Base64 encoded EC2 user data to associated with the host.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true, // relaunch instance
			Computed:    true,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent pool that this host is added to.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true, // relaunch instance
			Default:     0,
		},
		"is_ebs_optimized": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"allocated_public_ip": {
			Description: "Whether or not to allocate a public IP.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true, // relaunch instance
		},
		"cloud": {
			Description: "The numeric ID of the cloud provider to launch the host in.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true, // relaunch instance
		},
		"keypair_type": {
			Description: "The numeric ID of the keypair type being used." +
				"Should be one of:\n\n" +
				"   - `0` : Default\n" +
				"   - `1` : ED25519\n" +
				"   - `2` : RSA (deprecated - some operating systems no longer support it)\n",
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"encrypt_disk": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"status": {
			Description: "The current status of the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"identity_role": {
			Description: "The name of the IAM role associated with this host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_ip_address": {
			Description: "The primary private IP address assigned to the host.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"metadata": {
			Description: "Configuration metadata used when creating the host.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"tags": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true, // relaunch instance
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
		"network_interface": {
			Description: "An optional list of custom network interface configurations to use when creating the host.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true, // relaunch instance
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"network_interface_id": {
						Description: "The ID of an ENI to attach to this host.  Cannot be specified if `subnet_id` or `associate_public_ip` is specified.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"subnet_id": {
						Description: "The ID of a subnet in which to create a new ENI.  Cannot be specified if `network_interface_id` is specified.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"device_index": {
						Description: "The device index to pass to AWS for attaching the ENI.  Starts at zero.",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},
					"associate_public_ip": {
						Description: "Whether or not to associate a public IP with the newly created ENI.  Cannot be specified if `network_interface_id` is specified.",
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
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
	awsHostSchema := nativeHostSchema()

	awsHostSchema["wait_until_connected"] = &schema.Schema{
		Description:      "Whether or not to wait until Duplo can connect to the host, after creation.",
		Type:             schema.TypeBool,
		Optional:         true,
		ForceNew:         true,
		Default:          true,
		DiffSuppressFunc: diffSuppressWhenNotCreating,
	}

	return &schema.Resource{
		Description: "`duplocloud_aws_host` manages a native AWS host in Duplo.",

		ReadContext:   resourceAwsHostRead,
		CreateContext: resourceAwsHostCreate,
		DeleteContext: resourceAwsHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsHostSchema,
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

		// backend may return a 400 instead of a 404
		exists, err2 := c.NativeHostExists(tenantID, instanceID)
		if exists || err2 != nil {
			return diag.Errorf("Unable to retrieve AWS host '%s': %s", id, err)
		}
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
	var err error

	// Build a request.
	rq := expandNativeHost(d)
	log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): start", rq.TenantID, rq.FriendlyName)

	c := m.(*duplosdk.Client)

	// Set the NetworkInterfaces property as needed.
	diags := setNetworkInterfaces(rq, c)
	if diags != nil {
		return diags
	}

	// Create the host in Duplo.
	rp, err := c.NativeHostCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating AWS host '%s': %s", rq.FriendlyName, err)
	}
	if rp.InstanceID == "" {
		return diag.Errorf("Error creating AWS host '%s': no instance ID was received", rq.FriendlyName)
	}

	// Wait up to 60 seconds for Duplo to be able to return the host details.
	id := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", rp.TenantID, rp.InstanceID)
	diags = waitForResourceToBePresentAfterCreate(ctx, d, "AWS host", id, func() (interface{}, duplosdk.ClientError) {
		return c.NativeHostGet(rp.TenantID, rp.InstanceID)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// By default, wait until the host is completely ready.
	if d.Get("wait_until_connected") == nil || d.Get("wait_until_connected").(bool) {
		err = nativeHostWaitUntilReady(ctx, c, rp.TenantID, rp.InstanceID, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the host from the backend again.
	diags = resourceAwsHostRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): end", rq.TenantID, rq.FriendlyName)
	return diags
}

func setNetworkInterfaces(rq *duplosdk.DuploNativeHost, c *duplosdk.Client) diag.Diagnostics {
	// Handle subnet selection for hosts
	var subnetIds []string
	var err duplosdk.ClientError
	var orientation string

	if rq.AllocatedPublicIP {
		orientation = "external"
		subnetIds, err = c.TenantGetExternalSubnets(rq.TenantID)
	} else {
		orientation = "internal"
		subnetIds, err = c.TenantGetInternalSubnets(rq.TenantID)
	}

	if err != nil {
		return diag.Errorf("Error creating AWS host '%s': failed to get %s subnets for tenant '%s'"+
			"Internal error: %s", rq.FriendlyName, orientation, rq.TenantID, err)
	}

	if len(subnetIds) == 0 {
		return diag.Errorf("Error creating AWS host '%s': no %s subnets were found for tenant '%s'"+
			rq.FriendlyName, orientation, rq.TenantID)
	}

	if rq.Zone < 0 || rq.Zone >= len(subnetIds) {
		return diag.Errorf("Error creating AWS host '%s': zone %d is invalid. zone must be between 0 and %d.",
			rq.FriendlyName, rq.Zone, len(subnetIds))
	}

	subnetId := subnetIds[rq.Zone]

	// When AllocatedPublicIP is true, ensure there is at least one network interface
	if rq.AllocatedPublicIP && (rq.NetworkInterfaces == nil || len(*rq.NetworkInterfaces) == 0) {
		// No network interfaces, create a new one on the external subnet for the given zone.
		rq.NetworkInterfaces = &[]duplosdk.DuploNativeHostNetworkInterface{{
			SubnetID: subnetId,
		}}
	}

	if rq.NetworkInterfaces == nil {
		return nil
	}

	// Ensure all network interfaces without an ID are using the correct subnet
	for idx, niConfig := range *rq.NetworkInterfaces {
		if niConfig.NetworkInterfaceID != "" && niConfig.SubnetID != "" {
			return diag.Errorf("Error creating AWS host '%s': a subnetId on network interface %d cannot be specified since network_interface_id '%s' is provided",
				rq.FriendlyName, idx, niConfig.NetworkInterfaceID)
		}

		if niConfig.NetworkInterfaceID == "" {
			if niConfig.SubnetID == "" {
				niConfig.SubnetID = subnetId
			} else if niConfig.SubnetID != subnetId {
				return diag.Errorf("Error creating AWS host '%s': %s subnetId on network interface %d for zone %d must be '%s' instead of '%s'",
					rq.FriendlyName, orientation, idx, rq.Zone, subnetId, niConfig.SubnetID)
			}
		}
	}

	return nil
}

// UPDATE resource
/*
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
*/

// DELETE resource
func resourceAwsHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceAwsHostDelete(%s): start", id)
	tenantID, instanceID, err := nativeHostIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if the host exists
	c := m.(*duplosdk.Client)
	exists, err := c.NativeHostExists(tenantID, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}
	if exists {

		// Delete the host from Duplo
		err = c.NativeHostDelete(tenantID, instanceID)
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for the host to be missing
		diags = waitForResourceToBeMissingAfterDelete(ctx, d, "AWS host", id, func() (interface{}, duplosdk.ClientError) {
			if rp, err := c.NativeHostExists(tenantID, instanceID); rp || err != nil {
				return rp, err
			}
			return nil, nil
		})
	}

	log.Printf("[TRACE] resourceAwsHostDelete(%s): end", id)
	return diags
}

func expandNativeHost(d *schema.ResourceData) *duplosdk.DuploNativeHost {
	return &duplosdk.DuploNativeHost{
		TenantID:          d.Get("tenant_id").(string),
		InstanceID:        d.Get("instance_id").(string),
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
		KeyPairType:       d.Get("keypair_type").(int),
		EncryptDisk:       d.Get("encrypt_disk").(bool),
		MetaData:          keyValueFromState("metadata", d),
		Tags:              keyValueFromState("tags", d),
		MinionTags:        keyValueFromState("minion_tags", d),
		Volumes:           expandNativeHostVolumes("volume", d),
		NetworkInterfaces: expandNativeHostNetworkInterfaces("network_interface", d),
	}
}

func expandNativeHostVolumes(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostVolume {
	var result []duplosdk.DuploNativeHostVolume

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

func expandNativeHostNetworkInterfaces(key string, d *schema.ResourceData) *[]duplosdk.DuploNativeHostNetworkInterface {
	var result []duplosdk.DuploNativeHostNetworkInterface

	if rawlist, ok := d.GetOk(key); ok && rawlist != nil && len(rawlist.([]interface{})) > 0 {
		nics := rawlist.([]interface{})

		result = make([]duplosdk.DuploNativeHostNetworkInterface, 0, len(nics))
		for _, raw := range nics {
			nic := raw.(map[string]interface{})

			duplo := duplosdk.DuploNativeHostNetworkInterface{
				AssociatePublicIP: nic["associate_public_ip"].(bool),
				MetaData:          keyValueFromStateList("metadata", nic),
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
	d.Set("is_minion", duplo.IsMinion)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("is_ebs_optimized", duplo.IsEbsOptimized)
	d.Set("cloud", duplo.Cloud)
	d.Set("keypair_type", duplo.KeyPairType)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("status", duplo.Status)
	d.Set("identity_role", duplo.IdentityRole)
	d.Set("private_ip_address", duplo.PrivateIPAddress)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("minion_tags", keyValueToState("minion_tags", duplo.MinionTags))

	// If a network interface was customized, certain fields are not returned by the backend.
	if v, ok := d.GetOk("network_interface"); !ok || v == nil || len(v.([]interface{})) == 0 {
		d.Set("zone", duplo.Zone)
		d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	}

	// TODO:  The backend doesn't return these yet.
	// d.Set("metadata", keyValueToState("metadata", duplo.MetaData))
	// d.Set("volume", flattenNativeHostVolumes(duplo.Volumes))
	// d.Set("network_interface", flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces))
}

func flattenNativeHostVolumes(duplo *[]duplosdk.DuploNativeHostVolume) []map[string]interface{} {
	if duplo == nil {
		return []map[string]interface{}{}
	}

	list := make([]map[string]interface{}, len(*duplo))
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

	list := make([]map[string]interface{}, len(*duplo))
	for _, item := range *duplo {
		nic := map[string]interface{}{
			"associate_public_ip": item.AssociatePublicIP,
			"metadata":            keyValueToState("metadata", item.MetaData),
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
		return "", "", fmt.Errorf("invalid resource ID: %s", id)
	}
	return idParts[2], idParts[4], nil
}

// NativeHostWaitForCreation waits for creation of an AWS Host by the Duplo API
func nativeHostWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, instanceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
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
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func diffSuppressIfSame(k, old string, new string, d *schema.ResourceData) bool {
	if d.IsNewResource() {
		return true
	}

	oldFullName := d.Get("fullname").(string) // duploservices-tenant02-tftestasg01 (from Duplo API)

	// new: duploservices-tenant02-tftestasg01
	if strings.HasPrefix(new, "duploservices-") {
		log.Printf("[DEBUG]diffSuppressIfSame old: %s, new: %s)", oldFullName, new)
		return oldFullName == new
	}

	oldAccountName := d.Get("user_account").(string)
	prefix := strings.Join([]string{"duploservices", oldAccountName}, "-")
	oldName, _ := duplosdk.UnprefixName(prefix, oldFullName)

	log.Printf("[DEBUG]diffSuppressIfSame prefix: %s and new: %s, old: %s)", prefix, new, oldName)

	// new: tftestasg01
	return oldName == new
}
