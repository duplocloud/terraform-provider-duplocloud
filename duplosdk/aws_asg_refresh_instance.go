package duplosdk

import "fmt"

type DuploAwsRefreshInstancesRequest struct {
	AutoScalingGroupName string                                        `json:"AutoScalingGroupName"`
	DesiredConfiguration *DuploAwsRefreshInstancesDesiredConfiguration `json:"DesiredConfiguration,omitempty"`
	Preferences          *DuploRefreshInstancesPreference              `json:"Preferences,omitempty"`
}

type DuploAwsRefreshInstancesDesiredConfiguration struct {
	LaunchTemplate *DuploLaunchTemplate `json:"DuploLaunchTemplate,omitempty"`
}

type DuploLaunchTemplate struct {
	LaunchTemplateName string `json:"LaunchTemplateName"`
	Version            string `json:"Version"`
}

type DuploRefreshInstancesPreference struct {
	AutoRollback         bool `json:"AutoRollback"`
	InstanceWarmup       int  `json:"InstanceWarmup"`
	MaxHealthyPercentage int  `json:"MaxHealthyPercentage"`
	MinHealthyPercentage int  `json:"MinHealthyPercentage"`
}

func (c *Client) CreateAwsRefreshInstances(tenantId string, rq *DuploAwsRefreshInstancesRequest) ClientError {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("CreateAwsRefreshInstances(%s, %s)", tenantId, rq.AutoScalingGroupName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/refreshInstances", tenantId),
		rq,
		&rp,
	)
	return err
}

//func (c *Client) GetAwsRefreshInstances(tenantID string, templateName, asgName string) (*DuploAwsRefreshInstancesRequest, ClientError) {
//	rp := DuploAwsRefreshInstancesRequest{}
//	err := c.getAPI(
//		fmt.Sprintf("GetAwsRefreshInstances(%s, %s)", tenantID, templateName),
//		fmt.Sprintf("/v3/subscriptions/%s/aws/asg/%s/RefreshInstancesversions", tenantID, asgName),
//		&rp)
//	return &rp, err
//}
