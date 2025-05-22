package duplocloud

import (
	"context"
	"log"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func asgInstanceRefresh() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the asg will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"asg_name": {
			Description: "The fullname of the asg",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"refresh_identifier": {
			Description: "To identify instance refresh or invoke a instance refresh.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"auto_rollback": {
			Description: "Automatically rollback if instance refresh fails. This can be set only if update_launch_template_version is specified",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"instance_warmup": {
			Description: "Number of seconds until a newly launched instance is configured and ready to use. Default behavior is to use the Auto Scaling Group's health check grace period.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"max_healthy_percentage": {
			Description:  "Amount of capacity in the Auto Scaling group that can be in service and healthy, or pending, to support your workload when an instance refresh is in place, as a percentage of the desired capacity of the Auto Scaling group.",
			Type:         schema.TypeInt,
			Default:      100,
			Optional:     true,
			ValidateFunc: validation.IntBetween(100, 200),
		},
		"min_healthy_percentage": {
			Description: "Amount of capacity in the Auto Scaling group that must remain healthy during an instance refresh to allow the operation to continue, as a percentage of the desired capacity of the Auto Scaling group.",
			Type:        schema.TypeInt,
			Default:     90,
			Optional:    true,
		},
		"update_launch_template_version": {
			Description: "Launch template version to update",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}
func resourceAsgInstanceRefresh() *schema.Resource {
	return &schema.Resource{
		Description:   "duplocloud_asg_instance_refresh triggers the instance refresh of asg in duplo. Change in refresh identifier or any other configuration field will trigger the instance refresh for the ASG instantly\n\n Note: Following Instance Replacement Method configurations are: \n1. Launch Before Terminating - Min Healthy Percentage = 100\n2. Terminate And Launch - Max Healthy Percentage = 100",
		ReadContext:   resourceASGInstanceRefreshRead,
		CreateContext: resourceASGInstanceRefreshCreate,
		UpdateContext: resourceASGInstanceRefreshCreate,
		DeleteContext: resourceASGInstanceRefreshDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: asgInstanceRefresh(),
	}
}

func resourceASGInstanceRefreshRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
func resourceASGInstanceRefreshCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)

	rq := expandInstanceRefresh(d)
	if d.Get("refresh_identifier").(string) == "" {
		log.Println("Skipping resource creation as refresh_identifier is 0")
		d.SetId(tenantId + "/asg-refresh/" + rq.AutoScalingGroupName)
		return nil
	}
	c := m.(*duplosdk.Client)
	err := c.AsgInstanceRefresh(tenantId, &rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/asg-refresh/" + rq.AutoScalingGroupName)
	diag := resourceASGInstanceRefreshRead(ctx, d, m)
	return diag

}

func resourceASGInstanceRefreshDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

func expandInstanceRefresh(d *schema.ResourceData) duplosdk.DuploAsgInstanceRefresh {
	name := d.Get("asg_name").(string)
	ver := d.Get("update_launch_template_version").(string)
	config := duplosdk.DuploAsgInstanceRefreshDesiredConfiguration{}
	if ver != "" {
		config = duplosdk.DuploAsgInstanceRefreshDesiredConfiguration{
			LaunchTemplate: duplosdk.DuploAsgInstanceRefreshDesiredConfigurationLaunchTemplate{
				LaunchTemplateName: name,
				Version:            ver,
			},
		}
	}
	preferences := duplosdk.DuploAsgInstanceRefreshPreference{
		InstanceWarmup:       d.Get("instance_warmup").(int),
		MaxHealthyPercentage: d.Get("max_healthy_percentage").(int),
		MinHealthyPercentage: d.Get("min_healthy_percentage").(int),
		AutoRollback:         d.Get("auto_rollback").(bool),
	}
	return duplosdk.DuploAsgInstanceRefresh{
		AutoScalingGroupName: name,
		Preferences:          preferences,
		DesiredConfiguration: config,
	}

}
