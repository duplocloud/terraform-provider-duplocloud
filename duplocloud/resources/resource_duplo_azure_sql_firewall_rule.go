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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureSqlFirewallRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the sql firewall rule will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the firewall rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"server_name": {
			Description: "The name of the SQL Server on which to create the Firewall Rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"start_ip_address": {
			Description: "The starting IP address to allow through the firewall for this rule.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"end_ip_address": {
			Description: "The ending IP address to allow through the firewall for this rule.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"azure_id": {
			Description: "The ID of the SQL firewall rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureSqlFirewallRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_sql_firewall_rule` allows you to manage an Azure SQL Firewall Rule.",

		ReadContext:   resourceAzureSqlFirewallRuleRead,
		CreateContext: resourceAzureSqlFirewallRuleCreate,
		UpdateContext: resourceAzureSqlFirewallRuleUpdate,
		DeleteContext: resourceAzureSqlFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureSqlFirewallRuleSchema(),
	}
}

func resourceAzureSqlFirewallRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, ruleName, err := parseAzureSqlFirewallRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlFirewallRuleRead(%s, %s, %s): start", tenantID, serverName, ruleName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureSqlServerFirewallRuleGet(tenantID, serverName, ruleName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure sql firewall rule %s, %s : %s", tenantID, serverName, ruleName, clientErr)
	}

	flattenAzureSqlFirewallRule(d, duplo)
	d.Set("tenant_id", tenantID)
	d.Set("server_name", serverName)
	log.Printf("[TRACE] resourceAzureSqlFirewallRuleRead(%s, %s, %s): end", tenantID, serverName, ruleName)
	return nil
}

func resourceAzureSqlFirewallRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	serverName := d.Get("server_name").(string)
	ruleName := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureSqlFirewallRuleCreate(%s, %s, %s): start", tenantID, serverName, ruleName)
	c := m.(*duplosdk.Client)

	rq := expandAzureSqlFirewallRule(d)
	err = c.AzureSqlServerFirewallRuleCreate(tenantID, serverName, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure sql firewall rule '%s', '%s': %s", tenantID, serverName, ruleName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, serverName, ruleName)
	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "azure sql firewall rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureSqlServerFirewallRuleGet(tenantID, serverName, ruleName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureSqlFirewallRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureSqlFirewallRuleCreate(%s, %s, %s): end", tenantID, serverName, ruleName)
	return diags
}

func resourceAzureSqlFirewallRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureSqlFirewallRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, serverName, ruleName, err := parseAzureSqlFirewallRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlFirewallRuleDelete(%s, %s, %s): start", tenantID, serverName, ruleName)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureSqlServerFirewallRuleDelete(tenantID, serverName, ruleName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure sql firewall rule '%s', '%s': %s", tenantID, serverName, ruleName, clientErr)
	}

	diag := duplocloud.waitForResourceToBeMissingAfterDelete(ctx, d, "azure sql firewall rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureSqlServerFirewallRuleGet(tenantID, serverName, ruleName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureSqlFirewallRuleDelete(%s, %s, %s): end", tenantID, serverName, ruleName)
	return nil
}

func expandAzureSqlFirewallRule(d *schema.ResourceData) *duplosdk.AzureSqlServerFirewallRule {
	return &duplosdk.AzureSqlServerFirewallRule{
		Name:                     d.Get("name").(string),
		PropertiesStartIPAddress: d.Get("start_ip_address").(string),
		PropertiesEndIPAddress:   d.Get("end_ip_address").(string),
	}
}

func parseAzureSqlFirewallRuleIdParts(id string) (tenantID, serverName, ruleName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, serverName, ruleName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureSqlFirewallRule(d *schema.ResourceData, duplo *duplosdk.AzureSqlServerFirewallRule) {
	d.Set("name", duplo.Name)
	d.Set("start_ip_address", duplo.PropertiesStartIPAddress)
	d.Set("end_ip_address", duplo.PropertiesEndIPAddress)
	d.Set("azure_id", duplo.ID)
}
