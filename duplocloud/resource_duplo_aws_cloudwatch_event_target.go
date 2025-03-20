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

func awsCloudWatchEventTargetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the cloudwatch event target will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"rule_name": {
			Description: "The name of the rule you want to add targets to.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"target_arn": {
			Description: "The Amazon Resource Name (ARN) of the target.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"target_id": {
			Description: "The unique target assignment ID.",
			Required:    true,
			Type:        schema.TypeString,
			ForceNew:    true,
		},
		"role_arn": {
			Description: "The Amazon Resource Name (ARN) associated with the role that is used for target invocation.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"event_bus_name": {
			Description: "The event bus to associate with the rule. If you omit this, the default event bus is used.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"input": {
			Description: "Valid JSON text passed to the target. ",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
	}
}

func resourceAwsCloudWatchEventTarget() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_cloudwatch_event_target` manages an AWS cloudwatch event target in Duplo.",

		ReadContext:   resourceAwsCloudWatchEventTargetRead,
		CreateContext: resourceAwsCloudWatchEventTargetCreate,
		UpdateContext: resourceAwsCloudWatchEventTargetUpdate,
		DeleteContext: resourceAwsCloudWatchEventTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsCloudWatchEventTargetSchema(),
	}
}

func resourceAwsCloudWatchEventTargetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, ruleName, targetId, err := parseAwsCloudWatchEventTargetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetRead(%s, %s, %s): start", tenantID, ruleName, targetId)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploCloudWatchEventTargetGet(tenantID, ruleName, targetId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s cloudwatch event target'%s': %s", tenantID, targetId, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("arn", duplo.Arn)
	d.Set("role_arn", duplo.RoleArn)

	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetRead(%s, %s, %s): end", tenantID, ruleName, targetId)
	return nil
}

func resourceAwsCloudWatchEventTargetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	ruleName := d.Get("rule_name").(string)
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetCreate(%s, %s): start", tenantID, ruleName)
	c := m.(*duplosdk.Client)

	rq := expandCloudWatchEventTarget(d)
	_, err = c.DuploCloudWatchEventTargetsCreate(tenantID, &duplosdk.DuploCloudWatchEventTargets{
		Rule:         ruleName,
		EventBusName: d.Get("event_bus_name").(string),
		Targets:      &[]duplosdk.DuploCloudWatchEventTarget{*rq},
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloudwatch event target '%s': %s", tenantID, rq.Id, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, ruleName, rq.Id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloudwatch event rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchEventTargetGet(tenantID, ruleName, rq.Id)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsCloudWatchEventTargetRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetCreate(%s, %s): end", tenantID, ruleName)
	return diags
}

func resourceAwsCloudWatchEventTargetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsCloudWatchEventTargetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, ruleName, targetId, err := parseAwsCloudWatchEventTargetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetDelete(%s, %s, %s): start", tenantID, ruleName, targetId)

	c := m.(*duplosdk.Client)
	_, clientErr := c.DuploCloudWatchEventTargetsDelete(tenantID, duplosdk.DuploCloudWatchEventTargetsDeleteReq{
		Rule: ruleName,
		Ids:  []string{targetId},
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s cloudwatch event target '%s': %s", tenantID, ruleName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "cloudwatch event target", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploCloudWatchEventTargetGet(tenantID, ruleName, targetId)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetDelete(%s, %s, %s): end", tenantID, ruleName, targetId)
	return nil
}

func expandCloudWatchEventTarget(d *schema.ResourceData) *duplosdk.DuploCloudWatchEventTarget {
	return &duplosdk.DuploCloudWatchEventTarget{
		Id:    d.Get("target_id").(string),
		Arn:   d.Get("target_arn").(string),
		Input: d.Get("input").(string),
	}
}

func parseAwsCloudWatchEventTargetIdParts(id string) (tenantID, ruleName, targetId string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, ruleName, targetId = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
