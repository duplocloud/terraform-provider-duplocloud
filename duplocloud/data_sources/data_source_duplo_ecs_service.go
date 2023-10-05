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

func ecsServiceComputedSchema() map[string]*schema.Schema {
	itemSchema := duplocloud.ecsServiceSchema()

	for k, el := range itemSchema {
		if k != "tenant_id" && k != "name" {
			duplocloud.makeSchemaComputed(el)
		} else {
			el.Required = true
			el.Computed = false
			el.Optional = false
		}
	}

	return itemSchema
}

func dataSourceDuploEcsService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploEcsServiceRead,
		Schema:      ecsServiceComputedSchema(),
	}
}

func dataSourceDuploEcsServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourceDuploServiceRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsServiceGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s ECS service '%s': %s", tenantID, name, err)
	}
	if duplo == nil {
		return diag.Errorf("Unable to read tenant %s ECS service '%s': not found", tenantID, name)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	// Read the object into state
	duplocloud.flattenDuploEcsService(d, duplo, c)

	log.Printf("[TRACE] dataSourceDuploServiceRead: end")
	return nil
}
