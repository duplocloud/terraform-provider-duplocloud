package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsCloudWatchEventRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the cloudwatch event rule will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the event rule.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the event rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the event rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"schedule_expression": {
			Description: "The scheduling expression. For example, `cron(0 20 * * ? *)` or `rate(5 minutes)`. At least one of `schedule_expression` or `event_pattern` is required.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"event_pattern": {
			Description: "The event pattern described a JSON object. At least one of `schedule_expression` or `event_pattern` is required.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"description": {
			Description: "The description of the rule.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"role_arn": {
			Description: "The Amazon Resource Name (ARN) associated with the role that is used for target invocation.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"tag": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},

		"event_bus_name": {
			Description: "The event bus to associate with this rule. If you omit this, the default event bus is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"state": {
			Description: "Whether the rule should be enabled or disabled.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "ENABLED",
			ValidateFunc: validation.StringInSlice([]string{
				"DISABLED",
				"ENABLED",
			}, false),
		},
	}
}

func resourceAwsCloudWatchEventRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_cloudwatch_event_rule` manages an AWS cloudwatch event rule in Duplo.",

		ReadContext:   resourceAwsCloudWatchEventRuleRead,
		CreateContext: resourceAwsCloudWatchEventRuleCreate,
		UpdateContext: resourceAwsCloudWatchEventRuleUpdate,
		DeleteContext: resourceAwsCloudWatchEventRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsCloudWatchEventRuleSchema(),
	}
}

func resourceAwsCloudWatchEventRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullName, err := parseAwsCloudWatchEventRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleRead(%s, %s): start", tenantID, fullName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploCloudWatchEventRuleGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s cloudwatch event rule'%s': %s", tenantID, fullName, clientErr)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("fullname", fullName)
	d.Set("arn", duplo.Arn)
	d.Set("event_bus_name", duplo.EventBusName)
	d.Set("role_arn", duplo.RoleArn)
	d.Set("schedule_expression", duplo.ScheduleExpression)
	d.Set("event_pattern", duplo.EventPattern)
	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleRead(%s, %s): end", tenantID, fullName)
	return nil
}

func resourceAwsCloudWatchEventRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)

	rq := expandCloudWatchEventRule(d)

	_, err = c.DuploCloudWatchEventRuleCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloudwatch event rule '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloudwatch event rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchEventRuleGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsCloudWatchEventRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsCloudWatchEventRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)

	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleUpdate(%s, %s): start", tenantID, fullName)

	needsUpdate := needsAwsCloudWatchEventRuleUpdate(d)

	if needsUpdate {
		c := m.(*duplosdk.Client)
		rq := expandCloudWatchEventRule(d)

		_, err = c.DuploCloudWatchEventRuleCreate(tenantID, rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s cloudwatch event rule '%s': %s", tenantID, name, err)
		}

		id := fmt.Sprintf("%s/%s", tenantID, fullName)
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloudwatch event rule", id, func() (interface{}, duplosdk.ClientError) {
			return c.DuploCloudWatchEventRuleGet(tenantID, fullName)
		})
		if diags != nil {
			return diags
		}
		diags = resourceAwsCloudWatchEventRuleRead(ctx, d, m)
		log.Printf("[TRACE] resourceAwsCloudWatchEventRuleUpdate(%s, %s): end", tenantID, name)
		return diags
	}
	return nil
}

func resourceAwsCloudWatchEventRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullName, err := parseAwsCloudWatchEventRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleDelete(%s, %s): start", tenantID, fullName)

	c := m.(*duplosdk.Client)
	_, clientErr := c.DuploCloudWatchEventRuleDelete(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s cloudwatch event rule '%s': %s", tenantID, fullName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "cloudwatch event rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchEventRuleGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsCloudWatchEventRuleDelete(%s, %s): end", tenantID, fullName)
	return nil
}

func expandCloudWatchEventRule(d *schema.ResourceData) *duplosdk.DuploCloudWatchEventRule {
	return &duplosdk.DuploCloudWatchEventRule{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Tags:               keyValueFromState("tag", d),
		ScheduleExpression: d.Get("schedule_expression").(string),
		EventPattern:       d.Get("event_pattern").(string),
		State:              d.Get("state").(string),
		RoleArn:            d.Get("role_arn").(string),
		EventBusName:       d.Get("event_bus_name").(string),
	}
}

func parseAwsCloudWatchEventRuleIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func needsAwsCloudWatchEventRuleUpdate(d *schema.ResourceData) bool {
	return d.HasChange("schedule_expression") ||
		d.HasChange("event_pattern") ||
		d.HasChange("description") ||
		d.HasChange("role_arn") ||
		d.HasChange("event_bus_name") ||
		d.HasChange("state") ||
		d.HasChange("tag")
}
