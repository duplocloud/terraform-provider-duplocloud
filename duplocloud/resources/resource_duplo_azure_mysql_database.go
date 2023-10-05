package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureMysqlDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure mysql database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the MySQL Server. Changing this forces a new resource to be created. This needs to be globally unique within Azure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"sku_name": {
			Description: "Specifies the SKU Name for this MySQL Server.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"B_Gen4_1",
				"B_Gen4_2",
				"B_Gen5_1",
				"B_Gen5_2",
				"GP_Gen4_2",
				"GP_Gen4_4",
				"GP_Gen4_8",
				"GP_Gen4_16",
				"GP_Gen4_32",
				"GP_Gen5_2",
				"GP_Gen5_4",
				"GP_Gen5_8",
				"GP_Gen5_16",
				"GP_Gen5_32",
				"GP_Gen5_64",
				"MO_Gen5_2",
				"MO_Gen5_4",
				"MO_Gen5_8",
				"MO_Gen5_16",
				"MO_Gen5_32",
			}, false),
		},
		"administrator_login": {
			Description: "The Administrator Login for the MySQL Server.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},

		"administrator_login_password": {
			Description: "The Password associated with the `administrator_login` for the MySQL Server.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"storage_mb": {
			Description: " Max storage allowed for a server. Possible values are between `5120` MB(5GB) and `1048576` MB(1TB) for the Basic SKU and between `5120` MB(5GB) and `4194304` MB(4TB) for General Purpose/Memory Optimized SKUs.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.All(
				validation.IntBetween(5120, 4194304),
				validation.IntDivisibleBy(1024),
			),
		},
		"version": {
			Description: "Specifies the version of MySQL to use. Valid values are `5.7`, or `8.0`. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"5.7",
				"8.0",
			}, true),
			ForceNew: true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until MySql Server instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"backup_retention_days": {
			Description:  "Backup retention days for the server, supported values are between `7` and `35` days.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(7, 35),
		},
		"geo_redundant_backup": {
			Description: "Turn Geo-redundant server backups on/off.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Enabled",
				"Disabled",
			}, true),
		},
		"ssl_enforcement": {
			Description: "Specifies if SSL should be enforced on connections.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Enabled",
				"Disabled",
			}, true),
		},
		"fqdn": {
			Description: "The FQDN of the MySQL Server.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
	}
}

func resourceAzureMysqlDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_mysql_database` manages an Azure mysql database in Duplo.",

		ReadContext:   resourceAzureMysqlDatabaseRead,
		CreateContext: resourceAzureMysqlDatabaseCreate,
		UpdateContext: resourceAzureMysqlDatabaseUpdate,
		DeleteContext: resourceAzureMysqlDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureMysqlDatabaseSchema(),
	}
}

func resourceAzureMysqlDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureMysqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMysqlDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MySqlServerGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure mysql database %s : %s", tenantID, name, clientErr)
	}

	// TODO Set ccomputed attributes from duplo object to tf state.
	flattenAzureMysqlDatabase(d, duplo)

	log.Printf("[TRACE] resourceAzureMysqlDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureMysqlDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureMysqlDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureMysqlDatabase(d)
	err = c.MySqlDatabaseCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure mysql database '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "azure mysql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.MySqlServerGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the mysql server instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mysqlSeverWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureMysqlDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureMysqlDatabaseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureMysqlDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureMysqlDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureMysqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMysqlDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.MySqlDatabaseDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure mysql database '%s': %s", tenantID, name, clientErr)
	}

	diag := duplocloud.waitForResourceToBeMissingAfterDelete(ctx, d, "azure mysql database", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.MySqlServerExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureMysqlDatabaseDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureMysqlDatabase(d *schema.ResourceData) *duplosdk.DuploAzureMySqlRequest {
	return &duplosdk.DuploAzureMySqlRequest{
		Name:                d.Get("name").(string),
		Version:             d.Get("version").(string),
		StorageMB:           d.Get("storage_mb").(int),
		AdminUsername:       d.Get("administrator_login").(string),
		AdminPassword:       d.Get("administrator_login_password").(string),
		BackupRetentionDays: d.Get("backup_retention_days").(int),
		GeoRedundantBackup:  d.Get("geo_redundant_backup").(string),
		Size:                d.Get("sku_name").(string),
	}
}

func parseAzureMysqlDatabaseIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureMysqlDatabase(d *schema.ResourceData, duplo *duplosdk.DuploAzureMySqlServer) {
	d.Set("administrator_login", duplo.PropertiesAdministratorLogin)
	d.Set("storage_mb", duplo.PropertiesStorageProfile.StorageMB)
	d.Set("version", duplo.PropertiesVersion)
	d.Set("backup_retention_days", duplo.PropertiesStorageProfile.BackupRetentionDays)
	d.Set("geo_redundant_backup", duplo.PropertiesStorageProfile.GeoRedundantBackup)
	d.Set("ssl_enforcement", duplo.PropertiesSslEnforcement)
	d.Set("fqdn", duplo.PropertiesFullyQualifiedDomainName)
	d.Set("tags", duplo.Tags)

}

func mysqlSeverWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.MySqlServerGet(tenantID, name)
			log.Printf("[TRACE] MySql server user visible state is (%s).", rp.PropertiesUserVisibleState)
			status := "pending"
			if err == nil {
				if rp.PropertiesUserVisibleState == "Ready" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] mysqlSeverWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
