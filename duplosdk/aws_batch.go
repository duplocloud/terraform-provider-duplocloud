package duplosdk

import "fmt"

//  --------------- Scheduling Policies ---------------

type DuploAwsBatchSchedulingPolicy struct {
	FairsharePolicy *DuploAwsFairsharePolicy `json:"FairsharePolicy,omitempty"`
	Name            string                   `json:"Name"`
	Arn             string                   `json:"Arn,omitempty"`
	Tags            map[string]string        `json:"Tags,omitempty"`
}

type DuploAwsFairsharePolicy struct {
	ShareDecaySeconds  int                                         `json:"ShareDecaySeconds,omitempty"`
	ComputeReservation int                                         `json:"ComputeReservation,omitempty"`
	ShareDistribution  *[]DuploAwsFairsharePolicyShareDistribution `json:"ShareDistribution,omitempty"`
}

type DuploAwsFairsharePolicyShareDistribution struct {
	ShareIdentifier string  `json:"ShareIdentifier,omitempty"`
	WeightFactor    float64 `json:"WeightFactor,omitempty"`
}

func (c *Client) AwsBatchSchedulingPolicyCreate(tenantID string, rq *DuploAwsBatchSchedulingPolicy) ClientError {
	rp := ""
	return c.postAPI(
		fmt.Sprintf("AwsBatchSchedulingPolicyCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchSchedulingPolicy", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchSchedulingPolicyUpdate(tenantID string, rq *DuploAwsBatchSchedulingPolicy) ClientError {
	rp := ""
	return c.putAPI(
		fmt.Sprintf("AwsBatchSchedulingPolicyCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchSchedulingPolicy", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchSchedulingPolicyGet(tenantID string, name string) (*DuploAwsBatchSchedulingPolicy, ClientError) {
	list, err := c.AwsBatchSchedulingPolicyList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Name == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AwsBatchSchedulingPolicyList(tenantID string) (*[]DuploAwsBatchSchedulingPolicy, ClientError) {
	rp := []DuploAwsBatchSchedulingPolicy{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchSchedulingPolicyList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchSchedulingPolicy", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsBatchSchedulingPolicyDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchSchedulingPolicyDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchSchedulingPolicy/%s", tenantID, name),
		nil,
	)
}

//  --------------- Compute Environment ---------------

type DuploAwsBatchComputeEnvironment struct {
	ComputeResources       DuploAwsBatchComputeResource `json:"ComputeResources,omitempty"`
	ComputeEnvironmentName string                       `json:"ComputeEnvironmentName"`
}

type DuploAwsBatchComputeResource struct {
	Ec2Configuration   *[]DuploAwsBatchComputeEc2Configuration `json:"Ec2Configuration,omitempty"`
	Type               string                                  `json:"Type,omitempty"`
	AllocationStrategy string                                  `json:"AllocationStrategy,omitempty"`
	MaxvCpus           int                                     `json:"MaxvCpus,omitempty"`
	MinvCpus           int                                     `json:"MinvCpus,omitempty"`
	DesiredvCpus       int                                     `json:"DesiredvCpus,omitempty"`
	BidPercentage      int                                     `json:"BidPercentage,omitempty"`
	InstanceTypes      *[]string                               `json:"InstanceTypes,omitempty"`
}

type DuploAwsBatchComputeEc2Configuration struct {
	ImageType string `json:"ImageType"`
}
