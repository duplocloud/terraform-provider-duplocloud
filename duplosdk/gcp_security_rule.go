package duplosdk

import "fmt"

type DuploSecurityRule struct {
	Name            string `json:"Name"`
	Description     string `json:"Description"`
	ToPort          string `json:"ToPort,omitempty"`
	Ports           string `json:"Ports,omitempty"`
	FromPort        string `json:"FromPort,omitempty"`
	ServiceProtocol string `json:"ServiceProtocol"`
	SourceRanges    string `json:"SourceRanges"`
	RuleType        string `json:"RuleType"`
	TargetTenantId  string `json:"TargetSubscriptionId,omitempty"`
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

func (c *Client) GcpSecurityRuleGet(scopeName, ruleName string, tenantSide bool) (*DuploSecurityRule, ClientError) {
	rp := DuploSecurityRule{}
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
	patch := "infra/" + scopeName
	if tenantSide {
		patch = "tenant/" + scopeName
	}
	err := c.postAPI(
		fmt.Sprintf("GcpSecurityRuleUpdate(%s, %s)", scopeName, rq.Name),
		fmt.Sprintf("v3/admin/google/sgrule/%s/%s", patch, rq.Name),
		&rq,
		nil,
	)
	return err
}
