package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceTenantEksCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantEksCredentialsRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificate_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func dataSourceTenantEksCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantEksCredentialsRead ******** start")

	// Get the data from Duplo.
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	k8sCredentials, err := c.GetTenantK8sCredentials(tenantID)
	if err != nil {
		return fmt.Errorf("Failed to read tenant %s kubernetes credentials: %s", tenantID, err)
	}
	eksSecret, err := c.GetTenantEksSecret(tenantID)
	if err != nil {
		return fmt.Errorf("Failed to read tenant %s EKS CA certificate data: %s", tenantID, err)
	}
	d.SetId(tenantID)

	// Set the Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("name", k8sCredentials.Name)
	d.Set("endpoint", k8sCredentials.APIServer)
	d.Set("token", k8sCredentials.Token)
	d.Set("region", k8sCredentials.AwsRegion)
	d.Set("ca_certificate_data", eksSecret.Data["ca.crt"])

	log.Printf("[TRACE] dataSourceTenantEksCredentialsRead ******** end")
	return nil
}
