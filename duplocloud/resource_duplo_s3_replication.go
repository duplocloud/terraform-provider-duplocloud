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

func ruleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"destination_bucket": {
				Description: "name of destination bucket.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"name": {
				Description: "replication rule name for s3 source bucket",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fullname": {
				Description: "replication rule fullname for s3 source bucket",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"destination_arn": {
				Description: "destination bucket arn",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"priority": {
				Description: "replication priority. Priority must be unique between multiple rules.",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    false,
			},
			"delete_marker_replication": {
				Description:      "Whether or not to enable delete marker on replication. Can be set only during creation.",
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: diffSuppressWhenNotCreating,
			},
			"storage_class": {
				Description: "storage_class type: STANDARD, INTELLIGENT_TIERING, STANDARD_IA, ONEZONE_IA, GLACIER_IR, GLACIER, DEEP_ARCHIVE, REDUCED_REDUNDANCY. Can be set only during creation",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"STANDARD",
					"INTELLIGENT_TIERING",
					"STANDARD_IA",
					"ONEZONE_IA",
					"GLACIER_IR",
					"GLACIER",
					"DEEP_ARCHIVE",
					"REDUCED_REDUNDANCY",
				}, false),
				//DiffSuppressFunc: diffSuppressWhenNotCreating,
			},
		},
	}
}

func s3BucketReplicationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 bucket replication rule will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"rules": {
			Description: "replication rules for source bucket",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem:        ruleSchema(),
		},

		"source_bucket": {
			Description: "name of source bucket.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

// Resource for managing an S3 replication
func resourceS3BucketReplication() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource duplocloud_s3_bucket_replication is dependent on duplocloud_s3_bucket. This resource sets replication rules for source bucket",
		ReadContext:   resourceS3BucketReplicationRead,
		CreateContext: resourceS3BucketReplicationCreate,
		UpdateContext: resourceS3BucketReplicationUpdate,
		DeleteContext: resourceS3BucketReplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
	duplo, err := getS3BucketReplication(c, tenantID, name)
	if err != nil {
		d.SetId("")
		return diag.Errorf("resourceS3BucketReplicationRead: Unable to retrieve s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)

	}
	if duplo == nil {
		d.SetId("")
		return nil
	}
	// Get the object from Duplo
	d.Set("rules", duplo)
	d.Set("source_bucket", name)

	// Set simple fields first.

	log.Printf("[TRACE] resourceS3BucketReplicationRead ******** end")
	return nil
}

func getS3BucketReplication(c *duplosdk.Client, tenantID, name string) ([]map[string]interface{}, error) {
	duplo, err := c.TenantGetV3S3BucketReplication(tenantID, name)
	if err != nil {
		return nil, err
	}
	if duplo == nil || len(duplo.Rule) == 0 {
		return nil, nil
	}
	tenantInfo, err := c.TenantGetV2(tenantID)
	if err != nil {
		return nil, err
	}
	rules := make([]map[string]interface{}, 0, len(duplo.Rule))
	for _, data := range duplo.Rule {
		kv := make(map[string]interface{})
		kv["fullname"] = data.Rule
		kv["priority"] = data.Priority
		kv["delete_marker_replication"] = data.DeleteMarkerReplication.Status.Value == "Enabled"
		kv["destination_arn"] = data.DestinationBucket.BucketArn
		destTokens := strings.Split(data.DestinationBucket.BucketArn, ":")
		kv["destination_bucket"] = destTokens[len(destTokens)-1]
		ruleName := strings.Split(data.Rule, tenantInfo.AccountName+"-")
		kv["name"] = ruleName[len(ruleName)-1]
		kv["storage_class"] = "STANDARD"
		if data.DestinationBucket.StorageClass != nil && data.DestinationBucket.StorageClass.Value != "" {
			kv["storage_class"] = data.DestinationBucket.StorageClass.Value
		}
		rules = append(rules, kv)
	}
	return rules, nil
}

// CREATE resource
func resourceS3BucketReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketReplicationCreate ******** start")
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	rules := d.Get("rules").([]interface{})
	sourceBucket := d.Get("source_bucket").(string)
	// Create the request object.
	for _, rule := range rules {
		kv := rule.(map[string]interface{})

		duploObject := duplosdk.DuploS3BucketReplication{
			Rule:                    kv["name"].(string),
			DestinationBucket:       kv["destination_bucket"].(string),
			SourceBucket:            sourceBucket,
			Priority:                kv["priority"].(int),
			DeleteMarkerReplication: kv["delete_marker_replication"].(bool),
		}

		// Post the object to Duplo
		err := c.TenantCreateV3S3BucketReplication(tenantID, duploObject)
		if err != nil {
			return diag.Errorf("resourceS3BucketReplicationCreate: Unable to create s3 bucket replication for (tenant: %s, source bucket: %s: error: %s)", tenantID, duploObject.SourceBucket, err)
		}
	}

	id := fmt.Sprintf("%s/%s", tenantID, sourceBucket)
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
	c := m.(*duplosdk.Client)

	rules := d.Get("rules").([]interface{})

	for _, rule := range rules {
		kv := rule.(map[string]interface{})

		duploObject := duplosdk.DuploS3BucketReplication{
			Rule:                    kv["name"].(string),
			DestinationBucket:       kv["destination_bucket"].(string),
			SourceBucket:            d.Get("source_bucket").(string),
			Priority:                kv["priority"].(int),
			DeleteMarkerReplication: kv["delete_marker_replication"].(bool),
			StorageClass:            kv["storage_class"].(string),
		}
		ruleFullname := kv["fullname"].(string)
		// Post the object to Duplo
		err := c.TenantUpdateV3S3BucketReplication(tenantID, ruleFullname, duploObject)
		if err != nil {
			return diag.Errorf("resourceS3BucketReplicationUpdate: Unable to update s3 bucket using v3 api (tenant: %s, bucket: %s: rule: %s,error: %s)", tenantID, duploObject.SourceBucket, ruleFullname, err)
		}
	}
	diags := resourceS3BucketReplicationRead(ctx, d, m)

	log.Printf("[TRACE] resourceS3BucketReplicationUpdate ******** end")
	return diags
}

// DELETE resource
func resourceS3BucketReplicationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketReplicationDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3BucketReplicationDelete: Invalid resource (ID: %s)", id)
	}
	rules := d.Get("rules").([]interface{})

	for _, rule := range rules {
		kv := rule.(map[string]interface{})
		ruleName := kv["fullname"].(string)
		err := c.TenantDeleteV3S3BucketReplication(idParts[0], idParts[1], ruleName)
		if err != nil {
			return diag.Errorf("resourceS3BucketReplicationDelete: Unable to delete bucket replication rule (name:%s, error: %s)", ruleName, err)
		}
		// Wait up to 60 seconds for Duplo to delete the bucket replication.
		time.Sleep(60 * time.Second)
	}

	// Wait 10 more seconds to deal with consistency issues.

	log.Printf("[TRACE] resourceS3BucketDelete ******** end")
	return nil
}
