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

func gcpPubSubSubscriptionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the storage bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Name of the subscription.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"topic": {
			Description: "A reference to a Topic resource.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"big_query": {
			Description: "Default encryption settings for objects uploaded to the bucket.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"table": {
						Description: "The name of the table to which to write data.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"use_topic_schema": {
						Description: "When true, use the topic's schema as the columns to write to in BigQuery, if it exists. Only one of use_topic_schema and use_table_schema can be set.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"use_table_schema": {
						Description: "When true, write the subscription name, messageId, publishTime, attributes, and orderingKey to additional columns in the table.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"drop_unknown_fields": {
						Description: "When true and use_topic_schema or use_table_schema is true, any fields that are a part of the topic schema or message schema that are not part of the BigQuery table schema are dropped when writing to BigQuery. Otherwise, the schemas must be kept in sync and any messages with extra fields are not written and remain in the subscription's backlog",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"service_account_email": {
						Description: "The service account to use to write to BigQuery. If not specified, the Pub/Sub service agent, service-{project_number}@gcp-sa-pubsub.iam.gserviceaccount.com, is used.",
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
					},
					"write_metadata": {
						Description: "When true, write the subscription name, messageId, publishTime, attributes, and orderingKey as additional fields in the output.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"state": {
						Description: "An output-only field that indicates whether or not the subscription can receive messages.",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
		},
		"cloud_storage_config": {
			Description: "Default encryption settings for objects uploaded to the bucket.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"bucket": {
						Description: "User-provided name for the Cloud Storage bucket. The bucket must be created by the user. The bucket name must be without any prefix like 'gs://'.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"filename_prefix": {
						Description: "User-provided prefix for Cloud Storage filename.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"filename_suffix": {
						Description: "User-provided suffix for Cloud Storage filename. Must not end in '/'.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"filename_datetime_format": {
						Description: "User-provided format string specifying how to represent datetimes in Cloud Storage filenames",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"max_duration": {
						Description:  "The maximum duration that can elapse before a new Cloud Storage file is created. Min 1 minute, max 10 minutes, default 5 minutes. May not exceed the subscription's acknowledgement deadline",
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "300s",
						ValidateFunc: validateDurationBetween(60*time.Second, 600*time.Second, 9),
					},
					"max_bytes": {
						Description: "The maximum bytes that can be written to a Cloud Storage file before a new file is created. Min 1 KB, max 10 GiB. The maxBytes limit may be exceeded in cases where messages are larger than the limit.",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},
					"max_message": {
						Description:  "The maximum messages that can be written to a Cloud Storage file before a new file is created. Min 1000 messages.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      "1000",
						ValidateFunc: validateStringFormatedInt64Atleast("1000"),
					},
					"state": {
						Description: "An output-only field that indicates whether or not the subscription can receive messages.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"avro_config": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"write_metadata": {
									Description: "When true, write the subscription name, messageId, publishTime, attributes, and orderingKey as additional fields in the output.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
								},
								"use_topic_schema": {
									Description: "When true, the output Cloud Storage file will be serialized using the topic schema, if it exists.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
								},
							},
						},
					},
					"service_account_email": {
						Description: "The service account to use to write to BigQuery. If not specified, the Pub/Sub service agent, service-{project_number}@gcp-sa-pubsub.iam.gserviceaccount.com, is used.",
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
		"push_config": {
			Description: "Default encryption settings for objects uploaded to the bucket.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"push_endpoint": {
						Description: "User-provided name for the Cloud Storage bucket. The bucket must be created by the user. The bucket name must be without any prefix like 'gs://'.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"attributes": {
						Type:     schema.TypeMap,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"oidc_token": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"service_account_email ": {
									Description: "Service account email to be used for generating the OIDC token.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"audience": {
									Description: "Audience to be used when generating OIDC token. The audience claim identifies the recipients that the JWT is intended for. The audience value is a single case-sensitive string.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     false,
								},
							},
						},
					},
					"no_wrapper": {
						Description: "When set, the payload to the push endpoint is not wrapped",
						Type:        schema.TypeList,
						Optional:    true,
						Computed:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"write_metadata": {
									Description: "When true, writes the Pub/Sub message metadata to x-goog-pubsub-<KEY>:<VAL> headers of the HTTP request. Writes the Pub/Sub message attributes to <KEY>:<VAL> headers of the HTTP request.",
									Type:        schema.TypeBool,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},

		"ack_deadline_seconds": {
			Description:  "This value is the maximum time after a subscriber receives a message before the subscriber should acknowledge the message.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      10,
			ValidateFunc: validation.IntBetween(10, 600),
		},
		"message_retention_duration": {
			Description:  "How long to retain unacknowledged messages in the subscription's backlog, from the moment a message is published. If retain_acked_messages is true, then this also configures the retention of acknowledged messages, and thus configures how far back in time a subscriptions.seek can be done. Defaults to 7 days. Cannot be more than 7 days.",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "604800s",
			ValidateFunc: validateDurationBetween(600*time.Second, 604800*time.Second, 9),
		},

		"retain_acked_messages": {
			Description: "Indicates whether to retain acknowledged messages.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"filter": {
			Description: "The subscription only delivers the messages that match the filter. Pub/Sub automatically acknowledges the messages that don't match the filter. You can filter messages by their attributes. The maximum length of a filter is 256 bytes. After creating the subscription, you can't modify the filter.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"enable_message_ordering": {
			Description: " If true, messages published with the same orderingKey in PubsubMessage will be delivered to the subscribers in the order in which they are received by the Pub/Sub system. Otherwise, they may be delivered in any order.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_exactly_once_delivery": {
			Description: "If true, Pub/Sub provides the following guarantees for the delivery of a message with a given value of messageId on this Subscriptions': * The message sent to a subscriber is guaranteed not to be resent before the message's acknowledgement deadline expires. * An acknowledged message will not be resent to a subscriber. Note that subscribers may still receive multiple copies of a message when enable_exactly_once_delivery is true if the message was published multiple times by a publisher client. These copies are considered distinct by Pub/Sub and have distinct messageId values",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"expiration_policy": {
			Description: "A policy that specifies the conditions for this subscription's expiration. A subscription is considered active as long as any connected subscriber is successfully consuming messages from the subscription or is issuing operations on the subscription",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ttl": {
						Description: "Specifies the 'time-to-live' duration for an associated resource. The resource expires if it is not active for a period of ttl. If ttl is empty string, the associated resource never expires.  A duration in seconds with up to nine fractional digits, terminated by 's'. Example - '3.5s'.",
						Type:        schema.TypeString,
						Required:    true,
						Default:     "",
					},
				},
			},
		},
		"dead_letter_policy": {
			Description: "A policy that specifies the conditions for dead lettering messages in this subscription. If dead_letter_policy is not set, dead lettering is disabled",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"dead_letter_topic": {
						Description: "The name of the topic to which dead letter messages should be published.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"max_delivery_attempts": {
						Description:  "The maximum number of delivery attempts for any message. The value must be between 5 and 100. The number of delivery attempts is defined as 1 + (the sum of number of NACKs and number of times the acknowledgement deadline has been exceeded for the message)",
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(5, 100),
						Default:      5,
					},
				},
			},
		},
		"retry_policy": {
			Description: "A policy that specifies how Pub/Sub retries message delivery for this subscription. If not set, the default retry policy is applied. This generally implies that messages will be retried as soon as possible for healthy subscribers. RetryPolicy will be triggered on NACKs or acknowledgement deadline exceeded events for a given message",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"minimum_backoff": {
						Description:  "The minimum delay between consecutive deliveries of a given message. Value should be between 0 and 600 seconds. Defaults to 10 seconds",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validateDurationBetween(0, 600*time.Second, 9),
						Default:      "10s",
					},
					"maximum_backoff": {
						Description:  "The maximum delay between consecutive deliveries of a given message. Value should be between 0 and 600 seconds. Defaults to 600 seconds.",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validateDurationBetween(0, 600*time.Second, 9),
						Default:      "600s",
					},
				},
			},
		},

		"labels": {
			Description: "A set of key/value label pairs to assign to this Subscription",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceGCPPubSubSubscription() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceGCPPubSubSubscriptionRead,
		CreateContext: resourceGCPPubSubSubscriptionCreate,
		UpdateContext: resourceGCPPubSubSubscriptionUpdate,
		DeleteContext: resourceGCPPubSubSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: gcpPubSubSubscriptionSchema(),
	}
}

