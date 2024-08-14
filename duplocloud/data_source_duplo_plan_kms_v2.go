package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planKmsDataSourceSchema(single bool) map[string]*schema.Schema {

	// Create a fully computed schema.
	kms_schema := planKmsDataSchema()
	for k := range kms_schema {
		kms_schema[k].Required = false
		kms_schema[k].Computed = true
	}

	// For a single certifiate, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = kms_schema
		result["name"].Computed = false
		result["name"].Required = true

		// For a list of certificates, move the list under the result key.
	} else {
		result = map[string]*schema.Schema{
			"kms_keys": {
				Description: "The list of kms keys for this plan.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: kms_schema,
				},
			},
		}

	}

	// Always require the plan ID.
	result["plan_id"] = &schema.Schema{
		Description: "The plan ID",
		Type:        schema.TypeString,
		Required:    true,
	}

	return result
}

func dataSourcePlanKMSListV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms_key` retrieves a list of kms keys for a given plan.",

		ReadContext: dataSourcePlanKmsKeysReadV2,
		Schema:      planKmsDataSourceSchema(false),
	}
}

func dataSourcePlanKMSV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_kms` retrieves details of a specific kms for a given plan.",

		ReadContext: dataSourcePlanKmsReadV2,
		Schema:      planKmsDataSourceSchema(true),
	}
}

func dataSourcePlanKmsKeysReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanKmsKeysRead(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, err := getPlanKmsKeys(c, planID)
	if err != nil {
		return err
	}
	// Populate the results from the list.
	_ = d.Set("kms_keys", flattenPlanKmsKeysV2(all))

	d.SetId(planID + "/kms")

	log.Printf("[TRACE] dataSourcePlanKmsKeysRead(%s): end", planID)
	return nil
}

// READ/SEARCH resources
func dataSourcePlanKmsReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourcePlanKmsRead(%s, %s): start", planID, name)

	// Get the plan certificate from Duplo.
	c := m.(*duplosdk.Client)
	kms, diags := getPlanKms(c, planID, name)
	if diags != nil {
		return diags
	}
	d.SetId(planID + "/kms/" + kms.KeyName)
	_ = d.Set("name", kms.KeyName)
	_ = d.Set("id", kms.KeyId)
	_ = d.Set("arn", kms.KeyArn)

	log.Printf("[TRACE] dataSourcePlanKmsRead(): end")
	return nil
}

func getPlanKms(c *duplosdk.Client, planID, name string) (*duplosdk.DuploPlanKmsKeyInfo, diag.Diagnostics) {
	rsp, err := c.PlanGetKMSKey(planID, name)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan certificate for '%s/%s': %s", planID, name, err)
	}

	// If it failed, try the fallback method.
	if rsp == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan kms info: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		if plan.KmsKeyInfos != nil {
			for _, v := range *plan.KmsKeyInfos {
				if v.KeyName == name {
					rsp = &v
				}
			}
		}
	}

	if rsp == nil {
		return nil, diag.Errorf("failed to retrieve plan kms key for '%s/%s': %s", planID, name, err)
	}
	return rsp, nil
}

func getPlanKmsKeys(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanKmsKeyInfo, diag.Diagnostics) {
	resp, err := c.PlanKMSGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan certificates for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if resp == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		resp = plan.KmsKeyInfos
	}

	return resp, nil
}

func planKmsDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}
