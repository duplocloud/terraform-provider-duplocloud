package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	TOTALS3NAMELENGTH = 63
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
					"kms_key_id": {
						Description: "The tenant KMS key ID or ARN to use for encryption.  Only applicable when `method` is `TenantKms`.  " +
							"When omitted, the default tenant KMS key is used.",
						Type:             schema.TypeString,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: diffSuppressS3KmsKeyId,
					},
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
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
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
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema:        s3BucketSchema(),
		CustomizeDiff: customizeS3BucketEncryptionDiff,
	}
}

// s3ConfiguredKmsKeyId extracts default_encryption[0].kms_key_id from the raw
// config (works with both *schema.ResourceData and *schema.ResourceDiff via
// GetRawConfig). ok is false when the block or attribute is absent/unknown - this
// is what lets us tell "user dropped the key" (revert to default) apart from the
// stale value a Computed attribute would otherwise report through d.Get.
func s3ConfiguredKmsKeyId(raw cty.Value) (value string, ok bool) {
	if raw.IsNull() || !raw.IsKnown() {
		return "", false
	}
	de := raw.GetAttr("default_encryption")
	if de.IsNull() || !de.IsKnown() || de.LengthInt() == 0 {
		return "", false
	}
	block := de.AsValueSlice()[0]
	if !block.IsKnown() {
		return "", false
	}
	key := block.GetAttr("kms_key_id")
	if !key.IsKnown() || key.IsNull() {
		return "", false
	}
	return key.AsString(), true
}

// customizeS3BucketEncryptionDiff validates the encryption block and forces an
// update when an explicit CMK is dropped from config.
//
// It reads kms_key_id from GetRawConfig (not diff.Get) so it sees what the user
// actually wrote, not the value a Computed attribute carries over from state -
// otherwise switching away from TenantKms is blocked by a stale key (DUPLO-43356).
//
// default_encryption is a Computed block, so dropping kms_key_id would otherwise
// retain the old value and plan no change. When the key is removed from config but
// state still has one, we SetNew the block with an empty kms_key_id so the revert to
// the default tenant key actually runs. The SetNew is gated on a non-empty prior
// value: once state is empty there is nothing to clear and we leave the diff alone,
// so the plan converges (DUPLO-43358). SetNew (vs SetNewComputed) keeps method shown
// unchanged in the plan - only kms_key_id is flagged as recomputed.
func customizeS3BucketEncryptionDiff(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	method := diff.Get("default_encryption.0.method").(string)
	key, configured := s3ConfiguredKmsKeyId(diff.GetRawConfig())

	if configured {
		trimmed := strings.TrimSpace(key)
		switch {
		case method == "TenantKms" && trimmed == "":
			// DUPLO-43359: "" or whitespace-only is not a valid key.
			return fmt.Errorf("default_encryption.kms_key_id must be a non-empty KMS key ID or ARN when default_encryption.method is \"TenantKms\"")
		case key != trimmed:
			// DUPLO-43361: a padded ARN reaches AWS verbatim and fails with a misleading error.
			return fmt.Errorf("default_encryption.kms_key_id must not contain leading or trailing whitespace")
		case method != "TenantKms" && trimmed != "":
			// The backend only honors kms_key_id under TenantKms.
			return fmt.Errorf("default_encryption.kms_key_id can only be set when default_encryption.method is \"TenantKms\" (got %q)", method)
		}
		// A configured key is handled by the normal diff (diffSuppressS3KmsKeyId
		// absorbs bare-id vs ARN); nothing to force.
		return nil
	}

	// kms_key_id was dropped from config. If state still holds one, plan a clear so
	// the bucket reverts to the default tenant key.
	if old, _ := diff.GetChange("default_encryption.0.kms_key_id"); old != nil {
		if oldStr, ok := old.(string); ok && oldStr != "" {
			return diff.SetNew("default_encryption", []interface{}{
				map[string]interface{}{
					"method":     method,
					"kms_key_id": "",
				},
			})
		}
	}
	return nil
}

// diffSuppressS3KmsKeyId suppresses spurious drift between a bare KMS key ID supplied
// in config and the full key ARN the backend returns (which ends in "key/<key-id>").
func diffSuppressS3KmsKeyId(k, old, new string, d *schema.ResourceData) bool {
	if old == new {
		return true
	}
	if old == "" || new == "" {
		return false
	}
	// Treat an ARN and the bare key ID it embeds as equivalent.
	return strings.HasSuffix(old, "/"+new) || strings.HasSuffix(new, "/"+old)
}

// isS3KmsKeyError reports whether a v3 API error is about the KMS key rather than a
// genuinely missing endpoint. Such errors come back as 404s, which PossibleMissingAPI
// would otherwise misread as "v3 unavailable" and silently fall back to the old API.
func isS3KmsKeyError(err duplosdk.ClientError) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "kms")
}

