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

func duploAwsSqsQueueSchema() map[string]*schema.Schema {
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

func resourceAwsSqsQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sqs_queue` manages a SQS queuet in Duplo.",

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

	fullname, err := c.ExtractSqsFullname(tenantID, url)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): start", tenantID, url)

	queue, clientErr := c.TenantGetSQSQueue(tenantID, url)
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

	d.Set("tenant_id", tenantID)
	d.Set("url", queue.Name)
	d.Set("fullname", fullname)
	name, _ := duplosdk.UnprefixName(prefix, fullname)
	d.Set("name", name)
	// d.Set("fifo_queue", false) // TODO - Backend is not persisting this value.
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
	err = c.DuploSQSQueueCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s SQS queue '%s': %s", tenantID, name, err)
	}

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, name), func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetSqsQueueByName(tenantID, name)
	})
	if diags != nil {
		return diags
	}

	resp, err := c.TenantGetSqsQueueByName(tenantID, name)
	if err != nil {
		return diag.Errorf("Error reading tenant %s SQS queue '%s': %s", tenantID, name, err)
	}
	id := fmt.Sprintf("%s/%s", tenantID, resp.Name)
	d.SetId(id)

	diags = resourceAwsSqsQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSqsQueueCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsSqsQueueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	clientErr := c.DuploSQSQueueDelete(tenantID, url)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s sqs queue '%s': %s", tenantID, url, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SQS Queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetSQSQueue(tenantID, url)
	})
	if diag != nil {
		return diag
	}

	// Wait for 60 seconds.
	time.Sleep(time.Duration(60) * time.Second)

	log.Printf("[TRACE] resourceAwsSqsQueueDelete(%s, %s): end", tenantID, url)
	return nil
}

func expandAwsSqsQueue(d *schema.ResourceData) *duplosdk.DuploSQSQueue {
	queueType := 0
	if isFifo, ok := d.GetOk("fifo_queue"); ok && isFifo.(bool) {
		queueType = 1
	}
	return &duplosdk.DuploSQSQueue{
		Name:      d.Get("name").(string),
		QueueType: queueType,
	}
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
