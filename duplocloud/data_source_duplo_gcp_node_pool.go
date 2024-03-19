package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGCPNodePool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGCPNodePoolRead,
		Schema:      dataGcpK8NodePoolFunctionSchema(),
	}
}

func dataSourceGCPNodePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGCPNodePoolRead ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)

	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("error fetching detail name required ")

	}
	c := m.(*duplosdk.Client)

	fullName, clientErr := c.GetDuploServicesName(tenantID, name)
	if clientErr != nil {
		return diag.Errorf("error fetching tenant prefix for %s : %s", tenantID, clientErr)
	}
	duplo, err := c.GCPK8NodePoolGet(tenantID, fullName)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s GCP Node Pool Domain '%s': %s", tenantID, fullName, err)
	}

	if duplo == nil {
		d.SetId("") // object missing or deleted
		return nil
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	setGCPNodePoolStateField(d, duplo, tenantID)

	log.Printf("[TRACE] dataSourceGCPNodePoolRead ******** end")
	return nil
}
