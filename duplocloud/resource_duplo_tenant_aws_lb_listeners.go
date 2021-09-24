package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLoadBalancerListenerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the load balancer will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"load_balancer_name": {
			Description: "The short name of the load balancer.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"load_balancer_fullname": {
			Description: "The full name of the load balancer.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "Port on which the load balancer is listening.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"protocol": {
			Description: "Protocol for connections from clients to the load balancer.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"target_group_arn": {
			Description: "ARN of the Target Group to which to route traffic.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"certificate_arn": {
			Description: "The ARN of the certificate to attach to the listener.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"arn": {
			Description: "ARN of the listener.",
			Type:        schema.TypeString,
			Computed:    true,
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
		"ssl_policy": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Resource for managing an AWS load balancer listener
func resourceAwsLoadBalancerListener() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_load_balancer_listener` manages an AWS application load balancer listener in Duplo.",

		ReadContext:   resourceAwsLoadBalancerListenerRead,
		CreateContext: resourceAwsLoadBalancerListenerCreate,
		//UpdateContext: resourceAwsLoadBalancerListenerUpdate,
		DeleteContext: resourceAwsLoadBalancerListenerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: awsLoadBalancerListenerSchema(),
	}
}

/// READ resource
func resourceAwsLoadBalancerListenerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerListenerRead - start")

	tenantID := d.Get("tenant_id").(string)
	lbName := d.Get("load_balancer_name").(string)
	arn := d.Get("arn").(string)
	c := m.(*duplosdk.Client)

	// Get all listeners from duplo
	duplo, err := c.TenantListApplicationLbListeners(tenantID, lbName)
	if err != nil {
		return diag.FromErr(err)
	}
	lbFullName, err := c.TenantGetApplicationLbFullName(tenantID, lbName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve listener for tenant %s load balancer '%s': %s", tenantID, lbName, err)
	}

	for _, item := range *duplo {
		if arn == item.ListenerArn {
			flattenAwsLoadBalancerListener(d, tenantID, lbName, lbFullName, &item)
			break
		}
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, arn))

	log.Printf("[TRACE] dataSourceTenantAwsLbListenersRead - end")
	return nil
}

/// CREATE resource
func resourceAwsLoadBalancerListenerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerListenerCreate - start")
	// Create the request object.
	targetArn := d.Get("target_group_arn").(string)
	lbShortName := d.Get("load_balancer_name").(string)
	log.Printf("[TRACE] lbShortName - %s", lbShortName)
	rq := expandAwsLoadBalancerListener(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	lbName := d.Get("load_balancer_name").(string)
	lbFullName, err := c.TenantGetApplicationLbFullName(tenantID, lbShortName)

	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	err = c.TenantCreateApplicationLbListener(tenantID, lbFullName, rq)
	if err != nil {
		return diag.Errorf("Error while creating listener rule for tenant %s load balancer '%s': %s", tenantID, lbName, err)
	}
	listener, err := c.TenantApplicationLbListenersByTargetGrpArn(tenantID, lbFullName, targetArn)

	id := fmt.Sprintf("%s/%s", tenantID, listener.ListenerArn)

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "load balancer listener", id, func() (interface{}, duplosdk.ClientError) {
		listener, err = c.TenantApplicationLbListenersByTargetGrpArn(tenantID, lbFullName, targetArn)
		return listener, err
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	d.Set("arn", listener.ListenerArn)

	diags = resourceAwsLoadBalancerListenerRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLoadBalancerListenerCreate - end")
	return diags
}

/// UPDATE resource
// func resourceAwsLoadBalancerListenerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	return nil
// }

/// DELETE resource
func resourceAwsLoadBalancerListenerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceAwsLoadBalancerListenerDelete - start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)

	targetArn := d.Get("target_group_arn").(string)
	tenantId := d.Get("tenant_id").(string)
	lbFullName := d.Get("load_balancer_fullname").(string)
	listenerArn := d.Get("arn").(string)
	id := d.Id()

	err := c.TenantDeleteApplicationLbListener(tenantId, lbFullName, listenerArn)
	if err != nil {
		return diag.Errorf("Error deleting load balancer listener '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the load balancer listener.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "load balancer listener", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantApplicationLbListenersByTargetGrpArn(tenantId, lbFullName, targetArn)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsLoadBalancerListenerDelete - end")
	return nil
}

func expandAwsLoadBalancerListener(d *schema.ResourceData) duplosdk.DuploAwsLbListenerCreate {
	log.Printf("[TRACE] expandAwsLoadBalancerListener - start")
	targetArn := d.Get("target_group_arn").(string)
	cert := duplosdk.DuploAwsLbListenerCertificate{
		CertificateArn: d.Get("certificate_arn").(string),
		IsDefault:      false,
	}
	certs := []duplosdk.DuploAwsLbListenerCertificate{cert}
	log.Printf("[TRACE]  Certs : %+v -", certs)
	action := duplosdk.DuploAwsLbListenerActionCreate{
		TargetGroupArn: targetArn,
		Type:           "Forward",
	}
	log.Printf("[TRACE] action- %+v", action)
	protocol := d.Get("protocol").(string)
	actions := []duplosdk.DuploAwsLbListenerActionCreate{action}
	log.Printf("[TRACE] actions- %+v", actions)
	duploObject := duplosdk.DuploAwsLbListenerCreate{
		Port:           d.Get("port").(int),
		Protocol:       protocol,
		Certificates:   certs,
		DefaultActions: actions,
	}
	// data, err := json.Marshal(duploObject)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//fmt.Printf("%s\n", data)
	log.Printf("[TRACE] duploObject %+v - ", duploObject)
	log.Printf("[TRACE] expandAwsLoadBalancerListener - end")
	return duploObject
}

func flattenAwsLoadBalancerListener(d *schema.ResourceData, tenantID string, lbName string, lbFullname string, duplo *duplosdk.DuploAwsLbListener) {
	d.Set("tenant_id", tenantID)
	d.Set("load_balancer_name", lbName)
	d.Set("load_balancer_fullname", lbFullname)
	d.Set("arn", duplo.ListenerArn)
	d.Set("port", duplo.Port)
	if duplo.Protocol != nil {
		d.Set("protocol", duplo.Protocol.Value)
	}
	certs := make([]map[string]interface{}, 0, len(duplo.Certificates))
	for i := range duplo.Certificates {
		certs = append(certs, map[string]interface{}{
			"arn":        duplo.Certificates[i].CertificateArn,
			"is_default": duplo.Certificates[i].IsDefault,
		})
	}

	d.Set("certificates", certs)

	actions := make([]map[string]interface{}, 0, len(duplo.DefaultActions))
	for i := range duplo.DefaultActions {
		action := map[string]interface{}{
			"target_group_arn": duplo.DefaultActions[i].TargetGroupArn,
			"order":            duplo.DefaultActions[i].Order,
		}
		if duplo.DefaultActions[i].Type != nil {
			action["type"] = duplo.DefaultActions[i].Type.Value
		}
		actions = append(actions, action)
	}

	d.Set("default_actions", actions)
	d.Set("load_balancer_arn", duplo.LoadBalancerArn)
	d.Set("ssl_policy", duplo.SSLPolicy)
}
