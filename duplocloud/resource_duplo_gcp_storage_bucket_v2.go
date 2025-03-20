package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpStorageBucketV2Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the storage bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the storage bucket.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the storage bucket.",
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
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreIfCaseSensitive,
		},
		"labels": {
			Description: "The labels assigned to this storage bucket.",
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
	fullName, clientErr := c.GetDuploServicesNameWithGcp(tenantID, name, false)
	if clientErr != nil {
		return diag.Errorf("Error fetching tenant prefix for %s : %s", tenantID, clientErr)

	}
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
