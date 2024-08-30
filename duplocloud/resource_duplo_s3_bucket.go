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

func s3BucketSchema() map[string]*schema.Schema {
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
		"arn": {
			Description: "The ARN of the S3 bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"domain_name": {
			Description: "The domain name of the S3 bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_versioning": {
			Description: "Whether or not to enable versioning.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"enable_access_logs": {
			Description: "Whether or not to enable access logs.  When enabled, Duplo will send access logs to a centralized S3 bucket per plan.",
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
		"region": {
			Description: "The region of the S3 bucket.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"managed_policies": {
			Description: "Duplo can manage your S3 bucket policy for you, based on simple list of policy keywords:\n\n" +
				" - `\"ssl\"`: Require SSL / HTTPS when accessing the bucket.\n" +
				" - `\"ignore\"`: If this value is present, Duplo will not manage your bucket policy.\n",
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"tags": awsTagsKeyValueSchemaComputed(),
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceS3Bucket() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_s3_bucket` manages an s3 bucket in Duplo.",
		ReadContext:   resourceS3BucketRead,
		CreateContext: resourceS3BucketCreate,
		UpdateContext: resourceS3BucketUpdate,
		DeleteContext: resourceS3BucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: s3BucketSchema(),
	}
}

// READ resource
func resourceS3BucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("resourceS3BucketRead: Invalid resource (ID: %s)", id)
	}
	tenantID, name := idParts[0], idParts[1]

	c := m.(*duplosdk.Client)

	// Figure out the full resource name.
	features, _ := c.AdminGetSystemFeatures()
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceS3BucketRead: Unable to retrieve duplo service name (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}

	// Get the object from Duplo
	duplo, err := c.TenantGetV3S3Bucket(tenantID, fullName)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceS3BucketRead: Unable to retrieve s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}

	// **** fallback on older api ****
	if err != nil && err.PossibleMissingAPI() {
		duplo, err = c.TenantGetS3BucketSettings(tenantID, name)
		if duplo == nil {
			d.SetId("") // object missing
			return nil
		}
		if err != nil {
			return diag.Errorf("resourceS3BucketRead: Unable to retrieve s3 bucket settings (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
		}
	}

	// Set simple fields first.
	resourceS3BucketSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceS3BucketRead ******** end")
	return nil
}

// CREATE resource
func resourceS3BucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketCreate ******** start")
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)
	features, _ := c.AdminGetSystemFeatures()

	s3MaxLength := 63 - MAX_DUPLOSERVICES_AND_SUFFIX_LENGTH
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		s3MaxLength = 63
	}
	if len(name) > s3MaxLength {
		return diag.Errorf("resourceS3BucketCreate: Invalid s3 bucket name: %s, Length must be in the range: (1 - %d)", name, s3MaxLength)
	}
	tenantID := d.Get("tenant_id").(string)

	// prefix + name based on settings
	fullName, errname := c.GetDuploServicesNameWithAws(tenantID, name)
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}
	if errname != nil {
		return diag.Errorf("resourceS3BucketCreate: Unable to retrieve duplo service name (name: %s, error: %s)", name, errname.Error())
	}

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketSettingsRequest{
		Name: name,
	}
	errFill := fillS3BucketRequest(&duploObject, d)
	if errFill != nil {
		return diag.FromErr(errFill)
	}

	// Post the object to Duplo
	_, err := c.TenantCreateV3S3Bucket(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceS3BucketCreate: Unable to create s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	// **** fallback on older api ****
	if err != nil && err.PossibleMissingAPI() {
		return resourceS3BucketCreateOldApi(ctx, d, m)
	}

	// Wait up to 60 seconds for Duplo to be able to return the bucket's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "S3 bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetV3S3Bucket(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	duploObject.Name = fullName
	_, err = c.TenantUpdateV3S3Bucket(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	duplo, err := c.TenantGetV3S3Bucket(tenantID, fullName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("resourceS3BucketCreate: Unable to retrieve s3 bucket details using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}

	// Set simple fields first.
	resourceS3BucketSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceS3BucketCreate ******** end")
	return diags
}

// UPDATE resource
func resourceS3BucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketUpdate ******** start")

	fullname := d.Get("fullname").(string)
	name := d.Get("name").(string)

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketSettingsRequest{
		Name: fullname,
	}

	errName := fillS3BucketRequest(&duploObject, d)
	if errName != nil {
		return diag.FromErr(errName)
	}
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	resource, err := c.TenantUpdateV3S3Bucket(tenantID, duploObject)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("resourceS3BucketUpdate: Unable to update s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}
	// **** fallback on older api ****
	if err != nil && err.PossibleMissingAPI() {
		return resourceS3BucketUpdateOldApi(ctx, d, m)
	}

	resourceS3BucketSetData(d, tenantID, name, resource)

	log.Printf("[TRACE] resourceS3BucketUpdate ******** end")
	return nil
}

// DELETE resource
func resourceS3BucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

// CREATE resource older
func resourceS3BucketCreateOldApi(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketCreateOldApi ******** start")
	name := d.Get("name").(string)
	// Create the request object.
	duploObject := duplosdk.DuploS3BucketRequest{
		Name:           name,
		InTenantRegion: true,
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateS3Bucket(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("resourceS3BucketCreateOldApi: Unable to create s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the bucket's details.
	id := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "S3 bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetS3Bucket(tenantID, duploObject.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceS3BucketUpdateOldApi(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketCreateOldApi ******** end")
	return diags
}

// UPDATE resource older
func resourceS3BucketUpdateOldApi(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketUpdateOldApi ******** start")

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketSettingsRequest{
		Name: d.Get("name").(string),
	}

	// Set the object versioning
	if v, ok := d.GetOk("enable_versioning"); ok && v != nil {
		duploObject.EnableVersioning = v.(bool)
	}

	// Set the access logs flag
	if v, ok := d.GetOk("enable_access_logs"); ok && v != nil {
		duploObject.EnableAccessLogs = v.(bool)
	}

	// Set the public access block.
	if v, ok := d.GetOk("allow_public_access"); ok && v != nil {
		duploObject.AllowPublicAccess = v.(bool)
	}

	// Set the default encryption.
	defaultEncryption, err := getOptionalBlockAsMap(d, "default_encryption")
	if err != nil {
		return diag.FromErr(err)
	}
	if v, ok := defaultEncryption["method"]; ok && v != nil {
		duploObject.DefaultEncryption = v.(string)
	}

	// Set the managed policies.
	if v, ok := getAsStringArray(d, "managed_policies"); ok && v != nil {
		duploObject.Policies = *v
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	resource, err := c.TenantApplyS3BucketSettings(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("resourceS3BucketUpdateOldApi: Unable to update s3 bucket settings (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}
	resourceS3BucketSetData(d, tenantID, d.Get("name").(string), resource)

	log.Printf("[TRACE] resourceS3BucketUpdate ******** end")
	return nil
}

func resourceS3BucketSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploS3Bucket) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("domain_name", duplo.DomainName)
	d.Set("arn", duplo.Arn)
	d.Set("enable_versioning", duplo.EnableVersioning)
	d.Set("enable_access_logs", duplo.EnableAccessLogs)
	d.Set("allow_public_access", duplo.AllowPublicAccess)
	d.Set("default_encryption", []map[string]interface{}{{
		"method": duplo.DefaultEncryption,
	}})
	d.Set("managed_policies", duplo.Policies)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("region", duplo.Region)
}

func fillS3BucketRequest(duploObject *duplosdk.DuploS3BucketSettingsRequest, d *schema.ResourceData) error {
	log.Printf("[TRACE] fillS3BucketRequest ******** start")

	// Set the object versioning
	if v, ok := d.GetOk("enable_versioning"); ok && v != nil {
		duploObject.EnableVersioning = v.(bool)
	}

	// Set the access logs flag
	if v, ok := d.GetOk("enable_access_logs"); ok && v != nil {
		duploObject.EnableAccessLogs = v.(bool)
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
		duploObject.DefaultEncryption = v.(string)
	}

	if v, ok := d.GetOk("region"); ok && v != nil {
		duploObject.Region = v.(string)
	}

	if v, ok := d.GetOk("location"); ok && v != nil {
		duploObject.Location = v.(string)
	}

	// Set the managed policies.
	if v, ok := getAsStringArray(d, "managed_policies"); ok && v != nil {
		duploObject.Policies = *v
	}

	log.Printf("[TRACE] fillS3BucketRequest ******** end")
	return nil
}
