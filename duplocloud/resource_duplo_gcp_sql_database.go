package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpSqlDBInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the sql database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the sql database.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
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
			Description: "The MySQL, PostgreSQL or SQL Server version to use." +
				"Supported values include `MYSQL_5_6`,`MYSQL_5_7`, `MYSQL_8_0`, `POSTGRES_9_6`,`POSTGRES_10`," +
				"`POSTGRES_11`,`POSTGRES_12`, `POSTGRES_13`, `POSTGRES_14`, `POSTGRES_15`,`POSTGRES_16`,`POSTGRES_17`, `SQLSERVER_2017_STANDARD`,`SQLSERVER_2017_ENTERPRISE`," +
				"`SQLSERVER_2017_EXPRESS`, `SQLSERVER_2017_WEB`,`SQLSERVER_2019_STANDARD`, `SQLSERVER_2019_ENTERPRISE`, `SQLSERVER_2019_EXPRESS`," +
				"`SQLSERVER_2019_WEB`,`SQLSERVER_2022_WEB`,`SQLSERVER_2022_EXPRESS`,`SQLSERVER_2022_ENTERPRISE`,`SQLSERVER_2022_STANDARD`.[Database Version Policies](https://cloud.google.com/sql/docs/db-versions) includes an up-to-date reference of supported versions.",
			Type:     schema.TypeString,
			Required: true,
		},

		"tier": {
			Description: "The machine type to use. See tiers for more details and supported versions. " +
				"Postgres supports only shared-core machine types, and custom machine types such as `db-custom-2-13312`." +
				"See the [Custom Machine Type Documentation](https://cloud.google.com/compute/docs/instances/creating-instance-with-custom-machine-type#create) to learn about specifying custom machine types.",
			Type:     schema.TypeString,
			Required: true,
		},
		"disk_size": {
			Description: `The size of data disk, in GB. Size of a running instance cannot be reduced but can be increased. The minimum value is 10GB.`,
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"labels": {
			Description: "Map of string keys and values that can be used to organize and categorize this resource.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until sql database instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"root_password": {
			Description: "Provide root password for specific database versions.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"ip_address": {
			Description: "List of IP addresses of the database.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"connection_name": {
			Description: "Connection name of the database.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"need_backup": {
			Description: "Flag to enable backup process on delete of database",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceGcpSqlDBInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_sql_database_instance` manages a GCP SQL Database Instance in Duplo.",

		ReadContext:   resourceGcpSqlDBInstanceRead,
		CreateContext: resourceGcpSqlDBInstanceCreate,
		UpdateContext: resourceGcpSqlDBInstanceUpdate,
		DeleteContext: resourceGcpSqlDBInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: gcpSqlDBInstanceSchema(),
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIf("database_version", replaceOn)),
	}
}

// READ resource
func resourceGcpSqlDBInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSqlDBInstanceRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseGcpSqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	//fullName := d.Get("fullname").(string)
	fullName, clientErr := c.GetDuploServicesNameWithGcp(tenantID, name, false)
	if clientErr != nil {
		return diag.Errorf("Error fetching tenant prefix for %s : %s", tenantID, clientErr)
	}
	duplo, clientErr := c.GCPSqlDBInstanceGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s gcp sql database '%s': %s", tenantID, fullName, clientErr)
	}

	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	flattenGcpSqlDBInstance(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpSqlDBInstanceRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpSqlDBInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSqlDBInstanceCreate ******** start")
	// Create the request object.
	rq := expandGcpSqlDBInstance(d)
	tenantID := d.Get("tenant_id").(string)

	c := m.(*duplosdk.Client)

	// Validate the database version
	requestedVersion := d.Get("database_version").(string)
	if diags := validateDatabaseVersion(ctx, c, tenantID, requestedVersion); diags != nil {
		return diags
	}

	// Validate if password is needed
	needPassword, diags := passwordRequiredForVersion(ctx, c, tenantID, requestedVersion)
	if diags != nil {
		return diags
	}
	if needPassword && rq.RootPassword == "" {
		return diag.Errorf("root password is mandatory for database version %s", requestedVersion)
	}

	// Post the object to Duplo
	resp, err := c.GCPSqlDBInstanceCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s sql database '%s': %s", tenantID, resp.Name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, rq.Name)
	fullName := resp.Name

	diags = waitForResourceToBePresentAfterCreate(ctx, d, "gcp sql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPSqlDBInstanceGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err := gcpSqlDBInstanceWaitUntilReady(ctx, c, tenantID, fullName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	resourceGcpSqlDBInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpSqlDBInstanceCreate ******** end")
	return diags
}

// UPDATE resource
func resourceGcpSqlDBInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSqlDBInstanceUpdate ******** start")

	id := d.Id()
	tenantID, _, err := parseGcpSqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	fullName := d.Get("fullname").(string)
	// Post the object to Duplo
	duplo, clientErr := c.GCPSqlDBInstanceGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s gcp sql database '%s': %s", tenantID, fullName, clientErr)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if d.HasChanges("tier", "disk_size", "labels", "database_version") {
		requestedVersion := d.Get("database_version").(string)

		// Validate the database version
		if diags := validateDatabaseVersion(ctx, c, tenantID, requestedVersion); diags != nil {
			return diags
		}

		rq := expandGcpSqlDBInstance(d)
		rq.Name = fullName
		resp, err := c.GCPSqlDBInstanceUpdate(tenantID, rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s sql database '%s': %s", tenantID, resp.Name, err)
		}
		if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
			clientErr := gcpSqlDBInstanceWaitUntilReady(ctx, c, tenantID, fullName, d.Timeout("update"))
			if clientErr != nil {
				return diag.FromErr(clientErr)
			}
		}
		resourceGcpSqlDBInstanceRead(ctx, d, m)
	}

	log.Printf("[TRACE] resourceGcpSqlDBInstanceCreate ******** end")
	return nil
}

// DELETE resource
func resourceGcpSqlDBInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSqlDBInstanceDelete ******** start")
	id := d.Id()
	tenantID, _, err := parseGcpSqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	fullName := d.Get("fullname").(string)

	resp, clientErr := c.GCPSqlDBInstanceGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s gpc sql database instance %s : %s", tenantID, resp.Name, clientErr)
	}
	backup := d.Get("need_backup").(bool)
	clientErr = c.GCPSqlDBInstanceDelete(tenantID, fullName, backup)
	if clientErr != nil {
		return diag.Errorf("Error deleting gcp sql database '%s': %s", id, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "gcp sql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPSqlDBInstanceGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpSqlDBInstanceDelete ******** end")
	return nil
}

func flattenGcpSqlDBInstance(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGCPSqlDBInstance) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("self_link", duplo.SelfLink)
	d.Set("tier", duplo.Tier)
	d.Set("database_version", duplo.DatabaseVersion)
	d.Set("disk_size", duplo.DataDiskSizeGb)
	d.Set("ip_address", flattenStringList(duplo.IPAddress))
	d.Set("connection_name", duplo.ConnectionName)
	flattenGcpLabels(d, duplo.Labels)
	flattenIPAddress(d, duplo.IPAddress)

}

func expandGcpSqlDBInstance(d *schema.ResourceData) *duplosdk.DuploGCPSqlDBInstance {
	rq := &duplosdk.DuploGCPSqlDBInstance{
		Name:            d.Get("name").(string),
		DatabaseVersion: d.Get("database_version").(string),
		Tier:            d.Get("tier").(string),
		DataDiskSizeGb:  d.Get("disk_size").(int),
		ResourceType:    duplosdk.DuploGCPDatabaseInstanceResourceType,
		RootPassword:    d.Get("root_password").(string),
	}
	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		rq.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			rq.Labels[key] = value.(string)
		}
	}
	return rq
}

func parseGcpSqlDatabaseIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func gcpSqlDBInstanceWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	log.Printf("[DEBUG] gcpSqlDBInstanceWaitUntilReady(%s, %s)", tenantID, name)
	stateChange := false
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			log.Printf("[TRACE] gcpSqlDBInstanceWaitUntilReady - Refresh.")
			rp, err := c.GCPSqlDBInstanceGet(tenantID, name)
			log.Printf("[TRACE] Gcp sql database instance state is (%s).", rp.Status)
			status := "pending"
			if err == nil {
				if rp.Status == "RUNNABLE" && !stateChange {
					return rp, status, err
				} else {
					stateChange = true
				}
				if rp.Status == "RUNNABLE" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 10 * time.Second,
		Timeout:      timeout,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func replaceOn(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
	// Suppress diff if both database name and version are the same
	old, new := d.GetChange("database_version")
	oldParts := strings.Split(old.(string), "_")
	newParts := strings.Split(new.(string), "_")
	return oldParts[0] != newParts[0]
}

func validateDatabaseVersion(ctx context.Context, c *duplosdk.Client, tenantID, requestedVersion string) diag.Diagnostics {
	versions, err := c.GCPSqlDBInstanceVersionsList(tenantID)
	if err != nil {
		return diag.Errorf("Error fetching supported database versions: %s", err)
	}

	if !stringInSlice(requestedVersion, versions) {
		return diag.Errorf("Unsupported database version '%s'. Must be one of: %v", requestedVersion, versions)
	}

	return nil
}

func passwordRequiredForVersion(ctx context.Context, c *duplosdk.Client, tenantID, requestedVersion string) (bool, diag.Diagnostics) {
	versionsRequiringPassword, err := c.GCPSqlDBInstanceVersionsRequiringPasswordList(tenantID)
	if err != nil {
		return false, diag.Errorf("Error fetching versions requiring password: %s", err)
	}

	for _, v := range versionsRequiringPassword {
		if v == requestedVersion {
			return true, nil
		}
	}
	return false, nil
}
