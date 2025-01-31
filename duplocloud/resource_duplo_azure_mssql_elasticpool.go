package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureMssqlElasticPoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure mssql elastic pool will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"elastic_pool_id": {
			Description: "The ID of the MS SQL Elastic Pool.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the MS SQL elastic pool.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"server_name": {
			Description: "The name of the SQL Server on which to create the elastic pool.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"sku": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
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

func resourceAzureMssqlElasticPool() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_mssql_elasticpool` manages an azure mssql elastic pool in Duplo.",

		ReadContext:   resourceAzureMssqlElasticPoolRead,
		CreateContext: resourceAzureMssqlElasticPoolCreate,
		UpdateContext: resourceAzureMssqlElasticPoolUpdate,
		DeleteContext: resourceAzureMssqlElasticPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureMssqlElasticPoolSchema(),
	}
}

func resourceAzureMssqlElasticPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, epName, err := parseAzureMssqlElasticPoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlElasticPoolRead(%s, %s, %s): start", tenantID, serverName, epName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MsSqlElasticPoolGet(tenantID, serverName, epName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure mssql elastic pool %s : %s", tenantID, epName, clientErr)
	}

	err = flattenAzureMssqlElasticPool(d, duplo)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", epName)
	d.Set("server_name", serverName)
	log.Printf("[TRACE] resourceAzureMssqlElasticPoolRead(%s, %s): end", tenantID, epName)
	return nil
}

func resourceAzureMssqlElasticPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	epName := d.Get("name").(string)
	serverName := d.Get("server_name").(string)
	log.Printf("[TRACE] resourceAzureMssqlElasticPoolCreate(%s, %s): start", tenantID, epName)
	c := m.(*duplosdk.Client)

	rq := expandAzureMssqlElasticPool(d)
	err = c.MsSqlElasticPoolCreate(tenantID, serverName, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure mssql ElasticPool '%s': %s", tenantID, epName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, serverName, epName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure mssql elastic pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlElasticPoolGet(tenantID, serverName, epName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mssqlElasticPoolWaitUntilReady(ctx, c, tenantID, serverName, epName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureMssqlElasticPoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureMssqlElasticPoolCreate(%s, %s): end", tenantID, epName)
	return diags
}

func resourceAzureMssqlElasticPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureMssqlElasticPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, epName, err := parseAzureMssqlElasticPoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureMssqlElasticPoolDelete(%s, %s, %s): start", tenantID, serverName, epName)

	c := m.(*duplosdk.Client)
	clientErr := c.MsSqlElasticPoolDelete(tenantID, serverName, epName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure mssql ElasticPool '%s': %s", tenantID, epName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure mssql elastic pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.MsSqlElasticPoolGet(tenantID, serverName, epName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureMssqlElasticPoolDelete(%s, %s, %s): end", tenantID, serverName, epName)
	return nil
}

func expandAzureMssqlElasticPool(d *schema.ResourceData) *duplosdk.DuploAzureMsSqlElasticPoolRequest {
	return &duplosdk.DuploAzureMsSqlElasticPoolRequest{
		Name: d.Get("name").(string),
		Sku:  expandAzureMssqlElasticPoolSku(d),
	}
}

func expandAzureMssqlElasticPoolSku(d *schema.ResourceData) *duplosdk.DuploAzureMsSqlDatabaseSku {
	skuConfig := d.Get("sku").([]interface{})
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

func parseAzureMssqlElasticPoolIdParts(id string) (tenantID, serverName, epName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, serverName, epName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureMssqlElasticPool(d *schema.ResourceData, duplo *duplosdk.DuploAzureMsSqlElasticPool) error {
	d.Set("name", duplo.Name)
	d.Set("elastic_pool_id", duplo.ID)

	if err := d.Set("sku", flattenAzureMssqlElasticPoolSku(duplo.Sku)); err != nil {
		return fmt.Errorf("[DEBUG] setting `sku`: %#v", err)
	}
	return nil
}

func flattenAzureMssqlElasticPoolSku(sku *duplosdk.DuploAzureMsSqlDatabaseSku) []interface{} {
	result := make(map[string]interface{})
	result["name"] = sku.Name
	result["capacity"] = sku.Capacity

	if sku.Tier != "" {
		result["tier"] = sku.Tier
	}

	return []interface{}{result}
}

func mssqlElasticPoolWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, serverName string, epName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.MsSqlElasticPoolGet(tenantID, serverName, epName)
			log.Printf("[TRACE] MS Sql ElasticPool state is (%s).", rp.PropertiesState)
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
	log.Printf("[DEBUG] mssqlElasticPoolWaitUntilReady(%s, %s)", tenantID, epName)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
