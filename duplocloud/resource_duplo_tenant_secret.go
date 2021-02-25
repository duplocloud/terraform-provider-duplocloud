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

func tenantSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name_suffix": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"data": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Sensitive:        true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"rotation_enabled": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceTenantSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceTenantSecretRead,
		CreateContext: resourceTenantSecretCreate,
		DeleteContext: resourceTenantSecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: tenantSecretSchema(),
	}
}

/// READ resource
func resourceTenantSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] rresourceTenantSecretRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetSecretByName(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s secret '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, duplo.Name))
	d.Set("tenant_id", tenantID)
	d.Set("name", duplo.Name)
	d.Set("arn", duplo.Arn)
	d.Set("rotation_enabled", duplo.RotationEnabled)

	// Set name suffix.
	nameParts := strings.SplitN(duplo.Name, "-", 3)
	if len(nameParts) == 3 {
		d.Set("name_suffix", nameParts[2])
	}

	// Set tags
	d.Set("tags", duplosdk.KeyValueToState("tags", duplo.Tags))

	log.Printf("[TRACE] resourceTenantSecretRead ******** end")
	return nil
}

/// CREATE resource
func resourceTenantSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceTenantSecretCreate ******** start")

	// Create the request object.
	duploObject := duplosdk.DuploTenantSecretRequest{
		Name:         d.Get("name_suffix").(string),
		SecretString: d.Get("data").(string),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateSecret(tenantID, &duploObject)
	if err != nil {
		return diag.Errorf("Error creating tenant %s secret '%s': %s", tenantID, duploObject.Name, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the secret's details.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.TenantGetSecretByNameSuffix(tenantID, duploObject.Name)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting tenant %s secret '%s': %s", tenantID, duploObject.Name, err))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected tenant %s secret '%s' to be retrieved, but got: nil", tenantID, duploObject.Name))
		}

		// Finally, we can set the ID
		d.SetId(fmt.Sprintf("%s/%s", tenantID, resp.Name))
		return nil
	})

	diags := resourceTenantSecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantSecretCreate ******** end")
	return diags
}

/// DELETE resource
func resourceTenantSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceTenantSecretDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.TenantDeleteSecret(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting secret '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the secret.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.TenantGetSecretByName(idParts[0], idParts[1])

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting s secret '%s': %s", id, err))
		}

		if resp != nil {
			return resource.RetryableError(fmt.Errorf("Expected secret '%s' to be missing, but it still exists", id))
		}

		return nil
	})

	// Wait 60 more seconds to deal with consistency issues.
	time.Sleep(time.Minute)

	log.Printf("[TRACE] resourceTenantSecretDelete ******** end")
	return nil
}
