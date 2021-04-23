package duplocloud

import (
	"encoding/json"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func awsLbTargetGroupsComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"vpc_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"load_balancer_arns": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"target_type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"protocol": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"protocol_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"health_check": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"path": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"port": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"protocol": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"matcher": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"grpc_code": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"http_code": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
					"interval": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"timeout": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"healthy_threshold": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"unhealthy_threshold": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
		},
	}
}

// Data source listing target groups
func dataSourceTenantAwsLbTargetGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsLbTargetGroupsRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: awsLbTargetGroupsComputedSchema(),
				},
			},
		},
	}
}

/// READ resource
func dataSourceTenantAwsLbTargetGroupsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsLbTargetGroupsRead ******** 1 start")

	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	// Get all listeners from duplo
	duplo, err := c.TenantListApplicationLbTargetGroups(tenantID)
	if err != nil {
		return err
	}

	// Build a list of target groups
	targetGroups := make([]map[string]interface{}, 0, len(*duplo))
	for _, item := range *duplo {

		// First apply simple scalars.
		tg := map[string]interface{}{
			"tenant_id":          tenantID,
			"arn":                item.TargetGroupArn,
			"name":               item.TargetGroupName,
			"protocol_version":   item.ProtocolVersion,
			"load_balancer_arns": item.LoadBalancerArns,
			"vpc_id":             item.VpcID,
			"health_check": []map[string]interface{}{{
				"enabled":             item.HealthCheckEnabled,
				"path":                item.HealthCheckPath,
				"interval":            item.HealthCheckIntervalSeconds,
				"timeout":             item.HealthCheckTimeoutSeconds,
				"healthy_threshold":   item.HealthyThreshold,
				"unhealthy_threshold": item.UnhealthyThreshold,
			}},
		}
		if item.Protocol != nil {
			tg["protocol"] = item.Protocol.Value
		}
		if item.TargetType != nil {
			tg["target_type"] = item.TargetType.Value
		}
		if item.HealthCheckProtocol != nil {
			tg["health_check"].([]map[string]interface{})[0]["protocol"] = item.HealthCheckProtocol.Value
		}
		if item.HealthMatcher != nil {
			tg["health_check"].([]map[string]interface{})[0]["matcher"] = []map[string]interface{}{{
				"grpc_code": item.HealthMatcher.GrpcCode,
				"http_code": item.HealthMatcher.HttpCode,
			}}
		} else {
			tg["health_check"].([]map[string]interface{})[0]["matcher"] = []map[string]interface{}{{}}
		}

		targetGroups = append(targetGroups, tg)
	}
	d.SetId(tenantID)

	// Apply the result
	dump, _ := json.Marshal(targetGroups)
	log.Printf("[TRACE] dataSourceTenantAwsLbTargetGroupsRead ******** 2 dump: %s", dump)
	d.Set("target_groups", targetGroups)

	log.Printf("[TRACE] dataSourceTenantAwsLbTargetGroupsRead ******** 3 end")
	return nil
}
