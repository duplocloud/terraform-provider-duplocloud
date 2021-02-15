package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func resourceDuploEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploEcsTaskDefinitionRead,
		CreateContext: resourceDuploEcsTaskDefinitionCreate,
		//UpdateContext: resourceDuploEcsTaskDefinitionUpdate,
		DeleteContext: resourceDuploEcsTaskDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.DuploEcsTaskDefinitionSchema(),
	}
}

/// READ resource
func resourceDuploEcsTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsTaskDefinitionGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.EcsTaskDefToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", duplo.TenantId, duplo.Arn))

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcsTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcsTaskDefFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	arn, err := c.EcsTaskDefinitionCreate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", tenantID, arn))

	diags := resourceDuploEcsTaskDefinitionRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploEcsTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** start")
	// FIXME: NO-OP
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** end")
	return nil
}
