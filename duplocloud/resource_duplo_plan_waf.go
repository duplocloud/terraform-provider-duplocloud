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
func resourcePlanWaf() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_waf` manages the list of waf's avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual waf's for a specific plan.",
		DeprecationMessage: "duplocloud_plan_waf is deprecated. Use duplocloud_plan_waf_v2 instead.",
		ReadContext:        resourcePlanWafRead,
		CreateContext:      resourcePlanWafCreateOrUpdate,
		UpdateContext:      resourcePlanWafCreateOrUpdate,
		DeleteContext:      resourcePlanWafDelete,
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
			"waf_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"waf_arn": {
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

func resourcePlanWafRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idSplit := strings.SplitN(id, "/", 2)

	planID, name := idSplit[0], idSplit[1]
	log.Printf("[TRACE] resourcePlanWafRead(%s): start", planID)
	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan configs.
	duplo, err := c.PlanWAFGet(planID, name)
	if err != nil {
		return diag.Errorf("failed to retrieve plan waf for '%s': %s", planID, err)
	}
	d.Set("waf_name", duplo.WebAclName)
	d.Set("waf_arn", duplo.WebAclId)
	d.Set("dashboard_url", duplo.DashboardUrl)
	log.Printf("[TRACE] resourcePlanWafRead(%s): end", planID)
	return nil
}

func resourcePlanWafCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanWafCreateOrUpdate(%s): start", planID)

	// Get all of the plan configs from duplo.
	c := m.(*duplosdk.Client)
	rq := duplosdk.DuploPlanWAF{
		WebAclName:   d.Get("waf_name").(string),
		WebAclId:     d.Get("waf_arn").(string),
		DashboardUrl: d.Get("dashboard_url").(string),
	}

	err := c.PlanWAF(planID, rq)
	if err != nil {
		return diag.Errorf("Error creating plan wafs for '%s': %s", planID, err)
	}
	id := planID + "/" + rq.WebAclName
	d.SetId(id)

	diags := resourcePlanWafRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanWafCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanWafDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idSplit := strings.SplitN(id, "/", 2)
	planID, name := idSplit[0], idSplit[1]
	log.Printf("[TRACE] resourcePlanWafDelete(%s): start", planID)

	// Get all of the plan configs from duplo.
	c := m.(*duplosdk.Client)
	err := c.PlanWafDelete(planID, name)
	if err != nil {
		return diag.Errorf("Error deleting plan waf	 for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanWafDelete(%s): end", planID)
	return nil
}
