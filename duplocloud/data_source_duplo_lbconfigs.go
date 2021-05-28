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
func dataSourceDuploServiceLBConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceLBConfigsRead,
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
									"external_port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"health_check_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"certificate_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"is_native": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"is_internal": {
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

/// READ/SEARCH resources
func dataSourceDuploServiceLBConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServiceLBConfigsRead: start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	var id string

	// Get the object from Duplo, detecting a missing object
	var list *[]duplosdk.DuploServiceLBConfigs = nil
	var err error
	c := m.(*duplosdk.Client)
	if name == "" {
		id = tenantID
		list, err = c.DuploServiceLBConfigsGetList(tenantID)
	} else {
		id = fmt.Sprintf("%s/%s", tenantID, name)
		var service *duplosdk.DuploServiceLBConfigs
		service, err = c.DuploServiceLBConfigsGet(tenantID, name)
		if service != nil {
			list = &[]duplosdk.DuploServiceLBConfigs{*service}
		}
	}
	if err != nil {
		return diag.Errorf("Unable to list tenant %s load balancer configs: %s", tenantID, err)
	}
	d.SetId(id)

	// Apply the Terraform state for each service, or an empty list
	services := []interface{}{}
	if list != nil {
		services = make([]interface{}, 0, len(*list))
		for _, service := range *list {
			services = append(services, map[string]interface{}{
				"replication_controller_name": service.ReplicationControllerName,
				"status":                      service.Status,
				"arn":                         service.Arn,
				"lbconfigs":                   flattenDuploServiceLBConfigurations(service.LBConfigs),
			})
		}
	}
	if err = d.Set("services", services); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceDuploServiceLBConfigsRead: end")
	return nil
}
