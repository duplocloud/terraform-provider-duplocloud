package duplosdk

import "fmt"

type DuploCloudWatchEventRule struct {
	Name               string                 `json:"Name"`
	Description        string                 `json:"Description,omitempty"`
	ScheduleExpression string                 `json:"ScheduleExpression"`
	State              string                 `json:"State"`
	Tags               *[]DuploKeyStringValue `json:"Tags,omitempty"`
	EventBusName       string                 `json:"EventBusName,omitempty"`
	RoleArn            string                 `json:"RoleArn,omitempty"`
}

type DuploCloudWatchEventRuleGetReq struct {
	Name               string            `json:"Name"`
	Description        string            `json:"Description,omitempty"`
	ScheduleExpression string            `json:"ScheduleExpression"`
	EventBusName       string            `json:"EventBusName,omitempty"`
	RoleArn            string            `json:"RoleArn,omitempty"`
	Arn                string            `json:"Arn,omitempty"`
	State              *DuploStringValue `json:"State,omitempty"`
}

type DuploCloudWatchEventTargets struct {
	Rule         string                        `json:"Rule"`
	Targets      *[]DuploCloudWatchEventTarget `json:"Targets,omitempty"`
	EventBusName string                        `json:"EventBusName,omitempty"`
}

type DuploCloudWatchEventTargetsDeleteReq struct {
	Rule string   `json:"Rule,omitempty"`
	Ids  []string `json:"Ids,omitempty"`
}

type DuploCloudWatchEventTarget struct {
	Arn     string `json:"Arn"`
	Id      string `json:"Id,omitempty"`
	RoleArn string `json:"RoleArn,omitempty"`
	Input   string `json:"Input,omitempty"`
}

type DuploCloudWatchRunCommandTarget struct {
	Key    string   `json:"Key,omitempty"`
	Values []string `json:"Values,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

func (c *Client) DuploCloudWatchEventRuleCreate(tenantID string, rq *DuploCloudWatchEventRule) (*DuploCloudWatchEventRule, ClientError) {
	rp := DuploCloudWatchEventRule{}
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchEventRuleCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateAwsEventsRule", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventRuleDelete(tenantID string, ruleFullName string) (*map[string]interface{}, ClientError) {
	rq := map[string]string{"Name": ruleFullName}
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchEventRuleDelete(%s, %s)", tenantID, ruleFullName),
		fmt.Sprintf("subscriptions/%s/DeleteAwsEventsRule", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventRuleList(tenantID string) (*[]DuploCloudWatchEventRuleGetReq, ClientError) {
	rp := []DuploCloudWatchEventRuleGetReq{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchEventRuleList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAwsEventRules", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventRuleGet(tenantID string, ruleName string) (*DuploCloudWatchEventRuleGetReq, ClientError) {
	list, err := c.DuploCloudWatchEventRuleList(tenantID)

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
	return nil, err
}

func (c *Client) DuploCloudWatchEventTargetsCreate(tenantID string, rq *DuploCloudWatchEventTargets) (*map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchEventTargetCreate(%s, %s)", tenantID, rq.Rule),
		fmt.Sprintf("v3/subscriptions/%s/aws/eventTargets", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventTargetsDelete(tenantID string, rq DuploCloudWatchEventTargetsDeleteReq) (*map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchEventTargetsDelete(%s, %s)", tenantID, rq.Rule),
		fmt.Sprintf("v3/subscriptions/%s/aws/deleteEventTargets", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventTargetsList(tenantID string, ruleName string) (*[]DuploCloudWatchEventTarget, ClientError) {
	rp := []DuploCloudWatchEventTarget{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchEventTargetsList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/eventTargets/%s", tenantID, ruleName),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventTargetGet(tenantID, ruleName, targetId string) (*DuploCloudWatchEventTarget, ClientError) {
	list, err := c.DuploCloudWatchEventTargetsList(tenantID, ruleName)

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, target := range *list {
			if target.Id == targetId {
				return &target, nil
			}
		}
	}
	return nil, err
}
