package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource for managing an infrastructure's settings.
func resourceInfrastructureSetting() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_infrastructure_setting` manages a infrastructure's configuration in Duplo.\n\n" +
			"Infrastructure settings are initially populated by Duplo when an infrastructure is created.  This resource " +
			"allows you take control of individual configuration settings for a specific infrastructure.",

		ReadContext:   resourceInfrastructureSettingRead,
		CreateContext: resourceInfrastructureSettingCreateOrUpdate,
		UpdateContext: resourceInfrastructureSettingCreateOrUpdate,
		DeleteContext: resourceInfrastructureSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"infra_name": {
				Description: "The name of the infrastructure to configure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
			"custom_data": {
				Description: "A complete list of configuration settings for this infrastructure, even ones not being managed by this resource.",
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

func resourceInfrastructureSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	infraName := d.Id()
	log.Printf("[TRACE] resourceInfrastructureSettingRead(%s): start", infraName)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.InfrastructureGetSetting(infraName)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure settings for '%s': %s", infraName, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("infra_name", duplo.InfraName)
	d.Set("custom_data", keyValueToState("custom_data", duplo.CustomData))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		d.Set("setting", keyValueToState("setting", selectKeyValues(duplo.CustomData, *v)))
	}

	log.Printf("[TRACE] resourceInfrastructureSettingRead(%s): end", infraName)
	return nil
}

func resourceInfrastructureSettingCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	// Parse the identifying attributes
	infraName := d.Get("infra_name").(string)
	log.Printf("[TRACE] resourceInfrastructureSettingCreateOrUpdate(%s): start", infraName)

	// Collect the current state of settings specified by the user.
	c := m.(*duplosdk.Client)
	config, err := c.InfrastructureGetSetting(infraName)
	if err != nil {
		return diag.Errorf("Error retrieving infrastructure settings for '%s': %s", infraName, err)
	}
	var existing *[]duplosdk.DuploKeyStringValue
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		existing = selectKeyValues(config.CustomData, *v)
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
		err = c.InfrastructureReplaceSetting(duplosdk.DuploInfrastructureSetting{InfraName: infraName, CustomData: settings})
	} else {
		err = c.InfrastructureChangeSetting(infraName, existing, settings)
	}
	if err != nil {
		return diag.Errorf("Error updating infrastructure settings for '%s': %s", infraName, err)
	}
	d.SetId(infraName)

	diags := resourceInfrastructureSettingRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureSettingCreateOrUpdate(%s): end", infraName)
	return diags
}

func resourceInfrastructureSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	// Parse the identifying attributes
	infraName := d.Id()
	log.Printf("[TRACE] resourceInfrastructureSettingDelete(%s): start", infraName)

	// Delete the configuration with Duplo
	c := m.(*duplosdk.Client)
	all, err := c.InfrastructureGetSetting(infraName)

	if err != nil {
		return diag.Errorf("Error fetching infrastructure settings for '%s': %s", infraName, err)
	}

	// Get the previous and desired infrastructure settingss
	previous, _ := getInfrastructureSettingChange(all.CustomData, d)
	desired := &[]duplosdk.DuploKeyStringValue{}
	if d.Get("delete_unspecified_settings").(bool) {
		err = c.InfrastructureReplaceSetting(duplosdk.DuploInfrastructureSetting{InfraName: infraName})
	} else {
		err = c.InfrastructureChangeSetting(infraName, previous, desired)
	}

	if err != nil {
		return diag.Errorf("Error deleting infrastructure settings for '%s': %s", infraName, err)
	}

	log.Printf("[TRACE] resourceInfrastructureSettingDelete(%s): end", infraName)
	return nil
}

func getInfrastructureSettingChange(all *[]duplosdk.DuploKeyStringValue, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploKeyStringValue) {
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		previous = selectInfrastructureSettings(all, *v)
	} else {
		previous = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandInfrastructureSetting("setting", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.Key
	}

	// Track the change
	d.Set("specified_settings", specified)

	return
}

func expandInfrastructureSetting(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploKeyStringValue {
	var ary []duplosdk.DuploKeyStringValue

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandInfrastructureSetting ********: found %s", fieldName)
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
func selectInfrastructureSettings(all *[]duplosdk.DuploKeyStringValue, keys []string) *[]duplosdk.DuploKeyStringValue {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectInfrastructureSettingsFromMap(all, specified)
}

func selectInfrastructureSettingsFromMap(all *[]duplosdk.DuploKeyStringValue, keys map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	settings := make([]duplosdk.DuploKeyStringValue, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.Key]; ok {
			settings = append(settings, pc)
		}
	}

	return &settings
}
