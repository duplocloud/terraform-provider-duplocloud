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

func ecacheInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"identifier": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"host": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"port": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"cache_type": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			Default:  0,
		},
		"size": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			Default:  1,
		},
		"encryption_at_rest": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		"encryption_in_transit": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
			Default:  false,
		},
		"instance_status": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcacheInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploEcacheInstanceRead,
		CreateContext: resourceDuploEcacheInstanceCreate,
		DeleteContext: resourceDuploEcacheInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: ecacheInstanceSchema(),
	}
}

/// READ resource
func resourceDuploEcacheInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcacheInstanceRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcacheInstanceGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.EcacheInstanceToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/%s", duplo.TenantID, duplo.Name))

	log.Printf("[TRACE] resourceDuploEcacheInstanceRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcacheInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcacheInstanceFromState(d)
	if err != nil {
		return diag.Errorf("Internal error: %s", err)
	}

	// Populate the identifier field, and determine some other fields
	duploObject.Identifier = duploObject.Name
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/%s", tenantID, duploObject.Name)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.EcacheInstanceCreate(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error creating ECache instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.EcacheInstanceGet(id)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting ECache instance '%s': %s", id, err))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected ECache instance '%s' to be retrieved, but got: nil", id))
		}

		return nil
	})

	// Wait for the instance to become available.
	err = duplosdk.EcacheInstanceWaitUntilAvailable(c, id)
	if err != nil {
		return diag.Errorf("Error waiting for ECache instance '%s' to be available: %s", id, err)
	}

	diags := resourceDuploEcacheInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploEcacheInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcacheInstanceDelete ******** start")

	// Delete the object from Duplo
	id := d.Id()
	c := m.(*duplosdk.Client)
	_, err := c.EcacheInstanceDelete(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to show the object as deleted.
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, errget := c.EcacheInstanceGet(id)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting ECache instance '%s': %s", id, err))
		}

		if resp != nil && resp.InstanceStatus != "deleted" {
			return resource.RetryableError(fmt.Errorf("Expected ECache instance '%s' to be deleted", id))
		}

		return nil
	})

	log.Printf("[TRACE] resourceDuploEcacheInstanceDelete ******** end")
	return nil
}
