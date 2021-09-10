package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func autosalingGroupSchema() map[string]*schema.Schema {

	awsASGSchema := nativeHostSchema()
	delete(awsASGSchema, "instance_id")
	delete(awsASGSchema, "status")
	delete(awsASGSchema, "identity_role")
	delete(awsASGSchema, "private_ip_address")

	awsASGSchema["instance_count"] = &schema.Schema{
		Description: "The number of instances that should be running in the group.",
		Type:        schema.TypeInt,
		Computed:    true,
	}

	awsASGSchema["min_instance_count"] = &schema.Schema{
		Description: "The minimum size of the Auto Scaling Group.",
		Type:        schema.TypeInt,
		Computed:    true,
	}

	awsASGSchema["max_instance_count"] = &schema.Schema{
		Description: "The maximum size of the Auto Scaling Group.",
		Type:        schema.TypeInt,
		Computed:    true,
	}

	//TODO "wait_for_capacity_timeout" to be implemented

	return awsASGSchema
}

func resourceAwsASG() *schema.Resource {

	return &schema.Resource{
		Description: "`duplocloud_asg_profile` manages a ASG Profile in Duplo.",

		ReadContext:   resourceAwsASGRead,
		CreateContext: resourceAwsASGCreate,
		DeleteContext: resourceAwsASGDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: autosalingGroupSchema(),
	}
}

func resourceAwsASGCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	// Build a request.
	rq := expandAsgProfile(d)
	log.Printf("[TRACE] resourceAwsASGCreate(%s, %s): start", rq.TenantID, rq.FriendlyName)

	// Create the ASG Prfoile in Duplo.
	c := m.(*duplosdk.Client)
	rp, err := c.AsgProfileCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating ASG profile '%s': %s", rq.FriendlyName, err)
	}
	if rp.FriendlyName == "" {
		return diag.Errorf("Error creating ASG profile '%s': no friendly name was received", rq.FriendlyName)
	}

	// Wait up to 60 seconds for Duplo to be able to return the ASG details.
	id := fmt.Sprintf("%s/%s", rq.TenantID, rq.FriendlyName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "AWS host", id, func() (interface{}, duplosdk.ClientError) {
		return c.AsgProfileGet(rq.TenantID, rp.FriendlyName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// TODO- Implmenet API for wait for ASG Capacity
	// By default, wait until the host is completely ready.
	// if v, ok := d.GetOk("wait_until_connected"); !ok || v == nil || v.(bool) {
	// 	err = nativeHostWaitUntilReady(ctx, c, rp.TenantID, rp.FriendlyName, d.Timeout("create"))
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// }

	// Read the host from the backend again.
	//diags = resourceAwsHostRead(ctx, d, m)
	//log.Printf("[TRACE] resourceAwsHostCreate(%s, %s): end", rq.TenantID, rq.FriendlyName)
	return diags
}

func resourceAwsASGRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceAwsASGRead(%s): start", id)
	tenantID, friendlyName, err := asgProfileIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	profile, err := c.AsgProfileGet(tenantID, friendlyName)
	if err != nil {

		// backend may return a 400 instead of a 404
		exists, err2 := c.AsgProfileExists(tenantID, friendlyName)
		if exists || err2 != nil {
			return diag.Errorf("Unable to retrieve ASG profile '%s': %s", id, err)
		}
	}
	if profile == nil {
		d.SetId("") // object missing
		return nil
	}

	// Apply the data
	asgProfileToState(d, profile)

	log.Printf("[TRACE] resourceAwsASGRead(%s): end", id)
	return nil
}

func resourceAwsASGDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceAwsASGDelete(%s): start", id)
	tenantID, friendlyName, err := asgProfileIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if the ASG Profile exists
	c := m.(*duplosdk.Client)
	exists, err := c.AsgProfileExists(tenantID, friendlyName)
	if err != nil {
		return diag.FromErr(err)
	}
	if exists {

		// Delete the ASG Profile from Duplo
		err = c.AsgProfileDelete(tenantID, friendlyName)
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for the ASG profile to be missing
		diags = waitForResourceToBeMissingAfterDelete(ctx, d, "ASG Profile", id, func() (interface{}, duplosdk.ClientError) {
			if rp, err := c.AsgProfileExists(tenantID, friendlyName); rp || err != nil {
				return rp, err
			}
			return nil, nil
		})
	}

	log.Printf("[TRACE] resourceAwsASGDelete(%s): end", id)
	return diags
}

func asgProfileIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return tenantID, name, err
}

func asgProfileToState(d *schema.ResourceData, duplo *duplosdk.DuploAsgProfile) {
	d.Set("instance_count", duplo.DesiredCapacity)
	d.Set("min_instance_count", duplo.MinSize)
	d.Set("max_instance_count", duplo.MaxSize)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("friendly_name", duplo.FriendlyName)
	d.Set("capacity", duplo.Capacity)
	d.Set("is_minion", duplo.IsMinion)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("is_ebs_optimized", duplo.IsEbsOptimized)
	d.Set("cloud", duplo.Cloud)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("minion_tags", keyValueToState("minion_tags", duplo.MinionTags))

	// If a network interface was customized, certain fields are not returned by the backend.
	if v, ok := d.GetOk("network_interfaces"); !ok || v == nil || len(v.([]interface{})) == 0 {
		d.Set("zone", duplo.Zone)
		d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	}

	d.Set("metadata", keyValueToState("metadata", duplo.MetaData))
	d.Set("volumes", flattenNativeHostVolumes(duplo.Volumes))
	d.Set("network_interfaces", flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces))
}

func expandAsgProfile(d *schema.ResourceData) *duplosdk.DuploAsgProfile {
	return &duplosdk.DuploAsgProfile{
		TenantID:          d.Get("tenant_id").(string),
		AccountName:       d.Get("user_account").(string),
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
		MetaData:          keyValueFromState("metadata", d),
		Tags:              keyValueFromState("tag", d),
		MinionTags:        keyValueFromState("minion_tags", d),
		Volumes:           expandNativeHostVolumes("volume", d),
		NetworkInterfaces: expandNativeHostNetworkInterfaces("network_interface", d),
		DesiredCapacity:   d.Get("instance_count").(int),
		MinSize:           d.Get("min_instance_count").(int),
		MaxSize:           d.Get("max_instance_count").(int),
	}
}
