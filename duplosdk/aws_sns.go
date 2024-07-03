package duplosdk

import (
	"fmt"
	"strings"
	"time"
)

type DuploSnsTopic struct {
	Name                 string                        `json:"Name"`
	KmsKeyId             string                        `json:"KmsKeyId,omitempty"`
	ExtraTopicAttributes DuploSnsTopicAttributesCreate `json:"ExtraTopicAttributes,omitempty"`
}

type DuploSnsTopicResource struct {
	Name                 string                  `json:"Name"`
	ResourceType         int                     `json:"ResourceType,omitempty"`
	ExtraTopicAttributes DuploSnsTopicAttributes `json:"ExtraTopicAttributes,omitempty"`
}

type DuploSnsTopicAttributesCreate struct {
	DeliveryPolicy            string `json:"DeliveryPolicy,omitempty"`
	DisplayName               string `json:"DisplayName,omitempty"`
	FifoTopic                 bool   `json:"FifoTopic,omitempty"`
	Policy                    string `json:"Policy,omitempty"`
	SignatureVersion          string `json:"SignatureVersion,omitempty"`
	TracingConfig             string `json:"TracingConfig,omitempty"`
	KmsMasterKeyId            string `json:"KmsMasterKeyId,omitempty"`
	ArchivePolicy             string `json:"ArchivePolicy,omitempty"`
	BeginningArchiveTime      string `json:"BeginningArchiveTime,omitempty"`
	ContentBasedDeduplication string `json:"ContentBasedDeduplication,omitempty"`
}

type DuploSnsTopicAttributes struct {
	FifoTopic string `json:"FifoTopic,omitempty"`
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

func (c *Client) TenantGetSnsTopicAttributes(tenantID string, topicArn string) (*DuploSnsTopicAttributes, ClientError) {
	rp := DuploSnsTopicAttributes{}
	_, err := RetryWithExponentialBackoff(func() (interface{}, ClientError) {
		err := c.getAPI(
			fmt.Sprintf("TenantListSnsTopic(%s)", tenantID),
			fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic/%s/attributes", tenantID, topicArn),
			&rp,
		)
		return &rp, err
	},
		RetryConfig{
			MinDelay:  1 * time.Second,
			MaxDelay:  5 * time.Second,
			MaxJitter: 2000,
			Timeout:   60 * time.Second,
			IsRetryable: func(error ClientError) bool {
				return error.Status() == 400 || strings.Contains(error.Error(), "context deadline exceeded")
			},
		})

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
