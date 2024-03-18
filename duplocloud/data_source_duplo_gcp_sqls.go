package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGCPCloudSQLs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_sql_cloud` retrieves a gcp sql instance in Duplo.",

		ReadContext: dataSourceGCPSQLCloudList,

		Schema: gcpSqlDBInstanceSchema(),
	}
}

func dataSourceGCPSQLCloudList(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGCPSQLCloudList ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	rp, clientErr := c.GCPSqlDBInstanceList(tenantID)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to retrieve tenant %s gcp sql databases : %s", tenantID, clientErr)
	}
	if rp == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.SetId(tenantID)
	list := make([]map[string]interface{}, 0, len(*rp))

	for _, duplo := range *rp {
		sql := map[string]interface{}{
			"name":             duplo.Name,
			"self_link":        duplo.SelfLink,
			"tier":             duplo.Tier,
			"database_version": reverseGcpSQLDBVersionsMap()[duplo.DatabaseVersion],
			"disk_size":        duplo.DataDiskSizeGb,
			"labels":           flattenStringMap(duplo.Labels),
		}
		list = append(list, sql)
	}
	d.Set("sqls", list)
	log.Printf("[TRACE] dataSourceGCPSQLCloudList ******** end")
	return nil

}
