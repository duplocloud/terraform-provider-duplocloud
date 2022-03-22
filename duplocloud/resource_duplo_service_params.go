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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// duploServiceParamsSchema returns a Terraform resource schema for a service's parameters
func duploServiceParamsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that hosts the duplo service.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"replication_controller_name": {
			Description: "The name of the duplo service.",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
			ForceNew:    true, //switch service
		},
		"load_balancer_name": {
			Description: "The load balancer name.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"load_balancer_arn": {
			Description: "The load balancer ARN.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"webaclid": {
			Description: "The ARN of a web application firewall to associate this load balancer.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"dns_prfx": {
			Description: "The DNS prefix to assign to this service's load balancer.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"enable_access_logs": {
			Description: "Whether or not to enable access logs.  When enabled, Duplo will send access logs to a centralized S3 bucket per plan",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"drop_invalid_headers": {
			Description: "Whether or not to drop invalid HTTP headers received by the load balancer.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploServiceParams() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_duplo_service_lbconfigs` manages additional configuration for a container-based service in Duplo.\n\n" +
			"NOTE: For Amazon ECS services, see the `duplocloud_ecs_service` resource.",

		ReadContext:   resourceDuploServiceParamsRead,
		CreateContext: resourceDuploServiceParamsCreate,
		UpdateContext: resourceDuploServiceParamsUpdate,
		DeleteContext: resourceDuploServiceParamsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	var err error

	// Parse the identifying attributes
	tenantID, name := parseDuploServiceParamsIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}
	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, derr := getReplicationControllerIfHasAlb(c, tenantID, name)
	if derr != nil {
		return derr
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Get the WAF information.
	webAclId, err := c.ReplicationControllerLbWafGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	d.Set("replication_controller_name", duplo.Name)
	d.Set("tenant_id", tenantID)
	d.Set("dns_prfx", duplo.DnsPrfx)
	d.Set("webaclid", webAclId)

	// Next, look for load balancer settings.
	err = readDuploServiceAwsLbSettings(tenantID, duplo, d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceDuploServiceParamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, false)
}

/// UPDATE resource
func resourceDuploServiceParamsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, true)
}

func resourceDuploServiceParamsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, isUpdate bool) diag.Diagnostics {
	var err error

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
	_, err = c.DuploServiceParamsCreateOrUpdate(tenantID, &duplo, isUpdate)
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

func readDuploServiceAwsLbSettings(tenantID string, rpc *duplosdk.DuploReplicationController, d *schema.ResourceData, c *duplosdk.Client) error {

	// If we are not AWS, just return for now.
	if rpc.Template == nil || rpc.Template.Cloud != 0 {
		return nil
	}

	// Next, look for load balancer settings.
	details, err := c.TenantGetLbDetailsInService(tenantID, rpc.Name)
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

func updateDuploServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) duplosdk.ClientError {

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

func parseDuploServiceParamsIdParts(id string) (tenantID, name string) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	}
	return
}

func getReplicationControllerIfHasAlb(client *duplosdk.Client, tenantID, name string) (*duplosdk.DuploReplicationController, diag.Diagnostics) {

	// Get the object from Duplo.
	duplo, err := client.ReplicationControllerGet(tenantID, name)
	if err != nil {
		return nil, diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, name, err)
	}

	// Check for an application load balancer
	if duplo != nil && duplo.Template != nil {
		for _, lb := range duplo.Template.LBConfigurations {
			if lb.LbType == 2 {
				return duplo, nil
			}
		}
	}

	// Not found.
	return nil, nil
}