// READ resource
func resourceGCPPubSubSubscriptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPPubSubSubscriptionRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceGCPStorageBucketV2Read: Invalid resource (ID: %s)", id)
	}
	tenantID, name := idParts[0], idParts[1]

	c := m.(*duplosdk.Client)

	// Figure out the full resource name.
	fullName := d.Get("fullname").(string)

	// Get the object from Duplo
	duplo, err := c.GCPTenantGetV3StorageBucketV2(tenantID, fullName)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceGCPStorageBucketV2Read: Unable to retrieve storage bucket (tenant: %s, bucket: %s, error: %s)", tenantID, name, err)
	}

	// Set simple fields first.
	resourceGCPStorageBucketV2SetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceGCPStorageBucketV2Read ******** end")
	return nil
}

// CREATE resource
func resourceGCPPubSubSubscriptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPStorageBucketV2Create ******** start")
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)

	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	reqBody := expandPubSubSubscription(d)

	// Post the object to Duplo
	resp, err := c.GCPTenantCreateV3StorageBucketV2(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceGCPStorageBucketV2Create: Unable to create storage bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	fullName := resp.Name
	// Wait up to 60 seconds for Duplo to be able to return the bucket's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "storage bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPTenantGetV3StorageBucketV2(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	duplo, err := c.GCPTenantGetV3StorageBucketV2(tenantID, fullName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("resourceGCPStorageBucketV2Create: Unable to retrieve storage bucket details using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	d.SetId(id)

	// Set simple fields first.
	resourceGCPStorageBucketV2SetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGCPStorageBucketV2Create ******** end")
	return nil
}

