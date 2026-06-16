package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

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
			Description:   "Valid JSON text passed to the target. Conflicts with `input_transformer`.",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"input_transformer"},
		},
		"input_transformer": {
			Description:   "Settings used to transform the matched event before passing it to the target. Conflicts with `input`.",
			Type:          schema.TypeList,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"input"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"input_paths": {
						Description: "Map of variable names to JSONPath expressions that extract values from the event. The variable names can be referenced in `input_template` using `<name>` syntax.",
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"input_template": {
						Description: "Template that defines the payload passed to the target. Variables defined in `input_paths` are referenced using `<name>` syntax.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
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
			log.Printf("[TRACE] resourceAwsCloudWatchEventTargetRead(%s, %s, %s): not found", tenantID, ruleName, targetId)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s cloudwatch event target'%s': %s", tenantID, targetId, clientErr)
	}
	if duplo == nil {
		log.Printf("[TRACE] resourceAwsCloudWatchEventTargetRead(%s, %s, %s): not found", tenantID, ruleName, targetId)
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("rule_name", ruleName)
	d.Set("target_id", duplo.Id)
	d.Set("target_arn", duplo.Arn)
	d.Set("role_arn", duplo.RoleArn)
	d.Set("input", duplo.Input)
	d.Set("input_transformer", flattenCloudWatchInputTransformer(duplo.InputTransformer))

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
	err = c.DuploCloudWatchEventTargetsCreate(tenantID, &duplosdk.DuploCloudWatchEventTargets{
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
	tenantID := d.Get("tenant_id").(string)
	ruleName := d.Get("rule_name").(string)
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetUpdate(%s, %s): start", tenantID, ruleName)
	c := m.(*duplosdk.Client)

	rq := expandCloudWatchEventTarget(d)
	err := c.DuploCloudWatchEventTargetsCreate(tenantID, &duplosdk.DuploCloudWatchEventTargets{
		Rule:         ruleName,
		EventBusName: d.Get("event_bus_name").(string),
		Targets:      &[]duplosdk.DuploCloudWatchEventTarget{*rq},
	})
	if err != nil {
		return diag.Errorf("Error updating tenant %s cloudwatch event target '%s': %s", tenantID, rq.Id, err)
	}

	diags := resourceAwsCloudWatchEventTargetRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetUpdate(%s, %s): end", tenantID, ruleName)
	return diags
}

func resourceAwsCloudWatchEventTargetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, ruleName, targetId, err := parseAwsCloudWatchEventTargetIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudWatchEventTargetDelete(%s, %s, %s): start", tenantID, ruleName, targetId)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploCloudWatchEventTargetsDelete(tenantID, duplosdk.DuploCloudWatchEventTargetsDeleteReq{
		Rule: ruleName,
		Ids:  []string{targetId},
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsCloudWatchEventTargetDelete(%s, %s, %s): not found", tenantID, ruleName, targetId)
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
	target := &duplosdk.DuploCloudWatchEventTarget{
		Id:      d.Get("target_id").(string),
		Arn:     d.Get("target_arn").(string),
		RoleArn: d.Get("role_arn").(string),
	}

	// AWS EventBridge accepts at most one of Input / InputPath / InputTransformer per target.
	// Since `input` is Computed, its old value can linger in state when the user switches
	// to `input_transformer`, so we must ensure we never send both (otherwise PutTargets
	// returns a 400).
	if v, ok := d.GetOk("input_transformer"); ok {
		list := v.([]interface{})
		if len(list) > 0 && list[0] != nil {
			m := list[0].(map[string]interface{})
			it := &duplosdk.DuploCloudWatchInputTransformer{
				InputTemplate: m["input_template"].(string),
			}
			if paths, ok := m["input_paths"].(map[string]interface{}); ok && len(paths) > 0 {
				it.InputPathsMap = make(map[string]string, len(paths))
				for k, val := range paths {
					it.InputPathsMap[k] = val.(string)
				}
			}
			target.InputTransformer = it
		}
	} else {
		target.Input = d.Get("input").(string)
	}

	return target
}

func flattenCloudWatchInputTransformer(it *duplosdk.DuploCloudWatchInputTransformer) []interface{} {
	if it == nil {
		return nil
	}
	m := map[string]interface{}{
		"input_template": it.InputTemplate,
	}
	if len(it.InputPathsMap) > 0 {
		paths := make(map[string]interface{}, len(it.InputPathsMap))
		for k, v := range it.InputPathsMap {
			paths[k] = v
		}
		m["input_paths"] = paths
	}
	return []interface{}{m}
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