// validateS3CmkRegion guards against a CMK paired with an explicit non-tenant region.
// The backend only applies a tenant KMS key in the tenant's default region; with any
// other region it silently drops the CMK and creates the bucket in the tenant region
// anyway (DUPLO-43365). We can only know the tenant region by asking the API, so this
// runs at apply time. It is a no-op unless a kms_key_id is set alongside an explicitly
// configured region, and it never blocks when the region can't be determined.
func validateS3CmkRegion(c *duplosdk.Client, tenantID string, d *schema.ResourceData, req *duplosdk.DuploS3BucketSettingsRequest) error {
	if req.EncryptionKmsKeyId == "" {
		return nil
	}
	regionRaw := d.GetRawConfig().GetAttr("region")
	if !regionRaw.IsKnown() || regionRaw.IsNull() {
		return nil // region omitted -> tenant region is used -> no conflict
	}
	region := regionRaw.AsString()
	if region == "" {
		return nil
	}
	tenantRegion, err := c.TenantGetAwsRegion(tenantID)
	if err != nil || tenantRegion == "" {
		return nil // can't determine the tenant region; don't block the apply
	}
	if !strings.EqualFold(region, tenantRegion) {
		return fmt.Errorf("default_encryption.kms_key_id (method \"TenantKms\") is only supported for buckets in the tenant's default region (%s); remove kms_key_id or set region to %q", tenantRegion, tenantRegion)
	}
	return nil
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

	tenantID := d.Get("tenant_id").(string)

	// prefix + name based on settings
	fullName, errname := c.GetDuploServicesNameWithAws(tenantID, name)
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}
	if errname != nil {
		return diag.Errorf("resourceS3BucketCreate: Unable to retrieve duplo service name (name: %s, error: %s)", name, errname.Error())
	}
	if !validateStringLength(fullName, TOTALS3NAMELENGTH) {
		return diag.Errorf("resourceS3BucketCreate: fullname %s exceeds allowable bucket name length %d)", fullName, TOTALS3NAMELENGTH)

	}
	// Create the request object.
	duploObject := duplosdk.DuploS3BucketSettingsRequest{
		Name: name,
	}
	errFill := fillS3BucketRequest(&duploObject, d)
	if errFill != nil {
		return diag.FromErr(errFill)
	}
	if err := validateS3CmkRegion(c, tenantID, d, &duploObject); err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	_, err := c.TenantCreateV3S3Bucket(tenantID, duploObject)
	if err != nil {
		// A KMS error is a 404, which PossibleMissingAPI treats as "v3 missing" - do not
		// fall back to the old API for it, or the user sees an unrelated error (DUPLO-43362).
		if err.PossibleMissingAPI() && !isS3KmsKeyError(err) {
			return resourceS3BucketCreateOldApi(ctx, d, m)
		}
		return diag.Errorf("resourceS3BucketCreate: Unable to create s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
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

	if err := validateS3CmkRegion(c, tenantID, d, &duploObject); err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	_, err := c.TenantUpdateV3S3Bucket(tenantID, duploObject)
	if err != nil {
		// A KMS error is a 404, which PossibleMissingAPI treats as "v3 missing" - do not
		// fall back to the old API for it, or the user sees an unrelated error (DUPLO-43362).
		if err.PossibleMissingAPI() && !isS3KmsKeyError(err) {
			return resourceS3BucketUpdateOldApi(ctx, d, m)
		}
		return diag.Errorf("resourceS3BucketUpdate: Unable to update s3 bucket using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, duploObject.Name, err)
	}

	// Re-read so state reflects the reconciled values rather than the update
	// response, which can be stale (e.g. the KMS tag removal when reverting from a
	// CMK to default/Sse is not reflected in the echoed-back object).
	resource, err := c.TenantGetV3S3Bucket(tenantID, fullname)
	if err != nil {
		return diag.Errorf("resourceS3BucketUpdate: Unable to retrieve s3 bucket details using v3 api (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
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
		if err.Status() == 404 {
			log.Printf("[TRACE] resourceS3BucketDelete(%s): object not found", id)
			return nil
		}
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
	// Read kms_key_id from raw config (not the Computed-merged value) so dropping
	// it from config sends "" and the backend reverts to the default tenant key.
	if key, ok := s3ConfiguredKmsKeyId(d.GetRawConfig()); ok {
		duploObject.EncryptionKmsKeyId = strings.TrimSpace(key)
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
		"method":     duplo.DefaultEncryption,
		"kms_key_id": duplo.EncryptionKmsKeyId,
	}})
	userAdded, ok := d.GetOk("managed_policies")
	if ok {
		// ensure the key exists in the state
		// Filter out policies that do not exist in userAdded
		filteredPolicies := []string{}
		userPoliciesSet := make(map[string]struct{})
		for _, v := range userAdded.([]interface{}) {
			if s, ok := v.(string); ok {
				userPoliciesSet[s] = struct{}{}
			}
		}

		for _, policy := range duplo.Policies {
			if _, exists := userPoliciesSet[policy]; exists {
				filteredPolicies = append(filteredPolicies, policy)
			}
		}
		duplo.Policies = filteredPolicies
	}
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
	// Read kms_key_id from raw config (not the Computed-merged value) so dropping
	// it from config sends "" and the backend reverts to the default tenant key.
	if key, ok := s3ConfiguredKmsKeyId(d.GetRawConfig()); ok {
		duploObject.EncryptionKmsKeyId = strings.TrimSpace(key)
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