// UPDATE resource
func resourceGCPPubSubSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPStorageBucketV2Update ******** start")

	fullname := d.Get("fullname").(string)
	name := d.Get("name").(string)

	// Create the request object.
	duploObject := duplosdk.DuploGCPBucket{
		Name: fullname,
	}

	errName := fillGCPBucketRequest(&duploObject, d)
	if errName != nil {
		return diag.FromErr(errName)
	}
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	resource, err := c.GCPTenantUpdateV3StorageBucketV2(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("GCPTenantUpdateV3StorageBucketV2: Unable to update storage bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	resourceGCPStorageBucketV2SetData(d, tenantID, name, resource)

	log.Printf("[TRACE] GCPTenantUpdateV3StorageBucketV2 ******** end")
	return nil
}

// DELETE resource
func resourceGCPPubSubSubscriptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPStorageBucketV2Delete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceGCPStorageBucketV2Delete: Invalid resource (ID: %s)", id)
	}
	fullName := d.Get("fullname").(string)
	err := c.GCPTenantDeleteStorageBucketV2(idParts[0], idParts[1], fullName)
	if err != nil {
		return diag.Errorf("GCPTenantDeleteStorageBucketV2: Unable to delete bucket (name:%s, error: %s)", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the bucket.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPTenantGetV3StorageBucketV2(idParts[0], fullName)
	})
	if diag != nil {
		return diag
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceGCPStorageBucketV2Delete ******** end")
	return nil
}

func expandBigQuery(d *schema.ResourceData) *duplosdk.DuploPubSubBigQuery {
	return &duplosdk.DuploPubSubBigQuery{
		Table:               d.Get("table").(string),
		UseTopicSchema:      d.Get("use_topic_schema").(bool),
		UseTableSchema:      d.Get("use_table_schema").(bool),
		DropUnknownFields:   d.Get("drop_unknown_fields").(bool),
		ServiceAccountEmail: d.Get("service_account_email").(string),
		WriteMetadata:       d.Get("write_metadata").(bool),
	}

}

func flattenBigQuery(rb *duplosdk.DuploPubSubBigQuery) []interface{} {
	mp := map[string]interface{}{
		"table":                 rb.Table,
		"use_topic_schema":      rb.UseTopicSchema,
		"use_table_schema":      rb.UseTableSchema,
		"drop_unknown_fields":   rb.DropUnknownFields,
		"service_account_email": rb.ServiceAccountEmail,
		"write_metadata":        rb.WriteMetadata,
	}
	bq := make([]interface{}, 0, 1)
	bq = append(bq, mp)
	return bq
}

