package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpS3BucketSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the S3 bucket.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9._-]*$`), "Invalid S3 bucket name"),

				// NOTE: some validations are moot, because Duplo provides a prefix and suffix for the name:
				//
				// - Bucket names must begin and end with a letter or number.
				// - Bucket names must not be formatted as an IP address (for example, 192.168.5.4).
				// - Bucket names must not start with the prefix xn--.
				// - Bucket names must not end with the suffix -s3alias.
				//
				// Because Duplo automatically prefixes and suffixes bucket names, it is impossible to break any of the rules in the above bulleted list.
			),
		},
		"fullname": {
			Description: "The full name of the S3 bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"domain_name": {
			Description: "Bucket self link.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_versioning": {
			Description: "Whether or not to enable versioning.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},

		"allow_public_access": {
			Description: "Whether or not to remove the public access block from the bucket.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"default_encryption": {
			Description: "Default encryption settings for objects uploaded to the bucket.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"method": {
						Description:  "Default encryption method.  Must be one of: `None`, `Sse`, `AwsKms`, `TenantKms`.",
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "Sse",
						ValidateFunc: validation.StringInSlice([]string{"None", "Sse", "AwsKms", "TenantKms"}, false),
					},
					// RESERVED FOR THE FUTURE
					//
					//"kms_key_id": {
					//	Type:     schema.TypeString,
					//	Optional: true,
					//},
				},
			},
		},
		"location": {
			Description:      "The location is to set region/multi region, applicable for gcp cloud.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceGCPS3Bucket() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceGCPS3BucketRead,
		CreateContext: resourceGCPS3BucketCreate,
		UpdateContext: resourceGCPS3BucketUpdate,
		DeleteContext: resourceGCPS3BucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: gcpS3BucketSchema(),
	}
}

// READ resource
func resourceGCPS3BucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPS3BucketRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceGCPS3BucketRead: Invalid resource (ID: %s)", id)
	}
	tenantID, name := idParts[0], idParts[1]

	c := m.(*duplosdk.Client)

	// Figure out the full resource name.
	fullName := d.Get("fullname").(string)

	// Get the object from Duplo
	duplo, err := c.GCPTenantGetV3S3Bucket(tenantID, fullName)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceGCPS3BucketRead: Unable to retrieve s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}

	// Set simple fields first.
	resourceGcpS3BucketSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceGCPS3BucketRead ******** end")
	return nil
}

// CREATE resource
func resourceGCPS3BucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPS3BucketCreate ******** start")
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)

	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketSettingsRequest{
		Name: name,
	}
	errFill := fillS3BucketRequest(&duploObject, d)
	if errFill != nil {
		return diag.FromErr(errFill)
	}

	// Post the object to Duplo
	resp, err := c.GCPTenantCreateV3S3Bucket(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceGCPS3BucketCreate: Unable to create s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	fullName := resp.Name
	// Wait up to 60 seconds for Duplo to be able to return the bucket's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "S3 bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPTenantGetV3S3Bucket(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	duplo, err := c.GCPTenantGetV3S3Bucket(tenantID, fullName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("resourceGCPS3BucketCreate: Unable to retrieve s3 bucket details using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	d.SetId(id)

	// Set simple fields first.
	resourceGcpS3BucketSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGCPS3BucketCreate ******** end")
	return nil
}

// UPDATE resource
func resourceGCPS3BucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPS3BucketUpdate ******** start")

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
	resource, err := c.GCPTenantUpdateV3S3Bucket(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceGCPS3BucketUpdate: Unable to update s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	resourceGcpS3BucketSetData(d, tenantID, name, resource)

	log.Printf("[TRACE] resourceGCPS3BucketUpdate ******** end")
	return nil
}

// DELETE resource
func resourceGCPS3BucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPS3BucketDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceGCPS3BucketDelete: Invalid resource (ID: %s)", id)
	}
	fullName := d.Get("fullname").(string)
	err := c.GCPTenantDeleteS3Bucket(idParts[0], idParts[1], fullName)
	if err != nil {
		return diag.Errorf("resourceGCPS3BucketDelete: Unable to delete bucket (name:%s, error: %s)", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the bucket.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPTenantGetV3S3Bucket(idParts[0], fullName)
	})
	if diag != nil {
		return diag
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceGCPS3BucketDelete ******** end")
	return nil
}

func resourceGcpS3BucketSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGCPBucket) {
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
