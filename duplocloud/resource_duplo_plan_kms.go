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
			"kms": {
				Description: "A list of KMS key to manage.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        PlanKmsSchema(),
			},
			//"kms_keys": {
			//	Description: "A list of KMS key to manage.",
			//	Type:        schema.TypeList,
			//	Computed:    true,
			//	Elem:        PlanKmsSchema,
			//},
		},
	}
}

func PlanKmsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePlanKMSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	planID := info[0]

	log.Printf("[TRACE] resourcePlanKMSRead(%s): start", planID)
	c := m.(*duplosdk.Client)

	duplo, err := c.PlanKMSGetList(planID)
	if err != nil {
		return diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}

	if duplo == nil {
		d.SetId("")
		return nil
	}
	// Set the simple fields first.
	d.Set("kms", flattenPlanKms(duplo))
	// Build a list of current state, to replace the user-supplied settings.

	log.Printf("[TRACE] resourcePlanKMSRead(%s): end", planID)
	return nil
}
func flattenPlanKms(list *[]duplosdk.DuploPlanKmsKeyInfo) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, kms := range *list {
		result = append(result, map[string]interface{}{
			"name": kms.KeyName,
			"id":   kms.KeyId,
			"arn":  kms.KeyArn,
		})
	}

	return result
}

func isKMSUpdateOrCreateable(list []duplosdk.DuploPlanKmsKeyInfo, reqs []duplosdk.DuploPlanKmsKeyInfo) (bool, string) {
	nameMap := make(map[string]bool)
	idMap := make(map[string]bool)
	for _, val := range list {
		nameMap[val.KeyName] = true
		idMap[val.KeyId] = true
	}

	for _, rq := range reqs {
		if nameMap[rq.KeyName] {
			return false, "name : " + rq.KeyName
		} else if idMap[rq.KeyId] {
			return false, "id : " + rq.KeyName
		}
	}
	return true, ""
}
func resourcePlanKMSCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanKMSCreateOrUpdate(%s): start", planID)

	// Get all of the plan kms from duplo.
	desired := expandPlanKms("kms", d)
	c := m.(*duplosdk.Client)
	rp, _ := c.PlanKMSGetList(planID)
	if rp != nil {
		//	previous, desired := getPlanKmsChange(rp, d)
		createable, errStr := isKMSUpdateOrCreateable(*rp, *desired)
		if !createable {
			return diag.Errorf("Kms key with %s  already exist for plan %s", errStr, planID)
		}

	}
	clientErr := c.PlanCreateKMSKey(planID, *desired)
	if clientErr != nil {
		return diag.Errorf(clientErr.Error())
	}
	d.SetId(planID + "/kms")

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

func expandPlanKms(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanKmsKeyInfo {
	var ary []duplosdk.DuploPlanKmsKeyInfo

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandPlanKms ********: found %s", fieldName)
		ary = make([]duplosdk.DuploPlanKmsKeyInfo, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploPlanKmsKeyInfo{
				KeyId:   kv["id"].(string),
				KeyName: kv["name"].(string),
				KeyArn:  kv["arn"].(string),
			})
		}
	}

	return &ary
}

func getPlanKmsChange(all *[]duplosdk.DuploPlanKmsKeyInfo, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploPlanKmsKeyInfo) {
	if v, ok := getAsStringArray(d, "specified_kms"); ok && v != nil {
		previous = selectPlanKms(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanKmsKeyInfo{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanKms("kms", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.KeyName
	}

	// Track the change
	d.Set("specified_certificates", specified)

	return
}

func selectPlanKmsFromMap(all *[]duplosdk.DuploPlanKmsKeyInfo, keys map[string]interface{}) *[]duplosdk.DuploPlanKmsKeyInfo {
	certs := make([]duplosdk.DuploPlanKmsKeyInfo, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.KeyName]; ok {
			certs = append(certs, pc)
		}
	}

	return &certs
}

func selectPlanKms(all *[]duplosdk.DuploPlanKmsKeyInfo, keys []string) *[]duplosdk.DuploPlanKmsKeyInfo {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanKmsFromMap(all, specified)
}
