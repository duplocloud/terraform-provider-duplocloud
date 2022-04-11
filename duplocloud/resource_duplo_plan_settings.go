package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
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
			Description: "A list of metadata being managed by this resource.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
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

	duplo, err := c.PlanGet(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
	}
	if duplo == nil {
		return diag.Errorf("Plan could not be found. '%s'", planID)
	}

	flattenPlanSettings(d, duplo)
	log.Printf("[TRACE] resourcePlanSettingsRead(%s): end", planID)
	return nil
}

func resourcePlanSettingsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanSettingsCreateOrUpdate(%s): start", planID)

	c := m.(*duplosdk.Client)

	duplo, err := c.PlanGet(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
	}
	if duplo == nil {
		return diag.Errorf("Plan could not be found. '%s'", planID)
	}
	if v, ok := d.GetOk("unrestricted_ext_lb"); ok {
		duplo.UnrestrictedExtLB = v.(bool)
	}
	if v, ok := d.GetOk("dns_setting"); ok {
		expandDnsSetting(duplo, v.([]interface{})[0].(map[string]interface{}))
	}
	if _, ok := d.GetOk("metadata"); ok {
		log.Printf("[TRACE] Plan metadata from duplo :(%s)", duplo.MetaData)
		previous, desired := getPlanMetadataChange(duplo.MetaData, d)
		log.Printf("[TRACE] Plan metadata previous :(%s), desired(%s)", previous, desired)
		duplo.MetaData = getDesiredMetadataConfigs(duplo.MetaData, desired, previous)
	}

	err = c.PlanUpdate(duplo)
	if err != nil {
		return diag.Errorf("Error updating plan for '%s': %s", planID, err)
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

	duplo, err := c.PlanGet(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
	}

	// Skip if plan does not exist.
	if duplo == nil {
		return nil
	}
	if _, ok := d.GetOk("unrestricted_ext_lb"); ok {
		duplo.UnrestrictedExtLB = false
	}
	if _, ok := d.GetOk("dns_setting"); ok {
		duplo.DnsConfig = nil
	}
	if _, ok := d.GetOk("metadata"); ok {
		existing := duplo.MetaData
		specified := []string{}
		if v, ok := getAsStringArray(d, "specified_metadata"); ok && v != nil {
			specified = *v
		}
		ary := make([]duplosdk.DuploKeyStringValue, 0, len(*existing)-len(specified))
		log.Printf("[TRACE] existing (%s)", existing)
		log.Printf("[TRACE] Specified (%s)", specified)

		for _, e := range *existing {
			present := false
			for _, s := range specified {
				if e.Key == s {
					present = true
					break
				}
			}
			if !present {
				ary = append(ary, e)
			}
		}
		log.Printf("[TRACE] New metadata to be updated (%s)", ary)
		duplo.MetaData = &ary
	}

	err = c.PlanUpdate(duplo)
	if err != nil {
		return diag.Errorf("Error updating plan for '%s': %s", planID, err)
	}
	log.Printf("[TRACE] resourcePlanSettingsDelete(%s): end", planID)
	return nil
}

func flattenPlanSettings(d *schema.ResourceData, duplo *duplosdk.DuploPlan) {
	log.Printf("[TRACE] flattenPlanSettings(%s): start", duplo.Name)
	d.Set("plan_id", duplo.Name)
	d.Set("unrestricted_ext_lb", duplo.UnrestrictedExtLB)
	if duplo.DnsConfig != nil {
		d.Set("dns_setting", flattenDnsSetting(duplo.DnsConfig))
	}
	d.Set("all_metadata", keyValueToState("all_metadata", duplo.MetaData))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_metadata"); ok && v != nil {
		d.Set("metadata", keyValueToState("metadata", selectPlanMetadata(duplo.MetaData, *v)))
	}
	log.Printf("[TRACE] flattenPlanSettings(%s): end", duplo.Name)
}

func flattenDnsSetting(duplo *duplosdk.DuploPlanDnsConfig) []interface{} {
	log.Printf("[TRACE] flatterDnsSetting: start")
	m := make(map[string]interface{})
	m["domain_id"] = duplo.DomainId
	m["internal_dns_suffix"] = duplo.InternalDnsSuffix
	m["external_dns_suffix"] = duplo.ExternalDnsSuffix
	log.Printf("[TRACE] flatterDnsSetting: end")
	return []interface{}{m}
}

func expandDnsSetting(existingPlan *duplosdk.DuploPlan, m map[string]interface{}) {
	if existingPlan.DnsConfig == nil {
		existingPlan.DnsConfig = &duplosdk.DuploPlanDnsConfig{}
	}
	if v, ok := m["domain_id"]; ok {
		existingPlan.DnsConfig.DomainId = v.(string)
	}
	if v, ok := m["internal_dns_suffix"]; ok {
		existingPlan.DnsConfig.InternalDnsSuffix = v.(string)
	}
	if v, ok := m["external_dns_suffix"]; ok {
		existingPlan.DnsConfig.ExternalDnsSuffix = v.(string)
	}
}

func getPlanMetadataChange(all *[]duplosdk.DuploKeyStringValue, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploKeyStringValue) {
	log.Printf("[TRACE] getPlanMetadataChange(%s): start", all)
	if v, ok := getAsStringArray(d, "specified_metadata"); ok && v != nil {
		previous = selectPlanMetadata(all, *v)
	} else {
		previous = &[]duplosdk.DuploKeyStringValue{}
	}

	// Collect the desired state of settings specified by the user.
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

func getDesiredMetadataConfigs(existing, newMetadata, oldMetadata *[]duplosdk.DuploKeyStringValue) *[]duplosdk.DuploKeyStringValue {

	// Next, update all metada that are present, keeping a record of each one that is present
	log.Printf("[TRACE] existing-(%s), oldMetadata-(%s), newMetadata-(%s):", existing, oldMetadata, newMetadata)
	desired := make([]duplosdk.DuploKeyStringValue, 0, len(*existing)-len(*oldMetadata)+len(*newMetadata))

	for _, emd := range *existing {
		present := false
		for _, omd := range *oldMetadata {
			if emd.Key == omd.Key {
				present = true
				break
			}
		}
		if !present {
			desired = append(desired, emd)
		}
	}
	if len(*newMetadata) > 0 {
		desired = append(desired, *newMetadata...)
	}
	log.Printf("[TRACE] desired-(%s):", desired)
	return &desired
}
