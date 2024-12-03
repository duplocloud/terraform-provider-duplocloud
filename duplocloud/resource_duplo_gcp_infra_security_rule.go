package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
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
			Description: "Specify rule name",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "Duplocloud prefixed rule name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: "The description related to the rule",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"ports_and_protocols": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ports": {
						Description: "The list of ports to which this rule applies. This field is only applicable for UDP or TCP protocol.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"service_protocol": {
						Description: "The IP protocol to which this rule applies. The protocol type is required when creating a firewall rule. This value can either be one of the following well known protocol strings (tcp, udp, icmp, esp, ah, sctp, ipip, all), or the IP protocol number.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.Any(
							validation.StringInSlice([]string{"tcp", "udp", "icmp", "esp", "ah", "sctp", "ipip", "all"}, false),
							validation.StringMatch(regexp.MustCompile(`^((25[0-5])|(2[0-4][0-9])|([0-1]?[0-9][0-9]?)$)`), "exceptable value should be between 0-255"),
						),
					},
				},
			},
		},

		"source_ranges": {
			Description: "The lists of IPv4 or IPv6 addresses in CIDR format that specify the source of traffic for a firewall rule",
			Type:        schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Required: true,
		},
		"rule_type": {
			Description:  "Specify type of access rule (ALLOW , DENY)",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DENY"}, false),
			ForceNew:     true,
		},
		"direction": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"priority": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"kind": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"self_link": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"source_tags": {
			Type:     schema.TypeList,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Computed: true,
		},
	}
}

func infraSecurityRuleSchema() map[string]*schema.Schema {
	mp := schemaSecurityRule()
	mp["infra_name"] = &schema.Schema{
		Description: "The name of the infrastructure where rule gets applied",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	}
	return mp
}

// Resource for managing an infrastructure's settings.
func resourceGCPInfraSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_infra_security_rule` applies gcp security rule to  infra",

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

		Schema:        infraSecurityRuleSchema(),
		CustomizeDiff: validateGCPSecurityRuleAttribute,
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
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleUpdate(%s): start", infraName)
	rq, err := expandGCPSecurityRule(d)
	if err != nil {
		return diag.Errorf("resourceGcpInfraSecurityRuleUpdate cannot update security rule for infra %s error: %s", infraName, err.Error())
	}

	c := m.(*duplosdk.Client)
	rq.Name = d.Get("fullname").(string)

	err = c.GcpSecurityRuleUpdate(infraName, rq, false)
	if err != nil {
		return diag.Errorf("GcpSecurityRuleUpdate cannot update security rule for infra %s error: %s", infraName, err.Error())
	}
	d.SetId(infraName + "/security-rule/" + rq.Name)

	diags := resourceGcpInfraSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpInfraSecurityRuleUpdate(%s): end", infraName)
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
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		RuleType:    d.Get("rule_type").(string),
	}
	pps := []duplosdk.DuploSecurityRuleProtocolAndPorts{}

	for _, spp := range d.Get("ports_and_protocols").([]interface{}) {
		pp := duplosdk.DuploSecurityRuleProtocolAndPorts{}

		mpp := spp.(map[string]interface{})
		for _, p := range mpp["ports"].([]interface{}) {
			pp.Ports = append(pp.Ports, p.(string))
		}
		pp.ServiceProtocol = mpp["service_protocol"].(string)
		pps = append(pps, pp)

	}
	rq.ProtocolAndPorts = pps
	for _, sr := range d.Get("source_ranges").([]interface{}) {
		rq.SourceRanges = append(rq.SourceRanges, sr.(string))
	}
	return &rq, nil
}

func flattenGCPSecurityRule(d *schema.ResourceData, rb duplosdk.DuploSecurityRuleResponse) {
	d.Set("fullname", rb.Name)
	d.Set("description", rb.Description)

	if len(rb.Allowed) > 0 {
		ppI := make([]interface{}, 0, len(rb.Allowed))
		d.Set("rule_type", "ALLOW")
		for _, v := range rb.Allowed {
			mp := make(map[string]interface{})
			mp["ports"] = v.Ports
			mp["service_protocol"] = v.ServiceProtocol
			ppI = append(ppI, mp)
		}
		d.Set("ports_and_protocols", ppI)

	}
	if len(rb.Denied) > 0 {
		ppI := make([]interface{}, 0, len(rb.Allowed))
		d.Set("rule_type", "DENY")
		for _, v := range rb.Allowed {
			mp := make(map[string]interface{})
			mp["ports"] = v.Ports
			mp["service_protocol"] = v.ServiceProtocol
			ppI = append(ppI, mp)
		}
		d.Set("ports_and_protocols", ppI)
	}
	if len(rb.SourceServiceAccounts) > 0 {
		d.Set("source_service_account", rb.SourceServiceAccounts)
	}
	if len(rb.TargetServiceAccounts) > 0 {
		d.Set("target_service_account", rb.TargetServiceAccounts)
	}
	d.Set("source_ranges", rb.SourceRanges)
	d.Set("priority", rb.Priority)
	d.Set("kind", rb.Kind)
	d.Set("network", rb.Network)
	d.Set("self_link", rb.SelfLink)
	d.Set("source_tags", rb.SourceTags)

}

func validateGCPSecurityRuleAttribute(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	srs := diff.Get("source_ranges").([]interface{})
	dup := map[string]struct{}{}
	for _, v := range srs {
		if _, ok := dup[v.(string)]; ok {
			return fmt.Errorf("duplicate value in source_ranges not allowed")
		}
		dup[v.(string)] = struct{}{}
	}
	return nil
}
