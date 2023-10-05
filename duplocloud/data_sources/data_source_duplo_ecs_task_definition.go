package data_sources

import (
	"context"
	"fmt"
	"terraform-provider-duplocloud/duplocloud"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ecsTaskDefinitionComputedSchema() map[string]*schema.Schema {
	itemSchema := duplocloud.ecsTaskDefinitionSchema()

	for k, el := range itemSchema {
		if k != "tenant_id" && k != "arn" {
			duplocloud.makeSchemaComputed(el)
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

	// Read the object into state
	duplocloud.flattenEcsTaskDefinition(duplo, d)

	log.Printf("[TRACE] dataSourceDuploEcsTaskDefinitionRead(%s, %s): end", tenantID, arn)
	return nil
}
