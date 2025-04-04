package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ruleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"destination_bucket": {
				Description: "fullname of the destination bucket.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description:  "replication rule name for s3 source bucket",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\-_]*$`), "Invalid rule name: only alphabets, digits, underscores, and hyphens are allowed."),
				ForceNew:     true,
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
			},
			"delete_marker_replication": {
				Description: "Whether or not to enable delete marker on replication.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"storage_class": {
				Description: "storage_class type: STANDARD, INTELLIGENT_TIERING, STANDARD_IA, ONEZONE_IA, GLACIER_IR, GLACIER, DEEP_ARCHIVE, REDUCED_REDUNDANCY.",
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
			Required:    true,
			Elem:        ruleSchema(),
		},

		"source_bucket": {
			Description: "fullname of the source bucket.",
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
		Schema:        s3BucketReplicationSchema(),
		CustomizeDiff: validateTenantBucket,
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
	rp := []map[string]interface{}{}
	rules := expandRules(d)

	if len(rules) > 0 {

		for _, v := range duplo {
			mp := filterRule(v, rules)
			if mp != nil {
				rp = append(rp, mp)
			}
		}
	} else {
		rp = append(rp, duplo...)
	}

	// Get the object from Duplo
	d.Set("rules", rp)
	d.Set("source_bucket", name)

	// Set simple fields first.

	log.Printf("[TRACE] resourceS3BucketReplicationRead ******** end")
	return nil
}

func filterRule(m map[string]interface{}, filter []duplosdk.DuploS3BucketReplication) map[string]interface{} {
	for _, v := range filter {
		if name, ok := m["fullname"]; ok {
			if strings.Contains(name.(string), v.Rule) {
				return m
			}

		}
	}
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
	// Create the request object.
	rules := expandRules(d)
	sourceBucket := rules[0].SourceBucket
	for _, rule := range rules {
		err := c.TenantCreateV3S3BucketReplication(tenantID, rule)
		if err != nil {
			return diag.Errorf("resourceS3BucketReplicationCreate: Unable to create s3 bucket replication for (tenant: %s, source bucket: %s: error: %s)", tenantID, rule.SourceBucket, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
	id := fmt.Sprintf("%s/%s", tenantID, sourceBucket)

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "s3 replication rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetV3S3BucketReplication(tenantID, sourceBucket)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)
	diags = resourceS3BucketReplicationRead(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketReplicationCreate ******** end")
	return diags
}
func expandRules(d *schema.ResourceData) []duplosdk.DuploS3BucketReplication {
	obj := []duplosdk.DuploS3BucketReplication{}
	rules := d.Get("rules").([]interface{})
	sourceBucket := d.Get("source_bucket").(string)

	for _, rule := range rules {
		kv := rule.(map[string]interface{})
		duploObject := duplosdk.DuploS3BucketReplication{}

		duploObject.Rule = kv["name"].(string)
		duploObject.DestinationBucket = kv["destination_bucket"].(string)
		duploObject.SourceBucket = sourceBucket
		duploObject.Priority = kv["priority"].(int)
		duploObject.DeleteMarkerReplication = kv["delete_marker_replication"].(bool)
		duploObject.StorageClass = kv["storage_class"].(string)
		obj = append(obj, duploObject)
	}
	return obj
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
		time.Sleep(250 * time.Millisecond)
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
		cerr := s3replicaWaitUntilDelete(ctx, c, idParts[0], idParts[1], ruleName, d.Timeout("delete"))
		if cerr != nil {
			return diag.Errorf("%s", cerr.Error())
		}
		time.Sleep(250 * time.Millisecond)

	}
	log.Printf("[TRACE] resourceS3BucketDelete ******** end")
	return nil
}

func validateTenantBucket(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	tId := diff.Get("tenant_id").(string)
	sbucket := diff.Get("source_bucket").(string)
	c := m.(*duplosdk.Client)
	_, err := c.TenantGetAwsCloudResource(tId, 1, sbucket)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

func s3replicaWaitUntilDelete(ctx context.Context, c *duplosdk.Client, tenantID string, name string, ruleName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"deleted"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.TenantGetV3S3BucketReplication(tenantID, name)
			status := "pending"
			if err == nil {
				l := len(rp.Rule)
				c := 0
				if l == 0 {
					status = "deleted"
				} else {
					for _, r := range rp.Rule {
						if r.Rule != ruleName {
							c++
						}
					}
					if l == c {
						status = "deleted"
					} else {
						status = "pending"
					}
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] redisCacheWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
