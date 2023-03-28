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

func autosalingGroupSchema() map[string]*schema.Schema {

	awsASGSchema := nativeHostSchema()
	delete(awsASGSchema, "instance_id")
	delete(awsASGSchema, "status")
	delete(awsASGSchema, "identity_role")
	delete(awsASGSchema, "private_ip_address")

	awsASGSchema["instance_count"] = &schema.Schema{
		Description: "The number of instances that should be running in the group.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	}

	awsASGSchema["min_instance_count"] = &schema.Schema{
		Description: "The minimum size of the Auto Scaling Group.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	}

	awsASGSchema["max_instance_count"] = &schema.Schema{
		Description: "The maximum size of the Auto Scaling Group.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	}

	awsASGSchema["use_launch_template"] = &schema.Schema{
		Description: "Whether or not to use launch template.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	}

	awsASGSchema["wait_for_capacity"] = &schema.Schema{
		Description:      "Whether or not to wait until ASG instances to be healthy, after creation.",
		Type:             schema.TypeBool,
		Optional:         true,
		ForceNew:         true,
		Default:          true,
		DiffSuppressFunc: diffSuppressWhenNotCreating,
	}

	awsASGSchema["is_cluster_autoscaled"] = &schema.Schema{
		Description: "Whether or not to enable cluster autoscaler.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	}

	awsASGSchema["fullname"] = &schema.Schema{
		Description: "The full name of the ASG profile.",
		Type:        schema.TypeString,
		Computed:    true,
	}

	return awsASGSchema
}

func resourceAwsASG() *schema.Resource {

	return &schema.Resource{
		Description: "`duplocloud_asg_profile` manages a ASG Profile in Duplo.",

		ReadContext:   resourceAwsASGRead,
		CreateContext: resourceAwsASGCreate,
		DeleteContext: resourceAwsASGDelete,
		UpdateContext: resourceAwsASGUpdate,
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
	log.Printf("[TRACE] resourceAwsASGCreate(%s, %s): start", rq.TenantId, rq.FriendlyName)

	// Create the ASG Prfoile in Duplo.
	c := m.(*duplosdk.Client)
	rp, err := c.AsgProfileCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating ASG profile '%s': %s", rq.FriendlyName, err)
	}
	if rp == "" {
		return diag.Errorf("Error creating ASG profile '%s': no friendly name was received", rq.FriendlyName)
	}

	// Update minion tags once ASG is created
	fullName, _ := c.GetDuploServicesName(rq.TenantId, rq.FriendlyName)

	for _, raw := range *rq.MinionTags {
		err = c.TenantUpdateCustomData(rq.TenantId, duplosdk.CustomDataUpdate{
			ComponentId:   fullName,
			ComponentType: duplosdk.ASG,
			Key:           raw.Key,
			Value:         raw.Value,
		})

		if err != nil {
			return diag.Errorf("Error updating custom data using minion tags '%s': %s", fullName, err)
		}
	}

	id := fmt.Sprintf("%s/%s", rq.TenantId, rp)
	log.Printf("[DEBUG] ASG Profile Resource ID- (%s)", id)

	//Wait up to 60 seconds for Duplo to be able to return the ASG details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "ASG Profile", id, func() (interface{}, duplosdk.ClientError) {
		return c.AsgProfileGet(rq.TenantId, rp)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)
	//By default, wait until the ASG instances to be healthy.
	if d.Get("wait_for_capacity") == nil || d.Get("wait_for_capacity").(bool) {
		err = asgtWaitUntilCapacityReady(ctx, c, rq.TenantId, rq.MinSize, rp, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAwsASGRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsASGCreate(%s, %s): end", rq.TenantId, rq.FriendlyName)
	return diags
}

func resourceAwsASGUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
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

	if !exists {
		d.SetId("")
		return diag.Errorf("ASG profile does not exist with name '%s': %s", friendlyName, err)
	}

	needsUpdate := needsResourceAwsASGUpdate(d)
	if needsUpdate {
		// Build a request.
		rq := expandAsgProfile(d)
		log.Printf("[TRACE] resourceAwsASGUpdate(%s, %s): start", rq.TenantId, rq.FriendlyName)

		// Update the ASG Prfoile in Duplo.
		rp, err := c.AsgProfileUpdate(rq)
		if err != nil {
			return diag.Errorf("Error updating ASG profile '%s': %s", rq.FriendlyName, err)
		}
		if rp == "" {
			return diag.Errorf("Error updating ASG profile '%s': no friendly name was received", rq.FriendlyName)
		}

		//Wait up to 60 seconds for Duplo to be able to return the ASG details.
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "ASG Profile", id, func() (interface{}, duplosdk.ClientError) {
			return c.AsgProfileGet(rq.TenantId, rp)
		})
		if diags != nil {
			return diags
		}

		//By default, wait until the ASG instances to be healthy.
		if d.Get("wait_for_capacity") == nil || d.Get("wait_for_capacity").(bool) {
			err := asgtWaitUntilCapacityReady(ctx, c, rq.TenantId, rq.MinSize, rp, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diags := resourceAwsASGRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsASGUpdate(%s, %s): end", tenantID, friendlyName)
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
	d.Set("tenant_id", tenantID)
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

		time.Sleep(time.Duration(180) * time.Second)
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
	d.Set("use_launch_template", duplo.UseLaunchTemplate)
	d.Set("fullname", duplo.FriendlyName)
	d.Set("capacity", duplo.Capacity)
	d.Set("is_minion", duplo.IsMinion)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("is_ebs_optimized", duplo.IsEbsOptimized)
	d.Set("is_cluster_autoscaled", duplo.IsClusterAutoscaled)
	d.Set("cloud", duplo.Cloud)
	d.Set("keypair_type", duplo.KeyPairType)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("minion_tags", keyValueToState("minion_tags", duplo.MinionTags))

	// If a network interface was customized, certain fields are not returned by the backend.
	if v, ok := d.GetOk("network_interface"); !ok || v == nil || len(v.([]interface{})) == 0 {
		d.Set("zone", duplo.Zone)
		d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	}

	d.Set("metadata", keyValueToState("metadata", duplo.MetaData))
	d.Set("volume", flattenNativeHostVolumes(duplo.Volumes))
	d.Set("network_interface", flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces))
}

func expandAsgProfile(d *schema.ResourceData) *duplosdk.DuploAsgProfile {
	return &duplosdk.DuploAsgProfile{
		TenantId:            d.Get("tenant_id").(string),
		AccountName:         d.Get("user_account").(string),
		FriendlyName:        d.Get("friendly_name").(string),
		Capacity:            d.Get("capacity").(string),
		Zone:                d.Get("zone").(int),
		IsMinion:            d.Get("is_minion").(bool),
		ImageID:             d.Get("image_id").(string),
		Base64UserData:      d.Get("base64_user_data").(string),
		AgentPlatform:       d.Get("agent_platform").(int),
		IsEbsOptimized:      d.Get("is_ebs_optimized").(bool),
		IsClusterAutoscaled: d.Get("is_cluster_autoscaled").(bool),
		AllocatedPublicIP:   d.Get("allocated_public_ip").(bool),
		Cloud:               d.Get("cloud").(int),
		KeyPairType:         d.Get("keypair_type").(int),
		EncryptDisk:         d.Get("encrypt_disk").(bool),
		MetaData:            keyValueFromState("metadata", d),
		Tags:                keyValueFromState("tags", d),
		MinionTags:          keyValueFromState("minion_tags", d),
		Volumes:             expandNativeHostVolumes("volume", d),
		NetworkInterfaces:   expandNativeHostNetworkInterfaces("network_interface", d),
		DesiredCapacity:     d.Get("instance_count").(int),
		MinSize:             d.Get("min_instance_count").(int),
		MaxSize:             d.Get("max_instance_count").(int),
		UseLaunchTemplate:   d.Get("use_launch_template").(bool),
	}
}

func asgtWaitUntilCapacityReady(ctx context.Context, c *duplosdk.Client, tenantID string, minInstanceCount int, asgFriendlyName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			asgTagKey := []string{"aws:autoscaling:groupName"}
			log.Printf("[DEBUG] Fetching native hosts for Tenant-(%s)", tenantID)
			rp, err := c.NativeHostGetList(tenantID)
			status := "pending"
			count := 0
			if err == nil {
				for _, host := range *rp {
					log.Printf("[DEBUG] Duplo Host : %s", host.InstanceID)
					asgTag := selectKeyValues(host.Tags, asgTagKey)
					if asgTag != nil {
						asgHost := false
						for _, tag := range *asgTag {
							log.Printf("[DEBUG] Asg Host Tag : (%s)-(%s)-(%s).", host.InstanceID, tag.Key, tag.Value)
							if tag.Value == asgFriendlyName {
								asgHost = true
							}
						}
						if asgHost {
							count++
							log.Printf("[DEBUG] ASG profile found, Name-(%s)", asgFriendlyName)
							log.Printf("[DEBUG] Instance %s is in %s state.", host.InstanceID, host.Status)
							if minInstanceCount == 0 || host.Status == "running" {
								status = "ready"
							} else {
								status = "pending"
								break
							}
						}
					}
				}
				log.Printf("[DEBUG] Count, MinCount (%v-%v)", count, minInstanceCount)
				if minInstanceCount == 0 || (status == "ready" && count >= minInstanceCount) {
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
	log.Printf("[DEBUG] asgtWaitUntilCapacityReady(%s, %s)", tenantID, asgFriendlyName)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func needsResourceAwsASGUpdate(d *schema.ResourceData) bool {
	return d.HasChange("instance_count") ||
		d.HasChange("min_instance_count") ||
		d.HasChange("max_instance_count")
}