func expandCloudStorageConfig(d *schema.ResourceData) *duplosdk.DuploPubSubCloudStorageConfig {
	return &duplosdk.DuploPubSubCloudStorageConfig{
		Bucket:                 d.Get("bucket").(string),
		FilenamePrefix:         d.Get("filename_prefix").(string),
		FileNameSuffix:         d.Get("filename_suffix").(string),
		FileNameDateTimeFormat: d.Get("filename_datetime_format").(string),
		MaxDuration:            d.Get("max_duration").(string),
		MaxBytes:               d.Get("max_bytes").(string),
		MaxMessages:            d.Get("max_message").(string),
		AvroConfig: struct {
			WriteMetadata  bool `json:"writeMetadata"`
			UseTopicSchema bool `json:"useTopicSchema"`
		}{
			WriteMetadata:  d.Get("avro_config.0.write_metadata").(bool),
			UseTopicSchema: d.Get("avro_config.0.use_topic_schema").(bool),
		},

		ServiceAccountEmail: d.Get("service_account_email").(string),
	}
}

func flattenCloudStorageConfig(rb *duplosdk.DuploPubSubCloudStorageConfig) []interface{} {

	avro := make([]interface{}, 0, 1)
	amp := map[string]interface{}{
		"write_metadata":   rb.AvroConfig.WriteMetadata,
		"use_topic_schema": rb.AvroConfig.UseTopicSchema,
	}
	avro = append(avro, amp)
	mp := map[string]interface{}{
		"bucket":                   rb.Bucket,
		"filename_prefix":          rb.FilenamePrefix,
		"filename_suffix":          rb.FileNameSuffix,
		"filename_datetime_format": rb.FileNameDateTimeFormat,
		"max_duration":             rb.MaxDuration,
		"max_bytes":                rb.MaxBytes,
		"max_message":              rb.MaxMessages,
		"avro_config":              avro,
		"service_account_email":    rb.ServiceAccountEmail,
	}
	config := make([]interface{}, 0, 1)
	config = append(config, mp)
	return config
}

func expandPushConfig(d *schema.ResourceData) *duplosdk.DuploPubSubPushConfig {
	return &duplosdk.DuploPubSubPushConfig{
		PushEndpoint: d.Get("push_endpoint").(string),
		Attributes:   expandAsStringMap("attributes", d),
		OidcToken: struct {
			ServiceAccountEmail string `json:"serviceAccountEmail"`
			Audience            string `json:"audience"`
		}{
			ServiceAccountEmail: d.Get("oidc_token.0.service_account_email").(string),
			Audience:            d.Get("oidc_token.0.audience").(string),
		},

		NoWrapper: struct {
			WriteMetadata bool `json:"writeMetadata"`
		}{
			WriteMetadata: d.Get("no_wrapper.0.write_metadata").(bool),
		},
	}
}

func flattenPushConfig(rb *duplosdk.DuploPubSubPushConfig) []interface{} {
	omp := map[string]interface{}{
		"service_account_email": rb.OidcToken.ServiceAccountEmail,
		"audience":              rb.OidcToken.Audience,
	}
	ompList := make([]interface{}, 0, 1)
	ompList = append(ompList, omp)

	nwmp := map[string]interface{}{
		"write_metadata": rb.NoWrapper.WriteMetadata,
	}
	nwmpList := make([]interface{}, 0, 1)
	nwmpList = append(nwmpList, nwmp)
	mp := map[string]interface{}{
		"push_endpoint": rb.PushEndpoint,
		"attributes":    flattenStringMap(rb.Attributes),
		"oidc_token":    ompList,
		"no_wrapper":    nwmpList,
	}
	config := make([]interface{}, 0, 1)
	config = append(config, mp)
	return config
}

