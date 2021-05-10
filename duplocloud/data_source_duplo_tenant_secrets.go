package duplocloud

import (
	"fmt"
	"log"
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
					Schema: map[string]*schema.Schema{
						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"name_suffix": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"rotation_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Required: false,
							Elem:     KeyValueSchema(),
						},
					},
				},
			},
		},
	}
}

/// READ resource
func dataSourceTenantSecretsRead(d *schema.ResourceData, m interface{}) error {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceTenantSecretsRead(%s): start", tenantID)

	// List the secrets from Duplo.
	c := m.(*duplosdk.Client)
	duploSecrets, err := c.TenantListSecrets(tenantID)
	if err != nil {
		return fmt.Errorf("failed to list secrets: %s", err)
	}
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant prefix: %s", err)
	}
	d.SetId(tenantID)

	// Set the Terraform resource data
	secrets := make([]map[string]interface{}, 0, len(*duploSecrets))
	for _, duploSecret := range *duploSecrets {
		nameSuffix, _ := duplosdk.UnprefixName(prefix, duploSecret.Name)

		secrets = append(secrets, map[string]interface{}{
			"tenant_id":        tenantID,
			"arn":              duploSecret.Arn,
			"name":             duploSecret.Name,
			"name_suffix":      nameSuffix,
			"rotation_enabled": duploSecret.RotationEnabled,
			"tags":             keyValueToState("tags", duploSecret.Tags),
		})
	}

	d.Set("secrets", secrets)

	log.Printf("[TRACE] dataSourceTenantSecretsRead(%s): end", tenantID)
	return nil
}
