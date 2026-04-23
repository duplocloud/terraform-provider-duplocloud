package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsAsgWarmPoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that owns the ASG.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"asg_name": {
			Description: "The full name of the Auto Scaling Group (e.g. `duploservices-<tenant>-<shortname>`).",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"min_size": {
			Description:  "Minimum number of instances to maintain in the warm pool.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: validation.IntAtLeast(0),
		},
		"max_group_prepared_capacity": {
			Description:  "Maximum number of instances that are allowed to be in the warm pool or in any state except `Terminated` for the Auto Scaling group. Use `-1` to leave the value unspecified (no maximum).",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      -1,
			ValidateFunc: validation.IntAtLeast(-1),
		},
		"pool_state": {
			Description: "Sets the instance state to transition to after the lifecycle actions are complete. One of `Stopped`, `Running`, or `Hibernated`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Stopped",
			ValidateFunc: validation.StringInSlice([]string{
				"Stopped",
				"Running",
				"Hibernated",
			}, false),
		},
		"instance_reuse_policy": {
			Description: "Instance reuse policy for the warm pool.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"reuse_on_scale_in": {
						Description: "Whether instances in the warm pool can be returned to the warm pool on scale in.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
				},
			},
		},
		"instances": {
			Description: "Instances currently in the warm pool.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"instance_id":             {Type: schema.TypeString, Computed: true},
					"instance_type":           {Type: schema.TypeString, Computed: true},
					"availability_zone":       {Type: schema.TypeString, Computed: true},
					"health_status":           {Type: schema.TypeString, Computed: true},
					"lifecycle_state":         {Type: schema.TypeString, Computed: true},
					"protected_from_scale_in": {Type: schema.TypeBool, Computed: true},
					"launch_template_id":      {Type: schema.TypeString, Computed: true},
					"launch_template_name":    {Type: schema.TypeString, Computed: true},
					"launch_template_version": {Type: schema.TypeString, Computed: true},
				},
			},
		},
	}
}

