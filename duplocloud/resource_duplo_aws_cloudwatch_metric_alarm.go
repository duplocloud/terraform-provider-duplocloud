package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsCloudWatchMetricAlarmSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the cloudwatch metric alarm will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"fullname": {
			Description: "Duplo will generate name of the metric alarm.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"metric_name": {
			Description:  "The name for the alarm's associated metric.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(1, 255),
		},
		"comparison_operator": {
			Description: "The arithmetic operation to use when comparing the specified Statistic and Threshold. The specified Statistic value is used as the first operand. Either of the following is supported: `GreaterThanOrEqualToThreshold`, `GreaterThanThreshold`, `LessThanThreshold`, `LessThanOrEqualToThreshold`",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"GreaterThanOrEqualToThreshold", "GreaterThanThreshold",
				"LessThanThreshold", "LessThanOrEqualToThreshold",
			}, false),
		},
		"evaluation_periods": {
			Description: "The number of periods over which data is compared to the specified threshold.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"namespace": {
			Description: "The namespace for the alarm's associated metric.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringMatch(regexp.MustCompile(`[^:].*`), "must not contain colon characters"),
			),
		},
		"period": {
			Description: "The period in seconds over which the specified `statistic` is applied.",
			Type:        schema.TypeInt,
			Optional:    true,
			ValidateFunc: validation.Any(
				validation.IntInSlice([]int{10, 30}),
				validation.IntDivisibleBy(60),
			),
		},
		"threshold": {
			Description: "The value against which the specified statistic is compared. This parameter is required for alarms based on static thresholds, but should not be used for alarms based on anomaly detection models.",
			Type:        schema.TypeFloat,
			Optional:    true,
		},
		"dimensions": {
			Description: "The dimensions for the alarm's associated metric. For the list of available dimensions see the AWS documentation.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"statistic": {
			Description: "The statistic to apply to the alarm's associated metric. Either of the following is supported: `SampleCount`, `Average`, `Sum`, `Minimum`, `Maximum`",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"SampleCount", "Average",
				"Sum", "Minimum", "Maximum",
			}, false),
		},
	}
}

func resourceAwsCloudWatchMetricAlarm() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_cloudwatch_metric_alarm` manages an AWS cloudwatch metric alarm in Duplo.",

		ReadContext:   resourceAwsCloudWatchMetricAlarmRead,
		CreateContext: resourceAwsCloudWatchMetricAlarmCreate,
		UpdateContext: resourceAwsCloudWatchMetricAlarmUpdate,
		DeleteContext: resourceAwsCloudWatchMetricAlarmDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsCloudWatchMetricAlarmSchema(),
	}
}

func resourceAwsCloudWatchMetricAlarmRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullName, err := parseAwsCloudWatchMetricAlarmIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmRead(%s, %s): start", tenantID, fullName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploCloudWatchMetricAlarmGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s cloudwatch metric alarm'%s': %s", tenantID, fullName, clientErr)
	}

	d.Set("fullname", duplo.Name)

	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmRead(%s, %s): end", tenantID, fullName)
	return nil
}

func resourceAwsCloudWatchMetricAlarmCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	metricName := d.Get("metric_name").(string)
	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmCreate(%s, %s): start", tenantID, metricName)
	c := m.(*duplosdk.Client)

	rq := expandCloudWatchMetricAlarm(d)
	rq.State = "Create"
	err = c.DuploCloudWatchMetricAlarmCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloudwatch metric alarm '%s': %s", tenantID, metricName, err)
	}
	resourceIds := expandAwsCloudWatchMetricAlarmDimensionsResourceIds(d.Get("dimensions").(map[string]interface{}))
	resourceId := strings.Join(resourceIds, "-")
	id := fmt.Sprintf("%s/%s", tenantID, resourceId+"-"+rq.MetricName)
	log.Printf("[TRACE] Get alarm request(%s, %s): start", tenantID, id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloudwatch metric alarm", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchMetricAlarmGet(tenantID, resourceId)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsCloudWatchMetricAlarmRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmCreate(%s, %s): end", tenantID, id)
	return diags
}

func resourceAwsCloudWatchMetricAlarmUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	fullName := d.Get("fullname").(string)
	tenantID, resourceId, err := parseAwsCloudWatchMetricAlarmIdParts(id)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmUpdate(%s, %s): start", tenantID, fullName)

	needsUpdate := needsAwsCloudWatchMetricAlarmUpdate(d)

	if needsUpdate {
		c := m.(*duplosdk.Client)
		rq := expandCloudWatchMetricAlarm(d)
		rq.Name = fullName
		rq.State = "Create"
		err := c.DuploCloudWatchMetricAlarmCreate(rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s cloudwatch metric alarm '%s': %s", tenantID, fullName, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloudwatch metric alarm", id, func() (interface{}, duplosdk.ClientError) {
			return c.DuploCloudWatchMetricAlarmGet(tenantID, resourceId)
		})
		if diags != nil {
			return diags
		}
		diags = resourceAwsCloudWatchMetricAlarmRead(ctx, d, m)
		log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmUpdate(%s, %s): end", tenantID, fullName)
		return diags
	}
	return nil
}

func resourceAwsCloudWatchMetricAlarmDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullName, err := parseAwsCloudWatchMetricAlarmIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmDelete(%s, %s): start", tenantID, fullName)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploCloudWatchMetricAlarmDelete(tenantID, d.Get("fullname").(string))
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s cloudwatch metric alarm '%s': %s", tenantID, fullName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "cloudwatch metric alarm", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchMetricAlarmGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsCloudWatchMetricAlarmDelete(%s, %s): end", tenantID, fullName)
	return nil
}

func expandCloudWatchMetricAlarm(d *schema.ResourceData) *duplosdk.DuploCloudWatchMetricAlarm {
	return &duplosdk.DuploCloudWatchMetricAlarm{
		MetricName:         d.Get("metric_name").(string),
		Statistic:          d.Get("statistic").(string),
		ComparisonOperator: d.Get("comparison_operator").(string),
		Threshold:          d.Get("threshold").(float64),
		Period:             d.Get("period").(int),
		EvaluationPeriods:  d.Get("evaluation_periods").(int),
		TenantId:           d.Get("tenant_id").(string),
		Namespace:          d.Get("namespace").(string),
		Dimensions:         expandAwsCloudWatchMetricAlarmDimensions(d.Get("dimensions").(map[string]interface{})),
	}
}

func expandAwsCloudWatchMetricAlarmDimensions(dims map[string]interface{}) *[]duplosdk.DuploNameStringValue {
	var dimensions []duplosdk.DuploNameStringValue
	for k, v := range dims {
		dimensions = append(dimensions, duplosdk.DuploNameStringValue{
			Name:  k,
			Value: v.(string),
		})
	}
	return &dimensions
}

func expandAwsCloudWatchMetricAlarmDimensionsResourceIds(dims map[string]interface{}) []string {
	var dimensionsResourceIds []string
	for _, v := range dims {
		dimensionsResourceIds = append(dimensionsResourceIds, v.(string))
	}
	return dimensionsResourceIds
}

func parseAwsCloudWatchMetricAlarmIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func needsAwsCloudWatchMetricAlarmUpdate(d *schema.ResourceData) bool {
	return d.HasChange("metric_name") ||
		d.HasChange("comparison_operator") ||
		d.HasChange("evaluation_periods") ||
		d.HasChange("namespace") ||
		d.HasChange("period") ||
		d.HasChange("threshold") ||
		d.HasChange("dimensions") ||
		d.HasChange("statistic")
}
