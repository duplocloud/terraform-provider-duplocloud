package duplocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"
)

// SCHEMA for resource crud
func resourceDuploEcsService() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploEcsServiceRead,
		CreateContext: resourceDuploEcsServiceCreate,
		UpdateContext: resourceDuploEcsServiceUpdate,
		DeleteContext: resourceDuploEcsServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.DuploEcsServiceSchema(),
	}
}

/// READ resource
func resourceDuploEcsServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsServiceRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsServiceGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.EcsServiceToState(duplo, d)
	for key, _ := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", duplo.TenantId, duplo.Name))

	log.Printf("[TRACE] resourceDuploEcsServiceRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcsServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsServiceCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcsServiceFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	tenantId := d.Get("tenant_id").(string)
	rpObject, err := c.EcsServiceCreate(tenantId, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantId, rpObject.Name))

	diags := resourceDuploEcsServiceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsServiceCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploEcsServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsServiceUpdate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcsServiceFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Put the object to Duplo
	c := m.(*duplosdk.Client)
	tenantId := d.Get("tenant_id").(string)
	_, err = c.EcsServiceUpdate(tenantId, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceDuploEcsServiceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsServiceUpdate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploEcsServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsServiceDelete ******** start")

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	_, err := c.EcsServiceDelete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploEcsServiceDelete ******** end")
	return nil
}
