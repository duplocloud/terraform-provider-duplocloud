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
			"setting": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     KeyValueSchema(),
			},
			"delete_unspecified_settings": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     KeyValueSchema(),
			},
			"specified_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTenantConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
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

	// Set the simple fields first.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("metadata", duplosdk.KeyValueToState("metadata", duplo.Metadata))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		d.Set("setting", duplosdk.KeyValueToState("setting", selectTenantConfig(duplo.Metadata, *v)))
	}

	log.Printf("[TRACE] resourceTenantConfigRead(%s): end", tenantID)
	return nil
}

func resourceTenantConfigCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceTenantConfigCreateOrUpdate(%s): start", tenantID)

	// Collect the current state of settings specified by the user.
	c := m.(*duplosdk.Client)
	config, err := c.TenantGetConfig(tenantID)
	if err != nil {
		return diag.Errorf("Error retrieving tenant config for '%s': %s", tenantID, err)
	}
	var existing *[]duplosdk.DuploKeyStringValue
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		existing = selectTenantConfig(config.Metadata, *v)
	} else {
		existing = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of settings specified by the user.
	settings := duplosdk.KeyValueFromState("setting", d)
	specified := make([]string, len(*settings))
	for i, kv := range *settings {
		specified[i] = kv.Key
	}
	d.Set("specified_settings", specified)

	// Apply the changes via Duplo
	if d.Get("delete_unspecified_settings").(bool) {
		err = c.TenantReplaceConfig(duplosdk.DuploTenantConfig{TenantID: tenantID, Metadata: settings})
	} else {
		err = c.TenantChangeConfig(tenantID, existing, settings)
	}
	if err != nil {
		return diag.Errorf("Error updating tenant config for '%s': %s", tenantID, err)
	}
	d.SetId(tenantID)

	diags := resourceTenantConfigRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantConfigCreateOrUpdate(%s): end", tenantID)
	return diags
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

	diags := resourceTenantConfigRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantConfigDelete(%s): end", tenantID)
	return diags
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectTenantConfig(metadata *[]duplosdk.DuploKeyStringValue, keys []string) *[]duplosdk.DuploKeyStringValue {
	specified := map[string]struct{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	settings := make([]duplosdk.DuploKeyStringValue, 0, len(keys))
	for _, kv := range *metadata {
		if _, ok := specified[kv.Key]; ok {
			settings = append(settings, kv)
		}
	}

	return &settings
}
