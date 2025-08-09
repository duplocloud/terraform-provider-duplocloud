package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
			Description: "Postgres flexible server database name",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"principal_name": {
			Description: "Azure account user name",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"azure_tenant_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Azure GUID of the tenant",
		},
		"principal_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"User", "Group", "ServicePrincipal"}, false),
			Description:  "Specify the type of Azure AD identity being used for that administrator.",
		},
		"object_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The Azure Active Directory (AAD) Object ID of the user, group, or service principal. This is a globally unique identifier assigned by AAD to each identity, used to manage access and authentication across Azure resources.",
		},
	}
}

func resourceAzurePostgresqlFlexibleDatabaseAD() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_postgresql_flexible_db_ad_administrator` manages an azure postgresql flexible  active directory configuration. Supported with duplocloud_azure_postgresql_flexible_database_v2",

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
		Schema: duploAzurePostgresqlFlexibleDatabaseADSchema(),
	}
}

func resourceAzurePostgresqlFlexibleDatabaseADRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, name, objectId, principalType := idParts[0], idParts[1], idParts[2], idParts[3]

	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.PostgresqlFlexibleDatabaseADGet(tenantID, name, objectId, principalType)
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
	d.Set("tenant_id", tenantID)
	d.Set("object_id", objectId)
	d.Set("azure_tenant_id", duplo.ADTenantId)
	d.Set("principal_name", duplo.ADPrincipalName)
	d.Set("principal_type", duplo.ADPrincipalType)
	d.Set("db_name", name)
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzurePostgresqlFlexibleDatabaseADCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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
	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, name, objectId, adc.ADPrincipalType)
	d.SetId(id)

	//err = postgresqlFlexibleDBWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"), "create")
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	err := waitUntilADconfigured(ctx, c, tenantID, name, objectId, adc.ADPrincipalType, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("Error waiting for active directory configuration for tenant %s postgres flexible server %s: %s", tenantID, name, err)
	}
	diags := resourceAzurePostgresqlFlexibleDatabaseADRead(ctx, d, m)
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

func resourceAzurePostgresqlFlexibleDatabaseADDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, name, objectId := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	//err is set to nil for 404 so no need to check for it
	err := ensureADconfiguredDelete(ctx, c, tenantID, name, objectId, d.Timeout("delete"))
	if err != nil {
		return diag.Errorf("Unable to delete tenant %s azure postgresql database active directory authentication '%s': %s", tenantID, name, err)
	}

	log.Printf("[TRACE] resourceAzurePostgresqlFlexibleDatabaseADDelete(%s, %s): end", tenantID, name)
	return nil
}

func waitUntilADconfigured(ctx context.Context, c *duplosdk.Client, tenantID, name, objectId, principalType string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.PostgresqlFlexibleDatabaseADGet(tenantID, name, objectId, principalType)
			status := "pending"
			if rp != nil {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilADconfigured(%s, %s,%s)", tenantID, name, objectId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func ensureADconfiguredDelete(ctx context.Context, c *duplosdk.Client, tenantID, name, objectId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			err := c.PostgresqlFlexibleDatabaseADDelete(tenantID, name, objectId)
			if err == nil {
				log.Printf("Delete status ready")
				return "deleted", "ready", nil
			}
			if err.Status() == 404 {
				log.Printf("Delete status ready (404)")
				return "deleted", "ready", nil
			}
			if strings.Contains(err.Error(), "is busy with another operation") {
				return nil, "pending", nil
			}
			return nil, "pending", err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilADconfigured(%s, %s,%s)", tenantID, name, objectId)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
