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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsBatchComputeEnvironmentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws batch compute environment will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the compute environment.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the compute environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The Amazon Resource Name of the compute environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"ecs_cluster_arn": {
			Description: "The Amazon Resource Name (ARN) of the underlying Amazon ECS cluster used by the compute environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_role": {
			Description: "The full Amazon Resource Name (ARN) of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"wait_for_deployment": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"state": {
			Description: "The state of the compute environment. If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues. Compute environment must be created in `ENABLED` state. Valid items are `ENABLED` or `DISABLED`.",
			Type:        schema.TypeString,
			Optional:    true,
			StateFunc: func(val interface{}) string {
				return strings.ToUpper(val.(string))
			},
			ValidateFunc: validation.StringInSlice([]string{
				"ENABLED",
				"DISABLED",
			}, true),
			Default: "ENABLED",
		},
		"status": {
			Description: "The current status of the compute environment (for example, CREATING or VALID).",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status_reason": {
			Description: "A short, human-readable string to provide additional details about the current status of the compute environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: "The type of the compute environment. Valid items are `MANAGED` or `UNMANAGED`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			StateFunc: func(val interface{}) string {
				return strings.ToUpper(val.(string))
			},
			ValidateFunc: validation.StringInSlice([]string{
				"MANAGED",
				"UNMANAGED",
			}, true),
		},
		"compute_resources": {
			Description: "Details of the compute resources managed by the compute environment. This parameter is required for managed compute environments.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MinItems:    0,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allocation_strategy": {
						Description: "The allocation strategy to use for the compute resource in case not enough instances of the best fitting instance type can be allocated.",
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						StateFunc: func(val interface{}) string {
							return strings.ToUpper(val.(string))
						},
						ValidateFunc: validation.StringInSlice([]string{
							"BEST_FIT_PROGRESSIVE",
							"SPOT_CAPACITY_OPTIMIZED",
							"BEST_FIT",
						}, true),
					},
					"bid_percentage": {
						Description: "Integer of maximum percentage that a Spot Instance price can be when compared with the On-Demand price for that instance type before instances are launched.",
						Type:        schema.TypeInt,
						Optional:    true,
						ForceNew:    true,
					},
					"desired_vcpus": {
						Description: "The desired number of EC2 vCPUS in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},
					"ec2_configuration": {
						Description: "Provides information used to select Amazon Machine Images (AMIs) for EC2 instances in the compute environment.",
						Type:        schema.TypeList,
						Optional:    true,
						Computed:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"image_id_override": {
									Description:  "The AMI ID used for instances launched in the compute environment that match the image type. This setting overrides the `image_id` argument in the `compute_resources` block.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ForceNew:     true,
									ValidateFunc: validation.StringLenBetween(1, 256),
								},
								"image_type": {
									Description:  "The image type to match with the instance type to select an AMI.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringLenBetween(1, 256),
								},
							},
						},
					},
					"ec2_key_pair": {
						Description: "The EC2 key pair that is used for instances launched in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"image_id": {
						Description: "The Amazon Machine Image (AMI) ID used for instances launched in the compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified. (Deprecated, use ec2_configuration `image_id_override` instead)",
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
					},
					"instance_role": {
						Description: "The Amazon ECS instance role applied to Amazon EC2 instances in a compute environment. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"instance_type": {
						Description: "A list of instance types that may be launched. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeSet,
						Optional:    true,
						ForceNew:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"launch_template": {
						Description: "The launch template to use for your compute resources. See details below. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeList,
						Optional:    true,
						ForceNew:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"launch_template_id": {
									Description:   "ID of the launch template. You must specify either the launch template ID or launch template name in the request, but not both.",
									Type:          schema.TypeString,
									Optional:      true,
									ForceNew:      true,
									ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
								},
								"launch_template_name": {
									Description:   "Name of the launch template.",
									Type:          schema.TypeString,
									Optional:      true,
									ForceNew:      true,
									ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
								},
								"version": {
									Description: "The version number of the launch template. Default: The default version of the launch template.",
									Type:        schema.TypeString,
									Optional:    true,
									ForceNew:    true,
								},
							},
						},
					},
					"max_vcpus": {
						Description: "The maximum number of EC2 vCPUs that an environment can reach.",
						Type:        schema.TypeInt,
						Required:    true,
					},
					"min_vcpus": {
						Description: "The minimum number of EC2 vCPUs that an environment should maintain. For `EC2` or `SPOT` compute environments, if the parameter is not explicitly defined, a `0` default value will be set. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"security_group_ids": {
						Description: "A list of EC2 security group that are associated with instances launched in the compute environment. This parameter is required for Fargate compute environments.",
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"spot_iam_fleet_role": {
						Description: "The Amazon Resource Name (ARN) of the Amazon EC2 Spot Fleet IAM role applied to a SPOT compute environment. This parameter is required for SPOT compute environments. This parameter isn't applicable to jobs running on Fargate resources, and shouldn't be specified.",
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
					},
					"subnets": {
						Description: "A list of VPC subnets into which the compute resources are launched.",
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"tags": {
						Description: "Key-value map of resource tags.",
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Computed:    true,
					},
					"type": {
						Description: "The type of compute environment. Valid items are `EC2`, `SPOT`, `FARGATE` or `FARGATE_SPOT`.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
						StateFunc: func(val interface{}) string {
							return strings.ToUpper(val.(string))
						},
						ValidateFunc: validation.StringInSlice([]string{
							"EC2",
							"SPOT",
							"FARGATE",
							"FARGATE_SPOT",
						}, true),
					},
				},
			},
		},
		"tags": {
			Description: "Key-value map of resource tags.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    true,
		},
	}
}

func resourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_batch_compute_environment` manages an aws batch compute environment in Duplo.",

		ReadContext:   resourceAwsBatchComputeEnvironmentRead,
		CreateContext: resourceAwsBatchComputeEnvironmentCreate,
		UpdateContext: resourceAwsBatchComputeEnvironmentUpdate,
		DeleteContext: resourceAwsBatchComputeEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsBatchComputeEnvironmentSchema(),
	}
}

func resourceAwsBatchComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchComputeEnvironmentIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	ce, clientErr := c.AwsBatchComputeEnvironmentGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if ce == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenBatchComputeEnvironment(d, c, ce, tenantID)

	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsBatchComputeEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	rq := expandAwsBatchComputeEnvironment(d)
	err = c.AwsBatchComputeEnvironmentCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws batch compute environment '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch compute environment", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchComputeEnvironmentGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = batchComputeEnvironmentUntilValid(ctx, c, tenantID, fullName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAwsBatchComputeEnvironmentRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsBatchComputeEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	arn := d.Get("arn").(string)
	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	if d.HasChangesExcept("tags", "tags_all") {
		input := &duplosdk.DuploAwsBatchComputeEnvironment{
			ComputeEnvironment:    fullName,
			ComputeEnvironmentArn: arn,
		}

		if d.HasChange("service_role") {
			input.ServiceRole = d.Get("service_role").(string)
		}

		if d.HasChange("state") {
			input.State = &duplosdk.DuploStringValue{Value: d.Get("state").(string)}
		}

		if computeEnvironmentType := strings.ToUpper(d.Get("type").(string)); computeEnvironmentType == "MANAGED" {
			// "At least one compute-resources attribute must be specified"
			computeResourceUpdate := &duplosdk.DuploAwsBatchComputeResource{
				MaxvCpus: d.Get("compute_resources.0.max_vcpus").(int),
			}

			if d.HasChange("compute_resources.0.desired_vcpus") {
				computeResourceUpdate.DesiredvCpus = d.Get("compute_resources.0.desired_vcpus").(int)
			}

			if d.HasChange("compute_resources.0.min_vcpus") {
				computeResourceUpdate.MinvCpus = d.Get("compute_resources.0.min_vcpus").(int)
			}

			if d.HasChange("compute_resources.0.security_group_ids") {
				computeResourceUpdate.SecurityGroupIds = expandStringSet(d.Get("compute_resources.0.security_group_ids").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.subnets") {
				computeResourceUpdate.Subnets = expandStringSet(d.Get("compute_resources.0.subnets").(*schema.Set))
			}

			input.ComputeResources = computeResourceUpdate
		}

		log.Printf("[DEBUG] Updating Batch Compute Environment: %v", input)
		err := c.AwsBatchComputeEnvironmentUpdate(tenantID, input)
		if err != nil {
			return diag.Errorf("Error creating tenant %s aws batch compute environment '%s': %s", tenantID, name, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch compute environment", fullName, func() (interface{}, duplosdk.ClientError) {
			return c.AwsBatchComputeEnvironmentGet(tenantID, fullName)
		})
		if diags != nil {
			return diags
		}
		if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
			err := batchComputeEnvironmentUntilValid(ctx, c, tenantID, fullName, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceAwsBatchComputeEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchComputeEnvironmentIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)

	ce, err := c.AwsBatchComputeEnvironmentGet(tenantID, fullName)
	if err != nil {
		return diag.FromErr(err)
	}
	if ce == nil {
		return nil
	}

	if ce.State != nil && ce.State.Value == "ENABLED" {
		log.Printf("[TRACE] Disable batch compute environment before delete. (%s, %s): start", tenantID, fullName)

		clientErr := c.AwsBatchComputeEnvironmentDisable(tenantID, fullName)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				return nil
			}
			return diag.Errorf("Unable to disable tenant %s aws  batch compute environment '%s': %s", tenantID, fullName, clientErr)
		}
		err = batchComputeEnvironmentUntilDisabled(ctx, c, tenantID, fullName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
		// time.Sleep(time.Duration(20) * time.Second)
	}

	clientErr := c.AwsBatchComputeEnvironmentDelete(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws batch compute environment '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws batch compute environment", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchComputeEnvironmentGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsBatchComputeEnvironmentDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsBatchComputeEnvironment(d *schema.ResourceData) *duplosdk.DuploAwsBatchComputeEnvironment {

	input := &duplosdk.DuploAwsBatchComputeEnvironment{
		ComputeEnvironmentName: d.Get("name").(string),
		ServiceRole:            d.Get("service_role").(string),
		Type:                   &duplosdk.DuploStringValue{Value: d.Get("type").(string)},
	}

	if v, ok := d.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ComputeResources = expandComputeResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = &duplosdk.DuploStringValue{Value: v.(string)}
	}

	if v, ok := d.GetOk("tags"); ok && v != nil {
		input.Tags = expandAsStringMap("tags", d)
	}

	return input
}

func parseAwsBatchComputeEnvironmentIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenBatchComputeEnvironment(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploAwsBatchComputeEnvironment, tenantId string) diag.Diagnostics {
	log.Printf("[TRACE]flattenBatchComputeEnvironment... Start ")
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, duplo.ComputeEnvironmentName)
	d.Set("tenant_id", tenantId)
	d.Set("name", name)
	d.Set("arn", duplo.ComputeEnvironmentArn)
	d.Set("fullname", duplo.ComputeEnvironmentName)
	d.Set("tags", duplo.Tags)
	d.Set("ecs_cluster_arn", duplo.EcsClusterArn)
	d.Set("service_role", duplo.ServiceRole)
	if duplo.State != nil {
		d.Set("state", duplo.State.Value)
	}
	if duplo.Status != nil {
		d.Set("status", duplo.Status.Value)
	}
	d.Set("status_reason", duplo.StatusReason)
	if duplo.Type != nil {
		d.Set("type", duplo.Type.Value)
	}
	if duplo.ComputeResources != nil {
		if err := d.Set("compute_resources", []interface{}{flattenComputeResource(duplo.ComputeResources)}); err != nil {
			return diag.FromErr(err)
		}
	} else {
		d.Set("compute_resources", nil)
	}
	log.Printf("[TRACE]flattenBatchComputeEnvironment... End ")
	return nil
}

func flattenComputeResource(apiObject *duplosdk.DuploAwsBatchComputeResource) map[string]interface{} {
	log.Printf("[TRACE]flattenComputeResource... Start ")
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil && len(v.Value) > 0 {
		tfMap["allocation_strategy"] = v.Value
	}

	if v := apiObject.BidPercentage; v > 0 {
		tfMap["bid_percentage"] = v
	}

	if v := apiObject.DesiredvCpus; v > 0 {
		tfMap["desired_vcpus"] = v
	}

	if v := apiObject.Ec2Configuration; v != nil && len(*v) > 0 {
		tfMap["ec2_configuration"] = flattenEC2Configurations(v)
	}

	if v := apiObject.Ec2KeyPair; len(v) > 0 {
		tfMap["ec2_key_pair"] = v
	}

	if v := apiObject.ImageId; len(v) > 0 {
		tfMap["image_id"] = v
	}

	if v := apiObject.InstanceRole; len(v) > 0 {
		tfMap["instance_role"] = v
	}

	if v := apiObject.InstanceTypes; len(v) > 0 {
		tfMap["instance_type"] = v
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxvCpus; v > 0 {
		tfMap["max_vcpus"] = v
	}

	if v := apiObject.MinvCpus; v > 0 {
		tfMap["min_vcpus"] = v
	}

	if v := apiObject.SecurityGroupIds; len(v) > 0 {
		tfMap["security_group_ids"] = v
	}

	if v := apiObject.SpotIamFleetRole; len(v) > 0 {
		tfMap["spot_iam_fleet_role"] = v
	}

	if v := apiObject.Subnets; len(v) > 0 {
		tfMap["subnets"] = v
	}

	if v := apiObject.Tags; len(v) > 0 {
		tfMap["tags"] = v
	}

	if v := apiObject.Type; v != nil && len(v.Value) > 0 {
		tfMap["type"] = v.Value
	}
	log.Printf("[TRACE]flattenComputeResource... End ")
	return tfMap
}

func flattenEC2Configurations(apiObjects *[]duplosdk.DuploAwsBatchComputeEc2Configuration) []interface{} {
	log.Printf("[TRACE]flattenEC2Configurations... Start ")

	var tfList []interface{}

	for _, apiObject := range *apiObjects {
		tfList = append(tfList, flattenEC2Configuration(apiObject))
	}
	log.Printf("[TRACE]flattenEC2Configurations... End ")
	return tfList
}

func flattenEC2Configuration(apiObject duplosdk.DuploAwsBatchComputeEc2Configuration) map[string]interface{} {
	log.Printf("[TRACE]flattenEC2Configuration... Start ")
	tfMap := map[string]interface{}{}

	if v := apiObject.ImageIdOverride; len(v) > 0 {
		tfMap["image_id_override"] = v
	}

	if v := apiObject.ImageType; len(v) > 0 {
		tfMap["image_type"] = v
	}
	log.Printf("[TRACE]flattenEC2Configuration... End ")
	return tfMap
}

func flattenLaunchTemplateSpecification(apiObject *duplosdk.DuploAwsBatchLaunchTemplateConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; len(v) > 0 {
		tfMap["launch_template_id"] = v
	}

	if v := apiObject.LaunchTemplateName; len(v) > 0 {
		tfMap["launch_template_name"] = v
	}

	if v := apiObject.Version; len(v) > 0 {
		tfMap["version"] = v
	}

	return tfMap
}

func expandComputeResource(tfMap map[string]interface{}) *duplosdk.DuploAwsBatchComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap["type"].(string); ok && v != "" {
		computeResourceType = v
	}

	apiObject := &duplosdk.DuploAwsBatchComputeResource{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = &duplosdk.DuploStringValue{Value: v}
	}

	if v, ok := tfMap["bid_percentage"].(int); ok && v != 0 {
		apiObject.BidPercentage = v
	}

	if v, ok := tfMap["desired_vcpus"].(int); ok && v != 0 {
		apiObject.DesiredvCpus = v
	}

	if v, ok := tfMap["ec2_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.Ec2Configuration = expandEC2Configurations(v)
	}

	if v, ok := tfMap["ec2_key_pair"].(string); ok && v != "" {
		apiObject.Ec2KeyPair = v
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = v
	}

	if v, ok := tfMap["instance_role"].(string); ok && v != "" {
		apiObject.InstanceRole = v
	}

	if v, ok := tfMap["instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = expandStringSet(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplate = expandLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_vcpus"].(int); ok && v != 0 {
		apiObject.MaxvCpus = v
	}

	if v, ok := tfMap["min_vcpus"].(int); ok && v != 0 {
		apiObject.MinvCpus = v
	} else if computeResourceType := strings.ToUpper(computeResourceType); computeResourceType == "EC2" || computeResourceType == "SPOT" {
		apiObject.MinvCpus = 0
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = expandStringSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = v
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = expandStringSet(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		m := map[string]string{}
		for k, val := range v {
			if val == nil {
				m[k] = ""
			} else {
				m[k] = val.(string)
			}
		}
		apiObject.Tags = m
	}

	if computeResourceType != "" {
		apiObject.Type = &duplosdk.DuploStringValue{Value: computeResourceType}
	}

	return apiObject
}

func expandEC2Configurations(tfList []interface{}) *[]duplosdk.DuploAwsBatchComputeEc2Configuration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []duplosdk.DuploAwsBatchComputeEc2Configuration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEC2Configuration(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return &apiObjects
}

func expandEC2Configuration(tfMap map[string]interface{}) duplosdk.DuploAwsBatchComputeEc2Configuration {
	apiObject := duplosdk.DuploAwsBatchComputeEc2Configuration{}

	if v, ok := tfMap["image_id_override"].(string); ok && v != "" {
		apiObject.ImageIdOverride = v
	}

	if v, ok := tfMap["image_type"].(string); ok && v != "" {
		apiObject.ImageType = v
	}

	return apiObject
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *duplosdk.DuploAwsBatchLaunchTemplateConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &duplosdk.DuploAwsBatchLaunchTemplateConfiguration{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = v
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = v
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = v
	}

	return apiObject
}

func batchComputeEnvironmentUntilDisabled(ctx context.Context, c *duplosdk.Client, tenantID string, fullname string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsBatchComputeEnvironmentGet(tenantID, fullname)
			log.Printf("[TRACE] Batch compute environment state is (%s).", rp.State.Value)
			status := "pending"
			if err == nil {
				if rp.State != nil && rp.State.Value == "DISABLED" && rp.Status != nil && rp.Status.Value == "VALID" {
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
	log.Printf("[DEBUG] batchComputeEnvironmentUntilDisabled(%s, %s)", tenantID, fullname)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func batchComputeEnvironmentUntilValid(ctx context.Context, c *duplosdk.Client, tenantID string, fullname string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsBatchComputeEnvironmentGet(tenantID, fullname)
			log.Printf("[TRACE] Batch compute environment status is (%s).", rp.Status.Value)
			status := "pending"
			if err == nil {
				if rp.Status != nil && rp.Status.Value == "VALID" {
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
	log.Printf("[DEBUG] batchComputeEnvironmentUntilValid(%s, %s)", tenantID, fullname)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
