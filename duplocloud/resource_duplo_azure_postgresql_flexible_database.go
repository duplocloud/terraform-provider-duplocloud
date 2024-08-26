package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzurePostgresqlFlexibleDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure postgresql database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the PostgreSQL Server. Changing this forces a new resource to be created. This needs to be globally unique within Azure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},

		"service_tier": {
			Description: "Specify service_tier among `Burstable`, `GeneralPurpose` or `MemoryOptimized`",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{"Burstable", "GeneralPurpose", "MemoryOptimized"},
				false),
		},
		"hardware": {
			Description: "Specify compute based on service tier by prepending Standard_ keyword from following document https://azure.microsoft.com/en-in/pricing/details/postgresql/flexible-server",
			Type:        schema.TypeString,
			Required:    true,
		},
		"high_availability": {
			Description: "High availability options— Disabled, SameZone, and ZoneRedundant — are applicable if the service tier is set to GeneralPurpose or MemoryOptimized.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Disabled",
				"SameZone",
				"ZoneRedundant",
			}, false),
		},
		"administrator_login": {
			Description: "The Administrator Login for the PostgreSQL Server.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},

		"administrator_login_password": {
			Description: "The Password associated with the `administrator_login` for the PostgreSQL Server.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"storage_gb": {
			Description:  "Max storage allowed for a server. Possible values are `32`,`64`,`128`,`256`,`512`,`1024`,`2048`,`4096`,`8192`,`16384`,`32768` GB",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntInSlice([]int{32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768}),
		},
		"version": {
			Description: "Specifies the version of PostgreSQL Flexible DB to use. Valid values are `16`,`15`,`14`,`13`,`12`,`11`. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"16",
				"15",
				"14",
				"13",
				"12",
				"11",
			}, true),
			ForceNew: true,
		},
		"backup_retention_days": {
			Description:  "Backup retention days for the server, supported values are between `7` and `35` days.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(7, 35),
		},
		"geo_redundant_backup": {
			Description: "Turn Geo-redundant server backups Enabled/Disabled.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Enabled",
				"Disabled",
			}, false),
		},
		"subnet": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
		"location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until PostgreSQL Server instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceAzurePostgresqlFlexibleDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_postgresql_flexible_database` manages an azure postgresql flexible  database in Duplo.",

		ReadContext:   resourceAzurePostgresqlFlexibleDatabaseRead,
		CreateContext: resourceAzurePostgresqlFlexibleDatabaseCreate,
		UpdateContext: resourceAzurePostgresqlFlexibleDatabaseUpdate,
		DeleteContext: resourceAzurePostgresqlFlexibleDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzurePostgresqlFlexibleDatabaseSchema(),
	}
}

func resourceAzurePostgresqlFlexibleDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.PostgresqlFlexibleDatabaseGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure postgresql flexible database %s : %s", tenantID, name, clientErr)
	}
	// TODO Set ccomputed attributes from duplo object to tf state.
	flattenAzurePostgresqlFlexibleDatabase(d, duplo)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzurePostgresqlFlexibleDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzurePostgresqlFlexibleDatabase(d)
	rq.RequestMode = "Create"
	//password := rq.AdminLoginPassword
	_, err = c.PostgresqlFlexibleDatabaseCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure postgresql flexible database '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure postgresql flexible database", id, func() (interface{}, duplosdk.ClientError) {
		return c.PostgresqlFlexibleDatabaseGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready").(bool) {
		err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzurePostgresqlFlexibleDatabaseRead(ctx, d, m)
	//	if !diags.HasError() {
	//		d.Set("AdminLoginPassword", password)
	//	}
	log.Printf("[TRACE] resourceAzurePostgresqlDatabaseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzurePostgresqlFlexibleDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzurePostgresqlFlexibleDatabase(d)
	rq.RequestMode = "Update"
	//password := rq.AdminLoginPassword
	_, err = c.PostgresqlFlexibleDatabaseUpdate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure postgresql database '%s': %s", tenantID, name, err)
	}
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure postgresql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.PostgresqlServerGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}

	if d.Get("wait_until_ready").(bool) {
		err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzurePostgresqlFlexibleDatabaseRead(ctx, d, m)
	//if !diags.HasError() {
	//	d.Set("AdminLoginPassword", password)
	//}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzurePostgresqlFlexibleDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.PostgresqlFlexibleDatabaseDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure postgresql database '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure postgresql flexible database", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.PostgresqlServerExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzurePostgresqlFlexibleDatabase(d *schema.ResourceData) *duplosdk.DuploAzurePostgresqlFlexibleRequest {
	req := &duplosdk.DuploAzurePostgresqlFlexibleRequest{
		Name:               d.Get("name").(string),
		Version:            d.Get("version").(string),
		AdminUserName:      d.Get("administrator_login").(string),
		AdminLoginPassword: d.Get("administrator_login_password").(string),
	}
	req.Sku.Tier = d.Get("service_tier").(string)
	req.Sku.Name = d.Get("hardware").(string)
	if v, ok := d.GetOk("high_availability"); ok && (req.Sku.Tier == "GeneralPurpose" || req.Sku.Tier == "MemoryOptimized") {
		req.HighAvailability.Mode = v.(string)
	}
	req.Storage.StorageSize = d.Get("storage_gb").(int)
	req.Network.Subnet = d.Get("subnet").(string)
	req.BackUp.RetentionDays = d.Get("backup_retention_days").(int)
	req.BackUp.GeoRedundantBackUp = d.Get("geo_redundant_backup").(string)
	return req
}

func flattenAzurePostgresqlFlexibleDatabase(d *schema.ResourceData, duplo *duplosdk.DuploAzurePostgresqlFlexible) {
	d.Set("name", duplo.Name)
	d.Set("service_tier", duplo.Sku.Tier)
	d.Set("hardware", duplo.Sku.Name)
	d.Set("administrator_login", duplo.AdminUserName)
	d.Set("storage_gb", duplo.Storage.StorageSize)
	d.Set("version", duplo.Version)
	d.Set("backup_retention_days", duplo.BackUp.RetentionDays)
	d.Set("geo_redundant_backup", duplo.BackUp.GeoRedundantBackUp)
	d.Set("high_availability", duplo.HighAvailability.Mode)
	//d.Set("subnet", duplo.)
	d.Set("tags", duplo.Tags)
	d.Set("location", duplo.Location)

}

func postgresqlFlexibleDBWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.PostgresqlFlexibleDatabaseGet(tenantID, name)
			log.Printf("[TRACE] postgresql server user visible state is (%s).", rp.State)
			status := "pending"
			if err == nil {
				if rp.State == "Ready" {
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
	log.Printf("[DEBUG] postgresqlSeverWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
