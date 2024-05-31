package duplocloud

import (
	"context"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PlanKmsSchema returns a Terraform schema to represent a plan KMS

// Resource for managing an AWS ElasticSearch instance
func resourcePlanKMS() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms` manages the list of kms avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan kms for a specific plan.",

		ReadContext:   resourcePlanKMSRead,
		CreateContext: resourcePlanKMSCreateOrUpdate,
		UpdateContext: resourcePlanKMSCreateOrUpdate,
		DeleteContext: resourcePlanKMSDelete,
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
			"kms_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourcePlanKMSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	planID := info[0]
	name := info[2]

	log.Printf("[TRACE] resourcePlanKMSRead(%s): start", planID)
	c := m.(*duplosdk.Client)

	duplo, err := c.PlanGetKMSKey(planID, name)
	if err != nil || !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}

	// Set the simple fields first.
	d.Set("kms_id", duplo.KeyId)
	d.Set("kms_name", duplo.KeyName)
	d.Set("kms_arn", duplo.KeyArn)
	// Build a list of current state, to replace the user-supplied settings.

	log.Printf("[TRACE] resourcePlanKMSRead(%s): end", planID)
	return nil
}

func isKMSUpdateOrCreateable(list []duplosdk.DuploPlanKmsKeyInfo, rq duplosdk.DuploPlanKmsKeyInfo) bool {
	nameMap := make(map[string]bool)
	idMap := make(map[string]bool)
	for _, val := range list {
		nameMap[val.KeyName] = true
		idMap[val.KeyId] = true
	}
	if nameMap[rq.KeyName] {
		return false
	} else if idMap[rq.KeyId] {
		return false
	}
	return true
}
func resourcePlanKMSCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanKMSCreateOrUpdate(%s): start", planID)

	rq := duplosdk.DuploPlanKmsKeyInfo{
		KeyName: d.Get("kms_name").(string),
		KeyId:   d.Get("kms_id").(string),
		KeyArn:  d.Get("kms_arn").(string),
	}
	// Get all of the plan kms from duplo.
	c := m.(*duplosdk.Client)
	rp, _ := c.PlanKMSGetList(planID)
	if !isKMSUpdateOrCreateable(*rp, rq) {
		return diag.Errorf("Kms key with name %s or id %s already exist for plan %s", rq.KeyName, rq.KeyId, planID)
	}
	clientErr := c.PlanCreateKMSKey(planID, rq)
	if clientErr != nil {
		return diag.Errorf(clientErr.Error())
	}
	d.SetId(planID + "/kms/" + rq.KeyName)

	diags := resourcePlanKMSRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanKMSCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanKMSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	planID := info[0]
	name := info[2]

	c := m.(*duplosdk.Client)
	clientErr := c.PlanKMSDelete(planID, name)
	if clientErr != nil {
		return diag.Errorf(clientErr.Error())
	}

	return nil
}
