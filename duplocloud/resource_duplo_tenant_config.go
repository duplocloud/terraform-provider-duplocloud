package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an AWS ElasticSearch instance
func resourceTenantConfig() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant_config` manages a tenant's configuration in Duplo.\n\n" +
			"Tenant configuration is initially populated by Duplo when a tenant is created.  This resource " +
			"allows you take control of individual configuration settings for a specific tenant.",

		ReadContext:   resourceTenantConfigRead,
		CreateContext: resourceTenantConfigCreateOrUpdate,
		UpdateContext: resourceTenantConfigCreateOrUpdate,
		DeleteContext: resourceTenantConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant to configure.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"setting": {
				Description: "A list of configuration settings to manage, expressed as key / value pairs.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        KeyValueSchema(),
			},
			"delete_unspecified_settings": {
				Description: "Whether or not this resource should delete any settings not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"metadata": {
				Description: "A complete list of configuration settings for this tenant, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        KeyValueSchema(),
			},
			"specified_settings": {
				Description: "A list of configuration setting key being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
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

	// Set the simple fields first.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("metadata", keyValueToState("metadata", duplo.Metadata))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		d.Set("setting", keyValueToState("setting", selectKeyValues(duplo.Metadata, *v)))
	}

	log.Printf("[TRACE] resourceTenantConfigRead(%s): end", tenantID)
	return nil
}

func resourceTenantConfigCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

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
		existing = selectKeyValues(config.Metadata, *v)
	} else {
		existing = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of settings specified by the user.
	settings := keyValueFromState("setting", d)
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
	var err error
	// Parse the identifying attributes
	tenantID := d.Id()
	log.Printf("[TRACE] resourceTenantConfigDelete(%s): start", tenantID)

	// Delete the configuration with Duplo
	c := m.(*duplosdk.Client)
	all, err := c.TenantGetConfig(tenantID)

	if err != nil {
		return diag.Errorf("Error fetching tenant config for '%s': %s", tenantID, err)
	}

	// Get the previous and desired tenant configs
	previous, _ := getTenantConfigChange(all.Metadata, d)
	desired := &[]duplosdk.DuploKeyStringValue{}
	if d.Get("delete_unspecified_settings").(bool) {
		err = c.TenantReplaceConfig(duplosdk.DuploTenantConfig{TenantID: tenantID})
	} else {
		err = c.TenantChangeConfig(tenantID, previous, desired)
	}

	if err != nil {
		return diag.Errorf("Error deleting tenant config for '%s': %s", tenantID, err)
	}

	log.Printf("[TRACE] resourceTenantConfigDelete(%s): end", tenantID)
	return nil
}

func getTenantConfigChange(all *[]duplosdk.DuploKeyStringValue, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploKeyStringValue) {
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		previous = selectTenantConfigs(all, *v)
	} else {
		previous = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandTenantConfig("setting", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.Key
	}

	// Track the change
	d.Set("specified_settings", specified)

	return
}

func expandTenantConfig(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploKeyStringValue {
	var ary []duplosdk.DuploKeyStringValue

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandTenantConfig ********: found %s", fieldName)
		ary = make([]duplosdk.DuploKeyStringValue, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploKeyStringValue{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectTenantConfigs(all *[]duplosdk.DuploKeyStringValue, keys []string) *[]duplosdk.DuploKeyStringValue {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectTenantConfigsFromMap(all, specified)
}

func selectTenantConfigsFromMap(all *[]duplosdk.DuploKeyStringValue, keys map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	settings := make([]duplosdk.DuploKeyStringValue, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.Key]; ok {
			settings = append(settings, pc)
		}
	}

	return &settings
}
