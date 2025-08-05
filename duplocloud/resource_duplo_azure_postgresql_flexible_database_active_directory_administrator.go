package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzurePostgresqlFlexibleDatabaseADSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure postgresql database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"db_name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"principal_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"azure_tenant_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"principal_type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"object_id": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func resourceAzurePostgresqlFlexibleDatabaseAD() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_postgresql_flexible_db_ad_administrator` manages an azure postgresql flexible  active directory configuration.",

		ReadContext:   resourceAzurePostgresqlFlexibleDatabaseADRead,
		CreateContext: resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate,
		UpdateContext: resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate,
		DeleteContext: resourceAzurePostgresqlFlexibleDatabaseADDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema:        duploAzurePostgresqlFlexibleDatabaseADSchema(),
		CustomizeDiff: verifyPSQLFlexiV2Parameters,
	}
}

func resourceAzurePostgresqlFlexibleDatabaseADRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzurePostgresqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADRead(%s, %s): start", tenantID, name)

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
	//	flattenAzurePostgresqlFlexibleDatabasev2(d, duplo)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	objectId := d.Get("object_id").(string)
	name := d.Get("db_name").(string)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate(%s, %s): start", tenantID, objectId)
	c := m.(*duplosdk.Client)
	adc := expandADConfig(d)
	cerr := c.PostgresqlFlexibleDatabaseUpdateADConfig(tenantID, name, objectId, &adc)
	if cerr != nil {
		return diag.Errorf("Failed to Active directory user for tenantId %s postgres flexible server %s", tenantID, name)
	}
	id := fmt.Sprintf("%s/%s/%s", tenantID, name, objectId)
	d.SetId(id)

	err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"), "create")
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceAzurePostgresqlFlexibleDatabasev2Read(ctx, d, m)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate(%s, %s): end", tenantID, name)
	return diags
}

func expandADConfig(d *schema.ResourceData) duplosdk.DuploAzurePostgresqlFlexibleV2ADConfig {
	return duplosdk.DuploAzurePostgresqlFlexibleV2ADConfig{
		ADPrincipalName: d.Get("principal_name").(string),
		ADTenantId:      d.Get("azure_tenant_id").(string),
		ADPrincipalType: d.Get("principal_type").(string),
		ObjectId:        d.Get("object_id").(string),
	}
}

/*
func flattenAzurePostgresqlFlexibleDatabaseAD(d *schema.ResourceData, duplo *duplosdk.DuploAzurePostgresqlFlexibleV2) {
	d.Set("principal_name", duplo.Activ)
	d.Set("azure_tenant_id", duplo)
	d.Set("principal_type", duplo.Sku.Name)
	d.Set("administrator_login", duplo.AdminUserName)
	d.Set("storage_gb", duplo.Storage.StorageSize)
	d.Set("version", duplo.Version)
	d.Set("backup_retention_days", duplo.BackUp.RetentionDays)
	d.Set("geo_redundant_backup", duplo.BackUp.GeoRedundantBackUp)
	d.Set("high_availability", duplo.HighAvailability.Mode)
	//d.Set("subnet", duplo.)
	d.Set("tags", duplo.Tags)
	d.Set("location", duplo.Location)

}*/

func resourceAzurePostgresqlFlexibleDatabaseADDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split("/", id)
	tenantID, name, objectId := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.PostgresqlFlexibleDatabaseUpdateADConfig(tenantID, name, objectId, nil)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure postgresql database '%s': %s", tenantID, name, clientErr)
	}

	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADDelete(%s, %s): end", tenantID, name)
	return nil
}
