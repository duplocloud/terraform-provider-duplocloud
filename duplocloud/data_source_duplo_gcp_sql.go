package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceGCPCloudSQL() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_sql_database_instance` retrieves a gcp sql instance in Duplo.",

		ReadContext: dataSourceGCPSQLCloudRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the sql database will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The short name of the sql database.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fullname": {
				Description: "The full name of the sql database.",
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
			"wait_until_ready": {
				Description: "Whether or not to wait until sql database instance to be ready, after creation.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"ip_address": {
				Description: "IP address of the database.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connection_name": {
				Description: "Connection name  of the database.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
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
