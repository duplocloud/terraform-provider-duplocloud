package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceAwsAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAccountRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func dataSourceAwsAccountRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceAwsAccountRead ******** start")

	// Get the region from Duplo.
	c := m.(*duplosdk.Client)
	awsAccount, err := c.GetAwsAccountID()
	if err != nil {
		return fmt.Errorf("Failed to read AWS account ID: %s", err)
	}
	d.SetId(awsAccount)

	// Set the Terraform resource data
	d.Set("account_id", awsAccount)

	log.Printf("[TRACE] dataSourceAwsAccountRead ******** end")
	return nil
}
