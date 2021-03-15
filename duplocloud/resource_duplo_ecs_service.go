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

// DuploEcsServiceSchema returns a Terraform resource schema for an ECS Service
func ecsServiceSchema() map[string]*schema.Schema {
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
		"task_definition": {
			Type:     schema.TypeString,
			Required: true,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"health_check_grace_period_seconds": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"old_task_definition_buffer_size": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  10,
		},
		"is_target_group_only": {
			Type:     schema.TypeBool,
			ForceNew: true,
			Optional: true,
			Required: false,
			Default:  false,
		},
		"dns_prfx": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"load_balancer": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"replication_controller_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"lb_type": {
						Type:     schema.TypeInt,
						Optional: false,
						Required: true,
					},
					"port": {
						Type:     schema.TypeString,
						Optional: false,
						Required: true,
					},
					"protocol": {
						Type:     schema.TypeString,
						Optional: false,
						Required: true,
					},
					"external_port": {
						Type:     schema.TypeInt,
						Optional: false,
						Required: true,
					},
					"backend_protocol": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"is_internal": {
						Type:     schema.TypeBool,
						Optional: true,
						Required: false,
						Default:  false,
					},
					"health_check_url": {
						Type:     schema.TypeString,
						Optional: true,
						Required: false,
					},
					"certificate_arn": {
						Type:     schema.TypeString,
						Optional: true,
						Required: false,
					},
					"load_balancer_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"load_balancer_arn": {
						Type:     schema.TypeString,
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
					"webaclid": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

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
		Schema: ecsServiceSchema(),
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
	ecsServiceToState(duplo, d)
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", duplo.TenantID, duplo.Name))

	// Finally, retrieve any load balancer settings.
	err = readEcsServiceAwsLbSettings(duplo.TenantID, duplo.Name, d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploEcsServiceRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcsServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploEcsServiceCreateOrUpdate(ctx, d, m, false)
}

/// UPDATE resource
func resourceDuploEcsServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDuploEcsServiceCreateOrUpdate(ctx, d, m, true)
}

func resourceDuploEcsServiceCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, updating bool) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	// Create the request object.
	duplo := ecsServiceFromState(d)

	log.Printf("[TRACE] resourceDuploEcsServiceCreateOrUpdate(%s, %s): start", tenantID, duplo.Name)

	// Create the ECS service.
	c := m.(*duplosdk.Client)
	rpObject, err := c.EcsServiceCreateOrUpdate(tenantID, &duplo, updating)
	if err != nil {
		return diag.FromErr(err)
	}
	if !updating {
		d.SetId(fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantID, rpObject.Name))
	}

	// Next, we need to apply load balancer settings.
	err = updateEcsServiceAwsLbSettings(tenantID, rpObject.Name, d, c)
	if err != nil {
		return diag.Errorf("Error applying ECS load balancer settings '%s': %s", d.Id(), err)
	}

	diags := resourceDuploEcsServiceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsServiceCreate ******** end")
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

// EcsServiceFromState converts resource data respresenting an ECS Service to a Duplo SDK object.
func ecsServiceFromState(d *schema.ResourceData) duplosdk.DuploEcsService {
	duploObject := duplosdk.DuploEcsService{}

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.TaskDefinition = d.Get("task_definition").(string)
	duploObject.Replicas = d.Get("replicas").(int)
	duploObject.HealthCheckGracePeriodSeconds = d.Get("health_check_grace_period_seconds").(int)
	duploObject.OldTaskDefinitionBufferSize = d.Get("old_task_definition_buffer_size").(int)
	duploObject.IsTargetGroupOnly = d.Get("is_target_group_only").(bool)
	duploObject.DNSPrfx = d.Get("dns_prfx").(string)

	// Next, convert things into structured data.
	duploObject.LBConfigurations = ecsLoadBalancersFromState(d)

	return duploObject
}

