package duplocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsSqsQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sqs_queue` retrieves an SQS queue in a Duplo tenant.",
		ReadContext: dataSourceAwsSqsQueueRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the SQS queue resides in.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The short name of the SQS queue. Add `fifo` suffix for FIFO queues.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"arn": {
				Description: "The ARN of the SQS queue.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"fullname": {
				Description: "The full name of the SQS queue.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"url": {
				Description: "The URL for the SQS queue.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"fifo_queue": {
				Description: "Boolean designating a FIFO queue.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"message_retention_seconds": {
				Description: "The number of seconds Amazon SQS retains a message.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"visibility_timeout_seconds": {
				Description: "The visibility timeout for the queue.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"content_based_deduplication": {
				Description: "Enables content-based deduplication for FIFO queues.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"deduplication_scope": {
				Description: "Specifies whether message deduplication occurs at the message group or queue level.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"fifo_throughput_limit": {
				Description: "Specifies whether the FIFO queue throughput quota applies to the entire queue or per message group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"delay_seconds": {
				Description: "Postpone the delivery of new messages to consumers for a number of seconds.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"dead_letter_queue_configuration": {
				Description: "SQS dead letter queue configuration.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_sqs_dlq_name": {
							Description: "Name of the SQS queue meant to be the target dead letter queue.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"max_message_receive_attempts": {
							Description: "Maximum number of processing attempts before moving to the dead letter queue.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
			"receive_wait_time_seconds": {
				Description: "The time for which a ReceiveMessage call will wait for a message to arrive.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

func dataSourceAwsSqsQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] dataSourceAwsSqsQueueRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	// Get the prefix for the tenant
	prefix, err := c.GetDuploServicesPrefix(tenantID, "")
	if err != nil {
		return diag.FromErr(err)
	}
	searchName := name

	// Build the full name
	fullName := prefix + "-" + searchName

	// Get the queue from Duplo
	queue, clientErr := c.DuploSQSQueueGetV3(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return diag.Errorf("dataSourceAwsSqsQueueRead: SQS queue not found (tenant: %s, queue: %s)", tenantID, name)
		}
		return diag.Errorf("dataSourceAwsSqsQueueRead: Unable to retrieve SQS queue (tenant: %s, queue: %s: error: %s)", tenantID, name, clientErr)
	}

	if queue == nil {
		return diag.Errorf("dataSourceAwsSqsQueueRead: SQS queue not found (tenant: %s, queue: %s)", tenantID, name)
	}

	// Set the ID and data
	id := fmt.Sprintf("%s/%s", tenantID, queue.Url)
	d.SetId(id)
	flattenSqsQueueData(d, tenantID, name, fullName, queue)

	log.Printf("[TRACE] dataSourceAwsSqsQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func flattenSqsQueueData(d *schema.ResourceData, tenantID string, name string, fullName string, queue *duplosdk.DuploSQSQueue) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", fullName)
	d.Set("arn", queue.Arn)
	d.Set("url", queue.Url)
	d.Set("fifo_queue", queue.QueueType == 1)
	d.Set("content_based_deduplication", queue.ContentBasedDeduplication)
	d.Set("message_retention_seconds", queue.MessageRetentionPeriod)
	d.Set("visibility_timeout_seconds", queue.VisibilityTimeout)
	d.Set("delay_seconds", queue.DelaySeconds)
	d.Set("receive_wait_time_seconds", queue.ReceiveMessageWaitTimeSeconds)

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
	} else {
		// Clear FIFO-specific attributes for non-FIFO queues
		d.Set("deduplication_scope", "")
		d.Set("fifo_throughput_limit", "")
	}

	if queue.DeadLetterTargetQueueName != "" {
		dlqConfig := make(map[string]interface{})
		dlqConfig["target_sqs_dlq_name"] = queue.DeadLetterTargetQueueName
		dlqConfig["max_message_receive_attempts"] = queue.MaxMessageTimesReceivedBeforeDeadLetterQueue
		d.Set("dead_letter_queue_configuration", []interface{}{dlqConfig})
	} else {
		// Clear dead letter queue configuration when no DLQ is configured
		d.Set("dead_letter_queue_configuration", []interface{}{})
	}
}
