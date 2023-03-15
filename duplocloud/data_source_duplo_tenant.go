package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for tenants
func tenantSchemaComputed(nested bool) map[string]*schema.Schema {
	exactlyOneOf := []string{}
	if !nested {
		exactlyOneOf = []string{"id", "name"}
	}
	nameExactlyOneOf := exactlyOneOf // create a copy, to be safe

	return map[string]*schema.Schema{
		"id": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ExactlyOneOf: exactlyOneOf,
		},
		"name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ExactlyOneOf: nameExactlyOneOf,
		},
		"plan_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"infra_owner": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"policy": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allow_volume_mapping": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"block_external_ep": {
						Type:     schema.TypeBool,
						Computed: true,
					},
				},
			},
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// Data source listing tenants
func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantRead,

		Schema: tenantSchemaComputed(false),
	}
}

// READ resource
func dataSourceTenantRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantRead(): start")

	var tenantID, tenantName string
	if v, ok := d.GetOk("tenant_id"); ok && v != nil {
		tenantID = v.(string)
	} else if v, ok := d.GetOk("name"); ok && v != nil {
		tenantName = v.(string)
	}

	// Get the tenant from Duplo.
	var err error
	var duploTenant *duplosdk.DuploTenant
	c := m.(*duplosdk.Client)
	if tenantID != "" {
		duploTenant, err = c.GetTenantForUser(tenantID)
	} else if tenantName != "" {
		duploTenant, err = c.GetTenantByNameForUser(tenantName)
	}
	if err != nil {
		return fmt.Errorf("failed to get tenant: %s", err)
	}
	if duploTenant == nil {
		return nil // not found
	}

	// Set the Terraform resource data
	d.SetId(duploTenant.TenantID)
	d.Set("id", duploTenant.TenantID)
	d.Set("name", duploTenant.AccountName)
	d.Set("plan_id", duploTenant.PlanID)
	d.Set("infra_owner", duploTenant.InfraOwner)
	if duploTenant.TenantPolicy != nil {
		d.Set("policy", []map[string]interface{}{{
			"allow_volume_mapping": true,
			"block_external_ep":    true,
		}})
	}
	d.Set("tags", keyValueToState("tags", duploTenant.Tags))

	log.Printf("[TRACE] dataSourceTenantRead(): end")
	return nil
}
