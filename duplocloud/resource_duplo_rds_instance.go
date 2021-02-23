package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
		return diag.Errorf("Internal error: %s")
	}

	// Populate the identifier field, and determine some other fields
	duploObject.Identifier = duploObject.Name
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, duploObject.Name)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.RdsInstanceCreate(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error creating RDS DB instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.RdsInstanceGet(id)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting RDS DB instance '%s': %s", id, err))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected RDS DB instance '%s' to be retrieved, but got: nil", id))
		}

		return nil
	})

	// Wait for the instance to become available.
	err = duplosdk.RdsInstanceWaitUntilAvailable(c, id)
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

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
	id := d.Id()
	_, err = c.RdsInstanceUpdate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for the instance to become unavailable.
	err = duplosdk.RdsInstanceWaitUntilUnavailable(c, id)
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be unavailable: %s", id, err)
	}

	// Wait for the instance to become available.
	err = duplosdk.RdsInstanceWaitUntilAvailable(c, id)
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

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
