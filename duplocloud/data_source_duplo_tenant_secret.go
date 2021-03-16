package duplocloud

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for secrets
func tenantSecretSchemaComputed() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
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
	}
}

// Data source retrieving a secret
func dataSourceTenantSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantSecretRead,

		Schema: tenantSecretSchemaComputed(),
	}
}

/// READ resource
func dataSourceTenantSecretRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantSecretRead ******** start")

	// Get and validate the retrieval criteria.
	var arn, name, nameSuffix, secretID string
	tenantID := d.Get("tenant_id").(string)
	if v, ok := d.GetOk("arn"); ok {
		arn = v.(string)
		secretID = arn // for error reporting
	}
	if v, ok := d.GetOk("name"); ok {
		if arn != "" {
			return errors.New("specify only arn or name or name_suffix")
		}
		name = v.(string)
		secretID = name // for error reporting
	}
	if v, ok := d.GetOk("name_suffix"); ok {
		if arn != "" || name != "" {
			return errors.New("specify only arn or name or name_suffix")
		}
		nameSuffix = v.(string)
		secretID = nameSuffix // for error reporting
	}
	if arn == "" && name == "" && nameSuffix == "" {
		return errors.New("must specify either arn or name or name_suffix")
	}

	// List the secrets from Duplo.
	c := m.(*duplosdk.Client)
	duploSecrets, err := c.TenantListSecrets(tenantID)
	if err != nil {
		return fmt.Errorf("Failed to list secrets: %s", err)
	}

	// Set the Terraform resource data
	for _, duploSecret := range *duploSecrets {
		parts := strings.SplitN(duploSecret.Name, "-", 3)
		objNameSuffix := duploSecret.Name
		if len(parts) == 3 {
			objNameSuffix = parts[2]
		}

		if (arn != "" && duploSecret.Arn == arn) || (name != "" && duploSecret.Name == name) || (nameSuffix != "" && objNameSuffix == nameSuffix) {
			d.SetId(fmt.Sprintf("%s/%s", tenantID, arn))
			d.Set("tenant_id", tenantID)
			d.Set("arn", duploSecret.Arn)
			d.Set("name", duploSecret.Name)
			d.Set("name_suffix", objNameSuffix)
			d.Set("rotation_enabled", duploSecret.RotationEnabled)
			d.Set("tags", duplosdk.KeyValueToState("tags", duploSecret.Tags))
			break
		}
	}

	// Check for missing result
	if d.Id() == "" {
		return fmt.Errorf("Tenant secret '%s' not found", secretID)
	}

	log.Printf("[TRACE] dataSourceTenantSecretRead ******** end")
	return nil
}
