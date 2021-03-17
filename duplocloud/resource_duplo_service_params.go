package duplocloud

import (
	"context"
	"fmt"
	"strings"

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
		"load_balancer_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"load_balancer_arn": {
			Type:     schema.TypeString,
			Computed: true,
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
		"enable_access_logs": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"drop_invalid_headers": {
			Type:     schema.TypeBool,
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
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) < 5 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID := idParts[2]
	name := idParts[4]
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

	// Next, look for load balancer settings.
	err = readDuploServiceAwsLbSettings(tenantID, name, d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s): end", id)
	return nil
}

/// CREATE resource
func resourceDuploServiceParamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, true)
}

/// UPDATE resource
func resourceDuploServiceParamsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, true)
}

func resourceDuploServiceParamsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, isUpdate bool) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	duplo := duplosdk.DuploServiceParams{
		ReplicationControllerName: d.Get("replication_controller_name").(string),
		WebACLId:                  d.Get("webaclid").(string),
		DNSPrfx:                   d.Get("dns_prfx").(string),
	}

	log.Printf("[TRACE] resourceDuploServiceParamsCreateOrUpdate(%s, %s): start", tenantID, duplo.ReplicationControllerName)

	// Get or generate the ID.
	var id string
	if isUpdate {
		id = d.Id()
	} else {
		id = fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenantID, duplo.ReplicationControllerName)
	}

	// Create the service paramaters.
	c := m.(*duplosdk.Client)
	_, err := c.DuploServiceParamsCreateOrUpdate(tenantID, &duplo, isUpdate)
	if err != nil {
		return diag.Errorf("Error applying Duplo service params '%s': %s", id, err)
	}
	if !isUpdate {
		d.SetId(id)
	}

	// Next, we need to apply load balancer settings.
	err = updateDuploServiceAwsLbSettings(tenantID, duplo.ReplicationControllerName, d, c)
	if err != nil {
		return diag.Errorf("Error applying Duplo service params '%s': %s", id, err)
	}

	diags := resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceParamsCreateOrUpdate(%s, %s): end", tenantID, duplo.ReplicationControllerName)
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

func readDuploServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) error {

	// Next, look for load balancer settings.
	details, err := c.TenantGetLbDetailsInService(tenantID, name)
	if err != nil {
		return err
	}
	if details != nil && details.LoadBalancerArn != "" {

		// Populate load balancer details.
		d.Set("load_balancer_arn", details.LoadBalancerArn)
		d.Set("load_balancer_name", details.LoadBalancerName)

		settings, err := c.TenantGetApplicationLbSettings(tenantID, details.LoadBalancerArn)
		if err != nil {
			return err
		}
		if settings != nil && settings.LoadBalancerArn != "" {

			// Populate load balancer settings.
			d.Set("webaclid", settings.WebACLID)
			d.Set("enable_access_logs", settings.EnableAccessLogs)
			d.Set("drop_invalid_headers", settings.DropInvalidHeaders)
		}
	}

	return nil
}

func updateDuploServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) error {

	// Get any load balancer settings from the user.
	settings := duplosdk.DuploAwsLbSettingsUpdateRequest{}
	haveSettings := false
	if v, ok := d.GetOk("enable_access_logs"); ok && v != nil {
		settings.EnableAccessLogs = v.(bool)
		haveSettings = true
	}
	if v, ok := d.GetOk("drop_invalid_headers"); ok && v != nil {
		settings.DropInvalidHeaders = v.(bool)
		haveSettings = true
	}
	if v, ok := d.GetOk("webaclid"); ok && v != nil {
		settings.WebACLID = v.(string)
		haveSettings = true
	}

	// If we have load balancer settings, apply them.
	if haveSettings {
		details, err := c.TenantGetLbDetailsInService(tenantID, name)
		if err != nil {
			return err
		}

		if details != nil && details.LoadBalancerArn != "" {
			settings.LoadBalancerArn = details.LoadBalancerArn
			err = c.TenantUpdateApplicationLbSettings(tenantID, settings)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
