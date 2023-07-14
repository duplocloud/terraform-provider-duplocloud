package duplosdk

import (
	"fmt"
	"strings"
)

type DuploSQSQueue struct {
	Name                      string `json:"Name"`
	QueueType                 int    `json:"QueueType,omitempty"`
	State                     string `json:"State,omitempty"`
	MessageRetentionPeriod    int    `json:"MessageRetentionPeriod,omitempty"`
	VisibilityTimeout         int    `json:"VisibilityTimeout,omitempty"`
	Url                       string `json:"Url,omitempty"`
	ContentBasedDeduplication bool   `json:"ContentBasedDeduplication,omitempty"`
	DeduplicationScope        int    `json:"DeduplicationScope"`
	FifoThroughputLimit       int    `json:"FifoThroughputLimit"`
	ResourceType              int    `json:"ResourceType,omitempty"`
}

type DuploSQSQueueResource struct {
	Name         string `json:"Name"`
	ResourceType int    `json:"ResourceType,omitempty"`
}

func (c *Client) DuploSQSQueueCreate(tenantID string, rq *DuploSQSQueue) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploSQSQueueCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/SqsUpdate", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) DuploSQSQueueDelete(tenantID string, url string) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploSQSQueueDelete(%s, %s)", tenantID, url),
		fmt.Sprintf("subscriptions/%s/SqsUpdate", tenantID),
		&DuploSQSQueue{
			Name:  url,
			State: "delete",
		},
		nil,
	)
}

func (c *Client) DuploSQSQueueCreateV2(tenantID string, rq *DuploSQSQueue) (*DuploSQSQueue, ClientError) {
	resp := DuploSQSQueue{}
	err := c.postAPI(
		fmt.Sprintf("DuploSQSQueueCreateV2(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/sqs", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploSQSQueueUpdateV2(tenantID string, rq *DuploSQSQueue) (*DuploSQSQueue, ClientError) {
	resp := DuploSQSQueue{}
	err := c.putAPI(
		fmt.Sprintf("DuploSQSQueueUpdateV2(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/sqs/%s", tenantID, rq.Name),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploSQSQueueGetV2(tenantID string, fullname string) (*DuploSQSQueue, ClientError) {
	list, err := c.DuploSQSQueueListV2(tenantID, fullname)
	if err != nil {
		return nil, err
	}

	if list != nil && len(*list) > 0 {
		for _, element := range *list {
			if element.Name == fullname {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) DuploSQSQueueListV2(tenantID, fullname string) (*[]DuploSQSQueue, ClientError) {
	resp := []DuploSQSQueue{}
	err := c.getAPI(
		fmt.Sprintf("DuploSQSQueueListV2(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/aws/sqs", tenantID),
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploSQSQueueDeleteV2(tenantID string, fullname string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploSQSQueueDelete(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/aws/sqs/%s", tenantID, fullname),
		nil,
	)
}

func (c *Client) TenantGetSQSQueue(tenantID string, url string) (*DuploSQSQueueResource, ClientError) {
	resource, err := c.TenantGetAwsSqsQueueCloudResource(tenantID, url)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploSQSQueueResource{
		Name:         resource.Name,
		ResourceType: resource.Type,
	}, nil
}

func (c *Client) TenantGetSqsQueueByName(tenantID, name string) (*DuploAwsCloudResource, ClientError) {
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return nil, err
	}

	allResources, err := c.TenantListAwsCloudResources(tenantID)
	if err != nil {
		return nil, err
	}

	if allResources != nil {
		for _, resource := range *allResources {
			if resource.Type == ResourceTypeSQSQueue {
				resourceFullname, err := c.ExtractSqsFullname(tenantID, resource.Name)
				if err != nil {
					return nil, err
				}
				if resourceFullname == fullName {
					return &resource, nil
				}
			}
		}
	}
	return nil, nil
}

func (c *Client) TenantGetAwsSqsQueueCloudResource(tenantID string, name string) (*DuploAwsCloudResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific type and name.
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeSQSQueue && resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

func (c *Client) ExtractSqsFullname(tenantID string, sqsUrl string) (string, ClientError) {
	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return "", err
	}
	parts := strings.Split(sqsUrl, "/"+accountID+"/")
	fullname := parts[1]
	// fullname = strings.TrimSuffix(fullname, ".fifo")
	return fullname, nil
}
