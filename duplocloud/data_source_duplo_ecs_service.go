package duplocloud

import (
	"context"
	"fmt"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploEcsServiceComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that contains the service.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the service.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"task_definition": {
			Description: "The ARN of the task definition to use.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"replicas": {
			Description: "The number of container replicas to create.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"health_check_grace_period_seconds": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"old_task_definition_buffer_size": {
			Description: "The number of older task definitions to retain in AWS.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"is_target_group_only": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"target_group_arns": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"dns_prfx": {
			Description: "The DNS prefix to assign to this service's load balancer.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"load_balancer": {
			Description: "Zero or more load balancer configurations to associate with this service.",
			Type:        schema.TypeList,
			Computed:    true,
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
						Computed: false,
					},
					"port": {
						Description: "The backend port associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Computed:    false,
					},
					"protocol": {
						Description: "The frontend protocol associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Computed:    false,
					},
					"external_port": {
						Description: "The frontend port associated with this load balancer configuration.",
						Type:        schema.TypeInt,
						Computed:    false,
					},
					"target_group_count": {
						Description: "Number of Load Balancer target group to associate with the service.",
						Type:        schema.TypeInt,
						Computed:    false,
					},
					"backend_protocol_version": {
						Description: "The backend protocol version associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"is_internal": {
						Description: "Whether or not to create an internal load balancer.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"health_check_url": {
						Description: "The health check URL to associate with this load balancer configuration.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"certificate_arn": {
						Description: "The ARN of an ACM certificate to associate with this load balancer.  Only applicable for HTTPS.",
						Type:        schema.TypeString,
						Computed:    true,
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
						Computed:    true,
					},
					"drop_invalid_headers": {
						Description: "Whether or not to drop invalid HTTP headers received by the load balancer.",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"webaclid": {
						Description: "The ARN of a web application firewall to associate this load balancer.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"health_check_config": {
						Description: "Health check configuration for this load balancer.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"healthy_threshold_count": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"unhealthy_threshold_count": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"health_check_timeout_seconds": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"health_check_interval_seconds": {
									Type:     schema.TypeInt,
									Computed: true,
								},
								"http_success_code": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"grpc_success_code": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceDuploEcsService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploEcsServiceRead,
		Schema:      duploEcsServiceComputedSchema(),
	}
}

func dataSourceDuploEcsServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourceDuploServiceRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsServiceGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s ECS service '%s': %s", tenantID, name, err)
	}
	if duplo == nil {
		return diag.Errorf("Unable to read tenant %s ECS service '%s': not found", tenantID, name)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	// Read the object into state
	flattenDuploEcsService(d, duplo, c)

	log.Printf("[TRACE] dataSourceDuploServiceRead: end")
	return nil
}
