package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource for managing an AWS ElasticSearch instance
func resourceTenantConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceTenantConfigRead,
		CreateContext: resourceTenantConfigCreateOrUpdate,
		UpdateContext: resourceTenantConfigCreateOrUpdate,
		DeleteContext: resourceTenantConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     KeyValueSchema(),
			},
		},
	}
}

func resourceTenantConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Id()
	log.Printf("[TRACE] resourceTenantConfigRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetConfig(tenantID)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant config for '%s': %s", tenantID, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the fields
	d.Set("tenant_id", duplo.TenantID)
	d.Set("metadata", duplosdk.KeyValueToState("metadata", duplo.Metadata))

	log.Printf("[TRACE] resourceTenantConfigRead(%s): end", tenantID)
	return nil
}

func resourceTenantConfigCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Id()
	log.Printf("[TRACE] resourceTenantConfigCreate(%s): start", tenantID)

	// Build the request
	config := duplosdk.DuploTenantConfig{
		TenantID: tenantID,
		Metadata: duplosdk.KeyValueFromState("metadata", d),
	}

	// Apply the changes via Duplo
	c := m.(*duplosdk.Client)
	err := c.TenantReplaceConfig(config)
	if err != nil {
		return diag.Errorf("Error updating tenant config for '%s': %s", tenantID, err)
	}

	log.Printf("[TRACE] resourceTenantConfigCreate(%s): end", tenantID)
	return nil
}

func resourceTenantConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Id()
	log.Printf("[TRACE] resourceTenantConfigDelete(%s): start", tenantID)

	// Delete the configuration with Duplo
	c := m.(*duplosdk.Client)
	err := c.TenantReplaceConfig(duplosdk.DuploTenantConfig{TenantID: tenantID})
	if err != nil {
		return diag.Errorf("Error deleting tenant config for '%s': %s", tenantID, err)
	}

	log.Printf("[TRACE] resourceTenantConfigDelete(%s): end", tenantID)
	return nil
}
