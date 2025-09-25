package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzurePostgresqlFlexibleDatabaseSchemav2() map[string]*schema.Schema {
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
			Description: "Specify service_tier among `Burstable`, `GeneralPurpose` or `MemoryOptimized`. Note: should disable high_availability before updating to Burstable",
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
			Default:     "Disabled",
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
			Description:  "Max storage allowed for a server. Possible values are `32`,`64`,`128`,`256`,`512`,`1024`,`2048`,`4096`,`8192`,`16384`,`32768` GB. Note: Updation allowed on updating with higher storage size from current",
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
		"minor_version": {
			Description:  "Minor version",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "0",
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9]+$`), "Invalid minor version"),
		},
		"backup_retention_days": {
			Description:  "Backup retention days for the server, supported values are between `7` and `35` days. Note: Updation allowed on updating with higher retention days value from current",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(7, 35),
		},
		"geo_redundant_backup": {
			Description: "Turn Geo-redundant server backups Enabled/Disabled. To enable or disable geo_redundant_backup resource need to be recreated",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Enabled",
				"Disabled",
			}, false),
			ForceNew: true,
		},
		"delegated_subnet_id": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The ID of the virtual network subnet to create the PostgreSQL Flexible Server. The provided subnet should not have any other resource deployed in it and this subnet will be delegated to the PostgreSQL Flexible Server, if not already delegated. Changing this forces a new PostgreSQL Flexible Server to be created.",
		},
		"private_dns_zone_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the private DNS zone to create the PostgreSQL Flexible Server. The private_dns_zone_id will be required when setting a delegated_subnet_id. For existing flexible servers who don't want to be recreated, you need to provide the private_dns_zone_id to the service team to manually migrate to the specified private DNS zone. The azurerm_private_dns_zone should end with suffix .postgres.database.azure.com.",
		},
		"public_network_access": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Whether public network access is allowed for this server.",
			Default:      "Enabled",
			ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
		},
		"availability_zone": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"1", "2", "3"}, false),
			Description:  "The Azure Availability Zone where the primary PostgreSQL Flexible Server will be deployed. It enables zone-level redundancy for increased fault tolerance. When High Availability (HA) is enabled in ZoneRedundant mode, the standby replica is automatically placed in a different zone. Valid values are `1`, `2`, or `3`, depending on the region's supported zones",
		},
		"tags": {
			Description: "A mapping of tags which should be assigned to the PostgreSQL Flexible Server",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        schema.TypeString,
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
		"active_directory_authentication": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "Disabled",
			ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			Description:  "Whether Active Directory authentication is allowed to access the PostgreSQL Flexible Server",
		},
		"password_authentication": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "Enabled",
			ValidateFunc: validation.StringInSlice([]string{"Enabled", "Disabled"}, false),
			Description:  "Whether password authentication is allowed to access the PostgreSQL Flexible Server",
		},
		"azure_resource_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Id of postgresql flexible server returned from azure",
		},
		"private_connection_endpoints": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"active_directory_tenant_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Id of postgresql flexible server returned from azure",
		},
		"fully_qualified_domain_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAzurePostgresqlFlexibleDatabaseV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_postgresql_flexible_database_v2` manages an azure postgresql flexible  database in Duplo.",

		ReadContext:   resourceAzurePostgresqlFlexibleDatabasev2Read,
		CreateContext: resourceAzurePostgresqlFlexibleDatabasev2Create,
		UpdateContext: resourceAzurePostgresqlFlexibleDatabasev2Update,
		DeleteContext: resourceAzurePostgresqlFlexibleDatabasev2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema:        duploAzurePostgresqlFlexibleDatabaseSchemav2(),
		CustomizeDiff: verifyPSQLFlexiV2Parameters,
	}
}

