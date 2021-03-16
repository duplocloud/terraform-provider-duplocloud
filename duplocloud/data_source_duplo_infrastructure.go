package duplocloud

import (
	"context"
	"strconv"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInfrastructureRead,
		Schema: map[string]*schema.Schema{
			"filter": FilterSchema(), // todo: search specific to this object... may be api should support filter?
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"infra_name": {
							Type:     schema.TypeString,
							Computed: false,
						},
						"account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloud": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"azcount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enable_k8_cluster": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"address_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_cidr": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"status": {
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
func dataSourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceInfrastructureRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.InfrastructureGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.InfrastructuresFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceInfrastructureRead ******** end")

	return diags
}
