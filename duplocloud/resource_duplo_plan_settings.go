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
	if duplo == nil {
		return diag.Errorf("Plan could not be found. '%s'", planID)
	}
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
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
	if duplo == nil {
		return diag.Errorf("Plan could not be found. '%s'", planID)
	}
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
	}
	if v, ok := d.GetOk("unrestricted_ext_lb"); ok {
		duplo.UnrestrictedExtLB = v.(bool)
	}
	if v, ok := d.GetOk("dns_setting"); ok {
		expandDnsSetting(duplo, v.([]interface{})[0].(map[string]interface{}))
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
	// Skip if plan does not exist.
	if duplo == nil {
		return nil
	}
	if err != nil {
		return diag.Errorf("failed to retrieve plan for '%s': %s", planID, err)
	}
	if _, ok := d.GetOk("unrestricted_ext_lb"); ok {
		duplo.UnrestrictedExtLB = false
	}
	if _, ok := d.GetOk("dns_setting"); ok {
		duplo.DnsConfig = nil
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
	if v, ok := m["domain_id"]; ok && v != "" {
		existingPlan.DnsConfig.DomainId = v.(string)
	}
	if v, ok := m["internal_dns_suffix"]; ok && v != "" {
		existingPlan.DnsConfig.InternalDnsSuffix = v.(string)
	}
	if v, ok := m["external_dns_suffix"]; ok && v != "" {
		existingPlan.DnsConfig.ExternalDnsSuffix = v.(string)
	}
}
