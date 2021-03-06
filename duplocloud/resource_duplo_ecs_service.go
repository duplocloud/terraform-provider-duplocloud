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
			Description: "The GUID of the tenant that the service will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, //switch tenant
		},
		"name": {
			Description: "The name of the service to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"task_definition": {
			Description: "The ARN of the task definition to use.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"replicas": {
			Description: "The number of container replicas to create.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"health_check_grace_period_seconds": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"old_task_definition_buffer_size": {
			Description: "The number of older task definitions to retain in AWS.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     10,
		},
		"is_target_group_only": {
			Type:     schema.TypeBool,
			ForceNew: true,
			Optional: true,
			Default:  false,
		},
		"dns_prfx": {
			Description: "The DNS prefix to assign to this service's load balancer.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"load_balancer": {
			Description: "Zero or more load balancer configurations to associate with this service.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"replication_controller_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"lb_type": {
						Description: "The numerical index of the type of load balancer configuration to create.\n" +
							"Should be one of:\n\n" +
							"   - `0` : ELB (Classic Load Balancer)\n" +
							"   - `1` : ALB (Application Load Balancer)\n" +
							"   - `2` : Health-check Only (No Load Balancer)\n",
						Type:     schema.TypeInt,
						Optional: false,
						Required: true,
					},
					"port": {
						Description: "The backend port associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Optional:    false,
						Required:    true,
					},
					"protocol": {
						Description: "The frontend protocol associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Optional:    false,
						Required:    true,
					},
					"external_port": {
						Description: "The frontend port associated with this load balancer configuration.",
						Type:        schema.TypeInt,
						Optional:    false,
						Required:    true,
					},
					"backend_protocol": {
						Description: "The backend protocol associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"is_internal": {
						Description: "Whether or not to create an internal load balancer.",
						Type:        schema.TypeBool,
						Optional:    true,
						Required:    false,
						Default:     false,
					},
					"health_check_url": {
						Description: "The health check URL to associate with this load balancer configuration.",
						Type:        schema.TypeString,
						Optional:    true,
						Required:    false,
					},
					"certificate_arn": {
						Description: "The ARN of an ACM certificate to associate with this load balancer.  Only applicable for HTTPS.",
						Type:        schema.TypeString,
						Optional:    true,
						Required:    false,
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
					"webaclid": {
						Description: "The ARN of a web application firewall to associate this load balancer.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcsService() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecs_service` manages a Amazon ECS service in Duplo.",

		ReadContext:   resourceDuploEcsServiceRead,
		CreateContext: resourceDuploEcsServiceCreate,
		UpdateContext: resourceDuploEcsServiceUpdate,
		DeleteContext: resourceDuploEcsServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	var err error

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

	// First, convert things into simple scalars
	d.Set("tenant_id", duplo.TenantID)
	d.Set("name", duplo.Name)
	d.Set("task_definition", duplo.TaskDefinition)
	d.Set("replicas", duplo.Replicas)
	d.Set("health_check_grace_period_seconds", duplo.HealthCheckGracePeriodSeconds)
	d.Set("old_task_definition_buffer_size", duplo.OldTaskDefinitionBufferSize)
	d.Set("is_target_group_only", duplo.IsTargetGroupOnly)
	d.Set("dns_prfx", duplo.DNSPrfx)

	// Next, convert things into structured data.
	loadBalancers := ecsLoadBalancersToState(duplo.Name, duplo.LBConfigurations)

	// Retrieve the load balancer settings.
	if len(loadBalancers) > 0 {
		err = readEcsServiceAwsLbSettings(duplo.TenantID, duplo.Name, loadBalancers[0], c)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.Set("load_balancer", loadBalancers)

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
	var err error

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

	// Check if the object exists before attempting a delete.
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsServiceGet(d.Id())
	if err != nil || duplo != nil {
		err = c.EcsServiceDelete(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
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

	lb, err := getOptionalBlockAsMap(d, "load_balancer")
	if err != nil || lb == nil {
		return &ary
	}

	name := lb["replication_controller_name"].(string)
	if name == "" {
		name = d.Get("name").(string)
	}

	log.Printf("[TRACE] ecsLoadBalancersFromState ********: have data")

	ary = append(ary, duplosdk.DuploEcsServiceLbConfig{
		ReplicationControllerName: name,
		LbType:                    lb["lb_type"].(int),
		Port:                      lb["port"].(string),
		Protocol:                  lb["protocol"].(string),
		BackendProtocol:           lb["backend_protocol"].(string),
		ExternalPort:              lb["external_port"].(int),
		IsInternal:                lb["is_internal"].(bool),
		HealthCheckURL:            lb["health_check_url"].(string),
		CertificateArn:            lb["certificate_arn"].(string),
	})

	return &ary
}

func readEcsServiceAwsLbSettings(tenantID string, name string, lb map[string]interface{}, c *duplosdk.Client) error {
	// Next, look for load balancer settings.
	details, err := c.TenantGetLbDetailsInService(tenantID, name)
	if err != nil {
		return err
	}
	if details != nil && details.LoadBalancerArn != "" {

		// Populate load balancer details.
		lb["load_balancer_arn"] = details.LoadBalancerArn
		lb["load_balancer_name"] = details.LoadBalancerName

		settings, err := c.TenantGetApplicationLbSettings(tenantID, details.LoadBalancerArn)
		if err != nil {
			return err
		}
		if settings != nil && settings.LoadBalancerArn != "" {

			// Populate load balancer settings.
			lb["webaclid"] = settings.WebACLID
			lb["enable_access_logs"] = settings.EnableAccessLogs
			lb["drop_invalid_headers"] = settings.DropInvalidHeaders
		}
	}

	return nil
}

func updateEcsServiceAwsLbSettings(tenantID string, name string, d *schema.ResourceData, c *duplosdk.Client) error {
	var err error

	state, err := getOptionalBlockAsMap(d, "load_balancer")
	if err != nil || state == nil {
		return err
	}

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
		var details *duplosdk.DuploAwsLbDetailsInService
		details, err = c.TenantGetLbDetailsInService(tenantID, name)
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
