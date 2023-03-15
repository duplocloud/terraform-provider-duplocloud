package duplocloud

import (
	"context"
	"encoding/base64"
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
		"http_to_https_redirect": {
			Description: "Whether or not to enable http to https redirection.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"idle_timeout": {
			Description: "The time in seconds that the connection is allowed to be idle. Only valid for Load Balancers of type `application`.",
			Type:        schema.TypeInt,
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

// READ resource
func resourceDuploServiceParamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	var clientError duplosdk.ClientError
	// Parse the identifying attributes
	tenantID, name := parseDuploServiceParamsIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}
	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)

	// Get the object from Duplo.
	duplo, err := c.ReplicationControllerGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, name, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Get the WAF information.
	webAclId := ""
	if doesReplicationControllerHaveAlb(duplo) {
		webAclId, clientError = c.ReplicationControllerLbWafGet(tenantID, name)
		if clientError != nil {
			if clientError.Status() == 500 && duplo.Template.Cloud != 0 {
				log.Printf("[TRACE] Ignoring error %s for non AWS cloud.", clientError)
			} else {
				return diag.FromErr(err)
			}
		}
	}

	// Convert the object into Terraform resource data
	d.Set("replication_controller_name", duplo.Name)
	d.Set("tenant_id", tenantID)
	d.Set("dns_prfx", duplo.DnsPrfx)
	d.Set("webaclid", webAclId)

	// Next, look for load balancer settings.
	if doesReplicationControllerHaveAlbOrNlb(duplo) {
		err = readDuploServiceAwsLbSettings(tenantID, duplo, d, c)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] resourceDuploServiceParamsRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceDuploServiceParamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, false)
}

// UPDATE resource
func resourceDuploServiceParamsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploServiceParamsCreateOrUpdate(ctx, d, m, true)
}

func resourceDuploServiceParamsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, isUpdate bool) diag.Diagnostics {
	var err error
	var clientError duplosdk.ClientError
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("replication_controller_name").(string)
	log.Printf("[TRACE] resourceDuploServiceParamsCreateOrUpdate(%s, %s): start", tenantID, name)

	// Get the service from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.ReplicationControllerGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, name, err)
	}

	// Update the DNS.
	dns := d.Get("dns_prfx").(string)
	if dns != "" && dns != duplo.DnsPrfx {
		dnsRq := duplosdk.DuploLbDnsRequest{DnsPrfx: dns}
		err = c.ReplicationControllerLbDnsUpdate(tenantID, name, &dnsRq)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if dns == "" && duplo.DnsPrfx != "" {
		err = c.ReplicationControllerLbDnsDelete(tenantID, name)
		if err != nil {
			log.Printf("[TRACE] resourceDuploServiceParamsCreateOrUpdate(%s, %s): failed to delete DNS Prefix - continuing", tenantID, name)
		}
	}

	// Update the WAF.
	if doesReplicationControllerHaveAlb(duplo) {
		wafRq := duplosdk.DuploLbWafUpdateRequest{
			WebAclId:     d.Get("webaclid").(string),
			IsEcsLB:      false,
			IsPassThruLB: false,
		}
		if wafRq.WebAclId != "" {
			clientError = c.ReplicationControllerLbWafUpdate(tenantID, name, &wafRq)
		} else {
			wafRq.WebAclId, clientError = c.ReplicationControllerLbWafGet(tenantID, name)
			if clientError == nil && wafRq.WebAclId != "" {
				clientError = c.ReplicationControllerLbWafDelete(tenantID, name, &wafRq)
			}
		}
		if clientError != nil {
			if clientError.Status() == 500 && duplo.Template.Cloud != 0 {
				log.Printf("[TRACE] Ignoring error %s for non AWS cloud.", clientError)
			} else {
				return diag.FromErr(err)
			}
		}
	}

	// Update the ALB or NLB settings
	if doesReplicationControllerHaveAlbOrNlb(duplo) {

		// Check if the LB is active.
		details, err := getDuploServiceAwsLbSettings(tenantID, duplo, c)
		if err != nil {
			return diag.FromErr(err)
		}
		if duplo.Template.Cloud == 0 && !isDuploServiceAwsLbActive(details) {
			return diag.Errorf("Load balancer for tenant %s service '%s' is not active", tenantID, name)
		}

		// Next, we need to apply load balancer settings.
		clientError = updateDuploServiceAwsLbSettings(tenantID, details, d, c)
		if clientError != nil {
			if clientError.Status() == 500 && duplo.Template.Cloud != 0 {
				log.Printf("[TRACE] Ignoring error %s for non AWS cloud.", clientError)
			} else {
				return diag.Errorf("Error applying LB settings for tenant %s service '%s': %s", tenantID, name, err)
			}
		}
	}

	// Get or generate the ID.
	if !isUpdate {
		d.SetId(fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenantID, name))
	}

	// Finally, we read the information back.
	diags := resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceParamsCreateOrUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceDuploServiceParamsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var clientError duplosdk.ClientError
	// Parse the identifying attributes
	tenantID, name := parseDuploServiceParamsIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}
	log.Printf("[TRACE] resourceDuploServiceParamsDelete(%s, %s): start", tenantID, name)

	// Get the service from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.ReplicationControllerGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, name, err)
	}

	// We have a service.
	if duplo != nil {

		// Delete the DNS settings.
		if duplo.DnsPrfx != "" {
			err := c.ReplicationControllerLbDnsDelete(tenantID, name)
			if err != nil {
				log.Printf("[TRACE] resourceDuploServiceParamsDelete(%s, %s): failed to delete DNS Prefix - continuing", tenantID, name)
			}
		}

		// Detach the WAF.
		if doesReplicationControllerHaveAlb(duplo) {
			details, err := getDuploServiceAwsLbSettings(tenantID, duplo, c)
			if err != nil {
				return diag.FromErr(err)
			}

			// The ALB is active.
			if isDuploServiceAwsLbActive(details) {
				c := m.(*duplosdk.Client)

				wafRq := duplosdk.DuploLbWafUpdateRequest{
					IsEcsLB:      false,
					IsPassThruLB: false,
				}
				wafRq.WebAclId, clientError = c.ReplicationControllerLbWafGet(tenantID, name)
				if clientError == nil && wafRq.WebAclId != "" {
					clientError = c.ReplicationControllerLbWafDelete(tenantID, name, &wafRq)
				}
				if clientError != nil {
					if clientError.Status() == 500 && duplo.Template.Cloud != 0 {
						log.Printf("[TRACE] Ignoring error %s for non AWS cloud.", clientError)
					} else {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	log.Printf("[TRACE] resourceDuploServiceParamsDelete(%s, %s): start", tenantID, name)

	return nil
}

func getDuploServiceAwsLbSettings(tenantID string, rpc *duplosdk.DuploReplicationController, c *duplosdk.Client) (*duplosdk.DuploAwsLbDetailsInService, error) {

	if rpc.Template != nil && rpc.Template.Cloud == 0 {

		// Look for load balancer settings.
		details, err := c.TenantGetLbDetailsInService(tenantID, rpc.Name)
		if err != nil {
			return nil, err
		}
		if details != nil && details.LoadBalancerArn != "" {
			return details, nil
		}
	}

	// Nothing found.
	return nil, nil
}

func isDuploServiceAwsLbActive(details *duplosdk.DuploAwsLbDetailsInService) bool {
	return details != nil && details.State != nil && details.State.Code != nil && strings.ToLower(details.State.Code.Value) == "active"
}

func readDuploServiceAwsLbSettings(tenantID string, rpc *duplosdk.DuploReplicationController, d *schema.ResourceData, c *duplosdk.Client) error {
	details, err := getDuploServiceAwsLbSettings(tenantID, rpc, c)
	if details == nil || err != nil {
		return err
	}

	// Populate load balancer details.
	d.Set("load_balancer_arn", details.LoadBalancerArn)
	d.Set("load_balancer_name", details.LoadBalancerName)
	loadBalancerID := base64.URLEncoding.EncodeToString([]byte(details.LoadBalancerArn))

	settings, err := c.TenantGetLbSettings(tenantID, loadBalancerID)
	if err != nil {
		return err
	}
	if settings != nil {

		// Populate load balancer settings.
		if settings.SecurityPolicyId == nil {
			d.Set("web_acl_id", "")
		} else {
			d.Set("web_acl_id", *settings.SecurityPolicyId)
		}
		if settings.EnableAccessLogs == nil {
			d.Set("enable_access_logs", false)
		} else {
			d.Set("enable_access_logs", *settings.EnableAccessLogs)
		}
		if settings.Aws == nil {
			d.Set("drop_invalid_headers", false)
		} else {
			d.Set("drop_invalid_headers", settings.Aws.DropInvalidHeaders)
		}
		d.Set("idle_timeout", settings.Timeout)
		if settings.EnableHttpToHttpsRedirect == nil {
			d.Set("http_to_https_redirect", false)
		} else {
			d.Set("http_to_https_redirect", *settings.EnableHttpToHttpsRedirect)
		}
	}

	return nil
}

func updateDuploServiceAwsLbSettings(tenantID string, details *duplosdk.DuploAwsLbDetailsInService, d *schema.ResourceData, c *duplosdk.Client) duplosdk.ClientError {

	// Get any load balancer settings from the user.
	settings := &duplosdk.AgnosticLbSettings{
		Aws: &duplosdk.AgnosticLbSettingsAws{},
	}
	haveSettings := false
	if v, ok := d.GetOk("enable_access_logs"); ok && v != nil {
		enableAccessLogs := v.(bool)
		settings.EnableAccessLogs = &enableAccessLogs
		haveSettings = true
	}
	if v, ok := d.GetOk("drop_invalid_headers"); ok && v != nil {
		settings.Aws.DropInvalidHeaders = v.(bool)
		haveSettings = true
	}
	if v, ok := d.GetOk("http_to_https_redirect"); ok && v != nil {
		enableHttpToHttpsRedirect := v.(bool)
		settings.EnableHttpToHttpsRedirect = &enableHttpToHttpsRedirect
		haveSettings = true
	}
	if v, ok := d.GetOk("idle_timeout"); ok && v != nil {
		settings.Timeout = v.(int)
		haveSettings = true
	}
	if v, ok := d.GetOk("webaclid"); ok && v != nil {
		securityPolicyID := v.(string)
		settings.SecurityPolicyId = &securityPolicyID
		haveSettings = true
	}

	// If we have load balancer settings, apply them.
	if haveSettings {
		loadBalancerID := base64.URLEncoding.EncodeToString([]byte(details.LoadBalancerArn))
		_, err := c.TenantUpdateLbSettings(tenantID, loadBalancerID, settings)
		if err != nil {
			return err
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

func doesReplicationControllerHaveAlb(duplo *duplosdk.DuploReplicationController) bool {
	if duplo != nil && duplo.Template != nil {
		for _, lb := range duplo.Template.LBConfigurations {
			if lb.LbType == 1 || lb.LbType == 2 { // ALB or Healthcheck only
				return true
			}
		}
	}
	return false
}

func doesReplicationControllerHaveAlbOrNlb(duplo *duplosdk.DuploReplicationController) bool {
	if duplo != nil && duplo.Template != nil {
		for _, lb := range duplo.Template.LBConfigurations {
			if lb.LbType == 1 || lb.LbType == 2 || lb.LbType == 6 { // ALB, Healthcheck only or NLB
				return true
			}
		}
	}
	return false
}
