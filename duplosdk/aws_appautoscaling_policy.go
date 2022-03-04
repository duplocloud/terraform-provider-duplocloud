package duplosdk

import (
	"fmt"
)

type DuploAwsAutoscalingPolicyDeleteReq struct {
	ServiceNamespace  string `json:"ServiceNamespace,omitempty"`
	ScalableDimension string `json:"ScalableDimension,omitempty"`
	ResourceId        string `json:"ResourceId,omitempty"`
	PolicyName        string `json:"PolicyName,omitempty"`
}

type DuploAwsAutoscalingPolicyGetReq struct {
	ServiceNamespace  string   `json:"ServiceNamespace,omitempty"`
	ScalableDimension string   `json:"ScalableDimension,omitempty"`
	ResourceId        string   `json:"ResourceId,omitempty"`
	PolicyNames       []string `json:"PolicyNames,omitempty"`
}

type DuploPredefinedMetricSpecification struct {
	PredefinedMetricType *DuploStringValue `json:"PredefinedMetricType,omitempty"`
	ResourceLabel        string            `json:"ResourceLabel,omitempty"`
}

type DuploCustomizedMetricSpecification struct {
	MetricName string                  `json:"MetricName,omitempty"`
	Namespace  string                  `json:"Namespace,omitempty"`
	Unit       string                  `json:"Unit,omitempty"`
	Statistic  *DuploStringValue       `json:"Statistic,omitempty"`
	Dimensions *[]DuploNameStringValue `json:"Dimensions,omitempty"`
}

type DuploTargetTrackingScalingPolicyConfiguration struct {
	DisableScaleIn                bool                                `json:"DisableScaleIn"`
	CustomizedMetricSpecification *DuploCustomizedMetricSpecification `json:"CustomizedMetricSpecification,omitempty"`
	PredefinedMetricSpecification *DuploPredefinedMetricSpecification `json:"PredefinedMetricSpecification,omitempty"`
	ScaleInCooldown               int                                 `json:"ScaleInCooldown"`
	ScaleOutCooldown              int                                 `json:"ScaleOutCooldown"`
	TargetValue                   float64                             `json:"TargetValue"`
}

type DuploStepAdjustment struct {
	MetricIntervalLowerBound float64 `json:"MetricIntervalLowerBound,omitempty"`
	MetricIntervalUpperBound float64 `json:"MetricIntervalUpperBound"`
	ScalingAdjustment        int     `json:"ScalingAdjustment"`
}

type DuploStepScalingPolicyConfiguration struct {
	AdjustmentType         *DuploStringValue      `json:"AdjustmentType,omitempty"`
	Cooldown               int                    `json:"Cooldown"`
	MetricAggregationType  *DuploStringValue      `json:"MetricAggregationType,omitempty"`
	MinAdjustmentMagnitude int                    `json:"MinAdjustmentMagnitude,omitempty"`
	StepAdjustments        *[]DuploStepAdjustment `json:"StepAdjustments,omitempty"`
}
type DuploAwsAutoscalingPolicy struct {
	Alarms []struct {
		AlarmARN  string `json:"AlarmARN,omitempty"`
		AlarmName string `json:"AlarmName,omitempty"`
	}
	CreationTime                             string                                         `json:"CreationTime,omitempty"`
	PolicyARN                                string                                         `json:"PolicyARN,omitempty"`
	PolicyName                               string                                         `json:"PolicyName,omitempty"`
	PolicyType                               *DuploStringValue                              `json:"PolicyType,omitempty"`
	ResourceId                               string                                         `json:"ResourceId,omitempty"`
	ScalableDimension                        *DuploStringValue                              `json:"ScalableDimension,omitempty"`
	ServiceNamespace                         *DuploStringValue                              `json:"ServiceNamespace,omitempty"`
	TargetTrackingScalingPolicyConfiguration *DuploTargetTrackingScalingPolicyConfiguration `json:"TargetTrackingScalingPolicyConfiguration,omitempty"`
	StepScalingPolicyConfiguration           *DuploStepScalingPolicyConfiguration           `json:"StepScalingPolicyConfiguration,omitempty"`
}

func (c *Client) DuploAwsAutoscalingPolicyCreate(tenantID string, rq *DuploAwsAutoscalingPolicy) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingPolicyCreate(%s, %s)", tenantID, rq.PolicyName),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateScalingPolicy", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) DuploAwsAutoscalingPolicyGet(tenantID string, rq DuploAwsAutoscalingPolicyGetReq) (*DuploAwsAutoscalingPolicy, ClientError) {
	rp := []DuploAwsAutoscalingPolicy{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingPolicyGet(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalingPolicies", tenantID),
		&rq,
		&rp,
	)
	if len(rp) > 0 {
		return &rp[0], err
	}
	return nil, err
}

func (c *Client) DuploAwsAutoscalingPolicyList(tenantID string, rq DuploAwsAutoscalingPolicyGetReq) (*[]DuploAwsAutoscalingPolicy, ClientError) {
	rp := []DuploAwsAutoscalingPolicy{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingPolicyList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalingPolicies", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAwsAutoscalingPolicyExists(tenantID string, rq DuploAwsAutoscalingPolicyGetReq) (bool, ClientError) {
	list, err := c.DuploAwsAutoscalingPolicyList(tenantID, rq)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, element := range *list {
			if element.PolicyName == rq.PolicyNames[0] {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) DuploAwsAutoscalingPolicyDelete(tenantID string, rq DuploAwsAutoscalingPolicyDeleteReq) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingPolicyDelete(%s, %s)", tenantID, rq.PolicyName),
		fmt.Sprintf("subscriptions/%s/DeleteScalingPolicy", tenantID),
		&rq,
		nil,
	)
}
