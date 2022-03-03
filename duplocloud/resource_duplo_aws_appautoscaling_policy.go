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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsAppautoscalingPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws autoscaling policy will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description:  "The name of the policy. Must be between 1 and 255 characters in length.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(0, 255),
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"policy_type": {
			Description: "The policy type. Valid values are `StepScaling` and `TargetTrackingScaling`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "StepScaling",
		},
		"resource_id": {
			Description: "The resource type and unique identifier string for the resource associated with the scaling policy.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
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
		"step_scaling_policy_configuration": {
			Description: "Step scaling policy configuration, requires `policy_type = \"StepScaling\"`",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"adjustment_type": {
						Description: "Specifies whether the adjustment is an absolute number or a percentage of the current capacity. Valid values are `ChangeInCapacity`, `ExactCapacity`, and `PercentChangeInCapacity`.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"cooldown": {
						Description: "The amount of time, in seconds, after a scaling activity completes and before the next scaling activity can start.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"metric_aggregation_type": {
						Description: "The aggregation type for the policy's metrics. Valid values are \"Minimum\", \"Maximum\", and \"Average\". Without a value, AWS will treat the aggregation type as \"Average\".",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"min_adjustment_magnitude": {
						Description: "The minimum number to adjust your scalable dimension as a result of a scaling activity. If the adjustment type is PercentChangeInCapacity, the scaling policy changes the scalable dimension of the scalable target by this amount.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"step_adjustment": {
						Description: "A set of adjustments that manage scaling.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"metric_interval_lower_bound": {
									Description: "The lower bound for the difference between the alarm threshold and the CloudWatch metric.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"metric_interval_upper_bound": {
									Description: "The upper bound for the difference between the alarm threshold and the CloudWatch metric.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"scaling_adjustment": {
									Description: "The number of members by which to scale, when the adjustment bounds are breached.",
									Type:        schema.TypeInt,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
		"target_tracking_scaling_policy_configuration": {
			Description: "A target tracking policy, requires `policy_type = \"TargetTrackingScaling\"`",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"customized_metric_specification": {
						Type:          schema.TypeList,
						MaxItems:      1,
						Optional:      true,
						ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.predefined_metric_specification"},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"dimensions": {
									Type:     schema.TypeSet,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Description: "Name of the dimension.",
												Type:        schema.TypeString,
												Required:    true,
											},
											"value": {
												Description: "Value of the dimension.",
												Type:        schema.TypeString,
												Required:    true,
											},
										},
									},
								},
								"metric_name": {
									Description: "The name of the metric.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"namespace": {
									Description: "The namespace of the metric.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"statistic": {
									Description: "The statistic of the metric. Valid values: `Average`, `Minimum`, `Maximum`, `SampleCount`, and `Sum`.",
									Type:        schema.TypeString,
									Required:    true,
									ValidateFunc: validation.StringInSlice([]string{
										"Average",
										"Maximum",
										"Minimum",
										"SampleCount",
										"Sum",
									}, false),
								},
								"unit": {
									Description: "The unit of the metric.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"predefined_metric_specification": {
						Type:          schema.TypeList,
						MaxItems:      1,
						Optional:      true,
						ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification"},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"predefined_metric_type": {
									Description: "The metric type.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"resource_label": {
									Description:  "Reserved for future use. Must be less than or equal to 1023 characters in length.",
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringLenBetween(0, 1023),
								},
							},
						},
					},
					"disable_scale_in": {
						Description: "Indicates whether scale in by the target tracking policy is disabled.",
						Type:        schema.TypeBool,
						Default:     false,
						Optional:    true,
					},
					"scale_in_cooldown": {
						Description: "The amount of time, in seconds, after a scale in activity completes before another scale in activity can start.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"scale_out_cooldown": {
						Description: "The amount of time, in seconds, after a scale out activity completes before another scale out activity can start.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"target_value": {
						Description: "The target value for the metric.",
						Type:        schema.TypeFloat,
						Required:    true,
					},
				},
			},
		},
	}
}

func resourceAwsAppautoscalingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_appautoscaling_policy` manages an aws autoscaling policy in Duplo.",

		ReadContext:   resourceAwsAppautoscalingPolicyRead,
		CreateContext: resourceAwsAppautoscalingPolicyCreate,
		UpdateContext: resourceAwsAppautoscalingPolicyUpdate,
		DeleteContext: resourceAwsAppautoscalingPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsAppautoscalingPolicySchema(),
	}
}

func resourceAwsAppautoscalingPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, namespace, dimension, resourceId, name, err := parseAwsAppautoscalingPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploAwsAutoscalingPolicyGet(tenantID, duplosdk.DuploAwsAutoscalingPolicyGetReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceId:        resourceId,
		PolicyNames:       []string{name},
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws autoscaling policy %s : %s", tenantID, name, clientErr)
	}

	err = flattenAwsAppautoscalingPolicy(d, duplo)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsAppautoscalingPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq, inputErr := expandAwsAppautoscalingPolicy(d)
	if inputErr != nil {
		return diag.Errorf("Error creating aws autoscaling policy request - %s : %s", name, err)
	}
	err = c.DuploAwsAutoscalingPolicyCreate(tenantID, rq)

	if err != nil {
		return diag.Errorf("Error creating tenant %s aws autoscaling policy '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s/%s/%s", tenantID, rq.ServiceNamespace.Value, rq.ScalableDimension.Value, rq.ResourceId, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws autoscaling policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsAutoscalingPolicyGet(tenantID, duplosdk.DuploAwsAutoscalingPolicyGetReq{
			ServiceNamespace:  rq.ServiceNamespace.Value,
			ScalableDimension: rq.ScalableDimension.Value,
			ResourceId:        rq.ResourceId,
			PolicyNames:       []string{name},
		})
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsAppautoscalingPolicyRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsAppautoscalingPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsAppautoscalingPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, namespace, dimension, resourceId, name, err := parseAwsAppautoscalingPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	getRq := duplosdk.DuploAwsAutoscalingPolicyGetReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceId:        resourceId,
		PolicyNames:       []string{name},
	}

	clientErr := c.DuploAwsAutoscalingPolicyDelete(tenantID, duplosdk.DuploAwsAutoscalingPolicyDeleteReq{
		ServiceNamespace:  namespace,
		ScalableDimension: dimension,
		ResourceId:        resourceId,
		PolicyName:        name,
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws autoscaling policy '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws autoscaling policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsAutoscalingPolicyGet(tenantID, getRq)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsAppautoscalingPolicyDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsAppautoscalingPolicy(d *schema.ResourceData) (*duplosdk.DuploAwsAutoscalingPolicy, error) {
	var params = duplosdk.DuploAwsAutoscalingPolicy{
		PolicyName: d.Get("name").(string),
		ResourceId: d.Get("resource_id").(string),
	}

	if v, ok := d.GetOk("policy_type"); ok {
		params.PolicyType = &duplosdk.DuploStringValue{Value: v.(string)}
	}

	if v, ok := d.GetOk("service_namespace"); ok {
		params.ServiceNamespace = &duplosdk.DuploStringValue{Value: v.(string)}
	}

	if v, ok := d.GetOk("scalable_dimension"); ok {
		params.ScalableDimension = &duplosdk.DuploStringValue{Value: v.(string)}
	}

	if v, ok := d.GetOk("step_scaling_policy_configuration"); ok {
		params.StepScalingPolicyConfiguration = expandStepScalingPolicyConfiguration(v.([]interface{}))
	}

	if l, ok := d.GetOk("target_tracking_scaling_policy_configuration"); ok {
		v := l.([]interface{})
		if len(v) < 1 {
			return &params, fmt.Errorf("Empty target_tracking_scaling_policy_configuration block")
		}
		ttspCfg := v[0].(map[string]interface{})
		cfg := &duplosdk.DuploTargetTrackingScalingPolicyConfiguration{
			TargetValue: ttspCfg["target_value"].(float64),
		}

		if v, ok := ttspCfg["scale_in_cooldown"]; ok {
			cfg.ScaleInCooldown = v.(int)
		}

		if v, ok := ttspCfg["scale_out_cooldown"]; ok {
			cfg.ScaleOutCooldown = v.(int)
		}

		if v, ok := ttspCfg["disable_scale_in"]; ok {
			cfg.DisableScaleIn = v.(bool)
		}

		if v, ok := ttspCfg["customized_metric_specification"].([]interface{}); ok && len(v) > 0 {
			cfg.CustomizedMetricSpecification = expandAppautoscalingCustomizedMetricSpecification(v)
		}

		if v, ok := ttspCfg["predefined_metric_specification"].([]interface{}); ok && len(v) > 0 {
			cfg.PredefinedMetricSpecification = expandAppautoscalingPredefinedMetricSpecification(v)
		}

		params.TargetTrackingScalingPolicyConfiguration = cfg
	}

	return &params, nil
}

func parseAwsAppautoscalingPolicyIdParts(id string) (tenantID, namespace, dimension, resourceId, name string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 4 {
		tenantID, namespace, dimension, resourceId, name = idParts[0], idParts[1], idParts[2], idParts[3], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAwsAppautoscalingPolicy(d *schema.ResourceData, duplo *duplosdk.DuploAwsAutoscalingPolicy) error {
	d.Set("arn", duplo.PolicyARN)
	d.Set("name", duplo.PolicyName)
	d.Set("policy_type", duplo.PolicyType)
	d.Set("resource_id", duplo.ResourceId)
	d.Set("scalable_dimension", duplo.ScalableDimension.Value)
	d.Set("service_namespace", duplo.ServiceNamespace.Value)

	if err := d.Set("step_scaling_policy_configuration", flattenStepScalingPolicyConfiguration(duplo.StepScalingPolicyConfiguration)); err != nil {
		return fmt.Errorf("error setting step_scaling_policy_configuration: %s", err)
	}
	if err := d.Set("target_tracking_scaling_policy_configuration", flattenTargetTrackingScalingPolicyConfiguration(duplo.TargetTrackingScalingPolicyConfiguration)); err != nil {
		return fmt.Errorf("error setting target_tracking_scaling_policy_configuration: %s", err)
	}

	return nil
}

func flattenStepScalingPolicyConfiguration(cfg *duplosdk.DuploStepScalingPolicyConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if len(cfg.AdjustmentType.Value) > 0 {
		m["adjustment_type"] = cfg.AdjustmentType
	}
	m["cooldown"] = cfg.Cooldown
	if len(cfg.MetricAggregationType.Value) > 0 {
		m["metric_aggregation_type"] = cfg.MetricAggregationType
	}
	m["min_adjustment_magnitude"] = cfg.MinAdjustmentMagnitude
	if cfg.StepAdjustments != nil {
		stepAdjustmentsResource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric_interval_lower_bound": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"metric_interval_upper_bound": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"scaling_adjustment": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		}
		m["step_adjustment"] = schema.NewSet(schema.HashResource(stepAdjustmentsResource), flattenAppautoscalingStepAdjustments(cfg.StepAdjustments))
	}

	return []interface{}{m}
}

func flattenAppautoscalingStepAdjustments(adjs *[]duplosdk.DuploStepAdjustment) []interface{} {
	out := make([]interface{}, len(*adjs))

	for i, adj := range *adjs {
		m := make(map[string]interface{})

		m["scaling_adjustment"] = adj.ScalingAdjustment

		m["metric_interval_lower_bound"] = fmt.Sprintf("%g", adj.MetricIntervalLowerBound)
		m["metric_interval_upper_bound"] = fmt.Sprintf("%g", adj.MetricIntervalUpperBound)

		out[i] = m
	}

	return out
}

func flattenTargetTrackingScalingPolicyConfiguration(cfg *duplosdk.DuploTargetTrackingScalingPolicyConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := cfg.CustomizedMetricSpecification; v != nil {
		m["customized_metric_specification"] = flattenCustomizedMetricSpecification(v)
	}

	m["disable_scale_in"] = cfg.DisableScaleIn

	if v := cfg.PredefinedMetricSpecification; v != nil {
		m["predefined_metric_specification"] = flattenPredefinedMetricSpecification(v)
	}

	m["scale_in_cooldown"] = cfg.ScaleInCooldown

	m["scale_out_cooldown"] = cfg.ScaleOutCooldown

	m["target_value"] = cfg.TargetValue

	return []interface{}{m}
}

func flattenCustomizedMetricSpecification(cfg *duplosdk.DuploCustomizedMetricSpecification) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if cfg.Dimensions != nil && len(*cfg.Dimensions) > 0 {
		m["dimensions"] = flattenMetricDimensions(cfg.Dimensions)
	}

	if len(cfg.MetricName) > 0 {
		m["metric_name"] = cfg.MetricName
	}

	if len(cfg.Namespace) > 0 {
		m["namespace"] = cfg.Namespace
	}

	if cfg.Statistic != nil && len(cfg.Statistic.Value) > 0 {
		m["statistic"] = cfg.Statistic.Value
	}

	if len(cfg.Unit) > 0 {
		m["unit"] = cfg.Unit
	}

	return []interface{}{m}
}

func flattenMetricDimensions(ds *[]duplosdk.DuploNameStringValue) []interface{} {
	l := make([]interface{}, len(*ds))
	for i, d := range *ds {
		if ds == nil {
			continue
		}

		m := map[string]interface{}{}

		if len(d.Name) > 0 {
			m["name"] = d.Name
		}

		if len(d.Value) > 0 {
			m["value"] = d.Value
		}

		l[i] = m
	}
	return l
}

func flattenPredefinedMetricSpecification(cfg *duplosdk.DuploPredefinedMetricSpecification) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if cfg.PredefinedMetricType != nil && len(cfg.PredefinedMetricType.Value) > 0 {
		m["predefined_metric_type"] = cfg.PredefinedMetricType.Value
	}

	if len(cfg.ResourceLabel) > 0 {
		m["resource_label"] = cfg.ResourceLabel
	}

	return []interface{}{m}
}

func expandStepScalingPolicyConfiguration(cfg []interface{}) *duplosdk.DuploStepScalingPolicyConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	out := &duplosdk.DuploStepScalingPolicyConfiguration{}

	m := cfg[0].(map[string]interface{})
	if v, ok := m["adjustment_type"]; ok {
		out.AdjustmentType = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	if v, ok := m["cooldown"]; ok {
		out.Cooldown = v.(int)
	}
	if v, ok := m["metric_aggregation_type"]; ok {
		out.MetricAggregationType = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	if v, ok := m["min_adjustment_magnitude"].(int); ok && v > 0 {
		out.MinAdjustmentMagnitude = v
	}
	if v, ok := m["step_adjustment"].(*schema.Set); ok && v.Len() > 0 {
		out.StepAdjustments, _ = expandAppautoscalingStepAdjustments(v.List())
	}

	return out
}

func expandAppautoscalingStepAdjustments(configured []interface{}) (*[]duplosdk.DuploStepAdjustment, error) {
	var adjustments []duplosdk.DuploStepAdjustment

	// Loop over our configured step adjustments and create an array
	// of aws-sdk-go compatible objects. We're forced to convert strings
	// to floats here because there's no way to detect whether or not
	// an uninitialized, optional schema element is "0.0" deliberately.
	// With strings, we can test for "", which is definitely an empty
	// struct value.
	for _, raw := range configured {
		data := raw.(map[string]interface{})
		a := duplosdk.DuploStepAdjustment{
			ScalingAdjustment: data["scaling_adjustment"].(int),
		}
		if data["metric_interval_lower_bound"] != "" {
			bound := data["metric_interval_lower_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_lower_bound must be a float value represented as a string")
				}
				a.MetricIntervalLowerBound = f
			default:
				return nil, fmt.Errorf(
					"metric_interval_lower_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		if data["metric_interval_upper_bound"] != "" {
			bound := data["metric_interval_upper_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, fmt.Errorf(
						"metric_interval_upper_bound must be a float value represented as a string")
				}
				a.MetricIntervalUpperBound = f
			default:
				return nil, fmt.Errorf(
					"metric_interval_upper_bound isn't a string. This is a bug. Please file an issue.")
			}
		}
		adjustments = append(adjustments, a)
	}

	return &adjustments, nil
}

func expandAppautoscalingCustomizedMetricSpecification(configured []interface{}) *duplosdk.DuploCustomizedMetricSpecification {
	spec := &duplosdk.DuploCustomizedMetricSpecification{}

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		if v, ok := data["metric_name"]; ok {
			spec.MetricName = v.(string)
		}

		if v, ok := data["namespace"]; ok {
			spec.Namespace = v.(string)
		}

		if v, ok := data["unit"].(string); ok && v != "" {
			spec.Unit = v
		}

		if v, ok := data["statistic"]; ok {
			spec.Statistic = &duplosdk.DuploStringValue{Value: v.(string)}
		}

		if s, ok := data["dimensions"].(*schema.Set); ok && s.Len() > 0 {
			dimensions := make([]duplosdk.DuploNameStringValue, s.Len())
			for i, d := range s.List() {
				dimension := d.(map[string]interface{})
				dimensions[i] = duplosdk.DuploNameStringValue{
					Name:  dimension["name"].(string),
					Value: dimension["value"].(string),
				}
			}
			spec.Dimensions = &dimensions
		}
	}
	return spec
}

func expandAppautoscalingPredefinedMetricSpecification(configured []interface{}) *duplosdk.DuploPredefinedMetricSpecification {
	spec := duplosdk.DuploPredefinedMetricSpecification{}

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		if v, ok := data["predefined_metric_type"]; ok {
			spec.PredefinedMetricType = &duplosdk.DuploStringValue{Value: v.(string)}
		}

		if v, ok := data["resource_label"].(string); ok && v != "" {
			spec.ResourceLabel = v
		}
	}
	return &spec
}
