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
func resourcePlanKMSV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms_v2` manages the list of kms avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan kms for a specific plan.",

		ReadContext:   resourcePlanKMSReadV2,
		CreateContext: resourcePlanKMSCreateOrUpdateV2,
		UpdateContext: resourcePlanKMSCreateOrUpdateV2,
		DeleteContext: resourcePlanKMSDeleteV2,
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
				Elem:        planKmsSchema(),
			},
			"kms_keys": {
				Description: "A list of KMS key to manage.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        planKmsSchema(),
			},
			"delete_unspecified_kms_keys": {
				Description: "Whether or not this resource should delete any certificates not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"specified_kms_keys": {
				Description: "A list of certificate names being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func planKmsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourcePlanKMSReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	planID := info[0]

	log.Printf("[TRACE] resourcePlanKMSRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanKMSGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan kmskeys for '%s': %s", planID, err)
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

		duplo = plan.KmsKeyInfos
	}

	// Set the simple fields first.
	d.Set("kms_keys", flattenPlanKmsKeysV2(duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_kms_keys"); ok && v != nil {
		d.Set("kms", flattenPlanKmsKeysV2(selectPlanKms(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanCertificatesRead(%s): end", planID)
	return nil

}

func flattenPlanKmsKeysV2(list *[]duplosdk.DuploPlanKmsKeyInfo) []interface{} {
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

//	func isKMSV2UpdateOrCreateable(list []duplosdk.DuploPlanKmsKeyInfo, reqs []duplosdk.DuploPlanKmsKeyInfo) (bool, string) {
//		nameMap := make(map[string]bool)
//		idMap := make(map[string]bool)
//		for _, val := range list {
//			nameMap[val.KeyName] = true
//			idMap[val.KeyId] = true
//		}
//
//		for _, rq := range reqs {
//			if nameMap[rq.KeyName] {
//				return false, "name : " + rq.KeyName
//			} else if idMap[rq.KeyId] {
//				return false, "id : " + rq.KeyName
//			}
//		}
//		return true, ""
//	}
func resourcePlanKMSCreateOrUpdateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanKMSCreateOrUpdate(%s): start", planID)

	c := m.(*duplosdk.Client)
	rp, _ := c.PlanKMSGetList(planID)
	previous, desired := getPlanKmsChange(rp, d)

	//if rp != nil {
	//	//	previous, desired := getPlanKmsChange(rp, d)
	//	createable, errStr := isKMSV2UpdateOrCreateable(*rp, *desired)
	//	if !createable {
	//		return diag.Errorf("Kms key with %s  already exist for plan %s", errStr, planID)
	//	}
	//}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_kms_keys").(bool) {
		err = c.PlanReplaceKmsKeys(planID, rp, desired)
	} else {
		err = c.PlanChangeKmsKeys(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}
	d.SetId(planID + "/kms")

	diags := resourcePlanKMSReadV2(ctx, d, m)
	log.Printf("[TRACE] resourcePlanCertificatesCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanKMSDeleteV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	planID := info[0]
	log.Printf("[TRACE] resourcePlanKMSDelete(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, clientErrr := c.PlanKMSGetList(planID)
	if clientErrr != nil {
		return diag.Errorf(clientErrr.Error())
	}

	// Get the previous and desired plan certificates
	previous, _ := getPlanKmsChange(all, d)
	desired := &[]duplosdk.DuploPlanKmsKeyInfo{}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_kms_keys").(bool) {
		err = c.PlanReplaceKmsKeys(planID, all, desired)
	} else {
		err = c.PlanChangeKmsKeys(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan kmskeys for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanKMSDelete(%s): end", planID)
	return nil
}

func expandPlanKmsV2(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanKmsKeyInfo {
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
	if v, ok := getAsStringArray(d, "specified_kms_keys"); ok && v != nil {
		previous = selectPlanKms(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanKmsKeyInfo{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanKmsV2("kms", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.KeyName
	}

	// Track the change
	d.Set("specified_kms_keys", specified)

	return
}

func selectPlanKmsFromMap(all *[]duplosdk.DuploPlanKmsKeyInfo, keys map[string]interface{}) *[]duplosdk.DuploPlanKmsKeyInfo {
	kmsKeys := make([]duplosdk.DuploPlanKmsKeyInfo, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.KeyName]; ok {
			kmsKeys = append(kmsKeys, pc)
		}
	}

	return &kmsKeys
}

func selectPlanKms(all *[]duplosdk.DuploPlanKmsKeyInfo, keys []string) *[]duplosdk.DuploPlanKmsKeyInfo {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanKmsFromMap(all, specified)
}
