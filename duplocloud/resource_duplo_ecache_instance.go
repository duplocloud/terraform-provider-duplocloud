package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	jo := ecacheInstanceToState(duplo, d)
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
	duploObject, err := ecacheInstanceFromState(d)
	if err != nil {
		return diag.Errorf("Internal error: %s")
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
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "ECache instance", id, func() (interface{}, error) {
		return c.EcacheInstanceGet(id)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = ecacheInstanceWaitUntilAvailable(c, id)
	if err != nil {
		return diag.Errorf("Error waiting for ECache instance '%s' to be available: %s", id, err)
	}

	diags = resourceDuploEcacheInstanceRead(ctx, d, m)
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
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "ECache instance", id, func() (interface{}, error) {
		return c.EcacheInstanceGet(id)
	})
	log.Printf("[TRACE] resourceDuploEcacheInstanceDelete ******** end")
	return diag
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// ecacheInstanceFromState converts resource data respresenting an ECache instance to a Duplo SDK object.
func ecacheInstanceFromState(d *schema.ResourceData) (*duplosdk.DuploEcacheInstance, error) {
	duploObject := new(duplosdk.DuploEcacheInstance)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("identifier").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Endpoint = d.Get("endpoint").(string)
	duploObject.CacheType = d.Get("cache_type").(int)
	duploObject.Size = d.Get("size").(string)
	duploObject.Replicas = d.Get("replicas").(int)
	duploObject.EncryptionAtRest = d.Get("encryption_at_rest").(bool)
	duploObject.EncryptionInTransit = d.Get("encryption_in_transit").(bool)
	duploObject.InstanceStatus = d.Get("instance_status").(string)

	return duploObject, nil
}

// ecacheInstanceToState converts a Duplo SDK object respresenting an ECache instance to terraform resource data.
func ecacheInstanceToState(duploObject *duplosdk.DuploEcacheInstance, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-EcacheInstanceToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["name"] = duploObject.Name
	jo["identifier"] = duploObject.Identifier
	jo["arn"] = duploObject.Arn
	jo["endpoint"] = duploObject.Endpoint
	if duploObject.Endpoint != "" {
		uriParts := strings.SplitN(duploObject.Endpoint, ":", 2)
		jo["host"] = uriParts[0]
		if len(uriParts) == 2 {
			jo["port"], _ = strconv.Atoi(uriParts[1])
		}
	}
	jo["cache_type"] = duploObject.CacheType
	jo["size"] = duploObject.Size
	jo["replicas"] = duploObject.Replicas
	jo["encryption_at_rest"] = duploObject.EncryptionAtRest
	jo["encryption_in_transit"] = duploObject.EncryptionInTransit
	jo["instance_status"] = duploObject.InstanceStatus

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-EcacheInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}

// ecacheInstanceWaitUntilAvailable waits until an ECache instance is available.
//
// It should be usable both post-creation and post-modification.
func ecacheInstanceWaitUntilAvailable(c *duplosdk.Client, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"processing", "creating", "modifying", "rebooting cluster nodes", "snapshotting"},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.EcacheInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] EcacheInstanceWaitUntilAvailable (%s)", id)
	_, err := stateConf.WaitForState()
	return err
}
