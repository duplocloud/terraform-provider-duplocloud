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
			Type:      schema.TypeString,
			Required:  true,
			ForceNew:  true,
			Sensitive: true,

			// Supresses diffs for existing resources that were imported, so they have a blank secret data.
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return d.Id() != "" && (old == "" || old == new)
			},
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
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: tenantSecretSchema(),
	}
}

/// READ resource
func resourceTenantSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	log.Printf("[TRACE] resourceTenantSecretRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetSecretByName(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve secret '%s': %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
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

	log.Printf("[TRACE] resourceTenantSecretRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceTenantSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	duploObject := duplosdk.DuploTenantSecretRequest{
		Name:         d.Get("name_suffix").(string),
		SecretString: d.Get("data").(string),
	}

	log.Printf("[TRACE] resourceTenantSecretCreate(%s): start", duploObject.Name)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateSecret(tenantID, &duploObject)
	if err != nil {
		return diag.Errorf("Error creating tenant %s secret '%s': %s", tenantID, duploObject.Name, err)
	}
	tempID := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)

	// Wait for Duplo to be able to return the secret's details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "tenant secret", tempID, func() (interface{}, error) {
		rp, errget := c.TenantGetSecretByNameSuffix(tenantID, duploObject.Name)
		if errget == nil && rp != nil {
			d.SetId(fmt.Sprintf("%s/%s", tenantID, rp.Name))
		}
		return rp, errget
	})
	if diags == nil {
		diags = resourceTenantSecretRead(ctx, d, m)
	}
	log.Printf("[TRACE] resourceTenantSecretCreate(%s): end", duploObject.Name)
	return diags
}

/// DELETE resource
func resourceTenantSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	tenantID, name := idParts[0], idParts[1]

	log.Printf("[TRACE] resourceTenantSecretDelete(%s, %s): start", tenantID, name)

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	err := c.TenantDeleteSecret(tenantID, name)
	if err != nil {
		return diag.Errorf("Error deleting secret '%s': %s", id, err)
	}

	// Wait for Duplo to delete the secret.
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "tenant secret", id, func() (interface{}, error) {
		return c.TenantGetSecretByName(tenantID, name)
	})

	// Wait 60 more seconds to deal with consistency issues.
	if diags == nil {
		time.Sleep(time.Minute)
	}

	log.Printf("[TRACE] resourceTenantSecretDelete(%s, %s): end", tenantID, name)
	return diags
}