// EcsServiceToState converts a Duplo SDK object respresenting an ECS Service to terraform resource data.
func ecsServiceToState(duploObject *duplosdk.DuploEcsService, d *schema.ResourceData) {
	// First, convert things into simple scalars
	d.Set("tenant_id", duploObject.TenantID)
	d.Set("name", duploObject.Name)
	d.Set("task_definition", duploObject.TaskDefinition)
	d.Set("replicas", duploObject.Replicas)
	d.Set("health_check_grace_period_seconds", duploObject.HealthCheckGracePeriodSeconds)
	d.Set("old_task_definition_buffer_size", duploObject.OldTaskDefinitionBufferSize)
	d.Set("is_target_group_only", duploObject.IsTargetGroupOnly)
	d.Set("dns_prfx", duploObject.DNSPrfx)

	// Next, convert things into structured data.
	d.Set("load_balancer", ecsLoadBalancersToState(duploObject.Name, duploObject.LBConfigurations))
}

func ecsLoadBalancersToState(name string, lbcs *[]duplosdk.DuploEcsServiceLbConfig) []map[string]interface{} {
	if lbcs == nil {
		return nil
	}

	var ary []map[string]interface{}

	for _, lbc := range *lbcs {
		jo := make(map[string]interface{})
		jo["replication_controller_name"] = name
		jo["lb_type"] = lbc.LbType
		jo["port"] = lbc.Port
		jo["protocol"] = lbc.Protocol
		jo["backend_protocol"] = lbc.BackendProtocol
		if jo["backend_protocol"] == "" {
			jo["backend_protocol"] = "HTTP"
		}
		jo["external_port"] = lbc.ExternalPort
		jo["is_internal"] = lbc.IsInternal
		jo["health_check_url"] = lbc.HealthCheckURL
		jo["certificate_arn"] = lbc.CertificateArn
		ary = append(ary, jo)
	}

	return ary
}

func ecsLoadBalancersFromState(d *schema.ResourceData) *[]duplosdk.DuploEcsServiceLbConfig {
	var ary []duplosdk.DuploEcsServiceLbConfig

	slb := d.Get("load_balancer").(*schema.Set)
	if slb == nil {
		return nil
	}

	log.Printf("[TRACE] ecsLoadBalancersFromState ********: have data")

	for _, _lb := range slb.List() {
		lb := _lb.(map[string]interface{})
		ary = append(ary, duplosdk.DuploEcsServiceLbConfig{
			ReplicationControllerName: lb["replication_controller_name"].(string),
			LbType:                    lb["lb_type"].(int),
			Port:                      lb["port"].(string),
			Protocol:                  lb["protocol"].(string),
			BackendProtocol:           lb["backend_protocol"].(string),
			ExternalPort:              lb["external_port"].(int),
			IsInternal:                lb["is_internal"].(bool),
			HealthCheckURL:            lb["health_check_url"].(string),
			CertificateArn:            lb["certificate_arn"].(string),
		})
	}

	return &ary
}

func readEcsServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	slb := d.Get("load_balancer").(*schema.Set)
	if slb == nil || slb.Len() == 0 {
		return nil
	}
	state := slb.List()[0].(map[string]interface{})

	// Next, look for load balancer settings.
	details, err := c.TenantGetLbDetailsInService(tenantID, name)
	if err != nil {
		return err
	}
	if details != nil && details.LoadBalancerArn != "" {

		// Populate load balancer details.
		state["load_balancer_arn"] = details.LoadBalancerArn
		state["load_balancer_name"] = details.LoadBalancerName

		settings, err := c.TenantGetApplicationLbSettings(tenantID, details.LoadBalancerArn)
		if err != nil {
			return err
		}
		if settings != nil && settings.LoadBalancerArn != "" {

			// Populate load balancer settings.
			state["webaclid"] = settings.WebACLID
			state["enable_access_logs"] = settings.EnableAccessLogs
			state["drop_invalid_headers"] = settings.DropInvalidHeaders
		}
	}

	d.Set("load_balancer", []map[string]interface{}{state})

	return nil
}

func updateEcsServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	slb := d.Get("load_balancer").(*schema.Set)
	if slb == nil || slb.Len() == 0 {
		return nil
	}
	state := slb.List()[0].(map[string]interface{})

	// Get any load balancer settings from the user.
	settings := duplosdk.DuploAwsLbSettingsUpdateRequest{}
	haveSettings := false
	if v, ok := state["enable_access_logs"]; ok && v != nil {
		settings.EnableAccessLogs = v.(bool)
		haveSettings = true
	}
	if v, ok := state["drop_invalid_headers"]; ok && v != nil {
		settings.DropInvalidHeaders = v.(bool)
		haveSettings = true
	}
	if v, ok := state["webaclid"]; ok && v != nil {
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
