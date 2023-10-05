package resources

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource for managing an AWS ElasticSearch instance
func resourcePlanConfigs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_configs` manages the list of configs avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual configs for a specific plan.",

		ReadContext:   resourcePlanConfigsRead,
		CreateContext: resourcePlanConfigsCreateOrUpdate,
		UpdateContext: resourcePlanConfigsCreateOrUpdate,
		DeleteContext: resourcePlanConfigsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan to configure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"config": {
				Description: "A list of configs to manage.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        duplocloud.CustomDataExSchema(),
			},
			"delete_unspecified_configs": {
				Description: "Whether or not this resource should delete any configs not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"configs": {
				Description: "A complete list of configs for this plan, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        duplocloud.CustomDataExSchema(),
			},
			"specified_configs": {
				Description: "A list of config keys being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourcePlanConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanConfigsRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan configs.
	duplo, err := c.PlanConfigGetList(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan configs for '%s': %s", planID, err)
	}

	// Set the simple fields first.
	d.Set("configs", duplocloud.customDataExToState("configs", duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := duplocloud.getAsStringArray(d, "specified_configs"); ok && v != nil {
		d.Set("config", duplocloud.customDataExToState("configs", selectPlanConfigs(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanConfigsRead(%s): end", planID)
	return nil
}

func resourcePlanConfigsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanConfigsCreateOrUpdate(%s): start", planID)

	// Get all of the plan configs from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanConfigs(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan configs
	previous, desired := getPlanConfigsChange(all, d)

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_configs").(bool) {
		err = c.PlanReplaceConfigs(planID, desired)
	} else {
		err = c.PlanChangeConfigs(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan configs for '%s': %s", planID, err)
	}
	d.SetId(planID)

	diags = resourcePlanConfigsRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanConfigsCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanConfigsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanConfigsDelete(%s): start", planID)

	// Get all of the plan configs from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanConfigs(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan configs
	previous, _ := getPlanConfigsChange(all, d)
	desired := &[]duplosdk.DuploCustomDataEx{}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_configs").(bool) {
		err = c.PlanReplaceConfigs(planID, desired)
	} else {
		err = c.PlanChangeConfigs(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan configs for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanConfigsDelete(%s): end", planID)
	return nil
}

// Utiliy function to return a filtered list of plan configs, given the selected keys.
func selectPlanConfigsFromMap(all *[]duplosdk.DuploCustomDataEx, keys map[string]interface{}) *[]duplosdk.DuploCustomDataEx {
	certs := make([]duplosdk.DuploCustomDataEx, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.Key]; ok {
			certs = append(certs, pc)
		}
	}

	return &certs
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectPlanConfigs(all *[]duplosdk.DuploCustomDataEx, keys []string) *[]duplosdk.DuploCustomDataEx {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanConfigsFromMap(all, specified)
}

func expandPlanConfigs(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploCustomDataEx {
	var ary []duplosdk.DuploCustomDataEx

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandPlanConfigs ********: found %s", fieldName)
		ary = make([]duplosdk.DuploCustomDataEx, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploCustomDataEx{
				Key:   kv["key"].(string),
				Type:  kv["type"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

func getPlanConfigs(c *duplosdk.Client, planID string) (*[]duplosdk.DuploCustomDataEx, diag.Diagnostics) {
	duplo, err := c.PlanConfigGetList(planID)
	if err != nil {
		return nil, diag.Errorf("failed to retrieve plan configs for '%s': %s", planID, err)
	}
	return duplo, nil
}

func getPlanConfigsChange(all *[]duplosdk.DuploCustomDataEx, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploCustomDataEx) {
	if v, ok := duplocloud.getAsStringArray(d, "specified_configs"); ok && v != nil {
		previous = selectPlanConfigs(all, *v)
	} else {
		previous = &[]duplosdk.DuploCustomDataEx{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanConfigs("config", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.Key
	}

	// Track the change
	d.Set("specified_configs", specified)

	return
}
