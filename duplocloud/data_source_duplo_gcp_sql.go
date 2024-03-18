package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGCPCloudSQL() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_sql_cloud` retrieves a gcp sql instance in Duplo.",

		ReadContext: dataSourceGCPSQLCloudRead,

		Schema: gcpSqlDBInstanceSchema(),
	}
}

func dataSourceGCPSQLCloudRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGCPSQLCloudRead ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	c := m.(*duplosdk.Client)

	fullName, clientErr := c.GetDuploServicesName(tenantID, name)
	if clientErr != nil {
		return diag.Errorf("error fetching tenant prefix for %s : %s", tenantID, clientErr)
	}
	duplo, clientErr := c.GCPSqlDBInstanceGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve tenant %s gcp sql database '%s': %s", tenantID, fullName, clientErr)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	flattenGcpSqlDBInstance(d, tenantID, name, duplo)
	log.Printf("[TRACE] dataSourceGCPSQLCloudRead ******** end")
	return nil

}
