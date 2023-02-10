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
	ComputeResources       *DuploAwsBatchComputeResource `json:"ComputeResources,omitempty"`
	ComputeEnvironmentName string                        `json:"ComputeEnvironmentName,omitempty"`
	Tags                   map[string]string             `json:"Tags,omitempty"`
	ServiceRole            string                        `json:"ServiceRole,omitempty"`
	Type                   *DuploStringValue             `json:"Type,omitempty"`
	State                  *DuploStringValue             `json:"State,omitempty"`
	ComputeEnvironmentArn  string                        `json:"ComputeEnvironmentArn,omitempty"`
	EcsClusterArn          string                        `json:"EcsClusterArn,omitempty"`
	Status                 *DuploStringValue             `json:"Status,omitempty"`
	StatusReason           string                        `json:"StatusReason,omitempty"`
	ComputeEnvironment     string                        `json:"ComputeEnvironment,omitempty"`
}

type DuploAwsBatchComputeResource struct {
	Ec2Configuration   *[]DuploAwsBatchComputeEc2Configuration   `json:"Ec2Configuration,omitempty"`
	LaunchTemplate     *DuploAwsBatchLaunchTemplateConfiguration `json:"LaunchTemplate,omitempty"`
	Type               *DuploStringValue                         `json:"Type,omitempty"`
	AllocationStrategy *DuploStringValue                         `json:"AllocationStrategy,omitempty"`
	MaxvCpus           int                                       `json:"MaxvCpus,omitempty"`
	MinvCpus           int                                       `json:"MinvCpus,omitempty"`
	DesiredvCpus       int                                       `json:"DesiredvCpus,omitempty"`
	BidPercentage      int                                       `json:"BidPercentage,omitempty"`
	InstanceTypes      []string                                  `json:"InstanceTypes,omitempty"`
	Ec2KeyPair         string                                    `json:"Ec2KeyPair,omitempty"`
	ImageId            string                                    `json:"ImageId,omitempty"`
	InstanceRole       string                                    `json:"InstanceRole,omitempty"`
	PlacementGroup     string                                    `json:"PlacementGroup,omitempty"`
	SecurityGroupIds   []string                                  `json:"SecurityGroupIds,omitempty"`
	SpotIamFleetRole   string                                    `json:"SpotIamFleetRole,omitempty"`
	Subnets            []string                                  `json:"Subnets,omitempty"`
	Tags               map[string]string                         `json:"Tags,omitempty"`
}

type DuploAwsBatchComputeEc2Configuration struct {
	ImageType       string `json:"ImageType,omitempty"`
	ImageIdOverride string `json:"ImageIdOverride,omitempty"`
}

type DuploAwsBatchLaunchTemplateConfiguration struct {
	LaunchTemplateId   string `json:"LaunchTemplateId,omitempty"`
	LaunchTemplateName string `json:"LaunchTemplateName,omitempty"`
	Version            string `json:"Version,omitempty"`
}

