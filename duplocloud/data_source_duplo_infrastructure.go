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

func infrastructureSchemaComputed(single bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
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
	}

	if single {
		result["tenant_id"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"infra_name", "tenant_id"},
		}
		result["infra_name"] = &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ExactlyOneOf: []string{"infra_name", "tenant_id"},
		}
		result["vpc_id"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
		result["vpc_name"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
		result["private_subnets"] = &schema.Schema{
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     infrastructureVnetSubnetSchema(),
		}
		result["public_subnets"] = &schema.Schema{
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     infrastructureVnetSubnetSchema(),
		}
	} else {
		result["infra_name"] = &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		}
	}

	return result
}

func dataSourceInfrastructures() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInfrastructuresRead,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: infrastructureSchemaComputed(false),
				},
			},
		},
	}
}

func dataSourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInfrastructureRead,
		Schema:      infrastructureSchemaComputed(true),
	}
}

func dataSourceInfrastructuresRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceInfrastructuresRead(): start")

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.InfrastructureGetList()
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	data := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
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

	log.Printf("[TRACE] dataSourceInfrastructuresRead(): end")
	return nil
}

/// READ/SEARCH resources
func dataSourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceInfrastructureRead(): start")

	c := m.(*duplosdk.Client)
	var name string

	// Look up by tenant ID, if requested.
	if v, ok := d.GetOk("tenant_id"); ok && v != nil && v.(string) != "" {
		tenant, err := c.TenantGet(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		name = tenant.PlanID // plan ID always matches infra name
	} else {
		name = d.Get("infra_name").(string)
	}

	// Get the object from Duplo, detecting a missing object
	missing, err := infrastructureRead(c, d, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure '%s': %s", name, err)
	}
	if missing {
		return diag.Errorf("No infrastructure named '%s' was found", name)
	}

	log.Printf("[TRACE] dataSourceInfrastructureRead(): end")
	return nil
}
