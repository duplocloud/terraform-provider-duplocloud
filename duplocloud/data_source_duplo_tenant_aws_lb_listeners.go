package duplocloud

import (
	"encoding/json"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func awsLbListenersComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"load_balancer_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"certificates": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"arn": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"is_default": {
						Type:     schema.TypeBool,
						Computed: true,
					},
				},
			},
		},
		"default_actions": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"target_group_arn": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"type": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"order": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
		},
		"load_balancer_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"port": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"protocol": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ssl_policy": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Data source listing LB listeners
func dataSourceTenantAwsLbListeners() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsLbListenersRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"listeners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: awsLbListenersComputedSchema(),
				},
			},
		},
	}
}

// READ resource
func dataSourceTenantAwsLbListenersRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsLbListenersRead ******** 1 start")

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)

	// Get all listeners from duplo
	duplo, err := c.TenantListApplicationLbListeners(tenantID, name)
	if err != nil {
		return err
	}

	// Build a list of listeners
	listeners := make([]map[string]interface{}, 0, len(*duplo))
	for _, item := range *duplo {

		// First apply simple scalars.
		listener := map[string]interface{}{
			"tenant_id":          tenantID,
			"load_balancer_name": name,
			"arn":                item.ListenerArn,
			"load_balancer_arn":  item.LoadBalancerArn,
			"port":               item.Port,
			"ssl_policy":         item.SSLPolicy,
		}
		if item.Protocol != nil {
			listener["protocol"] = item.Protocol.Value
		}

		// Finally, apply lists.
		certs := make([]map[string]interface{}, 0, len(item.Certificates))
		for i := range item.Certificates {
			certs = append(certs, map[string]interface{}{
				"arn":        item.Certificates[i].CertificateArn,
				"is_default": item.Certificates[i].IsDefault,
			})
		}
		listener["certificates"] = certs
		actions := make([]map[string]interface{}, 0, len(item.DefaultActions))
		for i := range item.DefaultActions {
			action := map[string]interface{}{
				"target_group_arn": item.DefaultActions[i].TargetGroupArn,
				"order":            item.DefaultActions[i].Order,
			}
			if item.DefaultActions[i].Type != nil {
				action["type"] = item.DefaultActions[i].Type.Value
			}
			actions = append(actions, action)
		}
		listener["default_actions"] = actions

		listeners = append(listeners, listener)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	// Apply the result
	dump, _ := json.Marshal(listeners)
	log.Printf("[TRACE] dataSourceTenantAwsLbListenersRead ******** 2 dump: %s", dump)
	d.Set("listeners", listeners)

	log.Printf("[TRACE] dataSourceTenantAwsLbListenersRead ******** 3 end")
	return nil
}
