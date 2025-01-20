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

func s3EventNotificationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"bucket_name": {
			Description: "The fully qualified duplo name of the S3 bucket.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"destination_name": {
			Description: "The fully qualified duplo name of specified destination type.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"destination_type": {
			Description:  "The type of destination where event notification to be published.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"sns", "sqs", "lambda"}, false),
		},

		"destination_arn": {
			Description: "The ARN of the specified destination type.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_event_bridge": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"event_types": {
			Description: `Event types: 
			's3:TestEvent'<br>
			's3:ObjectCreated:*'<br>
			's3:ObjectCreated:Put'<br>
			's3:ObjectCreated:Post'<br>
			's3:ObjectCreated:Copy'<br>
			's3:ObjectCreated:CompleteMultipartUpload'<br>
			's3:ObjectRemoved:*'<br>
			's3:ObjectRemoved:Delete'<br>
			's3:ObjectRemoved:DeleteMarkerCreated'<br>
			's3:ObjectRestore:*'<br>
			's3:ObjectRestore:Post'<br>`,
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"s3:TestEvent", "s3:ObjectCreated:*", "s3:ObjectCreated:Put",
					"s3:ObjectCreated:Post", "s3:ObjectCreated:Copy", "s3:ObjectCreated:CompleteMultipartUpload",
					"s3:ObjectRemoved:*", "s3:ObjectRemoved:Delete", "s3:ObjectRemoved:DeleteMarkerCreated", "s3:ObjectRestore:*",
					"s3:ObjectRestore:Post"}, false),
			},
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceAWSS3EventNotification() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_s3_event_notificaiton` manages event configuration related to s3 bucket in Duplo.",
		ReadContext:   resourceS3EventNotificationRead,
		CreateContext: resourceS3EventNotificationCreateOrUpdate,
		UpdateContext: resourceS3EventNotificationCreateOrUpdate,
		DeleteContext: resourceS3EventNotificationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: s3EventNotificationSchema(),
	}
}

// READ resource
func resourceS3EventNotificationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3EventNotificationRead: Invalid resource (ID: %s)", id)
	}
	tenantID, name := idParts[0], idParts[1]

	c := m.(*duplosdk.Client)

	// Get the object from Duplo
	duplo, err := c.GetS3EventNotification(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceS3EventNotificationRead: Unable to retrieve s3 event notification (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}

	// Set simple fields first.
	flattenEventNotification(d, name, duplo)

	log.Printf("[TRACE] resourceS3BucketRead ******** end")
	return nil
}

// CREATE resource
func resourceS3EventNotificationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3EventNotificationCreateOrUpdate ******** start")
	name := d.Get("bucket_name").(string)
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	rq := expandEventNotification(d)

	// Post the object to Duplo
	err := c.UpdateS3EventNotification(tenantID, name, *rq)
	if err != nil {
		return diag.Errorf("resourceS3EventNotificationCreateOrUpdate: Unable to create or update s3 event notification using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/event_notification", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "S3 bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetS3EventNotification(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Set simple fields first.
	resourceS3EventNotificationRead(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketCreate ******** end")
	return diags
}

func flattenEventNotification(d *schema.ResourceData, name string, duplo *duplosdk.DuploS3EventNotificaitionResponse) {
	d.Set("bucket_name", name)
	if len(*duplo.Lambda) > 0 {
		for _, l := range *duplo.Lambda {
			d.Set("destination_type", "lambda")
			d.Set("destination_arn", l.LambdaARN)
			d.Set("event_types", StringValueSliceTolist(l.EventTypes))
			d.Set("destination_name", GetResourceNameFromARN(l.LambdaARN))
		}

	}
	if len(*duplo.SNS) > 0 {
		for _, s := range *duplo.SNS {
			d.Set("destination_type", "sns")
			d.Set("destination_arn", s.SNSARN)
			d.Set("event_types", StringValueSliceTolist(s.EventTypes))
			d.Set("destination_name", GetResourceNameFromARN(s.SNSARN))

		}
	}
	if len(*duplo.SQS) > 0 {
		for _, q := range *duplo.SQS {
			d.Set("destination_type", "sqs")
			d.Set("destination_arn", q.SQSARN)
			d.Set("event_types", StringValueSliceTolist(q.EventTypes))
			d.Set("destination_name", GetResourceNameFromARN(q.SQSARN))

		}
	}
}

func expandEventNotification(d *schema.ResourceData) *duplosdk.DuploS3EventNotificaition {
	destType := d.Get("destination_type").(string)
	obj := duplosdk.DuploS3EventNotificaition{}
	switch destType {
	case "lambda":
		obj.LambdaName = d.Get("destination_name").(string)
	case "sns":
		obj.SNSName = d.Get("destination_name").(string)

	case "sqs":
		obj.SQSName = d.Get("destination_name").(string)
	}
	obj.EventTypes = expandStringSet(d.Get("event_types").(*schema.Set))
	obj.EnableEventBridge = d.Get("enable_event_bridge").(bool)
	return &obj
}

func resourceS3EventNotificationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3EventNotificationDelete ******** start")

	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, name := idParts[0], idParts[1]
	rq := &duplosdk.DuploS3EventNotificaition{}
	c := m.(*duplosdk.Client)

	err := c.UpdateS3EventNotification(tenantID, name, *rq)
	if err != nil {
		return diag.Errorf("resourceS3EventNotificationCreateOrUpdate: Unable to create or update s3 event notification using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "s3 event notification", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetS3EventNotification(tenantID, name)
	})
	if diag != nil {
		return diag
	}
	log.Printf("[TRACE] resourceS3EventNotificationDelete ******** end")

	return nil
}
