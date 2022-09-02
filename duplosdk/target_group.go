package duplosdk

import "fmt"

type DuploTargetGroupMatcher struct {
	HTTPCode string `json:"HttpCode"`
	GRPCCode string `json:"GrpcCode"`
}
type DuploTargetGroup struct {
	HealthCheckEnabled         bool                     `json:"HealthCheckEnabled"`
	HealthCheckIntervalSeconds int                      `json:"HealthCheckIntervalSeconds,omitempty"`
	HealthCheckPath            string                   `json:"HealthCheckPath,omitempty"`
	HealthCheckPort            string                   `json:"HealthCheckPort,omitempty"`
	HealthCheckProtocol        *DuploStringValue        `json:"HealthCheckProtocol,omitempty"`
	HealthCheckTimeoutSeconds  int                      `json:"HealthCheckTimeoutSeconds,omitempty"`
	HealthyThresholdCount      int                      `json:"HealthyThresholdCount,omitempty"`
	IPAddressType              *DuploStringValue        `json:"IpAddressType,omitempty"`
	LoadBalancerArns           []string                 `json:"LoadBalancerArns,omitempty"`
	Matcher                    *DuploTargetGroupMatcher `json:"Matcher,omitempty"`
	Port                       int                      `json:"Port,omitempty"`
	Protocol                   *DuploStringValue        `json:"Protocol,omitempty"`
	ProtocolVersion            string                   `json:"ProtocolVersion,omitempty"`
	TargetGroupArn             string                   `json:"TargetGroupArn,omitempty"`
	TargetGroupName            string                   `json:"TargetGroupName,omitempty"`
	TargetType                 *DuploStringValue        `json:"TargetType"`
	UnhealthyThresholdCount    int                      `json:"UnhealthyThresholdCount,omitempty"`
	VpcID                      string                   `json:"VpcId,omitempty"`
	Name                       string                   `json:"Name,omitempty"`
}

type DuploTargetGroupUpdateReq struct {
	HealthCheckEnabled         bool                     `json:"HealthCheckEnabled"`
	HealthCheckIntervalSeconds int                      `json:"HealthCheckIntervalSeconds,omitempty"`
	HealthCheckPath            string                   `json:"HealthCheckPath,omitempty"`
	HealthCheckPort            string                   `json:"HealthCheckPort,omitempty"`
	HealthCheckProtocol        *DuploStringValue        `json:"HealthCheckProtocol,omitempty"`
	HealthCheckTimeoutSeconds  int                      `json:"HealthCheckTimeoutSeconds,omitempty"`
	HealthyThresholdCount      int                      `json:"HealthyThresholdCount,omitempty"`
	TargetGroupArn             string                   `json:"TargetGroupArn"`
	UnhealthyThresholdCount    int                      `json:"UnhealthyThresholdCount,omitempty"`
	Matcher                    *DuploTargetGroupMatcher `json:"Matcher,omitempty"`
}

func (c *Client) DuploTargetGroupCreate(tenantID string, rq *DuploTargetGroup) ClientError {
	rp := DuploTargetGroup{}
	return c.postAPI(
		fmt.Sprintf("DuploTargetGroupCreate(%s, %s)", tenantID, rq.TargetGroupName),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbTargetGroup", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploTargetGroupUpdate(tenantID, name string, rq *DuploTargetGroupUpdateReq) ClientError {
	rp := DuploTargetGroup{}
	return c.putAPI(
		fmt.Sprintf("DuploTargetGroupUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbTargetGroup", tenantID),
		&rq,
		&rp,
	)
}
func (c *Client) DuploTargetGroupGet(tenantID, name string) (*DuploTargetGroup, ClientError) {
	rp := DuploTargetGroup{}
	err := c.getAPI(
		fmt.Sprintf("DuploTargetGroupGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbTargetGroup/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploTargetGroupDelete(tenantID string, name string) ClientError {
	rp := DuploTargetGroup{}
	return c.deleteAPI(
		fmt.Sprintf("DuploTargetGroupDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbTargetGroup/%s", tenantID, name),
		&rp,
	)
}
