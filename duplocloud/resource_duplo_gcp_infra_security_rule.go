package duplocloud

import (
	"context"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func schemaSecurityRule() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"ports": {
			Description:   "",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"to_port", "from_port"},
		},
		"to_port": {
			Description:   "",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"ports"},
		},
		"from_port": {
			Description:   "",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"ports"},
		},
		"service_protocol": {
			Description: "",
			Type:        schema.TypeString,
			Required:    true,
		},
		"source_ranges": {
			Description: "",
			Type:        schema.TypeString,
			Required:    true,
		},
		"rule_type": {
			Description:  "",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DENY"}, false),
		},
	}
}

func infraSecurityRuleSchema() map[string]*schema.Schema {
	mp := schemaSecurityRule()
	mp["infra_name"] = &schema.Schema{
		Description: "The name of the infrastructure where maintenance windows need to be scheduled.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
	return mp
}

// Resource for managing an infrastructure's settings.
func resourceGCPInfraSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_infra_security_rule` applies security rule to gcp infra",

		ReadContext:   resourceGcpInfraSecurityRuleRead,
		CreateContext: resourceGcpInfraSecurityRuleCreate,
		UpdateContext: resourceGcpInfraSecurityRuleUpdate,
		DeleteContext: resourceGcpInfraSecurityRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: infraSecurityRuleSchema(),
	}
}

func resourceGcpInfraSecurityRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tokens := strings.Split(id, "/")
	infraName, ruleName := tokens[0], tokens[2]
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleRead(%s,%s): start", infraName, ruleName)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpSecurityRuleGet(infraName, ruleName, false)
	if err != nil {
		return diag.Errorf("Unable to retrieve rule name '%s' of '%s': %s", ruleName, infraName, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("infra_name", infraName)
	flattenGCPSecurityRule(d, *duplo)

	log.Printf("[TRACE] resourceGcpInfraSecurityRuleRead(%s,%s): end", infraName, ruleName)
	return nil
}

func resourceGcpInfraSecurityRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName := d.Get("infra_name").(string)
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleCreate(%s): start", infraName)

	rq, err := expandGCPSecurityRule(d)
	if err != nil {
		return diag.Errorf("resourceGcpInfraSecurityRuleCreate cannot create security rule for infra %s error: %s", infraName, err.Error())
	}

	c := m.(*duplosdk.Client)

	err = c.GcpSecurityRuleCreate(infraName, rq, false)
	if err != nil {
		return diag.Errorf("GcpSecurityRuleCreate cannot create security rule for infra %s error: %s", infraName, err.Error())
	}
	d.SetId(infraName + "/security-rule/" + rq.Name)

	diags := resourceGcpInfraSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleCreate(%s): end", infraName)
	return diags
}

func resourceGcpInfraSecurityRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	infraName := tokens[0]
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleCreate(%s): start", infraName)

	rq, err := expandGCPSecurityRule(d)
	if err != nil {
		return diag.Errorf("resourceGcpInfraSecurityRuleCreate cannot create security rule for infra %s error: %s", infraName, err.Error())
	}

	c := m.(*duplosdk.Client)

	err = c.GcpSecurityRuleUpdate(infraName, rq, false)
	if err != nil {
		return diag.Errorf("GcpSecurityRuleCreate cannot create security rule for infra %s error: %s", infraName, err.Error())
	}
	d.SetId(infraName + "/security-rule/" + rq.Name)

	diags := resourceGcpInfraSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleCreate(%s): end", infraName)
	return diags
}

func resourceGcpInfraSecurityRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	infraName, ruleName := tokens[0], tokens[2]

	log.Printf("[TRACE] resourceGcpInfraSecurityRuleDelete(%s,%s): start", infraName, ruleName)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	err := c.GcpSecurityRuleDelete(infraName, ruleName, false)
	if err != nil {
		return diag.Errorf("Unable to delete security rule '%s' for '%s': %s", ruleName, infraName, err)
	}

	log.Printf("[TRACE] resourceGcpInfraSecurityRuleDelete(%s,%s): end", infraName, ruleName)
	return nil
}

func expandGCPSecurityRule(d *schema.ResourceData) (*duplosdk.DuploSecurityRule, error) {
	rq := duplosdk.DuploSecurityRule{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		ServiceProtocol: d.Get("service_protocol").(string),
		RuleType:        d.Get("rule_type").(string),
		SourceRanges:    d.Get("source_ranges").(string),
	}
	if v, ok := d.GetOk("to_port"); ok {
		rq.ToPort = v.(string)
	}
	if v, ok := d.GetOk("from_port"); ok {
		rq.FromPort = v.(string)
	}
	if v, ok := d.GetOk("ports"); ok {
		rq.Ports = v.(string)
	}
	return &rq, nil
}

func flattenGCPSecurityRule(d *schema.ResourceData, rb duplosdk.DuploSecurityRule) {
	d.Set("name", rb.Name)
	d.Set("description", rb.Description)
	d.Set("service_protocol", rb.ServiceProtocol)
	d.Set("source_ranges", rb.SourceRanges)
	d.Set("rule_type", rb.RuleType)
	if rb.ToPort != "" {
		d.Set("to_port", rb.ToPort)
	}
	if rb.FromPort != "" {
		d.Set("from_port", rb.FromPort)
	}
	if rb.Ports != "" {
		d.Set("ports", rb.FromPort)
	}
	if rb.TargetTenantId != "" {
		d.Set("", rb.TargetTenantId)
	}

}
