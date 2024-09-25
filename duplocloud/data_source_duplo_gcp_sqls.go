package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceGCPCloudSQLs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_sql_database_instances` retrieves a gcp sql instance in Duplo.",

		ReadContext: dataSourceGCPSQLCloudList,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the sql database will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"databases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The name of the sql database.",
							Type:        schema.TypeString,
							Computed:    true,
						},

						"self_link": {
							Description: "The SelfLink of the sql database.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"database_version": {
							Description: "The MySQL, PostgreSQL orSQL Server version to use." +
								"Supported values include `MYSQL_5_6`,`MYSQL_5_7`, `MYSQL_8_0`, `POSTGRES_9_6`,`POSTGRES_10`," +
								"`POSTGRES_11`,`POSTGRES_12`, `POSTGRES_13`, `POSTGRES_14`, `POSTGRES_15`, `SQLSERVER_2017_STANDARD`,`SQLSERVER_2017_ENTERPRISE`," +
								"`SQLSERVER_2017_EXPRESS`, `SQLSERVER_2017_WEB`.`SQLSERVER_2019_STANDARD`, `SQLSERVER_2019_ENTERPRISE`, `SQLSERVER_2019_EXPRESS`," +
								"`SQLSERVER_2019_WEB`.[Database Version Policies](https://cloud.google.com/sql/docs/db-versions)includes an up-to-date reference of supported versions.",
							Type:     schema.TypeString,
							Computed: true,
						},
						"tier": {
							Description: "The machine type to use. See tiers for more details and supported versions. " +
								"Postgres supports only shared-core machine types, and custom machine types such as `db-custom-2-13312`." +
								"See the [Custom Machine Type Documentation](https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#create) to learn about specifying custom machine types.",
							Type:     schema.TypeString,
							Computed: true,
						},
						"disk_size": {
							Description: `The size of data disk, in GB. Size of a running instance cannot be reduced but can be increased. The minimum value is 10GB.`,
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"labels": {
							Description: "Map of string keys and values that can be used to organize and categorize this resource.",
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"ip_address": {
							Description: "IP address of the database.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"connection_name": {
							Description: "Connection name  of the database.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
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
			"ip_address":       flattenStringList(duplo.IPAddress),
			"connection_name":  duplo.ConnectionName,
		}
		list = append(list, sql)
	}
	d.Set("databases", list)
	log.Printf("[TRACE] dataSourceGCPSQLCloudList ******** end")
	return nil

}