func resourceAzurePostgresqlFlexibleDatabasev2Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.PostgresqlFlexibleDatabaseV2Get(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[DEBUG] resourceAzurePostgresqlFlexibleDatabasev2Read: Azure postgresql flexible database %s not found for tenantId %s, removing from state", name, tenantID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure postgresql flexible database %s : %s", tenantID, name, clientErr)
	}
	// TODO Set ccomputed attributes from duplo object to tf state.
	flattenAzurePostgresqlFlexibleDatabasev2(d, duplo)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzurePostgresqlFlexibleDatabasev2Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzurePostgresqlFlexibleDatabasev2(d)
	//password := rq.AdminLoginPassword
	_, err = c.PostgresqlFlexibleDatabaseV2Create(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure postgresql flexible database '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure postgresql flexible database", id, func() (interface{}, duplosdk.ClientError) {
		return c.PostgresqlFlexibleDatabaseV2Get(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready").(bool) {
		err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"), "create")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzurePostgresqlFlexibleDatabasev2Read(ctx, d, m)
	//	if !diags.HasError() {
	//		d.Set("AdminLoginPassword", password)
	//	}
	log.Printf("[TRACE] resourceAzurePostgresqlDatabaseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzurePostgresqlFlexibleDatabasev2Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzurePostgresqlFlexibleDatabasev2(d)
	//password := rq.AdminLoginPassword
	_, err = c.PostgresqlFlexibleDatabaseV2Update(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure postgresql database '%s': %s", tenantID, name, err)
	}

	if d.Get("wait_until_ready").(bool) {
		err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("update"), "update")
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := resourceAzurePostgresqlFlexibleDatabasev2Read(ctx, d, m)
	//if !diags.HasError() {
	//	d.Set("AdminLoginPassword", password)
	//}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseUpdate(%s, %s): end", tenantID, name)
	return diags
}

func expandAzurePostgresqlFlexibleDatabasev2(d *schema.ResourceData) *duplosdk.DuploAzurePostgresqlFlexibleV2Request {
	req := &duplosdk.DuploAzurePostgresqlFlexibleV2Request{
		Name:               d.Get("name").(string),
		Version:            d.Get("version").(string),
		MinorVersion:       d.Get("minor_version").(string),
		AdminUserName:      d.Get("administrator_login").(string),
		AdminLoginPassword: d.Get("administrator_login_password").(string),
	}
	req.Sku.Tier = d.Get("service_tier").(string)
	req.Sku.Name = d.Get("hardware").(string)
	if v, ok := d.GetOk("high_availability"); ok && (req.Sku.Tier == "GeneralPurpose" || req.Sku.Tier == "MemoryOptimized") {
		req.HighAvailability.Mode = v.(string)
	}
	req.Storage.StorageSize = d.Get("storage_gb").(int)
	req.Network.Subnet.ResourceId = d.Get("delegated_subnet_id").(string)
	req.Network.PrivateDnsZone.ResourceId = d.Get("private_dns_zone_id").(string)
	req.Network.PublicNetworkAccess = d.Get("public_network_access").(string)
	req.BackUp.RetentionDays = d.Get("backup_retention_days").(int)
	req.BackUp.GeoRedundantBackUp = d.Get("geo_redundant_backup").(string)
	req.AvailabilityZone = d.Get("availability_zone").(string)
	req.AuthConfig.ActiveDirectoryAuth = d.Get("active_directory_authentication").(string)
	req.AuthConfig.PasswordAuth = d.Get("password_authentication").(string)
	return req
}

func flattenAzurePostgresqlFlexibleDatabasev2(d *schema.ResourceData, duplo *duplosdk.DuploAzurePostgresqlFlexibleV2) {
	d.Set("name", duplo.Name)
	d.Set("service_tier", duplo.Sku.Tier)
	d.Set("hardware", duplo.Sku.Name)
	d.Set("administrator_login", duplo.AdminUserName)
	d.Set("storage_gb", duplo.Storage.StorageSize)
	d.Set("version", duplo.Version)
	d.Set("backup_retention_days", duplo.BackUp.RetentionDays)
	d.Set("geo_redundant_backup", duplo.BackUp.GeoRedundantBackUp)
	d.Set("high_availability", duplo.HighAvailability.Mode)
	d.Set("delegated_subnet_id", duplo.Network.Subnet)
	d.Set("private_dns_zone_id", duplo.Network.PrivateDnsZone)
	d.Set("public_network_access", duplo.Network.PublicNetworkAccess)
	d.Set("tags", duplo.Tags)
	d.Set("location", duplo.Location.Name)
	d.Set("active_directory_authentication", duplo.AuthConfig.ActiveDirectoryAuth)
	d.Set("password_authentication", duplo.AuthConfig.PasswordAuth)
	d.Set("active_directory_tenant_id", duplo.AuthConfig.TenantId)
	d.Set("azure_resource_id", duplo.AzureResourceId)
	d.Set("availability_zone", duplo.AvailabilityZone)
	d.Set("fully_qualified_domain_name", duplo.FullyQualifiedDomainName)
	con := []interface{}{}
	for _, pc := range duplo.PrivateEndpointConnections {
		con = append(con, pc)
	}
	d.Set("private_connection_endpoints", con)

}

func resourceAzurePostgresqlFlexibleDatabasev2Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
			log.Printf("[DEBUG] resourceAzurePostgresqlFlexibleDatabasev2Delete: Azure postgresql flexible database %s not found for tenantId %s, removing from state", name, tenantID)
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure postgresql database '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure postgresql flexible database", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.PostgresqlFlexibleDatabaseGet(tenantID, name); err != nil {
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

func verifyPSQLFlexiV2Parameters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	_, newS := diff.GetChange("service_tier")
	oldH, newH := diff.GetChange("high_availability")
	newHA := newH.(string)
	if newS == "Burstable" && newHA != "Disabled" {
		return fmt.Errorf("high_availability not supported with Burstable compute/service tier")
	}
	old, new := diff.GetChange("storage_gb")
	if new.(int) < old.(int) {
		return fmt.Errorf("storage downgrade not allowed")
	}
	oldR, newR := diff.GetChange("backup_retention_days")
	if newR.(int) < oldR.(int) {
		return fmt.Errorf("reducing the value of backup_retention_days is not allowed")
	}
	if oldH.(string) != "Disabled" && oldH.(string) != "" && newS.(string) == "Burstable" {
		return fmt.Errorf("disable high_availability before updating to Burstable compute/service tier")
	}

	if diff.Get("delegated_subnet_id").(string) != "" && diff.Get("private_dns_zone_id") == "" {
		return fmt.Errorf("private_dns_zone_id is required when delegated_subnet_id is set")
	}

	return nil
}