func resourceAwsAsgWarmPool() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_aws_asg_warm_pool` manages an AWS Auto Scaling Group warm pool in Duplo. A warm pool is a pool of pre-initialized EC2 instances that sits alongside an Auto Scaling group, ready to be placed into service when the group needs to scale out.",
		ReadContext:   resourceAwsAsgWarmPoolRead,
		CreateContext: resourceAwsAsgWarmPoolCreate,
		UpdateContext: resourceAwsAsgWarmPoolUpdate,
		DeleteContext: resourceAwsAsgWarmPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsAsgWarmPoolSchema(),
	}
}

func asgWarmPoolIdParts(id string) (tenantID, asgName string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) == 3 && parts[2] == "warmpool" {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("invalid ASG warm pool resource ID: %s", id)
}

func resourceAwsAsgWarmPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	asgName := d.Get("asg_name").(string)
	log.Printf("[TRACE] resourceAwsAsgWarmPoolCreate(%s, %s): start", tenantID, asgName)

	c := m.(*duplosdk.Client)
	rq := expandAwsAsgWarmPool(d)
	if err := c.AsgWarmPoolCreateOrUpdate(tenantID, rq); err != nil {
		return diag.Errorf("Error creating ASG warm pool for '%s': %s", asgName, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/warmpool", tenantID, asgName))
	if err := asgWarmPoolWaitUntilReady(ctx, c, tenantID, asgName, rq.MinSize, d.Timeout("create")); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ASG warm pool '%s' to be ready: %s", asgName, err))
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolCreate(%s, %s): end", tenantID, asgName)
	return resourceAwsAsgWarmPoolRead(ctx, d, m)
}

func resourceAwsAsgWarmPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, asgName, err := asgWarmPoolIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolUpdate(%s, %s): start", tenantID, asgName)

	c := m.(*duplosdk.Client)
	rq := expandAwsAsgWarmPool(d)
	if cerr := c.AsgWarmPoolCreateOrUpdate(tenantID, rq); cerr != nil {
		return diag.Errorf("Error updating ASG warm pool for '%s': %s", asgName, cerr)
	}

	if werr := asgWarmPoolWaitUntilReady(ctx, c, tenantID, asgName, rq.MinSize, d.Timeout("update")); werr != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ASG warm pool '%s' to be ready: %s", asgName, werr))
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolUpdate(%s, %s): end", tenantID, asgName)
	return resourceAwsAsgWarmPoolRead(ctx, d, m)
}

func resourceAwsAsgWarmPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, asgName, err := asgWarmPoolIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolRead(%s, %s): start", tenantID, asgName)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AsgWarmPoolGet(tenantID, asgName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsAsgWarmPoolRead(%s): warm pool not found", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve ASG warm pool '%s': %s", d.Id(), cerr)
	}
	if rp == nil || rp.WarmPoolConfiguration == nil {
		log.Printf("[TRACE] resourceAwsAsgWarmPoolRead(%s): warm pool configuration empty", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("tenant_id", tenantID)
	d.Set("asg_name", asgName)
	flattenAwsAsgWarmPool(d, rp)
	log.Printf("[TRACE] resourceAwsAsgWarmPoolRead(%s, %s): end", tenantID, asgName)
	return nil
}

func resourceAwsAsgWarmPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, asgName, err := asgWarmPoolIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolDelete(%s, %s): start", tenantID, asgName)

	c := m.(*duplosdk.Client)
	if cerr := c.AsgWarmPoolDelete(tenantID, asgName); cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsAsgWarmPoolDelete(%s): warm pool already gone", d.Id())
			return nil
		}
		return diag.Errorf("Error deleting ASG warm pool '%s': %s", d.Id(), cerr)
	}

	if werr := asgWarmPoolWaitUntilGone(ctx, c, tenantID, asgName, d.Timeout("delete")); werr != nil {
		return diag.FromErr(fmt.Errorf("error waiting for ASG warm pool '%s' to finish terminating: %s", asgName, werr))
	}
	log.Printf("[TRACE] resourceAwsAsgWarmPoolDelete(%s, %s): end", tenantID, asgName)
	return nil
}

func expandAwsAsgWarmPool(d *schema.ResourceData) *duplosdk.DuploAsgWarmPoolRequest {
	rq := &duplosdk.DuploAsgWarmPoolRequest{
		AutoScalingGroupName:     d.Get("asg_name").(string),
		MinSize:                  d.Get("min_size").(int),
		MaxGroupPreparedCapacity: d.Get("max_group_prepared_capacity").(int),
		PoolState:                d.Get("pool_state").(string),
	}

	if v, ok := d.GetOk("instance_reuse_policy"); ok {
		list := v.([]interface{})
		if len(list) > 0 && list[0] != nil {
			m := list[0].(map[string]interface{})
			rq.InstanceReusePolicy = &duplosdk.DuploAsgInstanceReusePolicy{
				ReuseOnScaleIn: m["reuse_on_scale_in"].(bool),
			}
		}
	}

	return rq
}

func flattenAwsAsgWarmPool(d *schema.ResourceData, rp *duplosdk.DuploAsgWarmPoolResponse) {
	cfg := rp.WarmPoolConfiguration
	d.Set("min_size", cfg.MinSize)
	d.Set("max_group_prepared_capacity", cfg.MaxGroupPreparedCapacity)
	if cfg.PoolState != nil {
		d.Set("pool_state", cfg.PoolState.Value)
	}
	if cfg.InstanceReusePolicy != nil {
		d.Set("instance_reuse_policy", []interface{}{
			map[string]interface{}{
				"reuse_on_scale_in": cfg.InstanceReusePolicy.ReuseOnScaleIn,
			},
		})
	} else {
		d.Set("instance_reuse_policy", []interface{}{})
	}

	instances := make([]interface{}, 0, len(rp.Instances))
	for _, inst := range rp.Instances {
		m := map[string]interface{}{
			"instance_id":             inst.InstanceId,
			"instance_type":           inst.InstanceType,
			"availability_zone":       inst.AvailabilityZone,
			"health_status":           inst.HealthStatus,
			"protected_from_scale_in": inst.ProtectedFromScaleIn,
		}
		if inst.LifecycleState != nil {
			m["lifecycle_state"] = inst.LifecycleState.Value
		}
		if inst.LaunchTemplate != nil {
			m["launch_template_id"] = inst.LaunchTemplate.LaunchTemplateId
			m["launch_template_name"] = inst.LaunchTemplate.LaunchTemplateName
			m["launch_template_version"] = inst.LaunchTemplate.Version
		}
		instances = append(instances, m)
	}
	d.Set("instances", instances)
}

// asgWarmPoolWaitUntilReady polls until at least `minSize` warm instances
// have reached a terminal warmed lifecycle state (Warmed:Stopped,
// Warmed:Running, or Warmed:Hibernated) and are "Healthy". Returns
// immediately when minSize <= 0.
//
// Note: AWS only applies PoolState changes to instances launched after the
// change, so we do not require instances to match the requested pool_state —
// an instance launched into the prior state is still "ready" from the warm
// pool's capacity perspective.
func asgWarmPoolWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, asgName string, minSize int, timeout time.Duration) error {
	if minSize <= 0 {
		return nil
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AsgWarmPoolGet(tenantID, asgName)
			if err != nil {
				return nil, "", err
			}
			if rp == nil {
				return nil, "pending", nil
			}
			ready := 0
			for _, inst := range rp.Instances {
				if inst.LifecycleState == nil {
					continue
				}
				switch inst.LifecycleState.Value {
				case "Warmed:Stopped", "Warmed:Running", "Warmed:Hibernated":
					if inst.HealthStatus == "Healthy" {
						ready++
					}
				}
			}
			log.Printf("[DEBUG] asgWarmPoolWaitUntilReady(%s, %s): ready=%d, minSize=%d", tenantID, asgName, ready, minSize)
			if ready >= minSize {
				return rp, "ready", nil
			}
			return rp, "pending", nil
		},
		PollInterval: 20 * time.Second,
		Timeout:      timeout,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// asgWarmPoolWaitUntilGone polls until the warm pool GET returns 404 (or an
// empty configuration / no lingering instances). AWS terminates warm pool
// instances asynchronously after the DELETE call, and a subsequent ASG delete
// will fail while those instances are still terminating, so we block until
// cleanup is complete.
func asgWarmPoolWaitUntilGone(ctx context.Context, c *duplosdk.Client, tenantID, asgName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"gone"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AsgWarmPoolGet(tenantID, asgName)
			if err != nil {
				if err.Status() == 404 {
					return struct{}{}, "gone", nil
				}
				return nil, "", err
			}
			if rp == nil || rp.WarmPoolConfiguration == nil || len(rp.Instances) == 0 {
				return rp, "gone", nil
			}
			log.Printf("[DEBUG] asgWarmPoolWaitUntilGone(%s, %s): %d instances still terminating", tenantID, asgName, len(rp.Instances))
			return rp, "pending", nil
		},
		PollInterval: 20 * time.Second,
		Timeout:      timeout,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
