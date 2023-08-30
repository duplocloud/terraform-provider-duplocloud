package duplosdk

import "fmt"

type DuploCloudWatchEventRule struct {
	Name               string                 `json:"Name"`
	Description        string                 `json:"Description,omitempty"`
	ScheduleExpression string                 `json:"ScheduleExpression,omitempty"`
	State              string                 `json:"State"`
	Tags               *[]DuploKeyStringValue `json:"Tags,omitempty"`
	EventBusName       string                 `json:"EventBusName,omitempty"`
	RoleArn            string                 `json:"RoleArn,omitempty"`
	EventPattern       string                 `json:"EventPattern,omitempty"`
}

type DuploCloudWatchEventRuleGetReq struct {
	Name               string            `json:"Name"`
	Description        string            `json:"Description,omitempty"`
	ScheduleExpression string            `json:"ScheduleExpression,omitempty"`
	EventBusName       string            `json:"EventBusName,omitempty"`
	RoleArn            string            `json:"RoleArn,omitempty"`
	Arn                string            `json:"Arn,omitempty"`
	State              *DuploStringValue `json:"State,omitempty"`
	EventPattern       string            `json:"EventPattern,omitempty"`
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

type DuploCloudWatchMetricAlarm struct {
	Statistic          string                  `json:"Statistic,omitempty"`
	MetricName         string                  `json:"MetricName,omitempty"`
	ComparisonOperator string                  `json:"ComparisonOperator,omitempty"`
	Threshold          float64                 `json:"Threshold,omitempty"`
	Period             int                     `json:"Period,omitempty"`
	EvaluationPeriods  int                     `json:"EvaluationPeriods,omitempty"`
	TenantId           string                  `json:"TenantId,omitempty"`
	Namespace          string                  `json:"Namespace,omitempty"`
	State              string                  `json:"State,omitempty"`
	Dimensions         *[]DuploNameStringValue `json:"Dimensions,omitempty"`
	AccountName        string                  `json:"AccountName,omitempty"`
	Name               string                  `json:"Name,omitempty"`
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

func (c *Client) DuploCloudWatchMetricAlarmCreate(rq *DuploCloudWatchMetricAlarm) ClientError {
	var rp = ""
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchMetricAlarmCreate(%s, %s)", rq.TenantId, rq.MetricName),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateAlarm", rq.TenantId),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) DuploCloudWatchMetricAlarmGet(tenantID, resourceId string) (*DuploCloudWatchMetricAlarm, ClientError) {
	rp := []DuploCloudWatchMetricAlarm{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchMetricAlarmList(%s, %s)", tenantID, resourceId),
		fmt.Sprintf("subscriptions/%s/%s/GetAlarms", tenantID, EncodePathParam(resourceId)),
		&rp,
	)
	if len(rp) == 0 {
		return nil, err
	}
	return &rp[0], err
}

func (c *Client) DuploCloudWatchMetricAlarmDelete(tenantId, fullName string) ClientError {
	var rp = ""
	rq := DuploCloudWatchMetricAlarm{
		Name:  fullName,
		State: "Delete",
	}
	err := c.postAPI(
		fmt.Sprintf("DuploCloudWatchMetricAlarmCreate(%s, %s)", tenantId, rq.MetricName),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateAlarm", tenantId),
		&rq,
		&rp,
	)
	return err
}
