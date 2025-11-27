package duplocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ecsTaskDefinitionComputedSchema() map[string]*schema.Schema {
	itemSchema := ecsTaskDefinitionSchema()

	for k, el := range itemSchema {
		if k != "tenant_id" && k != "arn" {
			makeSchemaComputed(el)
		} else {
			el.Required = true
			el.Computed = false
			el.Optional = false
		}
	}

	return itemSchema
}

func dataSourceDuploEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploEcsTaskDefinitionRead,
		Schema:      ecsTaskDefinitionComputedSchema(),
	}
}

func dataSourceDuploEcsTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	arn := d.Get("arn").(string)
	log.Printf("[TRACE] dataSourceDuploEcsTaskDefinitionRead(%s, %s): start", tenantID, arn)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsTaskDefinitionGet(tenantID, arn)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s ECS task definition '%s': %s", tenantID, arn, err)
	}
	if duplo == nil {
		return diag.Errorf("Unable to read tenant %s ECS task definition '%s': not found", tenantID, arn)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, arn))
	tenant, cerr := c.TenantGetV2(tenantID)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("Tenant %s not found", tenantID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Duplocloud resource tenant information'\n%s", err)
	}
	prefix, err := c.GetResourcePrefixWithoutTenant("duploservices")
	if err != nil {
		return diag.FromErr(err)
	}

	// Read the object into state
	flattenEcsTaskDefinition(duplo, d, tenant.AccountName, prefix)

	log.Printf("[TRACE] dataSourceDuploEcsTaskDefinitionRead(%s, %s): end", tenantID, arn)
	return nil
}
