package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func duploPlanSettingsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"plan_id": {
			Description: "The ID of the plan to configure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"unrestricted_ext_lb": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"dns_setting": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"domain_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"internal_dns_suffix": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"external_dns_suffix": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"ignore_global_dns": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"metadata": {
			Description: "A list of metadata for the plan to manage.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        KeyValueSchema(),
		},
		"all_metadata": {
			Description: "A complete list of metadata for this plan, even ones not being managed by this resource.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"specified_metadata": {
			Description:      "A list of metadata being managed by this resource.",
			Type:             schema.TypeList,
			Elem:             &schema.Schema{Type: schema.TypeString},
			Computed:         true,
			Optional:         true,
			DiffSuppressFunc: diffSuppressSpecifiedMetadata, //if removed it notifies the change when there is no change
		},
	}
}

func resourcePlanSettings() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_settings` manages an plan settings in Duplo.",

		ReadContext:   resourcePlanSettingsRead,
		CreateContext: resourcePlanSettingsCreateOrUpdate,
		UpdateContext: resourcePlanSettingsCreateOrUpdate,
		DeleteContext: resourcePlanSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploPlanSettingsSchema(),
	}
}

func resourcePlanSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanSettingsRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// Get "special" plan settings.
	settings, err := c.PlanGetSettings(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan settings for '%s': %s", planID, err)
	}

	// Get plan DNS config.  If the config is "global", that means there is no plan DNS config.
	dns, err := c.PlanGetDnsConfig(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan DNS config for '%s': %s", planID, err)
	}
	if dns != nil && dns.IsGlobalDNS {
		dns = nil
	}

	// Get plan metadata.
	allMetadata, err := c.PlanMetadataGetList(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan metadata for '%s': %s", planID, err)
	}

	// Set the simple fields first.
	d.Set("plan_id", planID)
	d.Set("all_metadata", keyValueToState("all_metadata", allMetadata))
	d.Set("unrestricted_ext_lb", settings.UnrestrictedExtLB)
	if dns != nil {
		d.Set("dns_setting", flattenDnsSetting(dns))
	}

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_metadata"); ok && v != nil {
		d.Set("metadata", keyValueToState("metadata", selectPlanMetadata(allMetadata, *v)))
	}

	log.Printf("[TRACE] resourcePlanSettingsRead(%s): end", planID)
	return nil
}

func resourcePlanSettingsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanSettingsCreateOrUpdate(%s): start", planID)

	c := m.(*duplosdk.Client)
	// Apply "special" plan settings.
	if d.HasChange("unrestricted_ext_lb") {
		unRestrictedExtLB := d.Get("unrestricted_ext_lb").(bool)
		settings := duplosdk.DuploPlanSettings{
			UnrestrictedExtLB: unRestrictedExtLB,
		}
		_, err := c.PlanUpdateSettings(planID, &settings)
		if err != nil {
			return diag.Errorf("failed to apply plan settings for '%s': %s", planID, err)
		}
	}

	// Apply plan DNS settings.
	if v, ok := d.GetOk("dns_setting"); ok {
		dns := expandDnsSetting(v.([]interface{})[0].(map[string]interface{}))
		_, err := c.PlanUpdateDnsConfig(planID, dns)
		if err != nil {
			return diag.Errorf("failed to apply plan DNS config for '%s': %s", planID, err)
		}
	}

	// Apply plan metadata
	if _, ok := d.GetOk("metadata"); ok || d.HasChange("metadata") {
		allMetadata, err := c.PlanMetadataGetList(planID)
		if err != nil {
			return diag.Errorf("failed to retrieve plan metadata for '%s': %s", planID, err)
		}

		previous, desired := getPlanMetadataChange(allMetadata, d)
		err = c.PlanChangeMetadata(planID, previous, desired)
		if err != nil {
			return diag.Errorf("Error updating plan configs for '%s': %s", planID, err)
		}
	}

	d.SetId(planID)
	diags := resourcePlanSettingsRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanSettingsCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanSettingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanSettingsDelete(%s): start", planID)

	c := m.(*duplosdk.Client)

	// Get "special" plan settings.
	settings, err := c.PlanGetSettings(planID)
	if err != nil && err.Status() != 404 {
		return diag.Errorf("failed to retrieve plan settings for '%s': %s", planID, err)
	}

	// Skip if plan does not exist.
	if settings == nil {
		return nil
	}

	// Undo the plan settings we can control here.
	if _, ok := d.GetOk("unrestricted_ext_lb"); ok {
		settings.UnrestrictedExtLB = false
		_, err := c.PlanUpdateSettings(planID, settings)
		if err != nil {
			return diag.Errorf("failed to remove plan settings for '%s': %s", planID, err)
		}
	}

	// Undo the plan DNS.
	if _, ok := d.GetOk("dns_setting"); ok {
		err := c.PlanDeleteDnsConfig(planID)
		if err != nil {
			return diag.Errorf("failed to remove plan DNS config for '%s': %s", planID, err)
		}
	}

	// Undo the plan metadata
	if _, ok := d.GetOk("metadata"); ok {
		allMetadata, err := c.PlanMetadataGetList(planID)
		if err != nil {
			return diag.Errorf("failed to retrieve plan metadata for '%s': %s", planID, err)
		}

		// Get the previous and desired plan configs
		previous, _ := getPlanMetadataChange(allMetadata, d)
		desired := &[]duplosdk.DuploKeyStringValue{}

		// Apply the changes via Duplo
		err = c.PlanChangeMetadata(planID, previous, desired)
		if err != nil {
			return diag.Errorf("failed to remove plan metadata for '%s': %s", planID, err)
		}
	}

	log.Printf("[TRACE] resourcePlanSettingsDelete(%s): end", planID)
	return nil
}

func flattenDnsSetting(duplo *duplosdk.DuploPlanDnsConfig) []interface{} {
	m := map[string]interface{}{
		"domain_id":           duplo.DomainId,
		"internal_dns_suffix": duplo.InternalDnsSuffix,
		"external_dns_suffix": duplo.ExternalDnsSuffix,
		"ignore_global_dns":   duplo.IgnoreGlobalDNS,
	}
	return []interface{}{m}
}

func expandDnsSetting(m map[string]interface{}) *duplosdk.DuploPlanDnsConfig {
	dns := duplosdk.DuploPlanDnsConfig{}

	if v, ok := m["domain_id"]; ok {
		dns.DomainId = v.(string)
	}
	if v, ok := m["internal_dns_suffix"]; ok {
		dns.InternalDnsSuffix = v.(string)
	}
	if v, ok := m["external_dns_suffix"]; ok {
		dns.ExternalDnsSuffix = v.(string)
	}
	if v, ok := m["ignore_global_dns"]; ok {
		dns.IgnoreGlobalDNS = v.(bool)
	}

	return &dns
}

func getPlanMetadataChange(all *[]duplosdk.DuploKeyStringValue, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploKeyStringValue) {
	log.Printf("[TRACE] getPlanMetadataChange(%s): start", all)
	if v, ok := getAsStringArray(d, "specified_metadata"); ok && v != nil {
		previous = selectPlanMetadata(all, *v)
	} else {
		previous = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of metadata specified by the user.
	desired = keyValueFromState("metadata", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.Key
	}
	log.Printf("[TRACE] specified_metadata - (%s):", specified)

	// Track the change
	d.Set("specified_metadata", specified)
	log.Printf("[TRACE] getPlanMetadataChange(%s): end", all)
	return
}

func selectPlanMetadata(all *[]duplosdk.DuploKeyStringValue, keys []string) *[]duplosdk.DuploKeyStringValue {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanMetadataFromMap(all, specified)
}

func selectPlanMetadataFromMap(all *[]duplosdk.DuploKeyStringValue, keys map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	mds := make([]duplosdk.DuploKeyStringValue, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.Key]; ok {
			mds = append(mds, pc)
		}
	}

	return &mds
}
