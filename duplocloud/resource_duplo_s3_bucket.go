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
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, duplo.Name))
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)

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
		return diag.Errorf("Error creating tenant %s bucket '%s': %s", tenantID, duploObject.Name, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the secret's details.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.TenantGetS3Bucket(tenantID, duploObject.Name)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting tenant %s bucket'%s': %s", tenantID, duploObject.Name, err))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected tenant %s bucket '%s' to be retrieved, but got: nil", tenantID, duploObject.Name))
		}

		// Finally, we can set the ID
		d.SetId(fmt.Sprintf("%s/%s", tenantID, duploObject.Name))
		return nil
	})

	diags := resourceS3BucketRead(ctx, d, m)
	log.Printf("[TRACE] resourceS3BucketCreate ******** end")
	return diags
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

	// Wait up to 60 seconds for Duplo to delete the secret.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.TenantGetS3Bucket(idParts[0], idParts[1])

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting a bucket '%s': %s", id, err))
		}

		if resp != nil {
			return resource.RetryableError(fmt.Errorf("Expected bucket '%s' to be missing, but it still exists", id))
		}

		return nil
	})

	// Wait 60 more seconds to deal with consistency issues.
	time.Sleep(time.Minute)

	log.Printf("[TRACE] resourceS3BucketDelete ******** end")
	return nil
}
