package duplocloud

/*
import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PlanCertificateSchema returns a Terraform schema to represent a plan certificate
func PlanKMSSchema() *schema.Resource {
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

// Resource for managing an AWS ElasticSearch instance
func resourcePlanKMS() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_certificates` manages the list of certificates avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan certificates for a specific plan.",

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
				Description: "A list of certificates to manage.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        PlanCertificateSchema(),
			},
			"delete_unspecified_kms": {
				Description: "Whether or not this resource should delete any certificates not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_s": {
				Description: "A complete list of certificates for this plan, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        PlanCertificateSchema(),
			},
			"specified_kms": {
				Description: "A list of certificate names being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourcePlanKMSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanKMSRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanKMSGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return diag.Errorf("failed to read plan kms: %s", err)
		}
		if plan == nil {
			return diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.KmsKeyInfos
	}

	// Set the simple fields first.
	d.Set("kms_s", flattenPlanKmsKeys(duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_kms"); ok && v != nil {
		d.Set("kms", flattenPlanKmsKeys(selectPlanKMS(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanCertificatesRead(%s): end", planID)
	return nil
}

func resourcePlanKMSCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanCertificatesCreateOrUpdate(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanKms(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan certificates
	previous, desired := getPlanKMSChange(all, d)

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_kms").(bool) {
		err = c.PlanReplaceKMS(planID, desired)
	} else {
		err = c.PlanChangeKMS(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}
	d.SetId(planID)

	diags = resourcePlanCertificatesRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanCertificatesCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanKMSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanCertificatesDelete(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanKms(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan certificates
	previous, _ := getPlanKMSChange(all, d)
	desired := &[]duplosdk.DuploPlanKmsKeyInfo{}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_certificates").(bool) {
		err = c.PlanReplaceKMS(planID, desired)
	} else {
		err = c.PlanChangeKMS(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanCertificatesDelete(%s): end", planID)
	return nil
}

// Utiliy function to return a filtered list of plan certificates, given the selected keys.
func selectPlanKMSFromMap(all *[]duplosdk.DuploPlanKmsKeyInfo, keys map[string]interface{}) *[]duplosdk.DuploPlanKmsKeyInfo {
	kms := make([]duplosdk.DuploPlanKmsKeyInfo, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.KeyName]; ok {
			kms = append(kms, pc)
		}
	}

	return &kms
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectPlanKMS(all *[]duplosdk.DuploPlanKmsKeyInfo, keys []string) *[]duplosdk.DuploPlanKmsKeyInfo {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanKMSFromMap(all, specified)
}

func expandPlanKMS(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanKmsKeyInfo {
	var ary []duplosdk.DuploPlanKmsKeyInfo

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandPlanCertificates ********: found %s", fieldName)
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

func getPlanKms(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanKmsKeyInfo, diag.Diagnostics) {

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanKMSGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan kms for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan kms: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.KmsKeyInfos
	}

	return duplo, nil
}

func getPlanKMSChange(all *[]duplosdk.DuploPlanKmsKeyInfo, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploPlanKmsKeyInfo) {
	if v, ok := getAsStringArray(d, "specified_kms"); ok && v != nil {
		previous = selectPlanKMS(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanKmsKeyInfo{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanKMS("kms", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.KeyName
	}

	// Track the change
	d.Set("specified_certificates", specified)

	return
}
*/
