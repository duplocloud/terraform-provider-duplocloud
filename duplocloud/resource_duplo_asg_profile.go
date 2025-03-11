package duplocloud

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func autoscalingGroupSchema() map[string]*schema.Schema {

	awsASGSchema := nativeHostSchema()
	delete(awsASGSchema, "instance_id")
	delete(awsASGSchema, "status")
	delete(awsASGSchema, "identity_role")
	delete(awsASGSchema, "private_ip_address")
	delete(awsASGSchema, "zone")

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

	awsASGSchema["use_spot_instances"] = &schema.Schema{
		Description: "Whether or not to use spot instances.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Default:     false,
	}

	awsASGSchema["max_spot_price"] = &schema.Schema{
		Description: "Maximum price to pay for a spot instance in dollars per unit hour.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
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
		ForceNew:    true,
	}

	awsASGSchema["can_scale_from_zero"] = &schema.Schema{
		Description: "Whether or not ASG should leverage duplocloud's scale from 0 feature",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
	}

	awsASGSchema["fullname"] = &schema.Schema{
		Description: "The full name of the ASG profile.",
		Type:        schema.TypeString,
		Computed:    true,
		ForceNew:    true,
	}

	awsASGSchema["enabled_metrics"] = &schema.Schema{
		Description: "List of metrics to collect for the ASG Specify one or more of the following metrics.`GroupMinSize`,`GroupMaxSize`,`GroupDesiredCapacity`,`GroupInServiceInstances`,`GroupPendingInstances`,`GroupStandbyInstances`,`GroupTerminatingInstances`,`GroupTotalInstances`,`GroupInServiceCapacity`,`GroupPendingCapacity`,`GroupStandbyCapacity`,`GroupTerminatingCapacity`,`GroupTotalCapacity`,`WarmPoolDesiredCapacity`,`WarmPoolWarmedCapacity`,`WarmPoolPendingCapacity`,`WarmPoolTerminatingCapacity`,`WarmPoolTotalCapacity`,`GroupAndWarmPoolDesiredCapacity`,`GroupAndWarmPoolTotalCapacity`.",
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity", "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances", "GroupInServiceCapacity", "GroupPendingCapacity", "GroupStandbyCapacity", "GroupTerminatingCapacity", "GroupTotalCapacity", "WarmPoolDesiredCapacity", "WarmPoolWarmedCapacity", "WarmPoolPendingCapacity", "WarmPoolTerminatingCapacity", "WarmPoolTotalCapacity", "GroupAndWarmPoolDesiredCapacity", "GroupAndWarmPoolTotalCapacity"}, true),
			ForceNew:     true,
		},
	}
	awsASGSchema["zones"] = &schema.Schema{
		Description: "The multi availability zone to launch the asg in, expressed as a number and starting at 0 - Zone A 1 - Zone B. For Automatic do not specify zones and zone field",
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type:     schema.TypeInt,
			ForceNew: true,
		},
		ConflictsWith:    []string{"zone"},
		DiffSuppressFunc: diffSuppressAsgZones,
		Computed:         true,
	}
	awsASGSchema["zone"] = &schema.Schema{

		Description:   "The availability zone to launch the host in, expressed as a numeric value and starting at 0 to 5. Recommended for environment on july release",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true, // relaunch instance
		Deprecated:    "zone has been deprecated instead use zones",
		ConflictsWith: []string{"zones"},
		ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[012345]{1}$`), "allowed values 0 - 5"),
		//DiffSuppressFunc: diffSuppressAsgZone,
	}

	return awsASGSchema
}

func validateMaxSpotPrice(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	useSpotInstances := diff.Get("use_spot_instances").(bool)
	maxSpotPrice := diff.Get("max_spot_price").(string)
	if maxSpotPrice != "" {
		if useSpotInstances {
			if _, err := strconv.ParseFloat(maxSpotPrice, 64); err != nil {
				return fmt.Errorf("'max_spot_price' must be a string representing a decimal number")
			}
		} else {
			return fmt.Errorf("'use_spot_instances' must be true when 'max_spot_price' is non-empty")
		}
	}

	return nil
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
		Schema: autoscalingGroupSchema(),
		CustomizeDiff: customdiff.All(
			diffUserData,
			validateMaxSpotPrice,
			validateTaintsSupport,
		),
	}
}

func resourceAwsASGCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	initUserDataOptions(d)

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
	rq := expandAsgProfile(d)
	if needsUpdate {
		// Build a request.
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

	needsAllocationTagsUpdate, newTags := checkAllocationTagsDiff(d)
	if needsAllocationTagsUpdate {
		updateRequest := duplosdk.CustomDataUpdate{
			ComponentId:   d.Get("fullname").(string),
			ComponentType: duplosdk.ASG,
			Key:           "AllocationTags",
			Value:         newTags,
		}
		err := c.TenantUpdateCustomData(rq.TenantId, updateRequest)
		if err != nil {
			return diag.Errorf("Error updating ASG profile '%s': %s", rq.FriendlyName, err)
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
	tenantID, _, err := asgProfileIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	friendlyName := d.Get("fullname").(string)
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
	d.Set("use_spot_instances", duplo.UseSpotInstances)
	d.Set("max_spot_price", duplo.MaxSpotPrice)
	d.Set("fullname", duplo.FriendlyName)
	d.Set("capacity", duplo.Capacity)
	d.Set("is_minion", duplo.IsMinion)
	d.Set("image_id", duplo.ImageID)
	d.Set("base64_user_data", duplo.Base64UserData)
	d.Set("prepend_user_data", duplo.PrependUserData)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("is_ebs_optimized", duplo.IsEbsOptimized)
	d.Set("is_cluster_autoscaled", duplo.IsClusterAutoscaled)
	d.Set("can_scale_from_zero", duplo.CanScaleFromZero)
	d.Set("cloud", duplo.Cloud)
	d.Set("keypair_type", duplo.KeyPairType)
	d.Set("encrypt_disk", duplo.EncryptDisk)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("minion_tags", keyValueToState("minion_tags", duplo.CustomDataTags))
	d.Set("enabled_metrics", duplo.EnabledMetrics)

	// If a network interface was customized, certain fields are not returned by the backend.
	if v, ok := d.GetOk("network_interface"); !ok || v == nil || len(v.([]interface{})) == 0 {
		_, zok := d.GetOk("zone")

		if len(duplo.Zones) > 0 && !zok {
			i := []interface{}{}
			for _, v := range duplo.Zones {
				i = append(i, v)
			}
			d.Set("zones", i)
		} else {
			if len(duplo.Zones) == 1 {
				d.Set("zone", duplo.Zones[0])
			}
		}
		d.Set("allocated_public_ip", duplo.AllocatedPublicIP)
	}

	d.Set("metadata", keyValueToState("metadata", duplo.MetaData))
	d.Set("volume", flattenNativeHostVolumes(duplo.Volumes))
	d.Set("network_interface", flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces))
	if duplo.Taints != nil {
		d.Set("taints", flattenTaints(*duplo.Taints))
	}
}

func expandAsgProfile(d *schema.ResourceData) *duplosdk.DuploAsgProfile {
	asgProfile := &duplosdk.DuploAsgProfile{
		TenantId:            d.Get("tenant_id").(string),
		AccountName:         d.Get("user_account").(string),
		FriendlyName:        d.Get("friendly_name").(string),
		Capacity:            d.Get("capacity").(string),
		IsMinion:            d.Get("is_minion").(bool),
		ImageID:             d.Get("image_id").(string),
		Base64UserData:      d.Get("base64_user_data").(string),
		PrependUserData:     d.Get("prepend_user_data").(bool),
		AgentPlatform:       d.Get("agent_platform").(int),
		IsEbsOptimized:      d.Get("is_ebs_optimized").(bool),
		IsClusterAutoscaled: d.Get("is_cluster_autoscaled").(bool),
		CanScaleFromZero:    d.Get("can_scale_from_zero").(bool),
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
		UseSpotInstances:    d.Get("use_spot_instances").(bool),
		MaxSpotPrice:        d.Get("max_spot_price").(string),
		ExtraNodeLabels:     keyValueFromMap(d.Get("custom_node_labels").(map[string]interface{})),
	}

	if v, ok := d.GetOk("enabled_metrics"); ok && len(v.([]interface{})) > 0 {
		metricList := make([]string, len(v.([]interface{})))
		for i, val := range v.([]interface{}) {
			metricList[i] = val.(string)
		}
		asgProfile.EnabledMetrics = &metricList
	}

	obj := []duplosdk.DuploTaints{}
	if val, ok := d.Get("taints").([]interface{}); ok {
		for _, dt := range val {
			m := dt.(map[string]interface{})
			taints := duplosdk.DuploTaints{
				Key:    m["key"].(string),
				Value:  m["value"].(string),
				Effect: m["effect"].(string),
			}
			obj = append(obj, taints)

		}
		asgProfile.Taints = &obj
	}
	z := []int{}
	if val, ok := d.Get("zones").([]interface{}); ok && len(val) > 0 {
		for _, dt := range val {
			z = append(z, dt.(int))
		}
		asgProfile.Zones = z
	} else if v, ok := d.GetOk("zone"); ok && v != nil {
		zn, _ := strconv.Atoi(d.Get("zone").(string))
		asgProfile.Zones = append(asgProfile.Zones, zn)
	} else {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		z = append(z, r.Intn(2))
		asgProfile.Zones = z
		d.Set("zones", z)

	}

	return asgProfile
}

func asgtWaitUntilCapacityReady(ctx context.Context, c *duplosdk.Client, tenantID string, minInstanceCount int, asgFriendlyName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
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
		d.HasChange("max_instance_count") ||
		d.HasChange("friendly_name") ||
		d.HasChange("enabled_metrics")
}

func checkAllocationTagsDiff(d *schema.ResourceData) (hasChange bool, tags string) {
	oldRevision, newRevision := d.GetChange("minion_tags")
	oldTags := oldRevision.([]interface{})
	newTags := newRevision.([]interface{})

	var oldAllocationTagValue, newAllocationTagValue string

	for _, tag := range oldTags {
		tagMap := tag.(map[string]interface{})
		if tagMap["key"].(string) == "AllocationTags" {
			oldAllocationTagValue = tagMap["value"].(string)
			break
		}
	}

	for _, tag := range newTags {
		tagMap := tag.(map[string]interface{})
		if tagMap["key"].(string) == "AllocationTags" {
			newAllocationTagValue = tagMap["value"].(string)
			break
		}
	}

	if oldAllocationTagValue != newAllocationTagValue {
		return true, newAllocationTagValue
	}

	return false, ""
}
