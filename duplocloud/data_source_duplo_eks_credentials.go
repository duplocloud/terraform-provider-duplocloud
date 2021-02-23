package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceEksCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEksCredentialsRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
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
func dataSourceEksCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceEksCredentialsRead ******** start")

	// Get the data from Duplo.
	planID := d.Get("plan_id").(string)
	c := m.(*duplosdk.Client)
	eksCredentials, err := c.GetEksCredentials(planID)
	if err != nil {
		return fmt.Errorf("Failed to read AWS account ID: %s", err)
	}
	d.SetId(eksCredentials.Name)

	// Set the Terraform resource data
	d.Set("plan_id", planID)
	d.Set("name", eksCredentials.Name)
	d.Set("endpoint", eksCredentials.APIServer)
	d.Set("token", eksCredentials.Token)
	d.Set("region", eksCredentials.AwsRegion)

	log.Printf("[TRACE] dataSourceEksCredentialsRead ******** end")
	return nil
}
