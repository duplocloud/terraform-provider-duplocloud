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
func resourceDuploRdsInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploRdsInstanceRead,
		CreateContext: resourceDuploRdsInstanceCreate,
		UpdateContext: resourceDuploRdsInstanceUpdate,
		DeleteContext: resourceDuploRdsInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.DuploRdsInstanceSchema(),
	}
}

/// READ resource
func resourceDuploRdsInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.RdsInstanceGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.RdsInstanceToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", duplo.TenantID, duplo.Name))

	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploRdsInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.RdsInstanceFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the identifier field
	duploObject.Identifier = duploObject.Name

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	rpObject, err := c.RdsInstanceCreate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, rpObject.Name))

	// Try to get the object for up to 60 seconds.
	// -- TODO --

	// Wait for the instance to become available.
	// -- TODO --

	diags := resourceDuploRdsInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploRdsInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceUpdate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.RdsInstanceFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Put the object to Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	_, err = c.RdsInstanceUpdate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for the instance to become "unavailable" for up to 60 seconds.
	// -- TODO --

	// Wait for the instance to become available.
	// -- TODO --

	diags := resourceDuploRdsInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploRdsInstanceUpdate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploRdsInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** start")

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	_, err := c.RdsInstanceDelete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** end")
	return nil
}
