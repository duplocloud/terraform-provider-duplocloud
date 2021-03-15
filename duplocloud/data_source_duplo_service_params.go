package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// duploServiceParamsSchema returns a Terraform resource schema for a service's parameters
func dataSourceDuploServiceParamSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"replication_controller_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"dns_prfx": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"webaclid": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// SCHEMA for resource data/search
func dataSourceDuploServiceParams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceParamsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_controller_name": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"dns_prfx": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"webaclid": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"result": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_controller_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_prfx": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"webaclid": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceDuploServiceParamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceDuploServiceParamsRead(%s): start", tenantID)

	// Get the result from duplo
	c := m.(*duplosdk.Client)
	allParams, err := c.DuploServiceParamsGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get all of the filters asked for by the user
	var name, dnsPrfx, webACLID string
	if v, ok := d.GetOk("replication_controller_name"); ok && v != nil {
		name = v.(string)
	}
	if v, ok := d.GetOk("dns_prfx"); ok && v != nil {
		dnsPrfx = v.(string)
	}
	if v, ok := d.GetOk("webaclid"); ok && v != nil {
		webACLID = v.(string)
	}

	// Build a list of selected params
	selectedParams := make([]map[string]interface{}, 0, len(*allParams))
	for _, param := range *allParams {

		if (name == "" || param.ReplicationControllerName == name) &&
			(dnsPrfx == "" || param.DNSPrfx == dnsPrfx) &&
			(webACLID == "" || param.WebACLId == webACLID) {

			selectedParams = append(selectedParams, map[string]interface{}{
				"tenant_id":                   param.TenantID,
				"replication_controller_name": param.ReplicationControllerName,
				"dns_prfx":                    param.DNSPrfx,
				"webaclid":                    param.WebACLId,
			})
		}
	}
	d.SetId(fmt.Sprintf("%s-%s", tenantID, strconv.FormatInt(time.Now().Unix(), 10)))

	// Apply the result
	dumpParams, _ := json.Marshal(selectedParams)
	log.Printf("[TRACE] dataSourceDuploServiceParamsRead(%s): dump: %s", tenantID, dumpParams)
	d.Set("result", selectedParams)

	log.Printf("[TRACE] dataSourceDuploServiceParamsRead(%s): end", tenantID)

	return nil
}
