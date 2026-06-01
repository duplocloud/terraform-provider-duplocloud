package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		Description: "The multi availability zone to launch the asg in, expressed as a number and starting at 0 - Zone A to 3 - Zone D, based on the infra setup",
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeInt,
			ForceNew:     true,
			ValidateFunc: validation.IntBetween(0, 3),
		},
		DiffSuppressFunc: diffSuppressAsgZones,
		Computed:         true,
	}
	awsASGSchema["zone"] = &schema.Schema{

		Description: "The availability zone to launch the host in is expressed as a numeric value ranging from 0 to 3. ",
		Type:        schema.TypeString,
		Optional:    true,
		//ForceNew:         true, // relaunch instance
		Deprecated:       "For environments on the July 2024 release or earlier, use zone. For environments on releases after July 2024, use zones, as zone has been deprecated and is non-functional on change.",
		Default:          0,
		DiffSuppressFunc: diffSuppressWhenNotCreating,
	}
	awsASGSchema["taints"] = &schema.Schema{
		Description: "Specify taints to attach to the nodes, to repel other nodes with different toleration",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    50,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}$|^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?/(.{1,253})$`), "Invalid key format: taint key must begin with a letter or number, can contain letters, numbers, hyphens(-), and periods(.), and be up to 63 characters long OR the taint key begins with a valid DNS subdomain prefix, followed by a single /, and includes a key of up to 253 characters"),
				},
				"value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,62}$`), "Invalid value format: taint value must begin with a letter or number, can contain letters, numbers, hyphens(-), and be up to 63 characters long"),
				},
				"effect": {
					Description: "Update strategy of the node. Effect types <br>      - NoSchedule<br>     - PreferNoSchedule<br>     - NoExecute",
					Type:        schema.TypeString,
					Required:    true,
					ValidateFunc: validation.StringInSlice([]string{
						"NoSchedule",
						"PreferNoSchedule",
						"NoExecute",
					}, false),
				},
			},
		},
	}

	awsASGSchema["arn"] = &schema.Schema{
		Description: "The ASG arn.",
		Type:        schema.TypeString,
		Computed:    true,
	}
	awsASGSchema["mixed_instances_policy"] = &schema.Schema{
		Description: "Configuration block for a mixed instances policy. This allows an ASG to use multiple instance types and a mix of On-Demand and Spot instances.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"launch_template": {
					Description: "Launch template configuration with instance type overrides.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"override": {
								Description: "List of instance type overrides for the launch template.",
								Type:        schema.TypeList,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"instance_type": {
											Description: "The instance type. Mutually exclusive with `instance_requirements`.",
											Type:        schema.TypeString,
											Optional:    true,
										},
										"weighted_capacity": {
											Description: "The number of capacity units provided by the instance type.",
											Type:        schema.TypeString,
											Optional:    true,
										},
										"instance_requirements": {
											Description: "Instance requirements for flexible instance selection.",
											Type:        schema.TypeList,
											Optional:    true,
											MaxItems:    1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"vcpu_count": {
														Description: "Range of vCPU counts.",
														Type:        schema.TypeList,
														Required:    true,
														MaxItems:    1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"min": {
																	Description: "Minimum number of vCPUs.",
																	Type:        schema.TypeInt,
																	Required:    true,
																},
																"max": {
																	Description: "Maximum number of vCPUs.",
																	Type:        schema.TypeInt,
																	Optional:    true,
																},
															},
														},
													},
													"memory_mib": {
														Description: "Range of memory in MiB.",
														Type:        schema.TypeList,
														Required:    true,
														MaxItems:    1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"min": {
																	Description: "Minimum memory in MiB.",
																	Type:        schema.TypeInt,
																	Required:    true,
																},
																"max": {
																	Description: "Maximum memory in MiB.",
																	Type:        schema.TypeInt,
																	Optional:    true,
																},
															},
														},
													},
													"allowed_instance_types": {
														Description: "List of allowed instance types (e.g. `m5.large`, `m5a.large`). Cannot be specified if `excluded_instance_types` is specified in the same launch template override.",
														Type:        schema.TypeList,
														Optional:    true,
														Computed:    true,
														Elem:        &schema.Schema{Type: schema.TypeString},
													},
													"excluded_instance_types": {
														Description: "List of excluded instance types. Cannot be specified if `allowed_instance_types` is specified in the same launch template override.",
														Type:        schema.TypeList,
														Optional:    true,
														Computed:    true,
														Elem:        &schema.Schema{Type: schema.TypeString},
													},
													"cpu_manufacturers": {
														Description: "List of CPU manufacturers (e.g. `intel`, `amd`, `amazon-web-services`).",
														Type:        schema.TypeList,
														Optional:    true,
														Computed:    true,

														Elem: &schema.Schema{Type: schema.TypeString},
													},
													"instance_generations": {
														Description: "List of instance generations (e.g. `current`, `previous`).",
														Type:        schema.TypeList,
														Optional:    true,
														Computed:    true,

														Elem: &schema.Schema{Type: schema.TypeString},
													},
													"spot_max_price_percentage_over_lowest_price": {
														Description: "Price protection threshold as a percentage over the lowest price.",
														Type:        schema.TypeInt,
														Optional:    true,
														Computed:    true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"instances_distribution": {
					Description: "Configuration for the distribution of On-Demand and Spot instances.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"on_demand_allocation_strategy": {
								Description: "Strategy for allocating On-Demand instances (e.g. `prioritized`).",
								Type:        schema.TypeString,
								Optional:    true,
								Computed:    true,
							},
							"on_demand_base_capacity": {
								Description: "Minimum number of On-Demand instances in the group.",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"on_demand_percentage_above_base_capacity": {
								Description: "Percentage of On-Demand instances above the base capacity (0-100).",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"spot_allocation_strategy": {
								Description: "Strategy for allocating Spot instances (e.g. `capacity-optimized`, `price-capacity-optimized`, `lowest-price`).",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"spot_instance_pools": {
								Description: "Number of Spot pools for allocation (only used with `lowest-price` strategy).",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"spot_max_price": {
								Description: "Maximum price per unit hour for Spot instances.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}

	awsASGSchema["capacity"].ForceNew = false
	awsASGSchema["image_id"].ForceNew = false
	awsASGSchema["capacity"].Required = false
	awsASGSchema["capacity"].Optional = true

	awsASGSchema["capacity"].DiffSuppressFunc = diffSuppressWhenNotCreating
	awsASGSchema["image_id"].DiffSuppressFunc = diffSuppressWhenNotCreating
	awsASGSchema["volume"].DiffSuppressFunc = diffSuppressWhenNotCreating
	awsASGSchema["minion_tags"].ConflictsWith = []string{"custom_data_tags"}
	awsASGSchema["minion_tags"].Deprecated = "minion_tags is deprecated and will be removed in a future release. Use custom_data_tags instead."
	awsASGSchema["minion_tags"].DiffSuppressFunc = diffSuppressWhenNotCreating

	awsASGSchema["custom_data_tags"] = &schema.Schema{
		Description:   "A map of tags to assign to the resource. Example - `AllocationTags` can be passed as tag key with any value.\n\n**Note:** When importing an ASG created using the minion_tags block, from v0.12.6 onwards, need to add a custom_data_tags block by replacing the minion_tags block with the same key and value as the minion_tags block to avoid drift.",
		Type:          schema.TypeList,
		Optional:      true,
		Computed:      true,
		Elem:          KeyValueSchema(),
		ConflictsWith: []string{"minion_tags"},
	}

	awsASGSchema["asg_tags"] = &schema.Schema{
		Description: "A map of arbitrary AWS tags applied to the ASG and its launched EC2 instances (routed via the backend's `TagsCsv` field). Use this for tags that aren't `AllocationTags` — those belong in `custom_data_tags`. Changes force replacement because the backend applies these tags only at create time.",
		Type:        schema.TypeMap,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
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

func validateCustomDataTags(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	if customDataTags, ok := diff.GetOk("custom_data_tags"); ok {
		tagsList := customDataTags.([]interface{})
		for _, tag := range tagsList {
			tagMap := tag.(map[string]interface{})
			key := tagMap["key"].(string)
			value := tagMap["value"].(string)

			if value == "" {
				return fmt.Errorf("custom_data_tags: tag value cannot be empty for key '%s'. To remove a tag, remove the entire key-value pair from the configuration", key)
			}
		}
	}
	return nil
}

func resourceAwsASG() *schema.Resource {

	return &schema.Resource{
		Description:   "`duplocloud_asg_profile` manages a ASG Profile in Duplo.\n\n**Note:** When updating a duplocloud_asg_profile resource in versions later than 0.10.53 use duplocloud_aws_launch_template which creates a new version of launch template. To set the new version as default version use duplocloud_aws_launch_template_default_version resource, this will avoid recreation of ASG profile. \nTo refresh the ASG with new launch template version use duplocloud_asg_instance_refresh resource.",
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
			validateCustomDataTags,
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
	prefix, err := c.GetDuploServicesPrefix(rq.TenantId, "")
	if err != nil {
		return diag.Errorf("Tenant details : %s", err)
	}
	list, err := c.AsgProfileGetList(rq.TenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, profile := range *list {
		if profile.FriendlyName == prefix+"-"+rq.FriendlyName {
			return diag.Errorf("ASG '%s' already exists.\n If you are using `create_before_destroy` to update ASG, use a duplocloud_aws_launch_template to create a new version and refresh the ASG with the duplocloud_asg_instance_refresh resource.\n To change the default version 'duplocloud_aws_launch_template_default_version' should be used", rq.FriendlyName)
		}
	}

	rp, err := c.AsgProfileCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating ASG profile '%s': %s", rq.FriendlyName, err)
	}
	if rp == "" {
		return diag.Errorf("Error creating ASG profile '%s': no friendly name was received", rq.FriendlyName)
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
	werr, executed := asgWaitUntilReady(ctx, c, rq.TenantId, rp, d.Timeout("create"))
	if werr != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ASG profile '%s' to be ready: %s", rp, werr))
	}
	if !executed {
		time.Sleep(5 * time.Minute) // Wait to ensure the ASG profile is created in Duplo if polling dint happened
		// to reduce impact of asg worker failure when tags are passed.
	}
	if v, ok := d.GetOk("minion_tags"); ok && len(v.([]interface{})) > 0 {
		fullName, _ := c.GetDuploServicesName(rq.TenantId, rq.FriendlyName)
		mTags := v.([]interface{})
		for _, raw := range mTags {
			m := raw.(map[string]interface{})
			err = c.TenantUpdateCustomData(rq.TenantId, duplosdk.CustomDataUpdate{
				ComponentId:   fullName,
				ComponentType: duplosdk.ASG,
				Key:           m["key"].(string),
				Value:         m["value"].(string),
			})

			if err != nil {
				return diag.Errorf("Error updating custom data using minion tags '%s': %s", fullName, err)
			}
		}
	}
	//By default, wait until the ASG instances to be healthy.
	if d.Get("wait_for_capacity") == nil || d.Get("wait_for_capacity").(bool) {
		err = asgtWaitUntilCapacityReady(ctx, c, rq.TenantId, rq.MinSize, rp, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	ctx = context.WithValue(ctx, flowContextKey, "normal")
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
		rq.FriendlyName = friendlyName
		log.Printf("[TRACE] resourceAwsASGUpdate(%s, %s): start", rq.TenantId, rq.FriendlyName)

		// Update the ASG Prfoile in Duplo.
		cerr := c.AsgProfileUpdate(rq)
		if cerr != nil {
			return diag.Errorf("Error updating ASG profile '%s': %s", rq.FriendlyName, err)
		}

		//Wait up to 60 seconds for Duplo to be able to return the ASG details.
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "ASG Profile", id, func() (interface{}, duplosdk.ClientError) {
			return c.AsgProfileGet(rq.TenantId, friendlyName)
		})
		if diags != nil {
			return diags
		}
		werr, _ := asgWaitUntilReady(ctx, c, rq.TenantId, friendlyName, d.Timeout("create"))
		if werr != nil {
			return diag.FromErr(fmt.Errorf("error waiting for ASG profile '%s' to be ready: %s", friendlyName, werr))
		}
		//By default, wait until the ASG instances to be healthy.
		if d.Get("wait_for_capacity") == nil || d.Get("wait_for_capacity").(bool) {
			err := asgtWaitUntilCapacityReady(ctx, c, rq.TenantId, rq.MinSize, friendlyName, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	customDataUpdates := getCustomDataTagsUpdates(d)
	for _, updateRequest := range customDataUpdates {
		updateRequest.ComponentId = d.Get("fullname").(string)
		updateRequest.ComponentType = duplosdk.ASG

		err := c.TenantUpdateCustomData(rq.TenantId, updateRequest)
		if err != nil {
			return diag.Errorf("Error updating ASG profile custom data '%s': %s", rq.FriendlyName, err)
		}
	}
	if d.HasChange("taints") {
		list, err := c.TenantListMinions(tenantID)
		if err != nil {
			return diag.Errorf("Error on updating taints from ASG profile '%s': %s", friendlyName, err)
		}
		privateAddress := ""
		for _, minion := range *list {
			if minion.AsgName == friendlyName {
				privateAddress = minion.PrivateIpAddress
				break
			}
		}
		k := []string{}

		obj := []duplosdk.DuploTaints{}
		if val, ok := d.Get("taints").([]interface{}); ok {
			for _, dt := range val {

				m := dt.(map[string]interface{})
				taints := duplosdk.DuploTaints{
					Key:    m["key"].(string),
					Value:  m["value"].(string),
					Effect: m["effect"].(string),
				}
				k = append(k, m["key"].(string))
				obj = append(obj, taints)

			}

		}
		cerr := c.DeleteASGTaints(tenantID, privateAddress, k)
		if cerr != nil {
			return diag.Errorf("Error updating taints from ASG profile '%s': %s", friendlyName, cerr)
		}
		time.Sleep(10 * time.Second)
		cerr = c.UpdateASGTaints(tenantID, privateAddress, obj)
		if cerr != nil {
			return diag.Errorf("Error updating taints from ASG profile '%s': %s", friendlyName, cerr)
		}
	}
	ctx = context.WithValue(ctx, flowContextKey, "normal")
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
	profile, cerr := c.AsgProfileGet(tenantID, friendlyName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsASGRead(%s): ASG profile not found", id)
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve ASG profile '%s': %s", id, cerr)
	}
	if profile == nil {
		log.Printf("[TRACE] resourceAwsASGRead(%s): ASG profile not found", id)
		d.SetId("") // object missing
		return nil
	}

	flow := "import"

	v := ctx.Value(flowContextKey)
	if v != nil {
		if f, ok := v.(string); ok {
			flow = f
		}
	}
	// Apply the data
	asgProfileToState(d, profile, flow)
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
	exists, cerr := c.AsgProfileExists(tenantID, friendlyName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsASGDelete(%s): ASG profile not found", id)
			return nil
		}
		return diag.FromErr(err)
	}
	if exists {
		// Delete the ASG Profile from Duplo
		cerr := c.AsgProfileDelete(tenantID, friendlyName)
		if cerr != nil {
			if cerr.Status() == 404 {
				log.Printf("[TRACE] resourceAwsASGDelete(%s): ASG profile not found", id)
				return nil
			}
			return diag.FromErr(cerr)
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

func asgProfileToState(d *schema.ResourceData, duplo *duplosdk.DuploAsgProfile, flow string) {
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
	d.Set("minion_tags", keyValueToState("minion_tags", duplo.MinionTags))
	d.Set("custom_data_tags", keyValueToState("custom_data_tags", duplo.CustomDataTags))
	parsedAsgTags := map[string]string{}
	if duplo.TagsCsv != "" {
		if err := json.Unmarshal([]byte(duplo.TagsCsv), &parsedAsgTags); err != nil {
			log.Printf("[WARN] Failed to unmarshal TagsCsv for asg_tags in ASG profile state: %v", err)
			parsedAsgTags = map[string]string{}
		}
	}
	d.Set("asg_tags", parsedAsgTags)
	d.Set("enabled_metrics", duplo.EnabledMetrics)
	d.Set("arn", duplo.Arn)
	// If a network interface was customized, certain fields are not returned by the backend.
	if v, ok := d.GetOk("network_interface"); !ok || v == nil || len(v.([]interface{})) == 0 {
		_, zok := d.GetOk("zones")
		if len(duplo.Zones) > 0 && flow == "import" {
			i := []interface{}{}
			for _, v := range duplo.Zones {
				i = append(i, v)
			}
			d.Set("zones", i)
		} else if len(duplo.Zones) > 0 && zok {
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
	if duplo.MixedInstancesPolicy != nil {
		d.Set("mixed_instances_policy", flattenAsgMixedInstancesPolicy(duplo.MixedInstancesPolicy))
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
		CustomDataTags:      keyValueFromState("custom_data_tags", d),
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
	}

	asgProfile.MixedInstancesPolicy = expandAsgMixedInstancesPolicy(d)

	// asg_tags is routed through the backend's TagsCsv field (JSON-encoded),
	// which is the only path that produces real AWS tags on the ASG and
	// launched EC2 instances.
	if v, ok := d.GetOk("asg_tags"); ok {
		raw := v.(map[string]interface{})
		if len(raw) > 0 {
			stringMap := make(map[string]string, len(raw))
			for k, val := range raw {
				stringMap[k] = val.(string)
			}
			if jsonBytes, err := json.Marshal(stringMap); err == nil {
				asgProfile.TagsCsv = string(jsonBytes)
			}
		}
	}

	return asgProfile
}

func expandAsgMixedInstancesPolicy(d *schema.ResourceData) *duplosdk.DuploAsgMixedInstancesPolicy {
	v, ok := d.GetOk("mixed_instances_policy")
	if !ok || len(v.([]interface{})) == 0 {
		return nil
	}
	policyMap := v.([]interface{})[0].(map[string]interface{})
	policy := &duplosdk.DuploAsgMixedInstancesPolicy{}

	if lt, ok := policyMap["launch_template"]; ok && len(lt.([]interface{})) > 0 {
		ltMap := lt.([]interface{})[0].(map[string]interface{})
		launchTemplate := &duplosdk.DuploAsgMixedInstancesLaunchTemplate{}

		if overrides, ok := ltMap["override"]; ok {
			for _, o := range overrides.([]interface{}) {
				oMap := o.(map[string]interface{})
				override := duplosdk.DuploAsgLaunchTemplateOverride{
					InstanceType:     oMap["instance_type"].(string),
					WeightedCapacity: oMap["weighted_capacity"].(string),
				}

				if ir, ok := oMap["instance_requirements"]; ok && len(ir.([]interface{})) > 0 {
					irMap := ir.([]interface{})[0].(map[string]interface{})
					req := &duplosdk.DuploAsgInstanceRequirements{}

					if vcpu, ok := irMap["vcpu_count"]; ok && len(vcpu.([]interface{})) > 0 {
						vcpuMap := vcpu.([]interface{})[0].(map[string]interface{})
						req.VCpuCount = &duplosdk.DuploAsgIntRange{
							Min: vcpuMap["min"].(int),
							Max: vcpuMap["max"].(int),
						}
					}
					if mem, ok := irMap["memory_mib"]; ok && len(mem.([]interface{})) > 0 {
						memMap := mem.([]interface{})[0].(map[string]interface{})
						req.MemoryMiB = &duplosdk.DuploAsgIntRange{
							Min: memMap["min"].(int),
							Max: memMap["max"].(int),
						}
					}
					if v, ok := irMap["allowed_instance_types"]; ok {
						req.AllowedInstanceTypes = expandStringList(v.([]interface{}))
					}
					if v, ok := irMap["excluded_instance_types"]; ok {
						req.ExcludedInstanceTypes = expandStringList(v.([]interface{}))
					}
					if v, ok := irMap["cpu_manufacturers"]; ok {
						req.CpuManufacturers = expandStringList(v.([]interface{}))
					}
					if v, ok := irMap["instance_generations"]; ok {
						req.InstanceGenerations = expandStringList(v.([]interface{}))
					}
					if v, ok := irMap["spot_max_price_percentage_over_lowest_price"]; ok && v.(int) > 0 {
						val := v.(int)
						req.SpotMaxPricePercentageOverLowestPrice = &val
					}
					override.InstanceRequirements = req
				}
				launchTemplate.Overrides = append(launchTemplate.Overrides, override)
			}
		}
		policy.LaunchTemplate = launchTemplate
	}

	if dist, ok := policyMap["instances_distribution"]; ok && len(dist.([]interface{})) > 0 {
		distMap := dist.([]interface{})[0].(map[string]interface{})
		distribution := &duplosdk.DuploAsgInstancesDistribution{
			OnDemandAllocationStrategy: distMap["on_demand_allocation_strategy"].(string),
			SpotAllocationStrategy:     distMap["spot_allocation_strategy"].(string),
			SpotMaxPrice:               distMap["spot_max_price"].(string),
		}
		if v, ok := distMap["on_demand_base_capacity"]; ok && v.(int) > 0 {
			val := v.(int)
			distribution.OnDemandBaseCapacity = &val
		}
		if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok {
			val := v.(int)
			distribution.OnDemandPercentageAboveBaseCapacity = &val
		}
		if v, ok := distMap["spot_instance_pools"]; ok && v.(int) > 0 {
			val := v.(int)
			distribution.SpotInstancePools = &val
		}
		policy.InstancesDistribution = distribution
	}

	return policy
}

func flattenAsgMixedInstancesPolicy(policy *duplosdk.DuploAsgMixedInstancesPolicy) []interface{} {
	if policy == nil {
		return nil
	}
	result := map[string]interface{}{}

	if policy.LaunchTemplate != nil {
		lt := map[string]interface{}{}
		var overrides []interface{}
		for _, o := range policy.LaunchTemplate.Overrides {
			override := map[string]interface{}{
				"instance_type":     o.InstanceType,
				"weighted_capacity": o.WeightedCapacity,
			}
			if o.InstanceRequirements != nil {
				ir := map[string]interface{}{}
				if o.InstanceRequirements.VCpuCount != nil {
					ir["vcpu_count"] = []interface{}{
						map[string]interface{}{
							"min": o.InstanceRequirements.VCpuCount.Min,
							"max": o.InstanceRequirements.VCpuCount.Max,
						},
					}
				}
				if o.InstanceRequirements.MemoryMiB != nil {
					ir["memory_mib"] = []interface{}{
						map[string]interface{}{
							"min": o.InstanceRequirements.MemoryMiB.Min,
							"max": o.InstanceRequirements.MemoryMiB.Max,
						},
					}
				}
				if o.InstanceRequirements.AllowedInstanceTypes != nil {
					ir["allowed_instance_types"] = o.InstanceRequirements.AllowedInstanceTypes
				}
				if o.InstanceRequirements.ExcludedInstanceTypes != nil {
					ir["excluded_instance_types"] = o.InstanceRequirements.ExcludedInstanceTypes
				}
				if o.InstanceRequirements.CpuManufacturers != nil {
					ir["cpu_manufacturers"] = o.InstanceRequirements.CpuManufacturers
				}
				if o.InstanceRequirements.InstanceGenerations != nil {
					ir["instance_generations"] = o.InstanceRequirements.InstanceGenerations
				}
				if o.InstanceRequirements.SpotMaxPricePercentageOverLowestPrice != nil {
					ir["spot_max_price_percentage_over_lowest_price"] = *o.InstanceRequirements.SpotMaxPricePercentageOverLowestPrice
				}
				override["instance_requirements"] = []interface{}{ir}
			}
			overrides = append(overrides, override)
		}
		lt["override"] = overrides
		result["launch_template"] = []interface{}{lt}
	}

	if policy.InstancesDistribution != nil {
		dist := map[string]interface{}{
			"on_demand_allocation_strategy": policy.InstancesDistribution.OnDemandAllocationStrategy,
			"spot_allocation_strategy":      policy.InstancesDistribution.SpotAllocationStrategy,
			"spot_max_price":                policy.InstancesDistribution.SpotMaxPrice,
		}
		if policy.InstancesDistribution.OnDemandBaseCapacity != nil {
			dist["on_demand_base_capacity"] = *policy.InstancesDistribution.OnDemandBaseCapacity
		}
		if policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity != nil {
			dist["on_demand_percentage_above_base_capacity"] = *policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity
		}
		if policy.InstancesDistribution.SpotInstancePools != nil {
			dist["spot_instance_pools"] = *policy.InstancesDistribution.SpotInstancePools
		}
		result["instances_distribution"] = []interface{}{dist}
	}

	return []interface{}{result}
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
		d.HasChange("enabled_metrics") ||
		d.HasChange("mixed_instances_policy")
}

func getCustomDataTagsUpdates(d *schema.ResourceData) []duplosdk.CustomDataUpdate {
	var updates []duplosdk.CustomDataUpdate

	if !d.HasChange("custom_data_tags") {
		return updates
	}

	oldRevision, newRevision := d.GetChange("custom_data_tags")
	oldTags := oldRevision.([]interface{})
	newTags := newRevision.([]interface{})

	// Create maps for easier lookup
	oldTagsMap := make(map[string]string)
	newTagsMap := make(map[string]string)

	for _, tag := range oldTags {
		tagMap := tag.(map[string]interface{})
		key := tagMap["key"].(string)
		value := tagMap["value"].(string)
		oldTagsMap[key] = value
	}

	for _, tag := range newTags {
		tagMap := tag.(map[string]interface{})
		key := tagMap["key"].(string)
		value := tagMap["value"].(string)
		newTagsMap[key] = value
	}

	// Check for updated or new tags
	for key, newValue := range newTagsMap {
		oldValue, existed := oldTagsMap[key]
		if !existed || oldValue != newValue {
			update := duplosdk.CustomDataUpdate{
				Key:   key,
				Value: newValue,
			}
			updates = append(updates, update)
		}
	}

	// Check for deleted tags
	for key := range oldTagsMap {
		if _, exists := newTagsMap[key]; !exists {
			update := duplosdk.CustomDataUpdate{
				Key:   key,
				Value: "",
				State: "delete",
			}
			updates = append(updates, update)
		}
	}

	return updates
}

func asgWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) (error, bool) {
	flag := true
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AsgProfileGet(tenantID, name)
			//			log.Printf("[TRACE] Dynamodb status is (%s).", rp.TableStatus.Value)
			status := "pending"
			if err == nil {
				if rp.Created == nil {
					status = "ready"
					flag = false
				} else {
					if *rp.Created && rp.Arn != "" {
						status = "ready"
					} else {
						status = "pending"
					}
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] asgWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err, flag
}
