package duplosdk

import (
	"fmt"
)

type DuploSnsTopic struct {
	Name     string `json:"Name"`
	KmsKeyId string `json:"KmsKeyId,omitempty"`
}

type DuploSnsTopicResource struct {
	Name         string `json:"Name"`
	ResourceType int    `json:"ResourceType,omitempty"`
}

func (c *Client) DuploSnsTopicCreate(tenantID string, rq *DuploSnsTopic) (*DuploSnsTopicResource, ClientError) {
	rp := &DuploSnsTopicResource{}
	err := c.postAPI(
		fmt.Sprintf("DuploSnsTopicCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) DuploSnsTopicDelete(tenantID string, arn string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploSnsTopicDelete(%s, %s)", tenantID, arn),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic/%s", tenantID, arn),
		nil,
	)
}

func (c *Client) TenantListSnsTopic(tenantID string) (*[]DuploSnsTopicResource, ClientError) {
	rp := []DuploSnsTopicResource{}
	err := c.getAPI(
		fmt.Sprintf("TenantListSnsTopic(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) TenantGetSnsTopic(tenantID string, arn string) (*DuploSnsTopicResource, ClientError) {
	list, err := c.TenantListSnsTopic(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, topic := range *list {
			if topic.Name == arn {
				return &topic, nil
			}
		}
	}
	return nil, nil
}
