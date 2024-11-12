package duplocloud

import (
	"context"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantSecurityRuleSchema() map[string]*schema.Schema {
	mp := schemaSecurityRule()
	mp["tenant_id"] = &schema.Schema{
		Description: "The GUID of the tenant.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
	mp["target_tenant_id"] = &schema.Schema{
		Description: "The GUID of the tenant to which security rule need to be applied",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
	return mp
}

// Resource for managing an infrastructure's settings.
func resourceGCPTenantSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_tenant_security_rule` applies security rule to gcp tenantId",

		ReadContext:   resourceGcpTenantSecurityRuleRead,
		CreateContext: resourceGcpTenantSecurityRuleCreate,
		UpdateContext: resourceGcpTenantSecurityRuleUpdate,
		DeleteContext: resourceGcpTenantSecurityRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: tenantSecurityRuleSchema(),
	}
}

func resourceGcpTenantSecurityRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tokens := strings.Split(id, "/")
	tenantId, ruleName := tokens[0], tokens[2]

	log.Printf("[TRACE] resourceGcpTenantSecurityRuleRead(%s,%s): start", tenantId, ruleName)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpSecurityRuleGet(tenantId, ruleName, false)
	if err != nil {
		return diag.Errorf("Unable to retrieve rule name '%s' of '%s': %s", ruleName, tenantId, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("tenant_id", tenantId)
	d.Set("target_tenant_id", duplo.TargetTenantId)
	flattenGCPSecurityRule(d, *duplo)

	log.Printf("[TRACE] resourceGcpTenantSecurityRuleRead(%s,%s): end", tenantId, ruleName)
	return nil
}

func resourceGcpTenantSecurityRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceGcpTenantSecurityRuleCreate(%s): start", tenantId)

	rq, err := expandGCPSecurityRule(d)
	if err != nil {
		return diag.Errorf("resourceGcpTenantSecurityRuleCreate cannot create security rule for tenant %s error: %s", tenantId, err.Error())
	}
	if v, ok := d.GetOk("target_tenant_id"); ok {
		rq.TargetTenantId = v.(string)
	}

	c := m.(*duplosdk.Client)

	err = c.GcpSecurityRuleCreate(tenantId, rq, false)
	if err != nil {
		return diag.Errorf("GcpSecurityRuleCreate cannot create security rule for tenant %s error: %s", tenantId, err.Error())
	}
	d.SetId(tenantId + "/security-rule/" + rq.Name)

	diags := resourceGcpInfraSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpTenantSecurityRuleCreate(%s): end", tenantId)
	return diags
}

func resourceGcpTenantSecurityRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	tenantId := tokens[0]
	log.Printf("[TRACE] resourceGcpTenantSecurityRuleCreate(%s): start", tenantId)

	rq, err := expandGCPSecurityRule(d)
	if err != nil {
		return diag.Errorf("resourceGcpTenantSecurityRuleCreate cannot create security rule for tenant %s error: %s", tenantId, err.Error())
	}
	if v, ok := d.GetOk("target_tenant_id"); ok {
		rq.TargetTenantId = v.(string)
	}

	c := m.(*duplosdk.Client)

	err = c.GcpSecurityRuleUpdate(tenantId, rq, false)
	if err != nil {
		return diag.Errorf("GcpSecurityRuleCreate cannot create security rule for tenant %s error: %s", tenantId, err.Error())
	}
	d.SetId(tenantId + "/security-rule/" + rq.Name)

	diags := resourceGcpInfraSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpTenantSecurityRuleCreate(%s): end", tenantId)
	return diags
}

func resourceGcpTenantSecurityRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	tenantId, ruleName := tokens[0], tokens[2]

	log.Printf("[TRACE] resourceGcpTenantSecurityRuleDelete(%s,%s): start", tenantId, ruleName)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	err := c.GcpSecurityRuleDelete(tenantId, ruleName, false)
	if err != nil {
		return diag.Errorf("Unable to delete security rule '%s' for '%s': %s", ruleName, tenantId, err)
	}

	log.Printf("[TRACE] resourceGcpTenantSecurityRuleDelete(%s,%s): end", tenantId, ruleName)
	return nil
}
