package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func duploComputedPlanSettingsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"plan_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"unrestricted_ext_lb": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"dns_setting": {
			Type:     schema.TypeList,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"domain_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"internal_dns_suffix": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"external_dns_suffix": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"ignore_global_dns": {
						Type:     schema.TypeBool,
						Computed: true,
					},
				},
			},
		},
		"metadata": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

func dataSourcePlanSettings() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_settings` manages an plan settings in Duplo.",

		ReadContext: dataSourcePlanSettingsRead,
		Schema:      duploComputedPlanSettingsSchema(),
	}
}

func dataSourcePlanSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanSettingsRead(%s): start", planID)

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
	d.SetId(planID)
	d.Set("metadata", keyValueToState("metadata", allMetadata))
	d.Set("unrestricted_ext_lb", settings.UnrestrictedExtLB)
	if dns != nil {
		d.Set("dns_setting", flattenDnsSetting(dns))
	}

	log.Printf("[TRACE] dataSourcePlanSettingsRead(%s): end", planID)
	return nil
}
