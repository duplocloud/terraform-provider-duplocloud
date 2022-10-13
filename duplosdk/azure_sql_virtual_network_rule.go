package duplosdk

import "fmt"

type AzureSqlServerVnetRule struct {
	PropertiesVirtualNetworkSubnetID           string `json:"properties.virtualNetworkSubnetId"`
	PropertiesIgnoreMissingVnetServiceEndpoint bool   `json:"properties.ignoreMissingVnetServiceEndpoint"`
	State                                      string `json:"properties.state"`
	ID                                         string `json:"id"`
	Name                                       string `json:"name"`
	Type                                       string `json:"type"`
}

func (c *Client) AzureSqlServerVnetRuleCreate(tenantID string, name string, rq *AzureSqlServerVnetRule) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureSqlServerVnetRuleCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/AddSqlServerVirtualNetworkRule/%s", tenantID, name),
		&rq,
		nil,
	)
}

func (c *Client) AzureSqlServerVnetRuleGet(tenantID string, serverName string, ruleName string) (*AzureSqlServerVnetRule, ClientError) {
	list, err := c.AzureSqlServerVnetRuleList(tenantID, serverName)
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

func (c *Client) AzureSqlServerVnetRuleList(tenantID, serverName string) (*[]AzureSqlServerVnetRule, ClientError) {
	rp := []AzureSqlServerVnetRule{}
	err := c.getAPI(
		fmt.Sprintf("AzureSqlServerVnetRuleList(%s, %s)", tenantID, serverName),
		fmt.Sprintf("subscriptions/%s/GetSqlServerVirtualNetworkRules/%s", tenantID, serverName),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureSqlServerVnetRuleDelete(tenantID string, serverName string, ruleName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureSqlServerVnetRuleDelete(%s, %s)", tenantID, serverName),
		fmt.Sprintf("subscriptions/%s/DeleteSqlServerVirtualNetworkRule/%s", tenantID, serverName),
		&AzureSqlServerVnetRule{Name: ruleName},
		nil,
	)
}
