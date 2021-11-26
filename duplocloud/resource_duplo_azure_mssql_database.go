package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureMssqlDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure mssql database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Microsoft SQL Server. This needs to be globally unique within Azure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"version": {
			Description: "The version for the new server. Valid values are: `2.0` (for v11 server) and `12.0` (for v12 server).",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"2.0",
				"12.0",
			}, true),
			ForceNew: true,
		},
		"administrator_login": {
			Description: "The Administrator Login for the  MS sql Server.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},

		"administrator_login_password": {
			Description: "The Password associated with the `administrator_login` for the MS sql Server.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"public_network_access": {
			Description: "Whether public network access is enabled or disabled for this server.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Enabled",
				"Disabled",
			}, true),
		},
		"minimum_tls_version": {
			Description: "The Minimum TLS Version for all SQL Database and SQL Data Warehouse databases associated with the server. Valid values are: `1.0`, `1.1` and `1.2`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"1.0",
				"1.1",
				"1.2",
			}, false),
		},
		"fqdn": {
			Description: "The fully qualified domain name of the Azure SQL Server.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until PostgreSQL Server instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
	}
}

func resourceAzureMssqlDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_mssql_database` manages an azure mssql database in Duplo.",

		ReadContext:   resourceAzureMssqlDatabaseRead,
		CreateContext: resourceAzureMssqlDatabaseCreate,
		UpdateContext: resourceAzureMssqlDatabaseUpdate,
		DeleteContext: resourceAzureMssqlDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureMssqlDatabaseSchema(),
	}
}

func resourceAzureMssqlDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureMssqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MsSqlServerGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure mssql database %s : %s", tenantID, name, clientErr)
	}

	// TODO Set ccomputed attributes from duplo object to tf state.
	flattenAzureMssqlDatabase(d, duplo)

	log.Printf("[TRACE] resourceAzureMssqlDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureMssqlDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureMssqlDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureMssqlDatabase(d)
	err = c.MsSqlDatabaseCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure mssql database '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure mssql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlServerGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the mssql server instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mssqlSeverWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureMssqlDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureMssqlDatabaseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureMssqlDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureMssqlDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureMssqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.MsSqlDatabaseDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure mssql database '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure mssql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlServerGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureMssqlDatabaseDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureMssqlDatabase(d *schema.ResourceData) *duplosdk.DuploAzureMsSqlRequest {
	return &duplosdk.DuploAzureMsSqlRequest{
		Name:                                 d.Get("name").(string),
		PropertiesVersion:                    d.Get("version").(string),
		PropertiesAdministratorLogin:         d.Get("administrator_login").(string),
		PropertiesAdministratorLoginPassword: d.Get("administrator_login_password").(string),
		PropertiesPublicNetworkAccess:        d.Get("public_network_access").(string),
		PropertiesMinimalTLSVersion:          d.Get("minimum_tls_version").(string),
	}
}

func parseAzureMssqlDatabaseIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureMssqlDatabase(d *schema.ResourceData, duplo *duplosdk.DuploAzureMsSqlServer) {
	d.Set("fqdn", duplo.PropertiesFullyQualifiedDomainName)
	d.Set("tags", duplo.Tags)
}

func mssqlSeverWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.MsSqlServerGet(tenantID, name)
			log.Printf("[TRACE] MS Sql server state is (%s).", rp.PropertiesState)
			status := "pending"
			if err == nil {
				if rp.PropertiesState == "Ready" {
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
	log.Printf("[DEBUG] mssqlSeverWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
