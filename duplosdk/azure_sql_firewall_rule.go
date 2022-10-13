package duplosdk

import "fmt"

type AzureSqlServerFirewallRule struct {
	Kind                     string `json:"kind"`
	Location                 string `json:"location"`
	PropertiesStartIPAddress string `json:"properties.startIpAddress"`
	PropertiesEndIPAddress   string `json:"properties.endIpAddress"`
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Type                     string `json:"type"`
}

func (c *Client) AzureSqlServerFirewallRuleCreate(tenantID string, name string, rq *AzureSqlServerFirewallRule) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureSqlServerFirewallRuleCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/AddSqlServerFirewallRule/%s", tenantID, name),
		&rq,
		nil,
	)
}

func (c *Client) AzureSqlServerFirewallRuleGet(tenantID string, serverName string, ruleName string) (*AzureSqlServerFirewallRule, ClientError) {
	list, err := c.AzureSqlServerFirewallRuleList(tenantID, serverName)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, rule := range *list {
			if rule.Name == ruleName {
				return &rule, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureSqlServerFirewallRuleList(tenantID, serverName string) (*[]AzureSqlServerFirewallRule, ClientError) {
	rp := []AzureSqlServerFirewallRule{}
	err := c.getAPI(
		fmt.Sprintf("AzureSqlServerFirewallRuleList(%s, %s)", tenantID, serverName),
		fmt.Sprintf("subscriptions/%s/GetSqlServerFirewallRules/%s", tenantID, serverName),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureSqlServerFirewallRuleDelete(tenantID string, serverName string, ruleName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureSqlServerFirewallRuleDelete(%s, %s, %s)", tenantID, serverName, ruleName),
		fmt.Sprintf("subscriptions/%s/DeleteSqlServerFirewallRule/%s", tenantID, serverName),
		&AzureSqlServerFirewallRule{
			Name: ruleName,
		},
		nil,
	)
}
