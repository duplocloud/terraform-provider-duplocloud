package duplocloud

import (
	"context"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsRefreshInstancesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the launch template will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the launch template",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"version": {
			Description: "The version of the launch template",
			Type:        schema.TypeString,
			Required:    true,
			//ForceNew:true,

		},
		"asg_name": {
			Description: "The name of the asg to which template should be attached",
			Type:        schema.TypeString,
			Required:    true,
		},
		"preferences": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"auto_rollback": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"instance_warmup": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"max_healthy_percentage": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"min_healthy_percentage": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
	}
}
func resourceAwsRefreshInstances() *schema.Resource {
	return &schema.Resource{
		Description: "",
		//ReadContext:   resourceAwsRefreshInstancesRead,
		CreateContext: resourceAwsRefreshInstancesCreate,
		//UpdateContext: resourceAwsRefreshInstancesUpdate,
		//DeleteContext: resourceAwsRefreshInstancesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: awsRefreshInstancesSchema(),
	}
}

//	func resourceAwsRefreshInstancesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//		return nil
//	}
func resourceAwsRefreshInstancesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

//func resourceAwsRefreshInstancesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	return nil
//
//}
//func resourceAwsRefreshInstancesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	return nil
//
//}

func expandRefreshInstances(d *schema.ResourceData) duplosdk.DuploAwsRefreshInstancesRequest {
	lt := duplosdk.DuploLaunchTemplate{}
	lt.LaunchTemplateName = d.Get("name").(string)
	lt.Version = d.Get("version").(string)

	return duplosdk.DuploAwsRefreshInstancesRequest{
		AutoScalingGroupName: d.Get("asg_name").(string),
		DesiredConfiguration: &duplosdk.DuploAwsRefreshInstancesDesiredConfiguration{
			LaunchTemplate: &lt,
		},
		Preferences: &duplosdk.DuploRefreshInstancesPreference{
			AutoRollback:         d.Get("preferences.0.auto_rollback").(bool),
			InstanceWarmup:       d.Get("preferences.0.instance_warmup").(int),
			MaxHealthyPercentage: d.Get("preferences.0.max_healthy_percentage").(int),
			MinHealthyPercentage: d.Get("preferences.0.min_healthy_percentage").(int),
		},
	}

}