func expandPubSubSubscription(d *schema.ResourceData) *duplosdk.DuploPubSubSubscription {
	return &duplosdk.DuploPubSubSubscription{
		Name:                      d.Get("name").(string),
		Topic:                     d.Get("topic").(string),
		BigQuery:                  expandBigQuery(d),
		CloudStorageConfig:        expandCloudStorageConfig(d),
		PushConfig:                expandPushConfig(d),
		AckDeadlineSeconds:        d.Get("ack_deadline_seconds").(int),
		MessageRetentionDuration:  d.Get("message_retention_duration").(string),
		RetainAckedMessages:       d.Get("retain_acked_messages").(bool),
		Filter:                    d.Get("filter").(string),
		EnableMessageOrdering:     d.Get("enable_message_ordering").(bool),
		EnableExactlyOnceDelivery: d.Get("enable_exactly_once_delivery").(bool),
		ExpirationPolicy:          expandExpirationPolicy(d),
		DeadLetterPolicy:          expandDeadLetterPolicy(d),
		RetryPolicy:               expandRetryPolicy(d),
		Labels:                    expandAsStringMap("labels", d),
	}
}
func flattenPubSubSubscription(rb *duplosdk.DuploPubSubSubscription) []interface{} {
	mp := map[string]interface{}{
		"name":                         rb.Name,
		"topic":                        rb.Topic,
		"big_query":                    flattenBigQuery(rb.BigQuery),
		"cloud_storage_config":         flattenCloudStorageConfig(rb.CloudStorageConfig),
		"push_config":                  flattenPushConfig(rb.PushConfig),
		"ack_deadline_seconds":         rb.AckDeadlineSeconds,
		"message_retention_duration":   rb.MessageRetentionDuration,
		"retain_acked_messages":        rb.RetainAckedMessages,
		"filter":                       rb.Filter,
		"enable_message_ordering":      rb.EnableMessageOrdering,
		"enable_exactly_once_delivery": rb.EnableExactlyOnceDelivery,
		"expiration_policy":            rb.ExpirationPolicy,
		"dead_letter_policy":           flattenDeadLetterPolicy(rb.DeadLetterPolicy),
		"retry_policy":                 flattenRetryPolicy(rb.RetryPolicy),
		"labels":                       flattenStringMap(rb.Labels),
	}
	p := make([]interface{}, 0, 1)
	p = append(p, mp)
	return p
}

func expandExpirationPolicy(d *schema.ResourceData) *duplosdk.DuplocloudPubSubExpirationPolicy {
	return &duplosdk.DuplocloudPubSubExpirationPolicy{
		Ttl: d.Get("expiration_policy.0.ttl").(string),
	}
}

func flattenExpirationPolicy(rb *duplosdk.DuplocloudPubSubExpirationPolicy) []interface{} {
	mp := map[string]interface{}{
		"ttl": rb.Ttl,
	}
	p := make([]interface{}, 0, 1)
	p = append(p, mp)
	return p
}

func expandDeadLetterPolicy(d *schema.ResourceData) *duplosdk.DuplocloudPubSubDeadLetterPolicy {
	return &duplosdk.DuplocloudPubSubDeadLetterPolicy{
		DeadLetterTopic:     d.Get("dead_letter_policy.0.dead_letter_topic").(string),
		MaxDeliveryAttempts: d.Get("dead_letter_policy.0.max_delivery_attempts").(int),
	}
}

func flattenDeadLetterPolicy(rb *duplosdk.DuplocloudPubSubDeadLetterPolicy) []interface{} {
	mp := map[string]interface{}{
		"dead_letter_topic":     rb.DeadLetterTopic,
		"max_delivery_attempts": rb.MaxDeliveryAttempts,
	}
	p := make([]interface{}, 0, 1)
	p = append(p, mp)
	return p
}

func expandRetryPolicy(d *schema.ResourceData) *duplosdk.DuplocloudPubSubRetryPolicy {
	return &duplosdk.DuplocloudPubSubRetryPolicy{
		MinimumBackoff: d.Get("minimum_backoff").(string),
		MaximumBackoff: d.Get("maximum_backoff").(string),
	}
}

func flattenRetryPolicy(rb *duplosdk.DuplocloudPubSubRetryPolicy) []interface{} {
	mp := map[string]interface{}{
		"minimum_backoff": rb.MinimumBackoff,
		"maximum_backoff": rb.MaximumBackoff,
	}
	p := make([]interface{}, 0, 1)
	p = append(p, mp)
	return p
}
