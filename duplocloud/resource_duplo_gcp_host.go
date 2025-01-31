package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func nativeGcpHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the host will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"friendly_name": {
			Description: "The name of the vm.",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			ForceNew:    true, // relaunch instance
		},

		"user_account": {
			Description: "The email id of the user.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // relaunch instance
		},

		"fullname": {
			Description: "The full name of the vm.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"capacity": {
			Description: "The machine type to create",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
		},
		"image_id": {
			Description:      "The image from which to initialize this vm",
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: diffSuppressGCPHostImageIdIfSame,
		},

		"instance_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent pool that this host is added to.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true, // relaunch instance
			Default:     0,
		},
		"tags": {
			Description: "List of network tags that can be added to the vm",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"allocated_public_ip": {
			Description: "Whether or not to allocate a public IP.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true, // relaunch instance
		},
		"zone": {
			Description: "The zone that the machine should be created in",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // relaunch instance
		},
		"accelerator_count": {
			Description: "The number of the guest accelerator cards exposed to this instance.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
			ForceNew:    true,
		},
		"metadata": {
			Description:      "Configuration, metadata used when creating the host.<br>*Note: To configure OS disk size OsDiskSize can be specified as Key and its size as value, size value should be atleast 10, Similarly to added start up script one can pass the key as startup_script and startup command as its value*",
			Type:             schema.TypeMap,
			Optional:         true,
			Computed:         true,
			Elem:             &schema.Schema{Type: schema.TypeString},
			DiffSuppressFunc: diffSuppressOnComputedDataOnMetadataBlock,
		},
		"labels": {
			Description:      "A set of key/value label pairs assigned to the vm",
			Type:             schema.TypeMap,
			Optional:         true,
			Computed:         true,
			Elem:             &schema.Schema{Type: schema.TypeString},
			DiffSuppressFunc: diffSuppressOnComputedDataOnLabelBlock,
		},
		"accelerator_type": {
			Description: "The accelerator type resource to expose to this instance",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"private_ip_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"public_ip_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"architecture": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"identity_role": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"self_link": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"wait_until_ready": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
	}
}

// SCHEMA for resource crud
func resourceGcpHost() *schema.Resource {

	return &schema.Resource{
		Description: "The duplocloud_gcp_host used to manage or configure virtual machine at gcp",

		ReadContext:   resourceGcpHostRead,
		CreateContext: resourceGcpHostCreate,
		UpdateContext: resourceGcpHostUpdate,
		DeleteContext: resourceGcpHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        nativeGcpHostSchema(),
		CustomizeDiff: validateGCPHostAttributes,
	}
}

func resourceGcpHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceGcpHostRead(%s): start", id)
	idparts := strings.Split(id, "/")
	tenantID, instanceID := idparts[0], idparts[2]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploGcpNativeHostGet(tenantID, instanceID)
	if err != nil {

		return diag.Errorf("Unable to retrieve gcp host '%s': %s", id, err)

	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Apply the data
	flattenGcpHost(d, duplo)

	log.Printf("[TRACE] resourceGcpHostRead(%s): end", id)
	return nil
}

func resourceGcpHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantId := d.Get("tenant_id").(string)
	// Build a request.
	rq := expandGcpHost(d)
	log.Printf("[TRACE] resourceGcpHostCreate(%s, %s): start", tenantId, rq.Name)

	c := m.(*duplosdk.Client)
	c.UserAccount = d.Get("user_account").(string)
	// Set the NetworkInterfaces property as needed.

	// Create the host in Duplo.
	rp, err := c.DuploGcpHostCreate(tenantId, &rq)
	if err != nil {
		return diag.Errorf("Error creating GCP host '%s': %s", rq.Name, err.Error())
	}

	// Wait up to 60 seconds for Duplo to be able to return the host details.
	id := fmt.Sprintf("%s/gcpHost/%s", tenantId, rp.InstanceId)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "Gcp host", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploGcpNativeHostGet(tenantId, rp.InstanceId)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// By default, wait until the host is completely ready.
	if d.Get("wait_until_ready").(bool) {
		err = gcpHostWaitUntilReady(ctx, c, tenantId, rp.InstanceId, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the host from the backend again.
	diags = resourceGcpHostRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpHostCreate(%s, %s): end", tenantId, rp.InstanceId)
	return diags
}

// UPDATE resource

func resourceGcpHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Build a request.
	tenantId := d.Get("tenant_id").(string)
	instanceId := d.Get("instance_id").(string)
	rq := expandGcpHostOnUpdate(d)
	log.Printf("[TRACE] resourceGcpHostUpdate(%s, %s): start", tenantId, instanceId)

	// Update the host in Duplo.
	c := m.(*duplosdk.Client)
	c.UserAccount = d.Get("user_account").(string)

	rp, err := c.DuploGcpHostUpdate(tenantId, instanceId, &rq)
	if err != nil {
		return diag.Errorf("Error creating GCP host '%s': %s", instanceId, err)
	}
	if d.Get("wait_until_ready").(bool) {
		err := gcpHostWaitUntilReady(ctx, c, tenantId, rp.InstanceId, d.Timeout("update"))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// Read the host from the backend again.
	diags := resourceGcpHostRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpHostUpdate(%s, %s): end", tenantId, instanceId)
	return diags
}

// DELETE resource
func resourceGcpHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceGcpHostDelete(%s): start", id)
	idParts := strings.Split(id, "/")
	tenantID, instanceID := idParts[0], idParts[2]

	// Delete the host from Duplo
	c := m.(*duplosdk.Client)

	err := c.DuploGcpHostDelete(tenantID, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for the host to be missing
	diags = waitForResourceToBeMissingAfterDelete(ctx, d, "GCP host", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.DuploGcpNativeHostGet(tenantID, instanceID); rp != nil || err != nil {
			return rp, err
		}
		return nil, nil
	})

	log.Printf("[TRACE] resourceGcpHostDelete(%s): end", id)
	return diags
}

func gcpHostWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, instanceID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploGcpNativeHostGet(tenantID, instanceID)
			status := "pending"
			if err == nil && rp.Status == "RUNNING" {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] gcpHostWaitUntilReady(%s, %s)", tenantID, instanceID)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func flattenGcpHost(d *schema.ResourceData, duplo *duplosdk.DuploGcpHost) {
	d.Set("fullname", duplo.Name)
	d.Set("capacity", duplo.Capacity)
	d.Set("image_id", duplo.ImageId)
	d.Set("instance_id", duplo.InstanceId)
	d.Set("agent_platform", duplo.AgentPlatform)
	if len(duplo.Tags) > 0 {
		d.Set("tags", trimStringsByPosition(duplo.Tags, 3))
	}
	d.Set("allocated_public_ip", duplo.EnablePublicIpAddress)
	d.Set("zone", duplo.Zone)
	d.Set("accelerator_count", duplo.AcceleratorCount)
	if len(duplo.Metadata) > 0 {
		d.Set("metadata", duplo.Metadata)
	}
	d.Set("labels", filterOutDefaultLabels(duplo.Labels))
	d.Set("accelerator_type", duplo.AcceleratorType)
	d.Set("private_ip_address", duplo.PrivateIpAddress)
	d.Set("public_ip_address", duplo.PublicIpAddress)
	d.Set("architecture", duplo.Arch)
	d.Set("identity_role", duplo.IdentityRole)
	d.Set("self_link", duplo.SelfLink)

	d.Set("status", duplo.Status)

}
func expandGcpHost(d *schema.ResourceData) duplosdk.DuploGcpHost {
	obj := duplosdk.DuploGcpHost{}
	obj.Name = d.Get("friendly_name").(string)
	obj.Capacity = d.Get("capacity").(string)
	obj.ImageId = d.Get("image_id").(string)
	if v, ok := d.GetOk("instance_id"); ok {
		obj.InstanceId = v.(string)
	}
	obj.AgentPlatform = d.Get("agent_platform").(int)
	tag := d.Get("tags").([]interface{})
	for _, v := range tag {
		obj.Tags = append(obj.Tags, v.(string))
	}
	obj.EnablePublicIpAddress = d.Get("allocated_public_ip").(bool)
	obj.Zone = d.Get("zone").(string)
	obj.AcceleratorCount = d.Get("accelerator_count").(int)
	if v, ok := d.GetOk("metadata"); ok {
		obj.Metadata = v.(map[string]interface{})
	}
	if v, ok := d.GetOk("labels"); ok {
		obj.Labels = make(map[string]string)
		for k, vl := range v.(map[string]interface{}) {
			obj.Labels[k] = vl.(string)

		}
	}
	if v, ok := d.GetOk("user_data"); ok && v.(string) != "" {
		obj.UserData = v.(string)
	}
	obj.AcceleratorType = d.Get("accelerator_type").(string)
	if v, ok := d.GetOk("private_ip_address"); ok && v.(string) != "" {
		obj.PrivateIpAddress = v.(string)
	}
	if v, ok := d.GetOk("public_ip_address"); ok && v.(string) != "" {
		obj.PublicIpAddress = v.(string)
	}
	if v, ok := d.GetOk("architecture"); ok && v.(string) != "" {
		obj.Arch = v.(string)
	}
	if v, ok := d.GetOk("identity_role"); ok && v.(string) != "" {
		obj.IdentityRole = v.(string)
	}
	if v, ok := d.GetOk("self_link"); ok && v.(string) != "" {
		obj.SelfLink = v.(string)
	}
	return obj
}

func expandGcpHostOnUpdate(d *schema.ResourceData) duplosdk.DuploGcpHost {
	obj := duplosdk.DuploGcpHost{}

	tag := d.Get("tags").([]interface{})
	for _, v := range tag {
		obj.Tags = append(obj.Tags, v.(string))
	}
	if v, ok := d.GetOk("metadata"); ok {
		obj.Metadata = v.(map[string]interface{})
	}
	if v, ok := d.GetOk("labels"); ok {
		obj.Labels = make(map[string]string)
		for k, vl := range v.(map[string]interface{}) {
			obj.Labels[k] = vl.(string)

		}

	}

	return obj
}

func validateGCPHostAttributes(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	accCount := diff.Get("accelerator_count").(int)
	accType := diff.Get("accelerator_type").(string)
	if accCount > 0 && accType == "" {
		return fmt.Errorf("missing accelerator_type")
	} else if accType != "" && accCount <= 0 {
		return fmt.Errorf("accelerator_count should be greater than zero")
	}
	return nil
}
