package duplocloud

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source listing secrets
func dataSourceTenantSecrets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantSecretsRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secrets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: tenantSecretSchemaComputed(),
				},
			},
		},
	}
}

/// READ resource
func dataSourceTenantSecretsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantSecretsRead ******** start")

	// List the secrets from Duplo.
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	duploSecrets, err := c.TenantListSecrets(tenantID)
	if err != nil {
		return fmt.Errorf("Failed to list secrets: %s", err)
	}
	d.SetId(tenantID)

	// Set the Terraform resource data
	secrets := make([]map[string]interface{}, 0, len(*duploSecrets))
	for _, duploSecret := range *duploSecrets {
		parts := strings.SplitN(duploSecret.Name, "-", 3)
		nameSuffix := duploSecret.Name
		if len(parts) == 3 {
			nameSuffix = parts[2]
		}

		secrets = append(secrets, map[string]interface{}{
			"tenant_id":        tenantID,
			"arn":              duploSecret.Arn,
			"name":             duploSecret.Name,
			"name_suffix":      nameSuffix,
			"rotation_enabled": duploSecret.RotationEnabled,
			"tags":             duplosdk.KeyValueToState("tags", duploSecret.Tags),
		})
	}

	d.Set("secrets", secrets)

	log.Printf("[TRACE] dataSourceTenantSecretsRead ******** end")
	return nil
}
