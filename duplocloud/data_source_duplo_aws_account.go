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
func dataSourceAwsAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAccountRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func dataSourceAwsAccountRead(d *schema.ResourceData, m interface{}) error {
	var err error
	var awsAccountID string

	log.Printf("[TRACE] dataSourceAwsAccountRead ******** start")

	// Get the account ID from Duplo.
	c := m.(*duplosdk.Client)
	if v, ok := d.GetOk("tenant_id"); ok && v != nil && v.(string) != "" {
		tenantID := v.(string)
		awsAccountID, err = c.TenantGetAwsAccountID(tenantID)
		d.SetId(fmt.Sprintf("%s-%s", tenantID, strconv.FormatInt(time.Now().Unix(), 10)))
	} else {
		awsAccountID, err = c.GetAwsAccountID()
		d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	}
	if err != nil {
		return fmt.Errorf("Failed to read AWS account ID: %s", err)
	}

	// Set the Terraform resource data
	d.Set("account_id", awsAccountID)

	log.Printf("[TRACE] dataSourceAwsAccountRead ******** end")
	return nil
}
