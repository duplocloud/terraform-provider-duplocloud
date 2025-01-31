package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource for managing an AWS ElasticSearch instance
func resourcePlanWafV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_waf_v2` manages the list of waf's avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual waf's for a specific plan.",

		ReadContext:   resourcePlanWafReadV2,
		CreateContext: resourcePlanWafCreateOrUpdateV2,
		UpdateContext: resourcePlanWafCreateOrUpdateV2,
		DeleteContext: resourcePlanWafDeleteV2,
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
				Description: "The ID of the plan for waf.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"waf": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     wafSchema(),
			},
			"delete_unspecified_wafs": {
				Description: "Whether or not this resource should delete any wafs not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wafs": {
				Description: "A complete list of wafs for this plan, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        wafSchema(),
			},
			"specified_wafs": {
				Description: "A list of wafs names being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func wafSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dashboard_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}

}
func resourcePlanWafReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idSplit := strings.SplitN(id, "/", 2)

	planID := idSplit[0]
	log.Printf("[TRACE] resourcePlanWafRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanWAFGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan certificates for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.WafInfos
	}

	// Set the simple fields first.
	d.Set("wafs", flattenPlanWafs(duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_wafs"); ok && v != nil {
		d.Set("waf", flattenPlanWafs(selectPlanWaf(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanCertificatesRead(%s): end", planID)
	return nil
}

func resourcePlanWafCreateOrUpdateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanWafCreateOrUpdate(%s): start", planID)

	// Get all of the plan configs from duplo.
	c := m.(*duplosdk.Client)

	all, clientError := c.PlanWAFGetList(planID)
	if clientError != nil {
		return diag.Errorf("%s", clientError.Error())
	}
	previous, desired := getPlanWafChange(all, d)
	if d.Get("delete_unspecified_wafs").(bool) {
		clientError = c.PlanReplaceWafs(planID, all, desired)
	} else {
		clientError = c.PlanChangeWafs(planID, previous, desired)
	}
	if clientError != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, clientError)
	}

	id := planID + "/waf"
	d.SetId(id)

	diags := resourcePlanWafReadV2(ctx, d, m)
	log.Printf("[TRACE] resourcePlanWafCreateOrUpdate(%s): end", planID)
	return diags
}

func expandWafV2(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanWAF {
	var ary []duplosdk.DuploPlanWAF

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		ary = make([]duplosdk.DuploPlanWAF, 0, len(kvs))

		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploPlanWAF{
				WebAclId:     kv["arn"].(string),
				WebAclName:   kv["name"].(string),
				DashboardUrl: kv["dashboard_url"].(string),
			})
		}
	}

	return &ary
}

func resourcePlanWafDeleteV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idSplit := strings.SplitN(id, "/", 2)
	planID := idSplit[0]
	log.Printf("[TRACE] resourcePlanWafDelete(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, err := c.PlanWAFGetList(planID)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}

	// Get the previous and desired plan certificates
	previous, _ := getPlanWafChange(all, d)
	desired := &[]duplosdk.DuploPlanWAF{}

	// Apply the changes via Duplo
	if d.Get("delete_unspecified_wafs").(bool) {
		err = c.PlanReplaceWafs(planID, all, desired)
	} else {
		err = c.PlanChangeWafs(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanCertificatesDelete(%s): end", planID)
	return nil
}

func getPlanWafChange(all *[]duplosdk.DuploPlanWAF, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploPlanWAF) {
	if v, ok := getAsStringArray(d, "specified_wafs"); ok && v != nil {
		previous = selectPlanWaf(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanWAF{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandWafV2("waf", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.WebAclName
	}

	// Track the change
	d.Set("specified_wafs", specified)

	return
}

func selectPlanWaf(all *[]duplosdk.DuploPlanWAF, keys []string) *[]duplosdk.DuploPlanWAF {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanWafFromMap(all, specified)
}

func selectPlanWafFromMap(all *[]duplosdk.DuploPlanWAF, keys map[string]interface{}) *[]duplosdk.DuploPlanWAF {
	certs := make([]duplosdk.DuploPlanWAF, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.WebAclName]; ok {
			certs = append(certs, pc)
		}
	}

	return &certs
}