func (c *Client) AwsBatchComputeEnvironmentCreate(tenantID string, rq *DuploAwsBatchComputeEnvironment) ClientError {
	rp := ""
	return c.postAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentCreate(%s, %s)", tenantID, rq.ComputeEnvironmentName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironment", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchComputeEnvironmentUpdate(tenantID string, rq *DuploAwsBatchComputeEnvironment) ClientError {
	rp := ""
	return c.putAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentUpdate(%s, %s)", tenantID, rq.ComputeEnvironmentName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironment", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchComputeEnvironmentGet(tenantID string, name string) (*DuploAwsBatchComputeEnvironment, ClientError) {
	list, err := c.AwsBatchComputeEnvironmentList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.ComputeEnvironmentName == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AwsBatchComputeEnvironmentList(tenantID string) (*[]DuploAwsBatchComputeEnvironment, ClientError) {
	rp := []DuploAwsBatchComputeEnvironment{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironment", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsBatchComputeEnvironmentDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironment/%s", tenantID, name),
		nil,
	)
}

func (c *Client) AwsBatchComputeEnvironmentDisable(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentDisable(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironmentDisable/%s", tenantID, name),
		nil,
	)
}

func (c *Client) AwsBatchComputeEnvironmentEnable(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchComputeEnvironmentEnable(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchComputeEnvironmentEnable/%s", tenantID, name),
		nil,
	)
}

//  --------------- Job Queue ---------------

type DuploAwsBatchJobQueue struct {
	Priority                int                                     `json:"Priority,omitempty"`
	ComputeEnvironmentOrder *[]DuploAwsBatchComputeEnvironmentOrder `json:"ComputeEnvironmentOrder,omitempty"`
	JobQueueName            string                                  `json:"JobQueueName,omitempty"`
	JobQueueArn             string                                  `json:"JobQueueArn,omitempty"`
	SchedulingPolicyArn     string                                  `json:"SchedulingPolicyArn,omitempty"`
	State                   *DuploStringValue                       `json:"State,omitempty"`
	Tags                    map[string]string                       `json:"Tags,omitempty"`
	Status                  *DuploStringValue                       `json:"Status,omitempty"`
	StatusReason            string                                  `json:"StatusReason,omitempty"`
	JobQueue                string                                  `json:"JobQueue,omitempty"`
}

type DuploAwsBatchComputeEnvironmentOrder struct {
	ComputeEnvironment string `json:"ComputeEnvironment,omitempty"`
	Order              int    `json:"Order,omitempty"`
}

func (c *Client) AwsBatchJobQueueCreate(tenantID string, rq *DuploAwsBatchJobQueue) ClientError {
	rp := ""
	return c.postAPI(
		fmt.Sprintf("AwsBatchJobQueueCreate(%s, %s)", tenantID, rq.JobQueueName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueue", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchJobQueueUpdate(tenantID string, rq *DuploAwsBatchJobQueue) ClientError {
	rp := ""
	return c.putAPI(
		fmt.Sprintf("AwsBatchJobQueueUpdate(%s, %s)", tenantID, rq.JobQueueName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueue", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchJobQueueGet(tenantID string, name string) (*DuploAwsBatchJobQueue, ClientError) {
	list, err := c.AwsBatchJobQueueList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.JobQueueName == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AwsBatchJobQueueList(tenantID string) (*[]DuploAwsBatchJobQueue, ClientError) {
	rp := []DuploAwsBatchJobQueue{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchJobQueueList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueue", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsBatchJobQueueDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchJobQueueDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueue/%s", tenantID, name),
		nil,
	)
}

func (c *Client) AwsBatchJobQueueDisable(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchJobQueueDisable(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueueDisable/%s", tenantID, name),
		nil,
	)
}

func (c *Client) AwsBatchJobQueueEnable(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchJobQueueEnable(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobQueueEnable/%s", tenantID, name),
		nil,
	)
}

//  --------------- Job Definition ---------------

type DuploAwsBatchJobDefinition struct {
	RetryStrategy        *DuploAwsBatchJobDefinitionRetryStrategy `json:"RetryStrategy,omitempty"`
	ContainerProperties  map[string]interface{}                   `json:"ContainerProperties,omitempty"`
	JobDefinitionName    string                                   `json:"JobDefinitionName,omitempty"`
	JobDefinitionArn     string                                   `json:"JobDefinitionArn,omitempty"`
	PlatformCapabilities []string                                 `json:"PlatformCapabilities,omitempty"`
	Timeout              *DuploAwsBatchJobDefinitionTimeout       `json:"Timeout,omitempty"`
	Parameters           map[string]string                        `json:"Parameters,omitempty"`
	Type                 *DuploStringValue                        `json:"Type,omitempty"`
	Tags                 map[string]string                        `json:"Tags,omitempty"`
	Revision             int                                      `json:"Revision,omitempty"`
}

type DuploAwsBatchJobDefinitionResp struct {
	RetryStrategy        *DuploAwsBatchJobDefinitionRetryStrategy `json:"RetryStrategy,omitempty"`
	ContainerProperties  map[string]interface{}                   `json:"ContainerProperties,omitempty"`
	JobDefinitionName    string                                   `json:"JobDefinitionName,omitempty"`
	JobDefinitionArn     string                                   `json:"JobDefinitionArn,omitempty"`
	PlatformCapabilities []string                                 `json:"PlatformCapabilities,omitempty"`
	Timeout              *DuploAwsBatchJobDefinitionTimeout       `json:"Timeout,omitempty"`
	Parameters           map[string]string                        `json:"Parameters,omitempty"`
	Type                 string                                   `json:"Type,omitempty"`
	Tags                 map[string]string                        `json:"Tags,omitempty"`
	Revision             int                                      `json:"Revision,omitempty"`
	Status               string                                   `json:"Status,omitempty"`
}

type DuploAwsBatchJobDefinitionTimeout struct {
	AttemptDurationSeconds int `json:"AttemptDurationSeconds,omitempty"`
}

type DuploAwsBatchJobDefinitionRetryStrategy struct {
	Attempts       int                                         `json:"Attempts,omitempty"`
	EvaluateOnExit *[]DuploAwsBatchJobDefinitionEvaluateOnExit `json:"EvaluateOnExit,omitempty"`
}
type DuploAwsBatchJobDefinitionEvaluateOnExit struct {
	Action         *DuploStringValue `json:"Action,omitempty"`
	OnExitCode     string            `json:"OnExitCode,omitempty"`
	OnReason       string            `json:"OnReason,omitempty"`
	OnStatusReason string            `json:"OnStatusReason,omitempty"`
}

// type DuploAwsBatchJobDefinitionContainerProperties struct {
// 	Command              []string `json:"Command,omitempty"`
// 	Image                string   `json:"Image,omitempty"`
// 	ResourceRequirements []struct {
// 		Type  *DuploStringValue `json:"Type,omitempty"`
// 		Value string            `json:"Value,omitempty"`
// 	} `json:"ResourceRequirements,omitempty"`
// 	InstanceType           string `json:"InstanceType,omitempty"`
// 	JobRoleArn             string `json:"JobRoleArn,omitempty"`
// 	ExecutionRoleArn       string `json:"ExecutionRoleArn,omitempty"`
// 	Memory                 int    `json:"Memory,omitempty"`
// 	Vcpus                  int    `json:"Vcpus,omitempty"`
// 	User                   string `json:"User,omitempty"`
// 	Privileged             bool   `json:"Privileged,omitempty"`
// 	ReadonlyRootFilesystem bool   `json:"ReadonlyRootFilesystem,omitempty"`
// }

func (c *Client) AwsBatchJobDefinitionCreate(tenantID string, rq *DuploAwsBatchJobDefinition) ClientError {
	rp := ""
	return c.postAPI(
		fmt.Sprintf("AwsBatchJobDefinitionCreate(%s, %s)", tenantID, rq.JobDefinitionName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobDefinition", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchJobDefinitionUpdate(tenantID string, rq *DuploAwsBatchJobDefinition) ClientError {
	rp := ""
	return c.putAPI(
		fmt.Sprintf("AwsBatchJobDefinitionUpdate(%s, %s)", tenantID, rq.JobDefinitionName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobDefinition", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsBatchJobDefinitionGet(tenantID string, name string) (*DuploAwsBatchJobDefinitionResp, ClientError) {
	list, err := c.AwsBatchJobDefinitionList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.JobDefinitionName == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AwsBatchJobDefinitionList(tenantID string) (*[]DuploAwsBatchJobDefinitionResp, ClientError) {
	rp := []DuploAwsBatchJobDefinitionResp{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchJobDefinitionList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobDefinition", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsBatchJobDefinitionDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchJobDefinitionDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobDefinition/%s", tenantID, name),
		nil,
	)
}

//  --------------- Submit Job Request ---------------

type DuploAwsBatchJobRequest struct {
	JobName                    string                                   `json:"JobName,omitempty"`
	JobDefinition              string                                   `json:"JobDefinition,omitempty"`
	JobQueue                   string                                   `json:"JobQueue,omitempty"`
	JobDefinitionArn           string                                   `json:"JobDefinitionArn,omitempty"`
	PlatformCapabilities       []string                                 `json:"PlatformCapabilities,omitempty"`
	Timeout                    *DuploAwsBatchJobDefinitionTimeout       `json:"Timeout,omitempty"`
	Parameters                 map[string]string                        `json:"Parameters,omitempty"`
	Type                       *DuploStringValue                        `json:"Type,omitempty"`
	Tags                       map[string]string                        `json:"Tags,omitempty"`
	Revision                   int                                      `json:"Revision,omitempty"`
	SchedulingPriorityOverride int                                      `json:"SchedulingPriorityOverride,omitempty"`
	RetryStrategy              *DuploAwsBatchJobDefinitionRetryStrategy `json:"RetryStrategy,omitempty"`
	ContainerOverrides         map[string]interface{}                   `json:"ContainerOverrides,omitempty"`
	ShareIdentifier            string                                   `json:"ShareIdentifier,omitempty"`
}

type DuploAwsBatchJobDetails struct {
	JobArn               string                                   `json:"JobArn,omitempty"`
	JobId                string                                   `json:"JobId,omitempty"`
	ShareIdentifier      string                                   `json:"ShareIdentifier,omitempty"`
	JobName              string                                   `json:"JobName,omitempty"`
	JobDefinition        string                                   `json:"JobDefinition,omitempty"`
	JobQueue             string                                   `json:"JobQueue,omitempty"`
	PlatformCapabilities []string                                 `json:"PlatformCapabilities,omitempty"`
	Timeout              *DuploAwsBatchJobDefinitionTimeout       `json:"Timeout,omitempty"`
	Parameters           map[string]string                        `json:"Parameters,omitempty"`
	Type                 *DuploStringValue                        `json:"Type,omitempty"`
	Tags                 map[string]string                        `json:"Tags,omitempty"`
	SchedulingPriority   int                                      `json:"SchedulingPriority,omitempty"`
	RetryStrategy        *DuploAwsBatchJobDefinitionRetryStrategy `json:"RetryStrategy,omitempty"`
	StatusReason         string                                   `json:"StatusReason,omitempty"`
	Status               *DuploStringValue                        `json:"Status,omitempty"`
	Container            map[string]interface{}                   `json:"Container,omitempty"`
	IsTerminated         bool                                     `json:"IsTerminated"`
	IsCancelled          bool                                     `json:"IsCancelled"`
}

func (c *Client) AwsBatchJobCreate(tenantID string, rq *DuploAwsBatchJobRequest) (string, ClientError) {
	rp := ""
	err := c.postAPI(
		fmt.Sprintf("AwsBatchJobCreate(%s, %s)", tenantID, rq.JobName),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobs", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) AwsBatchJobDescribe(tenantID string, name string) (*DuploAwsBatchJobDetails, ClientError) {
	rp := DuploAwsBatchJobDetails{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchJobDescribe(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobsDescribe/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsBatchJobTerminated(tenantID string, name string) (bool, ClientError) {
	rp := DuploAwsBatchJobDetails{}
	err := c.getAPI(
		fmt.Sprintf("AwsBatchJobDescribe(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobsDescribe/%s", tenantID, name),
		&rp,
	)
	if err != nil {
		return false, err
	}
	if rp.IsCancelled || rp.IsTerminated {
		return true, nil
	}
	return false, nil
}

func (c *Client) AwsBatchJobDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AwsBatchJobDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/batchJobs/%s", tenantID, name),
		nil,
	)
}
