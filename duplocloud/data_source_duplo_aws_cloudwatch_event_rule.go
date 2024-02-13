package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsCloudWatchEventRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDuploCloudWatchEventRuleRead,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

// Read resource
func dataSourceDuploCloudWatchEventRuleRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*duplosdk.Client)
	tenantID := ""
	ruleName := ""
	if v, ok := d.GetOk("tenant_id"); ok && v != nil {
		tenantID = v.(string)
	}

	if v, ok := d.GetOk("name"); ok && v != nil {
		ruleName = v.(string)
	}

	r, err := c.DuploCloudWatchEventRuleGet(tenantID, ruleName)
	if err != nil {
		return fmt.Errorf("failed to read CloudWatchEventRule: %s", err)

	}
	//resp, err := c.DuploCloudWatchEventRuleList(tenantID)
	//if err != nil {
	//	return fmt.Errorf("failed to read CloudWatchEventRule: %s", err)
	//
	//}
	log.Printf("Fetched data \n %+v", r)
	if r != nil {
		//for _, r := range *resp {
		d.Set("id", tenantID+"/"+r.Name)
		d.Set("name", r.Name)
		d.Set("arn", r.Arn)
		d.Set("schedule_expression", r.ScheduleExpression)
		d.Set("event_pattern", r.EventPattern)
		d.Set("description", r.Description)
		d.Set("role_arn", r.RoleArn)
		d.Set("event_bus_name", r.EventBusName)
		d.Set("state", r.State.Value)
		//}
	}
	log.Printf("[TRACE] dataSourceDuploCloudWatchEventRuleRead ******** end")

	return nil
}
