package duplocloud

import (
	"context"
	"fmt"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func k8sConfigMapSchemaComputed() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The name of the configmap.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"tenant_id": {
			Description: "The GUID of the tenant that the configmap will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"data": {
			Description: "A JSON encoded string representing the configmap data. " +
				"You can use the `jsondecode()` function to parse this, if needed.",
			Type:     schema.TypeString,
			Computed: true,
		},
		"metadata": {
			Description: "A JSON encoded string representing the configmap metadata. " +
				"You can use the `jsondecode()` function to parse this, if needed.",
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// SCHEMA for resource data/search
func dataSourceK8ConfigMap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8ConfigMapRead,
		Schema:      k8sConfigMapSchemaComputed(),
	}
}

/// READ/SEARCH resources
func dataSourceK8ConfigMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	id := fmt.Sprintf("%s/%s", tenantID, name)

	log.Printf("[TRACE] dataSourceK8ConfigMapRead(%s): start", id)

	// Get the result from Duplo, detecting a missing object.
	c := m.(*duplosdk.Client)
	rp, err := c.K8ConfigMapGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.Name == "" {
		return diag.Errorf("tenant configmap '%s' not found", id)
	}

	// Convert it into TF state.
	d.SetId(id)
	flattenK8sConfigMap(d, rp)

	log.Printf("[TRACE] dataSourceK8ConfigMapsRead(%s): end", id)
	return nil
}
