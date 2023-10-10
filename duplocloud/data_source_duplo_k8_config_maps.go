package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceK8ConfigMaps() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_config_maps` lists all kubernetes configmaps in a Duplo tenant.",

		ReadContext: dataSourceK8ConfigMapsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"config_maps": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: k8sConfigMapSchemaComputed(),
				},
			},
		},
	}
}

// READ/SEARCH resources
func dataSourceK8ConfigMapsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceK8ConfigMapsRead(%s): start", tenantID)

	// List from Duplo.
	c := m.(*duplosdk.Client)
	rp, err := c.K8ConfigMapGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the results into TF state.
	list := make([]map[string]interface{}, 0, len(*rp))
	for _, duplo := range *rp {

		// First, set the simple fields.
		cm := map[string]interface{}{"tenant_id": duplo.TenantID, "name": duplo.Name}

		// Next, set the JSON encoded strings.
		toJsonStringField("data", duplo.Data, cm)
		toJsonStringField("metadata", duplo.Metadata, cm)

		list = append(list, cm)
	}

	if err := d.Set("config_maps", list); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceK8ConfigMapsRead(%s): end", tenantID)
	return nil
}
