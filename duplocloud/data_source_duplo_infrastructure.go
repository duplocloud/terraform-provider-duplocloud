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
							Computed: true,
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
	log.Printf("[TRACE] dataSourceInfrastructureRead(): start")

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.InfrastructureGetList()
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	data := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {

		// TODO: ability to filter by tenant

		data = append(data, map[string]interface{}{
			"infra_name":        duplo.Name,
			"account_id":        duplo.AccountId,
			"cloud":             duplo.Cloud,
			"region":            duplo.Region,
			"azcount":           duplo.AzCount,
			"enable_k8_cluster": duplo.EnableK8Cluster,
			"address_prefix":    duplo.AddressPrefix,
			"subnet_cidr":       duplo.SubnetCidr,
			"status":            duplo.ProvisioningStatus,
		})
	}

	if err := d.Set("data", data); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] dataSourceInfrastructureRead(): start")
	return nil
}
