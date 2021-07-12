package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// PlanCertificateSchema returns a Terraform schema to represent a plan certificate
func PlanCertificateSchema() *schema.Resource {
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
				Computed: true,
			},
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourcePlanCertificates() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_certificates` manages the list of certificates avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual plan certificates for a specific plan.",

		ReadContext:   resourcePlanCertificatesRead,
		CreateContext: resourcePlanCertificatesCreateOrUpdate,
		UpdateContext: resourcePlanCertificatesCreateOrUpdate,
		DeleteContext: resourcePlanCertificatesDelete,
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
			"certificate": {
				Description: "A list of certificates to manage, expressed as key / value pairs.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        PlanCertificateSchema(),
			},
			"delete_unspecified_certificates": {
				Description: "Whether or not this resource should delete any certificates not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"certificates": {
				Description: "A complete list of certificates for this plan, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        PlanCertificateSchema(),
			},
			"specified_certificates": {
				Description: "A list of certificate names being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourcePlanCertificatesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanCertificatesRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanCertificateGetList(planID)
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

		duplo = plan.Certificates
	}

	// Set the simple fields first.
	d.Set("certificates", flattenPlanCertificates(duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_certificates"); ok && v != nil {
		d.Set("certificate", flattenPlanCertificates(selectPlanCertificates(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanCertificatesRead(%s): end", planID)
	return nil
}

func resourcePlanCertificatesCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanCertificatesCreateOrUpdate(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanCertificates(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan certificates
	previous, desired := getPlanCertificatesChange(all, d)

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_certificates").(bool) {
		err = c.PlanReplaceCertificates(planID, desired)
	} else {
		err = c.PlanChangeCertificates(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}
	d.SetId(planID)

	diags = resourcePlanCertificatesRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanCertificatesCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanCertificatesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanCertificatesDelete(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanCertificates(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan certificates
	previous, _ := getPlanCertificatesChange(all, d)
	desired := &[]duplosdk.DuploPlanCertificate{}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_certificates").(bool) {
		err = c.PlanReplaceCertificates(planID, desired)
	} else {
		err = c.PlanChangeCertificates(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan certificates for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanCertificatesDelete(%s): end", planID)
	return nil
}

// Utiliy function to return a filtered list of plan certificates, given the selected keys.
func selectPlanCertificatesFromMap(all *[]duplosdk.DuploPlanCertificate, keys map[string]interface{}) *[]duplosdk.DuploPlanCertificate {
	certs := make([]duplosdk.DuploPlanCertificate, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.CertificateName]; ok {
			certs = append(certs, pc)
		}
	}

	return &certs
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectPlanCertificates(all *[]duplosdk.DuploPlanCertificate, keys []string) *[]duplosdk.DuploPlanCertificate {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanCertificatesFromMap(all, specified)
}

func expandPlanCertificates(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanCertificate {
	var ary []duplosdk.DuploPlanCertificate

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandPlanCertificates ********: found %s", fieldName)
		ary = make([]duplosdk.DuploPlanCertificate, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploPlanCertificate{
				CertificateArn:  kv["id"].(string),
				CertificateName: kv["name"].(string),
			})
		}
	}

	return &ary
}

func getPlanCertificates(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanCertificate, diag.Diagnostics) {

	// First, try the newer method of getting the plan certificates.
	duplo, err := c.PlanCertificateGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan certificates for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.Certificates
	}

	return duplo, nil
}

func getPlanCertificatesChange(all *[]duplosdk.DuploPlanCertificate, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploPlanCertificate) {
	if v, ok := getAsStringArray(d, "specified_certificates"); ok && v != nil {
		previous = selectPlanCertificates(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanCertificate{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanCertificates("certificate", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.CertificateName
	}

	// Track the change
	d.Set("specified_certificates", specified)

	return
}
