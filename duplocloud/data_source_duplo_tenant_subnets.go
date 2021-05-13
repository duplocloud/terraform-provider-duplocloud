package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func schemaTenantSubnets() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"subnet_ids": {
			Description: "The list of subnet IDs.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func dataSourceTenantInternalSubnets() *schema.Resource {
	return &schema.Resource{
		Description: "The `duplocloud_tenant_internal_subnets` data source retrieves a list of tenant's internal subnet IDs.",

		ReadContext: dataSourceTenantInternalSubnetsRead,
		Schema:      schemaTenantSubnets(),
	}
}

func dataSourceTenantExternalSubnets() *schema.Resource {
	return &schema.Resource{
		Description: "The `duplocloud_tenant_external_subnets` data source retrieves a list of tenant's external subnet IDs.",

		ReadContext: dataSourceTenantExternalSubnetsRead,
		Schema:      schemaTenantSubnets(),
	}
}

func dataSourceTenantInternalSubnetsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceTenantInternalSubnetsRead(%s): start", tenantID)

	// List the secrets from Duplo.
	c := m.(*duplosdk.Client)
	subnetIDs, err := c.TenantGetInternalSubnets(tenantID)
	if err != nil {
		return diag.Errorf("failed to list subnets: %s", err)
	}
	d.SetId(tenantID)
	d.Set("subnet_ids", subnetIDs)

	log.Printf("[TRACE] dataSourceTenantInternalSubnetsRead(%s): end", tenantID)
	return nil
}

func dataSourceTenantExternalSubnetsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceTenantExternalSubnetsRead(%s): start", tenantID)

	// List the secrets from Duplo.
	c := m.(*duplosdk.Client)
	subnetIDs, err := c.TenantGetExternalSubnets(tenantID)
	if err != nil {
		return diag.Errorf("failed to list subnets: %s", err)
	}
	d.SetId(tenantID)
	d.Set("subnet_ids", subnetIDs)

	log.Printf("[TRACE] dataSourceTenantExternalSubnetsRead(%s): end", tenantID)
	return nil
}
