package duplosdk

import (
	"fmt"
)

type DuploDuploAwsAutoscalingSuspendedState struct {
	DynamicScalingInSuspended  bool `json:"DynamicScalingInSuspended,omitempty"`
	DynamicScalingOutSuspended bool `json:"DynamicScalingOutSuspended,omitempty"`
	ScheduledScalingSuspended  bool `json:"ScheduledScalingSuspended,omitempty"`
}

type DuploDuploAwsAutoscalingTargetGetReq struct {
	ServiceNamespace  string   `json:"ServiceNamespace,omitempty"`
	ScalableDimension string   `json:"ScalableDimension,omitempty"`
	ResourceIds       []string `json:"ResourceIds,omitempty"`
}

type DuploDuploAwsAutoscalingTarget struct {
	MaxCapacity       int                                     `json:"MaxCapacity,omitempty"`
	MinCapacity       int                                     `json:"MinCapacity,omitempty"`
	ResourceId        string                                  `json:"ResourceId,omitempty"`
	RoleARN           string                                  `json:"RoleARN,omitempty"`
	ScalableDimension *DuploStringValue                       `json:"ScalableDimension,omitempty"`
	ServiceNamespace  *DuploStringValue                       `json:"ServiceNamespace,omitempty"`
	SuspendedState    *DuploDuploAwsAutoscalingSuspendedState `json:"SuspendedState,omitempty"`
}

func (c *Client) DuploAwsAutoscalingTargetCreate(tenantID string, rq *DuploDuploAwsAutoscalingTarget) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetCreate(%s, %s)", tenantID, rq.ResourceId),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateScalableTarget", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) DuploAwsAutoscalingTargetGet(tenantID string, rq DuploDuploAwsAutoscalingTargetGetReq) (*DuploDuploAwsAutoscalingTarget, ClientError) {
	rp := []DuploDuploAwsAutoscalingTarget{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalableTargets", tenantID),
		&rq,
		&rp,
	)
	return &rp[0], err
}

func (c *Client) DuploAwsAutoscalingTargetList(tenantID string, rq DuploDuploAwsAutoscalingTargetGetReq) (*[]DuploDuploAwsAutoscalingTarget, ClientError) {
	rp := []DuploDuploAwsAutoscalingTarget{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetScalableTargets", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAwsAutoscalingTargetExists(tenantID string, rq DuploDuploAwsAutoscalingTargetGetReq) (bool, ClientError) {
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

func (c *Client) DuploAwsAutoscalingTargetDelete(tenantID string, rq DuploDuploAwsAutoscalingTargetGetReq) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsAutoscalingTargetDelete(%s, %s)", tenantID, rq.ResourceIds[0]),
		fmt.Sprintf("subscriptions/%s/DeleteScalableTarget", tenantID),
		&rq,
		nil,
	)
}
