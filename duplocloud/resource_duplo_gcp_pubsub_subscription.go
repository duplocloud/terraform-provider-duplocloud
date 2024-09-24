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
						Type:         schema.TypeBool,
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
						Default:      1000,
						ValidateFunc: validation.IntAtLeast(1000),
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
						Type:         schema.TypeBool,
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
						Default:      1000,
						ValidateFunc: validation.IntAtLeast(1000),
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
		"": {},
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
func resourceGCPStorageBucketV2() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceGCPStorageBucketV2Read,
		CreateContext: resourceGCPStorageBucketV2Create,
		UpdateContext: resourceGCPStorageBucketV2Update,
		DeleteContext: resourceGCPStorageBucketV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: gcpStorageBucketV2Schema(),
	}
}

// READ resource
func resourceGCPStorageBucketV2Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPStorageBucketV2Read ******** start")

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
func resourceGCPStorageBucketV2Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPStorageBucketV2Create ******** start")
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)

	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	duploObject := duplosdk.DuploGCPBucket{
		Name: name,
	}
	errFill := fillGCPBucketRequest(&duploObject, d)
	if errFill != nil {
		return diag.FromErr(errFill)
	}

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
func resourceGCPStorageBucketV2Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
func resourceGCPStorageBucketV2Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceGCPStorageBucketV2SetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGCPBucket) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("domain_name", duplo.DomainName)
	d.Set("enable_versioning", duplo.EnableVersioning)
	d.Set("allow_public_access", duplo.AllowPublicAccess)
	d.Set("default_encryption", []map[string]interface{}{{
		"method": encodeEncryption(duplo.DefaultEncryptionType),
	}})
	d.Set("location", duplo.Location)
	flattenGcpLabels(d, duplo.Labels)
}

func encodeEncryption(i int) string {
	m := map[int]string{
		0: "None",
		2: "Sse",
		3: "AwsKms",
		4: "TenantKms",
	}
	return m[i]
}

func decodeEncryption(i string) int {
	m := map[string]int{
		"None":      0,
		"Sse":       2,
		"AwsKms":    3,
		"TenantKms": 4,
	}
	return m[i]
}

func fillGCPBucketRequest(duploObject *duplosdk.DuploGCPBucket, d *schema.ResourceData) error {
	log.Printf("[TRACE] fillGCPBucketRequest ******** start")

	// Set the object versioning
	if v, ok := d.GetOk("enable_versioning"); ok && v != nil {
		duploObject.EnableVersioning = v.(bool)
	}

	// Set the public access block.
	if v, ok := d.GetOk("allow_public_access"); ok && v != nil {
		duploObject.AllowPublicAccess = v.(bool)
	}

	// Set the default encryption.
	defaultEncryption, err := getOptionalBlockAsMap(d, "default_encryption")
	if err != nil {
		return err
	}
	if v, ok := defaultEncryption["method"]; ok && v != nil {
		duploObject.DefaultEncryptionType = decodeEncryption(v.(string))
	}

	if v, ok := d.GetOk("location"); ok && v != nil {
		duploObject.Location = v.(string)
	}

	duploObject.Labels = expandAsStringMap("labels", d)
	log.Printf("[TRACE] fillGCPBucketRequest ******** end")
	return nil
}
