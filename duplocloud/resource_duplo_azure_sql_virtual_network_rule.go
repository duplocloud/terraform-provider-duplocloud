package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureSqlServerVnetRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the sql virtual network rule will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the SQL virtual network rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"server_name": {
			Description: "The name of the SQL Server to which this SQL virtual network rule will be applied to.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"subnet_id": {
			Description: "The ID of the subnet that the SQL server will be connected to.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"ignore_missing_vnet_service_endpoint": {
			Description: "Create the virtual network rule before the subnet has the virtual network service endpoint enabled.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until Redis cache instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"azure_id": {
			Description: "The ID of the SQL virtual network rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureSqlServerVnetRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_sql_virtual_network_rule` allows you to add, update, or remove an Azure SQL server to a subnet of a virtual network.",

		ReadContext:   resourceAzureSqlServerVnetRuleRead,
		CreateContext: resourceAzureSqlServerVnetRuleCreate,
		UpdateContext: resourceAzureSqlServerVnetRuleUpdate,
		DeleteContext: resourceAzureSqlServerVnetRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureSqlServerVnetRuleSchema(),
	}
}

func resourceAzureSqlServerVnetRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, ruleName, err := parseAzureSqlServerVnetRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleRead(%s, %s, %s): start", tenantID, serverName, ruleName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureSqlServerVnetRuleGet(tenantID, serverName, ruleName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure sql virtual network rule %s, %s : %s", tenantID, serverName, ruleName, clientErr)
	}

	flattenAzureSqlServerVnetRule(d, duplo)
	d.Set("tenant_id", tenantID)
	d.Set("server_name", serverName)
	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleRead(%s, %s, %s): end", tenantID, serverName, ruleName)
	return nil
}

func resourceAzureSqlServerVnetRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	serverName := d.Get("server_name").(string)
	ruleName := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleCreate(%s, %s, %s): start", tenantID, serverName, ruleName)
	c := m.(*duplosdk.Client)

	rq := expandAzureSqlServerVnetRule(d)

	err = c.AzureSqlServerVnetRuleCreate(tenantID, serverName, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure sql virtual network rule '%s', '%s': %s", tenantID, serverName, ruleName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, serverName, ruleName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure sql virtual network rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureSqlServerVnetRuleGet(tenantID, serverName, ruleName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = sqlServerVnetRuleeWaitUntilReady(ctx, c, tenantID, serverName, ruleName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureSqlServerVnetRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleCreate(%s, %s, %s): end", tenantID, serverName, ruleName)
	return diags
}

func resourceAzureSqlServerVnetRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureSqlServerVnetRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, ruleName, err := parseAzureSqlServerVnetRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleDelete(%s, %s, %s): start", tenantID, serverName, ruleName)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureSqlServerVnetRuleDelete(tenantID, serverName, ruleName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure sql virtual network rule '%s', '%s': %s", tenantID, serverName, ruleName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure sql virtual network rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureSqlServerVnetRuleGet(tenantID, serverName, ruleName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureSqlServerVnetRuleDelete(%s, %s, %s): end", tenantID, serverName, ruleName)
	return nil
}

func expandAzureSqlServerVnetRule(d *schema.ResourceData) *duplosdk.AzureSqlServerVnetRule {
	return &duplosdk.AzureSqlServerVnetRule{
		PropertiesVirtualNetworkSubnetID: d.Get("subnet_id").(string),
		Name:                             d.Get("name").(string),
		PropertiesIgnoreMissingVnetServiceEndpoint: d.Get("ignore_missing_vnet_service_endpoint").(bool),
	}
}

func parseAzureSqlServerVnetRuleIdParts(id string) (tenantID, serverName, ruleName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, serverName, ruleName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureSqlServerVnetRule(d *schema.ResourceData, duplo *duplosdk.AzureSqlServerVnetRule) {
	d.Set("name", duplo.Name)
	d.Set("subnet_id", duplo.PropertiesVirtualNetworkSubnetID)
	d.Set("ignore_missing_vnet_service_endpoint", duplo.PropertiesIgnoreMissingVnetServiceEndpoint)
	d.Set("azure_id", duplo.ID)
}

func sqlServerVnetRuleeWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, serverName string, ruleName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AzureSqlServerVnetRuleGet(tenantID, serverName, ruleName)
			log.Printf("[TRACE] Sql virtual network rule provisioning state is (%s).", rp.State)
			status := "pending"
			if err == nil {
				if rp.State == "Ready" || rp.State == "Failed" {
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
	log.Printf("[DEBUG] sqlServerVnetRuleeWaitUntilReady(%s, %s, %s)", tenantID, serverName, ruleName)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
