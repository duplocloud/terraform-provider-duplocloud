package duplocloud

import (
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceAdminAwsCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAdminAwsCredentialsRead,

		Schema: map[string]*schema.Schema{
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

// READ resource
func dataSourceAdminAwsCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceAdminAwsCredentialsRead ******** start")

	// Get the region from Duplo.
	c := m.(*duplosdk.Client)
	creds, err := c.AdminGetAwsCredentials()
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		return fmt.Errorf("failed to read admin AWS credentials: %s", err)
	}

	// Set the Terraform resource data
	d.Set("console_url", creds.ConsoleURL)
	d.Set("access_key_id", creds.AccessKeyID)
	d.Set("secret_access_key", creds.SecretAccessKey)
	d.Set("session_token", creds.SessionToken)
	d.Set("region", creds.Region)

	log.Printf("[TRACE] dataSourceAdminAwsCredentialsRead ******** end")
	return nil
}
