package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsSqsQueueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"arn": {
			Description: "The ARN of the SQS queue.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tenant_id": {
			Description:  "The GUID of the tenant that the SQS queue will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the queue. Queue names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and have up to 80 characters long which is inclusive of duplo prefix (duploservices-{tenant_name}-) appended by the system.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			//	ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]{0,46}$"), "invalid name format. Name can contain alphabet, numbers, hyphen and underscores, it should start with a alphabet and have up to 80 character long inclusive of duploservices-{max tenant name length}-"),
		},
		"fifo_queue": {
			Description: "Boolean designating a FIFO queue. If not set, it defaults to `false` making it standard.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"message_retention_seconds": {
			Description:  "The number of seconds Amazon SQS retains a message. Integer representing seconds, from 60 (1 minute) to 1209600 (14 days).",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(60, 1209600),
		},
		"visibility_timeout_seconds": {
			Description:  "The visibility timeout for the queue. An integer from 0 to 43200 (12 hours).",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(0, 43200),
		},
		"content_based_deduplication": {
			Description: "Enables content-based deduplication for FIFO queues.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"deduplication_scope": {
			Description: "Specifies whether message deduplication occurs at the message group or queue level. Valid values are `messageGroup` and `queue`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"messageGroup",
				"queue",
			}, false),
		},
		"fifo_throughput_limit": {
			Description: "Specifies whether the FIFO queue throughput quota applies to the entire queue or per message group. Valid values are `perQueue` (default) and `perMessageGroupId`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"perQueue",
				"perMessageGroupId",
			}, false),
		},
		"fullname": {
			Description: "The full name of the SQS queue.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"url": {
			Description: "The URL for the created Amazon SQS queue.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"delay_seconds": {
			Description:  "Postpone the delivery of new messages to consumers for a number of seconds seconds range [0-900]",
			Type:         schema.TypeInt,
			Computed:     true,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 900),
		},
		"dead_letter_queue_configuration": {
			Description: "SQS configuration for the SQS resource",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"target_sqs_dlq_name": {
						Description:  "Name of the SQS queue meant to be the target dead letter queue for this SQS resource (queues must belong to same tenant)",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"max_message_receive_attempts": {
						Description:  "Maximum number of processing attempts for a given message before it is moved to the dead letter queue",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 1000),
					},
				},
			},
		},
	}
}

func resourceAwsSqsQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sqs_queue` manages a SQS queue in Duplo.",

		ReadContext:   resourceAwsSqsQueueRead,
		CreateContext: resourceAwsSqsQueueCreate,
		UpdateContext: resourceAwsSqsQueueUpdate,
		DeleteContext: resourceAwsSqsQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsSqsQueueSchema(),
	}
}

func resourceAwsSqsQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	tenantID, url, err := parseAwsSqsQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	fullname, err := c.ExtractSqsFullname(tenantID, url)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): start", tenantID, url)

	queue, clientErr := c.DuploSQSQueueGetV3(tenantID, fullname)
	if queue == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s sqs queue %s : %s", tenantID, url, clientErr)
	}

	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", queue.Arn)
	d.Set("tenant_id", tenantID)
	d.Set("url", queue.Url)
	d.Set("fullname", fullname)
	d.Set("fifo_queue", queue.QueueType == 1)
	d.Set("content_based_deduplication", queue.ContentBasedDeduplication)
	d.Set("message_retention_seconds", queue.MessageRetentionPeriod)
	d.Set("visibility_timeout_seconds", queue.VisibilityTimeout)
	d.Set("delay_seconds", queue.DelaySeconds)
	if queue.QueueType == 1 {
		if queue.DeduplicationScope == 0 {
			d.Set("deduplication_scope", "queue")
		} else {
			d.Set("deduplication_scope", "messageGroup")
		}
		if queue.FifoThroughputLimit == 0 {
			d.Set("fifo_throughput_limit", "perQueue")
		} else {
			d.Set("fifo_throughput_limit", "perMessageGroupId")
		}
	}
	if queue.DeadLetterTargetQueueName != "" {
		dlq_config := make(map[string]interface{})
		dlq_config["target_sqs_dlq_name"] = queue.DeadLetterTargetQueueName
		dlq_config["max_message_receive_attempts"] = queue.MaxMessageTimesReceivedBeforeDeadLetterQueue
		d.Set("dead_letter_queue_configuration", []interface{}{dlq_config})
	}

	name, _ := duplosdk.UnwrapName(prefix, accountID, fullname, true)
	if queue.QueueType == 1 {
		name = strings.TrimSuffix(name, ".fifo")
	}
	d.Set("name", name)
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsSqsQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsSqsQueueCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	rq := expandAwsSqsQueue(d)
	err = validateSQSName(c, tenantID, name, rq.QueueType == 1)
	if err != nil {
		return diag.Errorf("%s", err)
	}

	resp, err := c.DuploSQSQueueCreateV3(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s SQS queue '%s': %s", tenantID, name, err)
	}
	fullname, err := c.ExtractSqsFullname(tenantID, resp.Url)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, name), func() (interface{}, duplosdk.ClientError) {
		resp, err = c.DuploSQSQueueGetV3(tenantID, fullname)
		// wait for an Arn to be available
		if err == nil && resp != nil && resp.Arn == "" {
			return nil, nil
		}
		return c.DuploSQSQueueGetV3(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}
	id := fmt.Sprintf("%s/%s", tenantID, resp.Url)
	d.SetId(id)

	diags = resourceAwsSqsQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSqsQueueCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsSqsQueueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	if d.HasChanges("message_retention_seconds", "visibility_timeout_seconds", "content_based_deduplication", "deduplication_scope", "fifo_throughput_limit", "delay_seconds", "dead_letter_queue_configuration") {
		var err error

		tenantID := d.Get("tenant_id").(string)
		fullname := d.Get("fullname").(string)
		url := d.Get("url").(string)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdate(%s, %s): start", tenantID, fullname)
		c := m.(*duplosdk.Client)

		rq := expandAwsSqsQueue(d)
		rq.Name = fullname
		rq.Url = url
		_, err = c.DuploSQSQueueUpdateV3(tenantID, rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s SQS queue '%s': %s", tenantID, fullname, err)
		}
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, fullname), func() (interface{}, duplosdk.ClientError) {
			return c.DuploSQSQueueGetV3(tenantID, fullname)
		})
		if diags != nil {
			return diags
		}

		diags = resourceAwsSqsQueueRead(ctx, d, m)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdate(%s, %s): end", tenantID, fullname)
		return diags
	}
	return nil
}

func resourceAwsSqsQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, url, err := parseAwsSqsQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueDelete(%s, %s): start", tenantID, url)

	c := m.(*duplosdk.Client)

	fullname, err := c.ExtractSqsFullname(tenantID, url)
	if err != nil {
		return diag.FromErr(err)
	}

	clientErr := c.DuploSQSQueueDeleteV3(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s sqs queue '%s': %s", tenantID, fullname, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SQS Queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploSQSQueueGetV3(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsSqsQueueDelete(%s, %s): end", tenantID, fullname)
	return nil
}

func expandAwsSqsQueue(d *schema.ResourceData) *duplosdk.DuploSQSQueue {
	queueType := 0
	if isFifo, ok := d.GetOk("fifo_queue"); ok && isFifo.(bool) {
		queueType = 1
	}
	req := &duplosdk.DuploSQSQueue{
		Name:                      d.Get("name").(string),
		QueueType:                 queueType,
		ContentBasedDeduplication: d.Get("content_based_deduplication").(bool),
	}
	if v, ok := d.GetOk("message_retention_seconds"); ok && v != nil {
		req.MessageRetentionPeriod = d.Get("message_retention_seconds").(int)
	}
	if v, ok := d.GetOk("visibility_timeout_seconds"); ok && v != nil {
		req.VisibilityTimeout = d.Get("visibility_timeout_seconds").(int)
	}
	if v, ok := d.GetOk("delay_seconds"); ok && v != nil {
		req.DelaySeconds = d.Get("delay_seconds").(int)
	}
	if v, ok := d.GetOk("deduplication_scope"); ok && v != nil {
		if v.(string) == "messageGroup" {
			req.DeduplicationScope = 1
		} else if v.(string) == "queue" {
			req.DeduplicationScope = 0
		}
	}
	if v, ok := d.GetOk("fifo_throughput_limit"); ok && v != nil {
		if v.(string) == "perMessageGroupId" {
			req.FifoThroughputLimit = 1
		} else if v.(string) == "perQueue" {
			req.FifoThroughputLimit = 0
		}
	}
	if value, ok := d.GetOk("dead_letter_queue_configuration"); ok && len(value.([]interface{})) > 0 {
		dlqConfig := value.([]interface{})[0].(map[string]interface{})
		req.DeadLetterTargetQueueName = dlqConfig["target_sqs_dlq_name"].(string)
		req.MaxMessageTimesReceivedBeforeDeadLetterQueue = dlqConfig["max_message_receive_attempts"].(int)
	}
	return req
}

func parseAwsSqsQueueIdParts(id string) (tenantID, url string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, url = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func validateSQSName(c *duplosdk.Client, tId, name string, fifo bool) error {
	rp, cerr := c.TenantGetV2(tId)
	if cerr != nil {
		return fmt.Errorf("%s", cerr.Error())
	}
	allowedLen := 80 - (15 + len([]rune(rp.AccountName)) - 1)

	if fifo {
		allowedLen = allowedLen - len([]rune(".fifo"))
	}
	regString := "^[a-zA-Z][a-zA-Z0-9_-]{0," + strconv.Itoa(allowedLen) + "}$"
	b, err := regexp.MatchString(regString, name)
	if err != nil {
		return err
	}

	if !b {
		return fmt.Errorf("invalid name format. Queue names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and have up to 80 characters long which is inclusive of duplo prefix (duploservices-{tenant_name}-) added by the system")
	}
	return nil
}
