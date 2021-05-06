package duplocloud

import (
	"fmt"
	"log"
	"strconv"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceTenantAwsCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTenantAwsCredentialsRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"console_url": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"access_key_id": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secret_access_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"session_token": {
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
func dataSourceTenantAwsCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceTenantAwsCredentialsRead ******** start")

	// Get the region from Duplo.
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	creds, err := c.TenantGetAwsCredentials(tenantID)
	d.SetId(fmt.Sprintf("%s-%s", tenantID, strconv.FormatInt(time.Now().Unix(), 10)))
	if err != nil {
		return fmt.Errorf("failed to read AWS credentials from tenant '%s': %s", tenantID, err)
	}

	// Set the Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("console_url", creds.ConsoleURL)
	d.Set("access_key_id", creds.AccessKeyID)
	d.Set("secret_access_key", creds.SecretAccessKey)
	d.Set("session_token", creds.SessionToken)
	d.Set("region", creds.Region)

	log.Printf("[TRACE] dataSourceTenantAwsCredentialsRead ******** end")
	return nil
}
