package duplosdk

import (
	"fmt"
)

type DuploTargetGroupAttributes struct {
	Attributes     *[]DuploKeyStringValue `json:"Attributes,omitempty"`
	IsEcsLB        bool                   `json:"IsEcsLB,omitempty"`
	IsPassThruLB   bool                   `json:"IsPassThruLB,omitempty"`
	Port           int                    `json:"Port,omitempty"`
	RoleName       string                 `json:"RoleName,omitempty"`
	TargetGroupArn string                 `json:"TargetGroupArn,omitempty"`
}

type DuploTargetGroupAttributesGetReq struct {
	IsEcsLB        bool   `json:"IsEcsLB,omitempty"`
	IsPassThruLB   bool   `json:"IsPassThruLB,omitempty"`
	Port           int    `json:"Port,omitempty"`
	RoleName       string `json:"RoleName,omitempty"`
	TargetGroupArn string `json:"TargetGroupArn,omitempty"`
}

func (c *Client) DuploAwsTargetGroupAttributesCreate(tenantID string, rq *DuploTargetGroupAttributes) ClientError {
	return c.putAPI(
		fmt.Sprintf("TargetGroupAttributesCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/targetGroupAttributes", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) DuploAwsTargetGroupAttributesGet(tenantID string, rq DuploTargetGroupAttributesGetReq) (*[]DuploKeyStringValue, ClientError) {
	rp := []DuploKeyStringValue{}
	err := c.postAPI(
		fmt.Sprintf("TargetGroupAttributesGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/targetGroupAttributes", tenantID),
		&rq,
		&rp,
	)
	filtered := []DuploKeyStringValue{}

	for _, v := range rp {
		if len(v.Key) > 0 && len(v.Value) > 0 {
			filtered = append(filtered, v)
		}
	}
	return &filtered, err
}

func (c *Client) DuploAwsTargetGroupAttributesExists(tenantID string, rq DuploTargetGroupAttributesGetReq) (bool, ClientError) {
	list, err := c.DuploAwsTargetGroupAttributesGet(tenantID, rq)
	if err != nil {
		return false, err
	}

	if list != nil && len(*list) > 0 {
		return true, nil
	}
	return false, nil
}
