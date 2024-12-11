package duplosdk

import "fmt"

type DuploSecurityRule struct {
	Name             string                              `json:"Name"`
	Description      string                              `json:"Description"`
	ProtocolAndPorts []DuploSecurityRuleProtocolAndPorts `json:"ProtocolAndPorts"`
	SourceRanges     []string                            `json:"SourceRanges"`
	RuleType         string                              `json:"Type"`
	TargetTenantId   string                              `json:"TargetSubscriptionId,omitempty"`
}
type DuploSecurityRuleProtocolAndPorts struct {
	Ports           []string `json:"Ports,omitempty"`
	ServiceProtocol string   `json:"ServiceProtocol"`
}

type DuploSecurityRuleProtocolAndPortsResponse struct {
	Ports           []string `json:"Ports,omitempty"`
	ServiceProtocol string   `json:"IPProtocol"`
}

type DuploSecurityRuleResponse struct {
	Name    string                                      `json:"name,omitempty"`
	Allowed []DuploSecurityRuleProtocolAndPortsResponse `json:"allowed,omitempty"`
	Denied  []DuploSecurityRuleProtocolAndPortsResponse `json:"denied,omitempty"`

	Description           string   `json:"description,omitempty"`
	Kind                  string   `json:"kind,omitempty"`
	Direction             string   `json:"direction,omitempty"`
	Network               string   `json:"network,omitempty"`
	SelfLink              string   `json:"selfLink,omitempty"`
	SourceTags            []string `json:"sourceTags,omitempty"`
	SourceRanges          []string `json:"sourceRanges,omitempty"`
	Priority              int      `json:"priority,omitempty"`
	TargetServiceAccounts []string `json:"targetServiceAccounts,omitempty"`
	SourceServiceAccounts []string `json:"sourceServiceAccounts,omitempty"`
}

func (c *Client) GcpSecurityRuleCreate(scopeName string, rq *DuploSecurityRule, tenantSide bool) ClientError {
	rp := DuploSecurityRule{}
	patch := "infra/" + scopeName
	if tenantSide {
		patch = "tenant/" + scopeName
	}
	err := c.postAPI(
		fmt.Sprintf("GcpSecurityRuleCreate(%s, %s)", scopeName, rq.Name),
		fmt.Sprintf("v3/admin/google/sgrule/%s", patch),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) GcpSecurityRuleDelete(scopeName, ruleName string, tenantSide bool) ClientError {
	patch := "infra/" + scopeName
	if tenantSide {
		patch = "tenant/" + scopeName
	}

	err := c.deleteAPI(
		fmt.Sprintf("GcpSecurityRuleDelete(%s, %s)", scopeName, ruleName),
		fmt.Sprintf("v3/admin/google/sgrule/%s/%s", patch, ruleName),
		nil)
	return err
}

func (c *Client) GcpSecurityRuleGet(scopeName, ruleName string, tenantSide bool) (*DuploSecurityRuleResponse, ClientError) {
	rp := DuploSecurityRuleResponse{}
	patch := "infra/" + scopeName
	if tenantSide {
		patch = "tenant/" + scopeName
	}

	err := c.getAPI(
		fmt.Sprintf("GcpSecurityRuleGet(%s, %s)", scopeName, ruleName),
		fmt.Sprintf("v3/admin/google/sgrule/%s/%s", patch, ruleName),
		&rp)
	return &rp, err
}

func (c *Client) GcpSecurityRuleUpdate(scopeName string, rq *DuploSecurityRule, tenantSide bool) ClientError {
	rp := DuploSecurityRule{}
	patch := "infra/" + scopeName
	if tenantSide {
		patch = "tenant/" + scopeName
	}
	err := c.postAPI(
		fmt.Sprintf("GcpSecurityRuleUpdate(%s, %s)", scopeName, rq.Name),
		fmt.Sprintf("v3/admin/google/sgrule/%s/%s", patch, rq.Name),
		&rq,
		&rp,
	)
	return err
}
