package duplocloud

import (
	"context"
	"log"

	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planCertDataSourceSchema(single bool) map[string]*schema.Schema {

	// Create a fully computed schema.
	certs_schema := planCertSchema()
	for k := range certs_schema {
		certs_schema[k].Required = false
		certs_schema[k].Computed = true
	}

	// For a single certifiate, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = certs_schema
		result["name"].Computed = false
		result["name"].Required = true

		// For a list of certificates, move the list under the result key.
	} else {
		result = map[string]*schema.Schema{
			"certificates": {
				Description: "The list of certificates for this plan.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: certs_schema,
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

func dataSourcePlanCerts() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_certificates` retrieves a list of cerificates for a given plan.",

		ReadContext: dataSourcePlanCertsRead,
		Schema:      planCertDataSourceSchema(false),
	}
}

func dataSourcePlanCert() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_certificate` retrieves details of a specific certificate for a given plan.",

		ReadContext: dataSourcePlanCertRead,
		Schema:      planCertDataSourceSchema(true),
	}
}

func dataSourcePlanCertsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanCertsRead(%s): start", planID)

	// Get all of the plan certificates from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanCerts(c, planID)
	if diags != nil {
		return diags
	}
	// Populate the results from the list.
	d.Set("certificates", flattenPlanCerts(all))
	d.Set("plan_id", planID)
	d.SetId(planID)

	log.Printf("[TRACE] dataSourcePlanCertsRead(%s): end", planID)
	return nil
}

// READ/SEARCH resources
func dataSourcePlanCertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourcePlanCertRead(%s, %s): start", planID, name)

	// Get the plan certificate from Duplo.
	c := m.(*duplosdk.Client)
	cert, diags := getPlanCert(c, planID, name)
	if diags != nil {
		return diags
	}
	d.SetId(cert.CertificateArn)
	d.Set("name", cert.CertificateName)
	d.Set("arn", cert.CertificateArn)
	d.Set("plan_id", planID)

	log.Printf("[TRACE] dataSourcePlanCertRead(): end")
	return nil
}

func getPlanCert(c *duplosdk.Client, planID, name string) (*duplosdk.DuploPlanCertificate, diag.Diagnostics) {
	rsp, err := c.PlanCertificateGet(planID, name)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan certificate for '%s/%s': %s", planID, name, err)
	}

	// If it failed, try the fallback method.
	if rsp == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan certificates: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		if plan.Certificates != nil {
			for _, v := range *plan.Certificates {
				if v.CertificateName == name {
					rsp = &v
				}
			}
		}
	}

	if rsp == nil {
		return nil, diag.Errorf("failed to retrieve plan certificate for '%s/%s': %s", planID, name, err)
	}
	return rsp, nil
}

func getPlanCerts(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanCertificate, diag.Diagnostics) {
	resp, err := c.PlanCertificateGetList(planID)
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

		resp = plan.Certificates
	}

	return resp, nil
}

func planCertSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A domain name for which the certificate should be issued",
		},
		"arn": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ARN of the certificate",
		},
	}
}

func flattenPlanCerts(list *[]duplosdk.DuploPlanCertificate) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, cert := range *list {
		result = append(result, map[string]interface{}{
			"name": cert.CertificateName,
			"arn":  cert.CertificateArn,
		})
	}
	return result
}
