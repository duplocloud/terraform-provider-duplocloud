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

func s3BucketReplicationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"rule": {
			Description: "replication rule name for s3 source bucket",
			Type:        schema.TypeString,
			Required:    true,
		},
		"destination_bucket": {
			Description: "name of destination bucket.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"source_bucket": {
			Description: "name of source bucket.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"priority": {
			Description: "replication priority. Priority must be unique between multiple rules.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"delete_marker_replication": {
			Description: "Whether or not to enable access logs.  When enabled, Duplo will send access logs to a centralized S3 bucket per plan.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Default:     false,
		},
		"storage_class": {
			Description: "storage_class type: Standard, IntelligentTiering, StandardInfrequentAccess, OneZoneInfrequentAccess, GlacierInstantRetrieval, Glacier, DeepArchive, ReducedRedundancy",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Standard",
				"IntelligentTiering",
				"StandardInfrequentAccess",
				"OneZoneInfrequentAccess",
				"GlacierInstantRetrieval",
				"Glacier",
				"DeepArchive",
				"ReducedRedundancy",
			}, false),
		},
	}
}

// Resource for managing an S3 replication
func resourceS3BucketReplication() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceS3BucketReplicationRead,
		CreateContext: resourceS3BucketReplicationCreate,
		//		UpdateContext: resourceS3BucketReplicationUpdate,
		//		DeleteContext: resourceS3BucketReplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: s3BucketReplicationSchema(),
	}
}

// READ resource
func resourceS3BucketReplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketReplicationRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3BucketReplicationRead: Invalid resource (ID: %s)", id)
	}
	tenantID, name := idParts[0], idParts[1]

	c := m.(*duplosdk.Client)

	// Get the object from Duplo
	duplo, err := c.TenantGetV3S3BucketReplication(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceS3BucketReplicationRead: Unable to retrieve s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	d.Set("source_bucket", duplo.SourceBucket)
	d.Set("destination_bucket", duplo.DestinationBucket)
	d.Set("priority", duplo.Priority)
	d.Set("rule", duplo.Rule)
	d.Set("delete_marker_replication", duplo.DeleteMarkerReplication)
	d.Set("storage_class", duplo.StorageClass)

	// Set simple fields first.

	log.Printf("[TRACE] resourceS3BucketReplicationRead ******** end")
	return nil
}

// CREATE resource
func resourceS3BucketReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketReplicationCreate ******** start")
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketReplication{
		Rule:                    d.Get("rule").(string),
		DestinationBucket:       d.Get("destination_bucket").(string),
		SourceBucket:            d.Get("source_bucket").(string),
		Priority:                d.Get("priority").(int),
		DeleteMarkerReplication: d.Get("delete_marker_replication").(bool),
	}

	// Post the object to Duplo
	_, err := c.TenantCreateV3S3BucketReplication(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("resourceS3BucketReplicationCreate: Unable to create s3 bucket replication for (tenant: %s, source bucket: %s: error: %s)", tenantID, duploObject.SourceBucket, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, duploObject.SourceBucket)
	d.SetId(id)
	diags := resourceS3BucketReplicationRead(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketReplicationCreate ******** end")
	return diags
}

// UPDATE resource
func resourceS3BucketReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketReplicationUpdate ******** start")

	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3BucketReplicationUpdate: Invalid resource (ID: %s)", id)
	}
	tenantID := idParts[0]

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketReplication{
		Rule:                    d.Get("rule").(string),
		DestinationBucket:       d.Get("destination_bucket").(string),
		SourceBucket:            d.Get("source_bucket").(string),
		Priority:                d.Get("priority").(int),
		DeleteMarkerReplication: d.Get("delete_marker_replication").(bool),
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err := c.TenantUpdateV3S3BucketReplication(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceS3BucketUpdate: Unable to update s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.SourceBucket, err)
	}
	diags := resourceS3BucketReplicationRead(ctx, d, m)

	log.Printf("[TRACE] resourceS3BucketReplicationUpdate ******** end")
	return diags
}

// DELETE resource
func resourceS3BucketReplecationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3BucketDelete: Invalid resource (ID: %s)", id)
	}
	err := c.TenantDeleteS3Bucket(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("resourceS3BucketDelete: Unable to delete bucket (name:%s, error: %s)", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the bucket.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetS3Bucket(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceS3BucketDelete ******** end")
	return nil
}
