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

func duploAwsSqsQueueSchemaV2() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the SQS queue will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the queue. Queue names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and must be between 1 and 80 characters long.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"fifo_queue": {
			Description: "Boolean designating a FIFO queue. If not set, it defaults to `false` making it standard.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
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
	}
}

func resourceAwsSqsQueueV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sqs_queue_v2` manages a SQS queue in Duplo.",

		ReadContext:   resourceAwsSqsQueueReadV2,
		CreateContext: resourceAwsSqsQueueCreateV2,
		UpdateContext: resourceAwsSqsQueueUpdateV2,
		DeleteContext: resourceAwsSqsQueueDeleteV2,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsSqsQueueSchemaV2(),
	}
}

func resourceAwsSqsQueueReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullname, err := parseAwsSqsQueueIdPartsV2(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	log.Printf("[TRACE] resourceAwsSqsQueueReadV2(%s, %s): start", tenantID, fullname)

	queue, clientErr := c.DuploSQSQueueGetV2(tenantID, fullname)
	if queue == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s sqs queue %s : %s", tenantID, fullname, clientErr)
	}

	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("tenant_id", tenantID)
	d.Set("fullname", queue.Name)
	d.Set("url", queue.Url)
	d.Set("fifo_queue", queue.QueueType == 1)
	d.Set("content_based_deduplication", queue.ContentBasedDeduplication)
	d.Set("message_retention_seconds", queue.MessageRetentionPeriod)
	d.Set("visibility_timeout_seconds", queue.VisibilityTimeout)
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

	name, _ := duplosdk.UnprefixName(prefix, fullname)
	if queue.QueueType == 1 {
		name = strings.TrimSuffix(name, ".fifo")
	}
	d.Set("name", name)
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsSqsQueueCreateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsSqsQueueCreateV2(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAwsSqsQueueV2(d)
	resp, err := c.DuploSQSQueueCreateV2(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s SQS queue '%s': %s", tenantID, name, err)
	}
	fullname := resp.Name
	if resp.QueueType == 1 {
		fullname += ".fifo"
	}
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, fullname), func() (interface{}, duplosdk.ClientError) {
		return c.DuploSQSQueueGetV2(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullname)
	d.SetId(id)

	diags = resourceAwsSqsQueueReadV2(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSqsQueueCreateV2(%s, %s): end", tenantID, fullname)
	return diags
}

func resourceAwsSqsQueueUpdateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChanges("message_retention_seconds", "visibility_timeout_seconds", "content_based_deduplication", "deduplication_scope", "fifo_throughput_limit") {
		var err error

		tenantID := d.Get("tenant_id").(string)
		fullname := d.Get("fullname").(string)
		url := d.Get("url").(string)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdateV2(%s, %s): start", tenantID, fullname)
		c := m.(*duplosdk.Client)

		rq := expandAwsSqsQueueV2(d)
		rq.Name = fullname
		rq.Url = url
		_, err = c.DuploSQSQueueUpdateV2(tenantID, rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s SQS queue '%s': %s", tenantID, fullname, err)
		}
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, fullname), func() (interface{}, duplosdk.ClientError) {
			return c.DuploSQSQueueGetV2(tenantID, fullname)
		})
		if diags != nil {
			return diags
		}

		id := fmt.Sprintf("%s/%s", tenantID, fullname)
		d.SetId(id)

		diags = resourceAwsSqsQueueReadV2(ctx, d, m)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdateV2(%s, %s): end", tenantID, fullname)
		return diags
	}
	return nil
}

func resourceAwsSqsQueueDeleteV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullname, err := parseAwsSqsQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueDeleteV2(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploSQSQueueDeleteV2(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s sqs queue '%s': %s", tenantID, fullname, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SQS Queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploSQSQueueGetV2(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsSqsQueueDeleteV2(%s, %s): end", tenantID, fullname)
	return nil
}

func expandAwsSqsQueueV2(d *schema.ResourceData) *duplosdk.DuploSQSQueue {
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
	return req
}

func parseAwsSqsQueueIdPartsV2(id string) (tenantID, fullname string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, fullname = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
