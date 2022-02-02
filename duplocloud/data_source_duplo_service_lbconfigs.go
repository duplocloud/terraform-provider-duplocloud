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

func duploLbConfigSchemaComputed() map[string]*schema.Schema {
	x := duploLbConfigSchema()
	for _, v := range x {
		v.Computed = true
		v.Required = false
		v.Optional = false
		v.ForceNew = false
	}
	return x
}

// SCHEMA for resource search/data
func dataSourceDuploServiceLbConfigs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_duplo_service_lbconfigs` retrieves load balancer configuration(s) for container-based service(s) in Duplo.\n\n" +
			"NOTE: For Amazon ECS services, see the `duplocloud_ecs_services` data source.",

		ReadContext: dataSourceDuploServiceLbConfigsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that hosts the duplo service.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The name of the duplo service.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the duplo service.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"replication_controller_name": {
							Description: "The name of the duplo service.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"arn": {
							Description: "The ARN (or ID) of the cloud load balancer (if applicable).",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "The status of the cloud load balancer (if applicable).",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"lbconfigs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: duploLbConfigSchemaComputed(),
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

				if lbdetails != nil {
					svc["arn"] = lbdetails.LoadBalancerArn
					if lbdetails.State != nil && lbdetails.State.Code != nil {
						svc["status"] = lbdetails.State.Code.Value
					}
				}
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
