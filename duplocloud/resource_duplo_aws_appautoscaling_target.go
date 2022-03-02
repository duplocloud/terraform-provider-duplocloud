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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsAppautoscalingTargetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws autoscaling target will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"max_capacity": {
			Description: "The max capacity of the scalable target.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"min_capacity": {
			Description: "The min capacity of the scalable target.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"resource_id": {
			Description: "Resource name associated with the scaling policy.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"full_resource_id": {
			Description: "The resource type and unique identifier string for the resource associated with the scaling policy.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role_arn": {
			Description: "The ARN of the IAM role that allows Application AutoScaling to modify your scalable target on your behalf.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"scalable_dimension": {
			Description: "The scalable dimension of the scalable target.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"service_namespace": {
			Description: "The AWS service namespace of the scalable target.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

func resourceAwsAppautoscalingTarget() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_appautoscaling_target` manages an aws autoscaling target in Duplo.",

		ReadContext:   resourceAwsAppautoscalingTargetRead,
		CreateContext: resourceAwsAppautoscalingTargetCreate,
		UpdateContext: resourceAwsAppautoscalingTargetUpdate,
		DeleteContext: resourceAwsAppautoscalingTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsAppautoscalingTargetSchema(),
	}
}

func resourceAwsAppautoscalingTargetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, namespace, dimension, name, err := parseAwsAppautoscalingTargetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAppautoscalingTargetRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploAwsAutoscalingTargetGet(tenantID, duplosdk.DuploDuploAwsAutoscalingTargetGetReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceIds:       []string{name},
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws autoscaling target %s : %s", tenantID, name, clientErr)
	}

	flattenAwsAppautoscalingTarget(d, duplo)
	d.Set("tenant_id", tenantID)
	d.Set("resource_id", name)

	log.Printf("[TRACE] resourceAwsAppautoscalingTargetRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsAppautoscalingTargetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	rq := expandAwsAppautoscalingTarget(d)
	err = c.DuploAwsAutoscalingTargetCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws autoscaling target '%s': %s", tenantID, rq.ResourceId, err)
	}

	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, rq.ServiceNamespace.Value, rq.ScalableDimension.Value, rq.ResourceId)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws autoscaling target", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsAutoscalingTargetGet(tenantID, duplosdk.DuploDuploAwsAutoscalingTargetGetReq{
			ServiceNamespace:  rq.ServiceNamespace.Value,
			ScalableDimension: rq.ScalableDimension.Value,
			ResourceIds:       []string{rq.ResourceId},
		})
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsAppautoscalingTargetRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsAppautoscalingTargetCreate(%s, %s): end", tenantID, rq.ResourceId)
	return diags
}

func resourceAwsAppautoscalingTargetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAwsAppautoscalingTargetCreate(ctx, d, m)
}

func resourceAwsAppautoscalingTargetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, namespace, dimension, name, err := parseAwsAppautoscalingTargetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAppautoscalingTargetDelete(%s, %s): start", tenantID, name)
	getRq := duplosdk.DuploDuploAwsAutoscalingTargetGetReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceIds:       []string{d.Get("resource_id").(string)},
	}
	c := m.(*duplosdk.Client)
	clientErr := c.DuploAwsAutoscalingTargetDelete(tenantID, duplosdk.DuploDuploAwsAutoscalingTargetDeleteReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceId:        name,
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws autoscaling target '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws autoscaling target", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsAutoscalingTargetGet(tenantID, getRq)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsAppautoscalingTargetDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsAppautoscalingTarget(d *schema.ResourceData) *duplosdk.DuploDuploAwsAutoscalingTarget {
	return &duplosdk.DuploDuploAwsAutoscalingTarget{
		MaxCapacity: d.Get("max_capacity").(int),
		MinCapacity: d.Get("min_capacity").(int),
		ResourceId:  d.Get("resource_id").(string),
		ScalableDimension: &duplosdk.DuploStringValue{
			Value: d.Get("scalable_dimension").(string),
		},
		ServiceNamespace: &duplosdk.DuploStringValue{
			Value: d.Get("service_namespace").(string),
		},
		RoleARN: d.Get("role_arn").(string),
	}
}

func parseAwsAppautoscalingTargetIdParts(id string) (tenantID, namespace, dimension, name string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, namespace, dimension, name = idParts[0], idParts[1], idParts[2], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAwsAppautoscalingTarget(d *schema.ResourceData, duplo *duplosdk.DuploDuploAwsAutoscalingTarget) {
	d.Set("max_capacity", duplo.MaxCapacity)
	d.Set("min_capacity", duplo.MinCapacity)
	d.Set("full_resource_id", duplo.ResourceId)
	d.Set("scalable_dimension", duplo.ScalableDimension.Value)
	d.Set("service_namespace", duplo.ServiceNamespace.Value)
	d.Set("role_arn", duplo.RoleARN)
}
