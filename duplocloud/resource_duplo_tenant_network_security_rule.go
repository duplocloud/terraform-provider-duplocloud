package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an AWS ElasticSearch instance
func resourceTenantSecurityRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant_network_security_rule` manages a single network single rule for a Duplo tenant.",

		ReadContext:   resourceTenantNetworkSecurityRuleRead,
		CreateContext: resourceTenantNetworkSecurityRuleCreate,
		DeleteContext: resourceTenantNetworkSecurityRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant that the network security rule will be created in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"protocol": {
				Description:  "The network protocol.  Must be one of:  `tcp`, `udp`, `icmp`",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "tcp",
				ValidateFunc: validation.StringInSlice([]string{"tcp", "icmp", "udp"}, false),
			},
			"source_tenant": {
				Description: "The source tenant name (*not* GUID) to allow traffic from. " +
					"Only one of `source_tenant` or `source_address` may be specified.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_address"},
			},
			"source_address": {
				Description: "The source CIDR block to allow traffic from. " +
					"Only one of `source_tenant` or `source_address` may be specified.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_tenant"},
			},
			"from_port": {
				Description:  "The start of a port range to allow traffic to.",
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"to_port": {
				Description:  "The end of a port range to allow traffic to.",
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      65535,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"description": {
				Description: "A description for this rule.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceTenantNetworkSecurityRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleRead(%s): start", id)
	rq, err := duploTenantNetworkSecurityRuleFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetExtConnSecurityGroupRule(rq)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant security group rule '%s': %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("from_port", duplo.FromPort)
	d.Set("to_port", duplo.ToPort)
	d.Set("protocol", duplo.Protocol)
	if rq.Type == duplosdk.SGSourceTypeTenant {
		d.Set("source_address", nil)
		d.Set("source_tenant", (*duplo.Sources)[0].Value)
	} else {
		d.Set("source_address", (*duplo.Sources)[0].Value)
		d.Set("source_tenant", nil)
	}
	if (*duplo.Sources)[0].Description != "" {
		d.Set("description", (*duplo.Sources)[0].Description)
	}

	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleRead(%s): end", id)
	return nil
}

func resourceTenantNetworkSecurityRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Start building the request.
	source := []duplosdk.DuploTenantExtConnSecurityGroupSource{{Description: d.Get("description").(string)}}
	rq := duplosdk.DuploTenantExtConnSecurityGroupRule{
		TenantID: d.Get("tenant_id").(string),
		Protocol: d.Get("protocol").(string),
		FromPort: d.Get("from_port").(int),
		ToPort:   d.Get("to_port").(int),
		Sources:  &source,
	}
	sourceTenant := d.Get("source_tenant").(string)
	sourceAddress := d.Get("source_address").(string)
	if sourceTenant != "" {
		source[0].Type = duplosdk.SGSourceTypeTenant
		source[0].Value = sourceTenant
	} else if sourceAddress != "" {
		source[0].Type = duplosdk.SGSourceTypeIPAddress
		source[0].Value = sourceAddress
	} else {
		return diag.Errorf("Must specify one of: source_tenant, source_address")
	}
	rq.Type = source[0].Type

	// Build the ID
	id := fmt.Sprintf("%s/%d/%s/%s/%d/%d", rq.TenantID, rq.Type, source[0].Value, rq.Protocol, rq.FromPort, rq.ToPort)
	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleCreate(%s): start", id)

	// Create the rule in Duplo.
	c := m.(*duplosdk.Client)
	err := c.TenantUpdateExtConnSecurityGroupRule(&rq)
	if err != nil {
		return diag.Errorf("Error creating tenant network security rule '%s': %s", id, err)
	}
	d.SetId(id)

	diags := resourceTenantNetworkSecurityRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleCreate(%s): end", id)
	return diags
}

func resourceTenantNetworkSecurityRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleRead(%s): start", id)
	rq, err := duploTenantNetworkSecurityRuleFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}
	// Delete the rule with Duplo
	c := m.(*duplosdk.Client)
	err = c.TenantDeleteExtConnSecurityGroupRule(rq)
	if err != nil {
		return diag.Errorf("Error deleting tenant network security rule '%s': %s", id, err)
	}

	log.Printf("[TRACE] resourceTenantNetworkSecurityRuleDelete(%s): end", id)
	return nil
}

func duploTenantNetworkSecurityRuleFromId(id string) (*duplosdk.DuploTenantExtConnSecurityGroupRule, error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) < 6 {
		return nil, fmt.Errorf("invalid resource ID: %s", id)
	}

	ruleType, err := strconv.Atoi(idParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid resource ID: %s: type: %s", id, err)
	}
	fromPort, err := strconv.Atoi(idParts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid resource ID: %s: fromPort: %s", id, err)
	}
	toPort, err := strconv.Atoi(idParts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid resource ID: %s: toPort: %s", id, err)
	}

	return &duplosdk.DuploTenantExtConnSecurityGroupRule{
		TenantID: idParts[0],
		Type:     ruleType,
		Sources:  &[]duplosdk.DuploTenantExtConnSecurityGroupSource{{Type: ruleType, Value: idParts[2]}},
		Protocol: idParts[3],
		FromPort: fromPort,
		ToPort:   toPort,
	}, nil
}
