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
			Description: "The cloud account ID.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cloud": {
			Description: "The numerical index of cloud provider to use for the infrastructure.\n" +
				"Will be one of:\n\n" +
				"   - `0` : AWS\n" +
				"   - `2` : Azure\n",
			Type:     schema.TypeInt,
			Computed: true,
		},
		"region": {
			Description: "The cloud provider region.  The Duplo portal must have already been configured to support this region.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"azcount": {
			Description: "The number of availability zones.  Will be one of: `2`, `3`, or `4`.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"enable_k8_cluster": {
			Description: "Whether or not a kubernetes cluster is provisioned.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"address_prefix": {
			Description: "The CIDR for the VPC or VNet.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"subnet_cidr": {
			Description: "The CIDR subnet size (in bits) of the automatically created subnets.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"status": {
			Description: "The status of the infrastructure.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	if single {
		result["tenant_id"] = &schema.Schema{
			Description:  "The ID of the tenant to look up the infrastructure for. Must be specified if `infra_name` is blank.",
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"infra_name", "tenant_id"},
		}
		result["infra_name"] = &schema.Schema{
			Description:  "The name of the infrastructure to look up. Must be specified if `tenant_id` is blank.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ExactlyOneOf: []string{"infra_name", "tenant_id"},
		}
		result["vpc_id"] = &schema.Schema{
			Description: "The VPC or VNet ID.",
			Type:        schema.TypeString,
			Computed:    true,
		}
		result["vpc_name"] = &schema.Schema{
			Description: "The VPC or VNet name.",
			Type:        schema.TypeString,
			Computed:    true,
		}
		result["private_subnets"] = &schema.Schema{
			Description: "The private subnets for the VPC or VNet.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        infrastructureVnetSubnetSchema(),
		}
		result["public_subnets"] = &schema.Schema{
			Description: "The public subnets for the VPC or VNet.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        infrastructureVnetSubnetSchema(),
		}
		result["security_groups"] = &schema.Schema{
			Description: "The security groups for the VPC or VNet.",
			Type:        schema.TypeSet,
			Computed:    true,
			Elem:        infrastructureVnetSecurityGroupsSchema(),
		}
	} else {
		result["infra_name"] = &schema.Schema{
			Description: "The name of the infrastructure.",
			Type:        schema.TypeString,
			Computed:    true,
		}
	}

	return result
}

func dataSourceInfrastructures() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_infrastructures` retrieves a list of infrastructures in Duplo.",

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
		Description: "`duplocloud_infrastructure` retrieves details of an infrastructure in Duplo.",

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
	d.SetId("-")

	log.Printf("[TRACE] dataSourceInfrastructureRead(): end")
	return nil
}
