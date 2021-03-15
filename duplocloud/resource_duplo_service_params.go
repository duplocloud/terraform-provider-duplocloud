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

// duploServiceParamsSchema returns a Terraform resource schema for a service's parameters
func duploServiceParamsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"replication_controller_name": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch service
		},
		"webaclid": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"dns_prfx": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploServiceParams() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploServiceParamsRead,
		CreateContext: resourceDuploServiceParamsCreate,
		UpdateContext: resourceDuploServiceParamsUpdate,
		DeleteContext: resourceDuploServiceParamsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploServiceParamsSchema(),
	}
}

/// READ resource
func resourceDuploServiceParamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s): start", id)

	// Get the object from Duplo, handling a missing object
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("replication_controller_name").(string)
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploServiceParamsGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}

	// Convert the object into Terraform resource data
	d.Set("replication_controller_name", duplo.ReplicationControllerName)
	d.Set("webaclid", duplo.WebACLId)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("dns_prfx", duplo.DNSPrfx)

	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s): end", id)
	return nil
}

/// CREATE resource
func resourceDuploServiceParamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	duplo := duploServiceParamsFromState(d)
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] resourceDuploServiceParamsCreate(%s, %s): start", tenantID, duplo.ReplicationControllerName)

	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenantID, duplo.ReplicationControllerName)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err := c.DuploServiceParamsCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error creating Duplo service params instance '%s': %s", id, err)
	}
	d.SetId(id)

	diags := resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceParamsCreate(%s, %s): end", tenantID, duplo.ReplicationControllerName)
	return diags
}

/// UPDATE resource
func resourceDuploServiceParamsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	duplo := duploServiceParamsFromState(d)
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] resourceDuploServiceParamsUpdate(%s, %s): start", tenantID, duplo.ReplicationControllerName)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err := c.DuploServiceParamsCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error updating Duplo service params instance '%s': %s", d.Id(), err)
	}

	diags := resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceParamsUpdate(%s, %s): end", tenantID, duplo.ReplicationControllerName)
	return diags
}

/// DELETE resource
func resourceDuploServiceParamsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceDuploServiceParamsDelete(%s): start", id)

	// Delete the object from Duplo
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("replication_controller_name").(string)
	c := m.(*duplosdk.Client)
	err := c.DuploServiceParamsDelete(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploServiceParamsDelete(%s): end", id)

	return nil
}

// duploServiceParamsFromState converts resource data respresenting a service's parameters to a Duplo SDK object.
func duploServiceParamsFromState(d *schema.ResourceData) *duplosdk.DuploServiceParams {
	duploObject := new(duplosdk.DuploServiceParams)

	duploObject.ReplicationControllerName = d.Get("replication_controller_name").(string)
	duploObject.WebACLId = d.Get("webaclid").(string)
	duploObject.DNSPrfx = d.Get("dns_prfx").(string)

	return duploObject
}
