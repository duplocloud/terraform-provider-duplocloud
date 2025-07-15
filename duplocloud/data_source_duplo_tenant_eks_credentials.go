package duplocloud

import (
	"encoding/base64"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

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
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ resource
func dataSourceTenantEksCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantEksCredentialsRead ******** start")

	// Get the data from Duplo.
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	caCertificateData := ""

	// First, try the newer method of obtaining a JIT access token.
	k8sConfig, err := c.GetTenantK8sJitAccess(tenantID)
	if err != nil && !err.PossibleMissingAPI() {
		return fmt.Errorf("failed to get tenant %s kubernetes JIT access: %s", tenantID, err)
	}
	if k8sConfig != nil {
		bytes, err := base64.StdEncoding.DecodeString(k8sConfig.CertificateAuthorityDataBase64)
		if err == nil {
			caCertificateData = string(bytes)
		}

		// If it failed, try the fallback method.
	} else {
		k8sConfig, err = c.GetTenantK8sCredentials(tenantID)
		if err != nil {
			return fmt.Errorf("failed to read tenant %s kubernetes config: %s", tenantID, err)
		}
		k8sSecret, err := c.GetTenantEksSecret(tenantID)
		if err != nil {
			return fmt.Errorf("failed to read tenant %s EKS service account token: %s", tenantID, err)
		}

		k8sConfig.Token = k8sSecret.Data["token"]
		k8sConfig.DefaultNamespace = k8sSecret.Data["namespace"]
		caCertificateData = k8sSecret.Data["ca.crt"]
	}
	d.SetId(tenantID)

	// Set the Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("name", k8sConfig.Name)
	d.Set("endpoint", k8sConfig.APIServer)
	d.Set("region", k8sConfig.AwsRegion)
	d.Set("version", k8sConfig.K8sVersion)
	d.Set("token", k8sConfig.Token)
	d.Set("ca_certificate_data", caCertificateData)
	d.Set("namespace", k8sConfig.DefaultNamespace)

	log.Printf("[TRACE] dataSourceTenantEksCredentialsRead ******** end")
	return nil
}
