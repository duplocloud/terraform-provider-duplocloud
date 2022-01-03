package duplocloud

import (
	"context"
	"fmt"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource search/data
func dataSourceDuploServiceLbConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceLbConfigsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_controller_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lbconfigs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"replication_controller_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"lb_type": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"port": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"host_port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"external_port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"is_infra_deployment": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"certificate_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cloud_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"health_check_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"external_traffic_policy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"backend_protocol_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"frontend_ip": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"is_internal": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"is_native": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceDuploServiceLbConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServiceLbConfigsRead: start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	var id string

	// Get the object from Duplo, detecting a missing object
	var list *[]duplosdk.DuploLbConfiguration
	var err error
	c := m.(*duplosdk.Client)
	if name == "" {
		id = tenantID
		list, err = c.LbConfigurationList(tenantID)
	} else {
		id = fmt.Sprintf("%s/%s", tenantID, name)
		list, err = c.ReplicationControllerLbConfigurationList(tenantID, name)
	}
	if err != nil {
		return diag.Errorf("Unable to list tenant %s load balancer configs: %s", tenantID, err)
	}
	d.SetId(id)

	// Apply the Terraform state for each LB config, or an empty list
	services := []interface{}{}
	if list != nil {

		// Track all of the services we've seen.
		services = make([]interface{}, 0, len(*list))
		svcMap := map[string]int{}

		// Handle each LB config
		for _, lb := range *list {

			// Handle a service we haven't seen in this loop yet
			svcIdx, ok := svcMap[lb.ReplicationControllerName]
			if !ok {

				// Add the service
				services = append(services, map[string]interface{}{
					"name":                        lb.ReplicationControllerName,
					"replication_controller_name": lb.ReplicationControllerName,
					"lbconfigs":                   make([]interface{}, 0, len(*list)),
				})

				// Record the index of the service
				svcIdx = len(svcMap)
				svcMap[lb.ReplicationControllerName] = svcIdx
			}

			// Get the service element.
			svc := services[svcIdx].(map[string]interface{})

			// Handle a cloud load balancer we haven't seen in this loop yet
			if svc["arn"] == nil && (lb.LbType != 2 && lb.LbType != 3 && lb.LbType != 4) {
				lbdetails, lberr := c.TenantGetLbDetailsInService(tenantID, svc["name"].(string))
				if lberr != nil && lberr.Status() != 404 {
					return diag.FromErr(err)
				}

				svc["arn"] = lbdetails.LoadBalancerArn
				svc["status"] = lbdetails.State.Code.Value
			}

			// Add this LB to the service's list of LB configs.
			svc["lbconfigs"] = append(svc["lbconfigs"].([]interface{}), flattenDuploServiceLbConfiguration(&lb))
		}
	}
	if err = d.Set("services", services); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceDuploServiceLbConfigsRead: end")
	return nil
}
