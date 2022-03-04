package duplosdk

import (
	"fmt"
)

type DuploAwsAutoscalingSuspendedState struct {
	DynamicScalingInSuspended  bool `json:"DynamicScalingInSuspended,omitempty"`
	DynamicScalingOutSuspended bool `json:"DynamicScalingOutSuspended,omitempty"`
	ScheduledScalingSuspended  bool `json:"ScheduledScalingSuspended,omitempty"`
}

type DuploAwsAutoscalingTargetGetReq struct {
	ServiceNamespace  string   `json:"ServiceNamespace,omitempty"`
	ScalableDimension string   `json:"ScalableDimension,omitempty"`
	ResourceIds       []string `json:"ResourceIds,omitempty"`
}

type DuploAwsAutoscalingTargetDeleteReq struct {
	ServiceNamespace  string `json:"ServiceNamespace,omitempty"`
	ScalableDimension string `json:"ScalableDimension,omitempty"`
	ResourceId        string `json:"ResourceId,omitempty"`
}

type DuploAwsAutoscalingTarget struct {
	MaxCapacity       int                                `json:"MaxCapacity,omitempty"`
	MinCapacity       int                                `json:"MinCapacity,omitempty"`
	ResourceId        string                             `json:"ResourceId,omitempty"`
	RoleARN           string                             `json:"RoleARN,omitempty"`
	ScalableDimension *DuploStringValue                  `json:"ScalableDimension,omitempty"`
	ServiceNamespace  *DuploStringValue                  `json:"ServiceNamespace,omitempty"`
	SuspendedState    *DuploAwsAutoscalingSuspendedState `json:"SuspendedState,omitempty"`
}

func (c *Client) DuploAwsAutoscalingTargetCreate(tenantID string, rq *DuploAwsAutoscalingTarget) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetCreate(%s, %s)", tenantID, rq.ResourceId),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateScalableTarget", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) DuploAwsAutoscalingTargetGet(tenantID string, rq DuploAwsAutoscalingTargetGetReq) (*DuploAwsAutoscalingTarget, ClientError) {
	rp := []DuploAwsAutoscalingTarget{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalableTargets", tenantID),
		&rq,
		&rp,
	)
	if len(rp) > 0 {
		return &rp[0], err
	}
	return nil, err
}

func (c *Client) DuploAwsAutoscalingTargetList(tenantID string, rq DuploAwsAutoscalingTargetGetReq) (*[]DuploAwsAutoscalingTarget, ClientError) {
	rp := []DuploAwsAutoscalingTarget{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalableTargets", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAwsAutoscalingTargetExists(tenantID string, rq DuploAwsAutoscalingTargetGetReq) (bool, ClientError) {
	list, err := c.DuploAwsAutoscalingTargetList(tenantID, rq)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, server := range *list {
			if server.ResourceId == rq.ResourceIds[0] {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) DuploAwsAutoscalingTargetDelete(tenantID string, rq DuploAwsAutoscalingTargetDeleteReq) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetDelete(%s, %s)", tenantID, rq.ResourceId),
		fmt.Sprintf("subscriptions/%s/DeleteScalableTarget", tenantID),
		&rq,
		nil,
	)
}
