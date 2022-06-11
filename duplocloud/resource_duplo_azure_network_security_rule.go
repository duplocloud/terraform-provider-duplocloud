package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureNetworkSgRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description:  "The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(2, 12),
		},
		"network_security_group_name": {
			Description: "The name of the Network Security Group that we want to attach the rule to.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"name": {
			Description: "The name of the security group rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The fullname of the security group rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"source_rule_type": {
			Description: "Type of the source security rule. Possible values include `0(IP Address)`, `1(Service Tag)`, `2(Application Security Group)`.",
			Type:        schema.TypeInt,
			Required:    true,
			ValidateFunc: validation.IntInSlice([]int{
				0, 1, 2,
			}),
		},
		"destination_rule_type": {
			Description: "Type of the destination security rule. Possible values include `0(IP Address)`, `1(Service Tag)`, `2(Application Security Group)`.",
			Type:        schema.TypeInt,
			Required:    true,
			ValidateFunc: validation.IntInSlice([]int{
				0, 1, 2,
			}),
		},
		"protocol": {
			Description: "Network protocol this rule applies to. Possible values include `tcp`, `udp`, `icmp`, `esp`, `ah` or `*` (which matches all).",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"tcp", "udp", "*",
			}, false),
		},
		"source_port_range": {
			Description: "Source Port or Range. Integer or range between `0` and `65535` or `*` to match any. ",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"destination_port_range": {
			Description: "Destination Port or Range. Integer or range between `0` and `65535` or `*` to match any.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"source_address_prefix": {
			Description: "CIDR or source IP range or * to match any IP.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"destination_address_prefix": {
			Description: "CIDR or destination IP range or * to match any IP.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"access": {
			Description: "Specifies whether network traffic is allowed or denied. Possible values are `Allow` and `Deny`.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Allow",
				"Deny",
			}, false),
		},
		"priority": {
			Description: "Specifies the priority of the rule.",
			Type:        schema.TypeInt,
			Required:    true,
			//ValidateFunc: validation.IntBetween(100, 4096),
		},
		"direction": {
			Description: "The direction specifies if rule will be evaluated on incoming or outgoing traffic. Possible values are `Inbound` and `Outbound`.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Inbound",
				"Outbound",
			}, false),
		},
	}
}

func resourceAzureNetworkSgRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_network_security_rule` manages an Azure security group rule in Duplo.",

		ReadContext:   resourceAzureNetworkSgRuleRead,
		CreateContext: resourceAzureNetworkSgRuleCreate,
		UpdateContext: resourceAzureNetworkSgRuleUpdate,
		DeleteContext: resourceAzureNetworkSgRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureNetworkSgRuleSchema(),
	}
}

func resourceAzureNetworkSgRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, sgName, ruleName, err := parseAzureNetworkSgRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureNetworkSgRuleRead(%s, %s, %s): start", infraName, sgName, ruleName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.NetworkSgRuleGet(infraName, sgName, ruleName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve infra security group %s rule %s : %s", sgName, ruleName, clientErr)
	}

	flattenAzureNetworkSgRule(infraName, sgName, d, duplo)

	log.Printf("[TRACE] resourceAzureNetworkSgRuleRead(%s, %s,%s): end", infraName, sgName, ruleName)
	return nil
}

func resourceAzureNetworkSgRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	sgName := d.Get("network_security_group_name").(string)
	ruleName := d.Get("name").(string)
	ruleFullName := "duploservices-" + ruleName + "-" + sgName
	log.Printf("[TRACE] resourceAzureNetworkSgRuleCreate(%s, %s, %s): start", infraName, sgName, ruleName)
	c := m.(*duplosdk.Client)

	rq := expandAzureNetworkSgRule(d)
	err = c.NetworkSgRuleCreateOrDelete(&duplosdk.InfrastructureSgUpdate{
		Name:   infraName,
		SgName: sgName,
		RulesToAdd: &[]duplosdk.DuploInfrastructureVnetSGRule{
			*rq,
		},
	})
	if err != nil {
		return diag.Errorf("Error creating security group rule '%s': %s", ruleName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", infraName, sgName, ruleFullName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure security group rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.NetworkSgRuleGet(infraName, sgName, ruleFullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureNetworkSgRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureNetworkSgRuleCreate(%s, %s, %s): end", infraName, sgName, ruleName)
	return diags
}

func resourceAzureNetworkSgRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureNetworkSgRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, sgName, ruleName, err := parseAzureNetworkSgRuleIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureNetworkSgRuleDelete(%s, %s, %s): start", infraName, sgName, ruleName)

	c := m.(*duplosdk.Client)
	clientErr := c.NetworkSgRuleCreateOrDelete(&duplosdk.InfrastructureSgUpdate{
		Name:          infraName,
		SgName:        sgName,
		RulesToRemove: []string{d.Get("fullname").(string)},
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete security group rule '%s': %s", ruleName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure security group rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.NetworkSgRuleGet(infraName, sgName, ruleName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureNetworkSgRuleDelete(%s, %s, %s): end", infraName, sgName, ruleName)
	return nil
}

func expandAzureNetworkSgRule(d *schema.ResourceData) *duplosdk.DuploInfrastructureVnetSGRule {
	return &duplosdk.DuploInfrastructureVnetSGRule{
		Name:                 d.Get("name").(string),
		SrcRuleType:          d.Get("source_rule_type").(int),
		DstRuleType:          d.Get("destination_rule_type").(int),
		Protocol:             d.Get("protocol").(string),
		RuleAction:           d.Get("access").(string),
		Direction:            d.Get("direction").(string),
		Priority:             d.Get("priority").(int),
		SourcePortRange:      d.Get("source_port_range").(string),
		DestinationPortRange: d.Get("destination_port_range").(string),
		SrcAddressPrefix:     d.Get("source_address_prefix").(string),
		DstAddressPrefix:     d.Get("destination_address_prefix").(string),
	}
}

func parseAzureNetworkSgRuleIdParts(id string) (infraName, sgName, ruleName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		infraName, sgName, ruleName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureNetworkSgRule(infraName, sgName string, d *schema.ResourceData, duplo *duplosdk.DuploInfrastructureVnetSGRule) {

	d.Set("infra_name", infraName)
	d.Set("network_security_group_name", sgName)
	parts := strings.Split(duplo.Name, "-"+sgName)
	d.Set("name", parts[0][len("duploservices-"):len(parts[0])])
	d.Set("fullname", duplo.Name)
	d.Set("source_rule_type", duplo.SrcRuleType)
	d.Set("destination_rule_type", duplo.DstRuleType)
	d.Set("protocol", duplo.Protocol)
	d.Set("source_port_range", duplo.SourcePortRange)
	d.Set("destination_port_range", duplo.DestinationPortRange)
	d.Set("source_address_prefix", duplo.SrcAddressPrefix)
	d.Set("destination_address_prefix", duplo.DstAddressPrefix)
	d.Set("access", duplo.RuleAction)
	d.Set("priority", duplo.Priority)
	d.Set("direction", duplo.Direction)
}
