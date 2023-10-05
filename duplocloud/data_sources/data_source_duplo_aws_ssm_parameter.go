package data_sources

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Data source retrieving an SSM parameter
func dataSourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSsmParameterRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allowed_pattern": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ resource
func dataSourceSsmParameterRead(d *schema.ResourceData, m interface{}) error {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] dataSourceSsmParameterRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	c := m.(*duplosdk.Client)
	ssmParam, err := c.SsmParameterGet(tenantID, name)
	if err != nil {
		return fmt.Errorf("failed to get SSM parameter: %s: %s", id, err)
	}

	// Check for missing result
	if ssmParam == nil {
		return fmt.Errorf("SSM parameter '%s' not found", id)
	}

	d.SetId(id)
	d.Set("type", ssmParam.Type)
	d.Set("value", ssmParam.Value)
	d.Set("key_id", ssmParam.KeyId)
	d.Set("description", ssmParam.Description)
	d.Set("allowed_pattern", ssmParam.AllowedPattern)
	d.Set("last_modified_user", ssmParam.LastModifiedUser)
	d.Set("last_modified_date", ssmParam.LastModifiedDate)

	log.Printf("[TRACE] dataSourceSsmParameterRead(%s, %s): end", tenantID, name)
	return nil
}
