package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func s3BucketSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"fullname": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"enable_versioning": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"enable_access_logs": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"allow_public_access": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"default_encryption": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"method": {
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
		"managed_policies": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"in_tenant_region": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,

			// Supresses diffs for existing resources that were imported, so they have a blank region flag.
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return d.Id() != "" && old == ""
			},
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceS3Bucket() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceS3BucketRead,
		CreateContext: resourceS3BucketCreate,
		UpdateContext: resourceS3BucketUpdate,
		DeleteContext: resourceS3BucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: s3BucketSchema(),
	}
}

/// READ resource
func resourceS3BucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetS3Bucket(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s bucket '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceS3BucketSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceS3BucketRead ******** end")
	return nil
}

/// CREATE resource
func resourceS3BucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketCreate ******** start")

	// Create the request object.
	duploObject := duplosdk.DuploS3BucketRequest{
		Name:           d.Get("name").(string),
		InTenantRegion: d.Get("in_tenant_region").(bool),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateS3Bucket(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error applying tenant %s bucket '%s': %s", tenantID, duploObject.Name, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the bucket's details.
	err = resource.RetryContext(ctx, d.Timeout("create"), func() *resource.RetryError {
		resp, errget := c.TenantGetS3Bucket(tenantID, duploObject.Name)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting tenant %s bucket '%s': %s", tenantID, duploObject.Name, errget))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected tenant %s bucket '%s' to be retrieved, but got: nil", tenantID, duploObject.Name))
		}

		// Finally, we can set the ID
		d.SetId(fmt.Sprintf("%s/%s", tenantID, duploObject.Name))
		return nil
	})
	if err != nil {
		return diag.Errorf("Error applying tenant %s bucket '%s': %s", tenantID, duploObject.Name, err)
	}

	diags := resourceS3BucketUpdate(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceS3BucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketUpdate ******** start")

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
		duploObject.Policies = v
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	resource, err := c.TenantApplyS3BucketSettings(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error applying tenant %s bucket '%s': %s", tenantID, duploObject.Name, err)
	}
	resourceS3BucketSetData(d, tenantID, d.Get("name").(string), resource)

	log.Printf("[TRACE] resourceS3BucketUpdate ******** end")
	return nil
}

/// DELETE resource
func resourceS3BucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceS3BucketDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.TenantDeleteS3Bucket(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting bucket '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the bucket.
	err = resource.RetryContext(ctx, d.Timeout("delete"), func() *resource.RetryError {
		resp, errget := c.TenantGetS3Bucket(idParts[0], idParts[1])

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting a bucket '%s': %s", id, errget))
		}

		if resp != nil {
			return resource.RetryableError(fmt.Errorf("Expected bucket '%s' to be missing, but it still exists", id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("Error deleting s bucket '%s': %s", id, err)
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceS3BucketDelete ******** end")
	return nil
}

func resourceS3BucketSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploS3Bucket) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)
	d.Set("enable_versioning", duplo.EnableVersioning)
	d.Set("enable_access_logs", duplo.EnableAccessLogs)
	d.Set("allow_public_access", duplo.AllowPublicAccess)
	d.Set("default_encryption", []map[string]interface{}{{
		"method": duplo.DefaultEncryption,
	}})
	d.Set("managed_policies", duplo.Policies)
}
