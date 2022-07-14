package duplosdk

import (
	"fmt"
	"strings"
)

type DuploSQSQueue struct {
	Name      string `json:"Name"`
	QueueType int    `json:"QueueType,omitempty"`
	State     string `json:"State,omitempty"`
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
	fullname = strings.TrimSuffix(fullname, ".fifo")
	return fullname, nil
}
