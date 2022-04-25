package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLoadBalancerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the load balancer will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the load balancer.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the load balancer.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"load_balancer_type": {
			Description: "The type of load balancer to create. Possible values are `Application` or `Network`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Application",
			ValidateFunc: validation.StringInSlice([]string{
				"Application",
				"Network",
			}, true),
		},
		"arn": {
			Description: "The ARN of the load balancer.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"is_internal": {
			Description: "Whether or not the load balancer is internal (non internet-facing).",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"enable_access_logs": {
			Description: "Whether or not access logs should be enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"drop_invalid_headers": {
			Description:      "Whether or not the load balancer should drop invalid HTTP headers. Only valid for Load Balancers of type `Application`",
			Type:             schema.TypeBool,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: suppressIfLBType("Network"),
		},
		"web_acl_id": {
			Description: "The ARN of a WAF to attach to the load balancer.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"tags": {
			Description: "The tags assigned to this load balancer.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"dns_name": {
			Description: "The DNS name of the load balancer.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceAwsLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_load_balancer` manages an AWS application load balancer in Duplo.",

		ReadContext:   resourceAwsLoadBalancerRead,
		CreateContext: resourceAwsLoadBalancerCreate,
		UpdateContext: resourceAwsLoadBalancerUpdate,
		DeleteContext: resourceAwsLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: awsLoadBalancerSchema(),
	}
}

/// READ resource
func resourceAwsLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetApplicationLB(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s load balancer '%s': %s", tenantID, name, err)
	}

	// Next, get the load balancer settings.
	settings, err := c.TenantGetApplicationLbSettings(tenantID, duplo.Arn)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s load balancer '%s' WAF: %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceAwsLoadBalancerSetData(d, tenantID, name, duplo, settings)
	log.Printf("[TRACE] resourceAwsLoadBalancerRead ******** end")
	return nil
}

/// CREATE resource
func resourceAwsLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerCreate ******** start")

	// Create the request object.
	duploObject := duplosdk.DuploAwsLBConfiguration{
		Name:          d.Get("name").(string),
		IsInternal:    d.Get("is_internal").(bool),
		LbTypeSyncApi: d.Get("load_balancer_type").(string),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateApplicationLB(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error applying tenant %s load balancer '%s': %s", tenantID, duploObject.Name, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the load balancer's details.
	var resource *duplosdk.DuploApplicationLB
	id := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "load balancer", id, func() (interface{}, duplosdk.ClientError) {
		resource, err = c.TenantGetApplicationLB(tenantID, duploObject.Name)
		return resource, err
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	d.Set("arn", resource.Arn) // Update needs the ARN

	diags = resourceAwsLoadBalancerUpdate(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLoadBalancerCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceAwsLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerUpdate ******** start")

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	// apply load balancer settings
	log.Printf("[TRACE] resourceAwsLoadBalancerUpdate ******** update settings: start")
	settingsRq := duplosdk.DuploAwsLbSettingsUpdateRequest{
		LoadBalancerArn:    d.Get("arn").(string),
		EnableAccessLogs:   d.Get("enable_access_logs").(bool),
		DropInvalidHeaders: d.Get("drop_invalid_headers").(bool),
		WebACLID:           d.Get("web_acl_id").(string),
	}
	err := c.TenantUpdateApplicationLbSettings(tenantID, settingsRq)
	if err != nil {
		return diag.Errorf("Error configuring load balancer %s settings: %s", settingsRq.LoadBalancerArn, err)
	}
	log.Printf("[TRACE] resourceAwsLoadBalancerUpdate ******** update settings: end")

	// Get the result from Duplo.
	resource, err := c.TenantGetApplicationLB(tenantID, name)
	if err != nil {
		return diag.Errorf("Error retrieving load balancer '%s/%s': %s", tenantID, name, err)
	}
	settings, err := c.TenantGetApplicationLbSettings(tenantID, settingsRq.LoadBalancerArn)
	if err != nil {
		return diag.Errorf("Error retrieving load balancer %s settings: %s", settingsRq.LoadBalancerArn, err)
	}
	resourceAwsLoadBalancerSetData(d, tenantID, name, resource, settings)

	log.Printf("[TRACE] resourceAwsLoadBalancerUpdate ******** end")
	return nil
}

/// DELETE resource
func resourceAwsLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.TenantDeleteApplicationLB(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting load balancer '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the load balancer.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "load balancer", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetApplicationLB(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsLoadBalancerDelete ******** end")
	return nil
}

func resourceAwsLoadBalancerSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploApplicationLB, settings *duplosdk.DuploAwsLbSettings) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	if duplo.LbType != nil {
		d.Set("load_balancer_type", duplo.LbType.Value)
	}
	d.Set("arn", duplo.Arn)
	d.Set("is_internal", duplo.IsInternal)
	d.Set("enable_access_logs", settings.EnableAccessLogs)
	d.Set("drop_invalid_headers", settings.DropInvalidHeaders)
	d.Set("web_acl_id", settings.WebACLID)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("dns_name", duplo.DNSName)
}

func suppressIfLBType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return strings.EqualFold(d.Get("load_balancer_type").(string), t)
	}
}
