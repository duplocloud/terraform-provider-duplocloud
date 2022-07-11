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
			Description: "The name of the MS SQL Database.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"server_name": {
			Description: "The name of the MS SQL Server on which to create the database.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"elastic_pool_id": {
			Description:   "Specifies the id of the elastic pool containing this database.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"sku"},
		},
		"collation": {
			Description: "Specifies the collation of the database.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"sku": {
			Type:          schema.TypeList,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"elastic_pool_id"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringIsNotEmpty,
					},
					"tier": {
						Type:             schema.TypeString,
						DiffSuppressFunc: CaseDifference,
						Optional:         true,
						Computed:         true,
					},
					"capacity": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(0),
					},
				},
			},
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
	tenantID, serverName, dbName, err := parseAzureMssqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlDatabaseRead(%s, %s, %s): start", tenantID, serverName, dbName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MsSqlDatabaseGet(tenantID, serverName, dbName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure mssql database %s : %s", tenantID, dbName, clientErr)
	}

	err = flattenAzureMssqlDatabase(d, duplo)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", dbName)
	d.Set("server_name", serverName)
	log.Printf("[TRACE] resourceAzureMssqlDatabaseRead(%s, %s): end", tenantID, dbName)
	return nil
}

func resourceAzureMssqlDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	dbName := d.Get("name").(string)
	serverName := d.Get("server_name").(string)
	log.Printf("[TRACE] resourceAzureMssqlDatabaseCreate(%s, %s): start", tenantID, dbName)
	c := m.(*duplosdk.Client)

	rq := expandAzureMssqlDatabase(d)
	err = c.MsSqlDatabaseCreate(tenantID, serverName, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure mssql database '%s': %s", tenantID, dbName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, serverName, dbName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure mssql database", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlDatabaseGet(tenantID, serverName, dbName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mssqlDatabaseWaitUntilReady(ctx, c, tenantID, serverName, dbName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureMssqlDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureMssqlDatabaseCreate(%s, %s): end", tenantID, dbName)
	return diags
}

func resourceAzureMssqlDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureMssqlDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, dbName, err := parseAzureMssqlDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlDatabaseDelete(%s, %s, %s): start", tenantID, serverName, dbName)

	c := m.(*duplosdk.Client)
	clientErr := c.MsSqlDatabaseDelete(tenantID, serverName, dbName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure mssql database '%s': %s", tenantID, dbName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure mssql Server", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlDatabaseGet(tenantID, serverName, dbName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureMssqlDatabaseDelete(%s, %s, %s): end", tenantID, serverName, dbName)
	return nil
}

func expandAzureMssqlDatabase(d *schema.ResourceData) *duplosdk.DuploAzureMsSqlDatabaseRequest {
	return &duplosdk.DuploAzureMsSqlDatabaseRequest{
		Name:                    d.Get("name").(string),
		PropertiesCollation:     d.Get("collation").(string),
		PropertiesElasticPoolId: d.Get("elastic_pool_id").(string),
		Sku:                     expandAzureMssqlDatabaseSku(d),
	}
}

func expandAzureMssqlDatabaseSku(d *schema.ResourceData) *duplosdk.DuploAzureMsSqlDatabaseSku {
	skuConfig := d.Get("sku").([]interface{})
	if len(skuConfig) == 0 {
		return nil
	}

	config := skuConfig[0].(map[string]interface{})

	sku := &duplosdk.DuploAzureMsSqlDatabaseSku{
		Name:     config["name"].(string),
		Capacity: config["capacity"].(int),
	}

	if tier, ok := config["tier"].(string); ok && tier != "" {
		sku.Tier = tier
	}

	return sku
}

func parseAzureMssqlDatabaseIdParts(id string) (tenantID, serverName, dbName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, serverName, dbName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureMssqlDatabase(d *schema.ResourceData, duplo *duplosdk.DuploAzureMsSqlDatabase) error {
	d.Set("name", duplo.Name)
	d.Set("collation", duplo.PropertiesCollation)
	if len(duplo.PropertiesElasticPoolId) > 0 {
		d.Set("elastic_pool_id", duplo.PropertiesElasticPoolId)
	} else {
		if err := d.Set("sku", flattenAzureMssqlDatabaseSku(duplo.PropertiesCurrentSku)); err != nil {
			return fmt.Errorf("[DEBUG] setting `sku`: %#v", err)
		}
	}
	return nil
}

func flattenAzureMssqlDatabaseSku(sku *duplosdk.DuploAzureMsSqlDatabaseSku) []interface{} {
	result := make(map[string]interface{})
	result["name"] = sku.Name
	result["capacity"] = sku.Capacity

	if sku.Tier != "" {
		result["tier"] = sku.Tier
	}

	return []interface{}{result}
}

func mssqlDatabaseWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, serverName string, dbName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.MsSqlDatabaseGet(tenantID, serverName, dbName)
			log.Printf("[TRACE] MS Sql database status is (%s).", rp.PropertiesStatus)
			status := "pending"
			if err == nil {
				if rp.PropertiesStatus == "Online" {
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
	log.Printf("[DEBUG] mssqlDatabaseWaitUntilReady(%s, %s)", tenantID, dbName)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
