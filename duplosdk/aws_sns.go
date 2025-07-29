package duplosdk

import (
	"fmt"
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
	ContentBasedDeduplication bool   `json:"ContentBasedDeduplication,omitempty"`
}

type DuploSnsTopicAttributes struct {
	Policy                    string `json:"Policy,omitempty"`
	Owner                     string `json:"Owner,omitempty"`
	SubscriptionsPending      string `json:"SubscriptionsPending,omitempty"`
	TopicArn                  string `json:"TopicArn,omitempty"`
	EffectiveDeliveryPolicy   string `json:"EffectiveDeliveryPolicy,omitempty"`
	SubscriptionsConfirmed    string `json:"SubscriptionsConfirmed,omitempty"`
	FifoTopic                 string `json:"FifoTopic,omitempty"`
	KmsMasterKeyId            string `json:"KmsMasterKeyId,omitempty"`
	DisplayName               string `json:"DisplayName,omitempty"`
	ArchivePolicy             string `json:"ArchivePolicy,omitempty"`
	ContentBasedDeduplication string `json:"ContentBasedDeduplication,omitempty"`
	SubscriptionsDeleted      string `json:"SubscriptionsDeleted,omitempty"`
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
