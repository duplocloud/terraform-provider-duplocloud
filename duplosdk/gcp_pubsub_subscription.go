package duplosdk

import "fmt"

type State string

const (
	STATE_UNSPECIFIED               State = "STATE_UNSPECIFIED"
	ACTIVE                          State = "ACTIVE"
	PERMISSION_DENIED               State = "PERMISSION_DENIED"
	NOT_FOUND                       State = "NOT_FOUND"
	SCHEMA_MISMATCH                 State = "SCHEMA_MISMATCH"
	IN_TRANSIT_LOCATION_RESTRICTION State = "IN_TRANSIT_LOCATION_RESTRICTION"
)

type DuploPubSubBigQuery struct {
	Table               string `json:"table"`
	State               State  `json:"state"`
	UseTopicSchema      bool   `json:"useTopicSchema"`
	WriteMetadata       bool   `json:"writeMetadata"`
	DropUnknownFields   bool   `json:"dropUnknownFields"`
	UseTableSchema      bool   `json:"useTableSchema"`
	ServiceAccountEmail string `json:"serviceAccountEmail"`
}

type DuploPubSubCloudStorageConfig struct {
	Bucket                 string `json:"bucket"`
	FilenamePrefix         string `json:"filenamePrefix"`
	FileNameSuffix         string `json:"filenameSuffix"`
	FileNameDateTimeFormat string `json:"filenameDatetimeFormat"`
	MaxDuration            string `json:"maxDuration"`
	MaxBytes               string `json:"maxBytes"`
	MaxMessages            string `json:"maxMessages"`
	State                  State  `json:"state"`
	ServiceAccountEmail    string `json:"serviceAccountEmail"`
	AvroConfig             struct {
		WriteMetadata  bool `json:"writeMetadata"`
		UseTopicSchema bool `json:"useTopicSchema"`
	} `json:"avroConfig"`
}

type DuploPubSubPushConfig struct {
	PushEndpoint string            `json:"pushEndpoint"`
	Attributes   map[string]string `json:"attributes"`
	OidcToken    struct {
		ServiceAccountEmail string `json:"serviceAccountEmail"`
		Audience            string `json:"audience"`
	} `json:"oidcToken"`

	NoWrapper struct {
		WriteMetadata bool `json:"writeMetadata"`
	} `json:"noWrapper"`
}

type DuploPubSubSubscription struct {
	Name                      string                         `json:"name"`
	Topic                     string                         `json:"topic"`
	BigQuery                  *DuploPubSubBigQuery           `json:"bigqueryConfig"`
	CloudStorageConfig        *DuploPubSubCloudStorageConfig `json:"cloudStorageConfig"`
	PushConfig                *DuploPubSubPushConfig         `json:"pushConfig"`
	AckDeadlineSeconds        int                            `json:"ackDeadlineSeconds"`
	MessageRetentionDuration  string                         `json:"messageRetentionDuration"`
	RetainAckedMessages       bool                           `json:"retainAckedMessages"`
	Filter                    string                         `json:"filter"`
	EnableMessageOrdering     bool                           `json:"enableMessageOrdering"`
	EnableExactlyOnceDelivery bool                           `json:"enableExactlyOnceDelivery"`
	ExpirationPolicy          *DuploPubSubExpirationPolicy   `json:"expirationPolicy"`
	DeadLetterPolicy          *DuploPubSubDeadLetterPolicy   `json:"deadLetterPolicy"`
	RetryPolicy               *DuploPubSubRetryPolicy        `json:"retryPolicy"`
	Labels                    map[string]string              `json:"labels"`
}

type DuploPubSubSubscriptionResponse struct {
	Name                      string                               `json:"name"`
	Topic                     string                               `json:"topic"`
	BigQuery                  *DuploPubSubBigQuery                 `json:"bigqueryConfig"`
	CloudStorageConfig        *DuploPubSubCloudStorageConfig       `json:"cloudStorageConfig"`
	PushConfig                *DuploPubSubPushConfig               `json:"pushConfig"`
	AckDeadlineSeconds        int                                  `json:"ackDeadlineSeconds"`
	MessageRetentionDuration  string                               `json:"messageRetentionDuration"`
	RetainAckedMessages       bool                                 `json:"retainAckedMessages"`
	Filter                    string                               `json:"filter"`
	EnableMessageOrdering     bool                                 `json:"enableMessageOrdering"`
	EnableExactlyOnceDelivery bool                                 `json:"enableExactlyOnceDelivery"`
	ExpirationPolicy          *DuploPubSubExpirationPolicyResponse `json:"expirationPolicy"`
	DeadLetterPolicy          *DuploPubSubDeadLetterPolicy         `json:"deadLetterPolicy"`
	RetryPolicy               *DuploPubSubRetryPolicy              `json:"retryPolicy"`
	Labels                    map[string]string                    `json:"labels"`
}

type DuploPubSubExpirationPolicy struct {
	Ttl string `json:"ttl"`
}

type DuploPubSubExpirationPolicyResponse struct {
	Ttl struct {
		Seconds int `json:"seconds"`
	} `json:"ttl"`
}

type DuploPubSubDeadLetterPolicy struct {
	DeadLetterTopic     string `json:"deadLetterTopic"`
	MaxDeliveryAttempts int    `json:"maxDeliveryAttempts"`
}

type DuploPubSubRetryPolicy struct {
	MinimumBackoff string `json:"minimumBackoff"`
	MaximumBackoff string `json:"maximumBackoff"`
}

func (c *Client) GCPTenantCreatePubSubSubscription(tenantID string, duplo DuploPubSubSubscription) (*DuploPubSubSubscription, ClientError) {

	resp := DuploPubSubSubscription{}

	// Create the bucket via Duplo.
	err := c.postAPI(
		fmt.Sprintf("GCPTenantCreatePubSubSubscription(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/subscription/%s", tenantID, duplo.Topic),
		&duplo,
		&resp)

	if err != nil || resp.Name == "" {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GCPTenantGetPubSubSubscription(tenantID string, topic string) (*DuploPubSubSubscription, ClientError) {
	rp := DuploPubSubSubscription{}
	err := c.getAPI(fmt.Sprintf("GCPTenantGetPubSubSubscription(%s, %s)", tenantID, topic),
		fmt.Sprintf("v3/subscriptions/%s/google/subscription/%s", tenantID, topic),
		&rp)
	if err != nil { //|| rp.Arn == "" {
		return nil, err
	}
	return &rp, err
}

func (c *Client) GCPTenantDeletePubSubSubscription(tenantID string, name string) ClientError {

	// Delete the bucket via Duplo.
	return c.deleteAPI(fmt.Sprintf("GCPTenantDeletePubSubSubscription(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/subscription/%s", tenantID, name),
		nil)

}

func (c *Client) GCPTenantUpdatePubSubSubscription(tenantID string, duplo DuploPubSubSubscription) (*DuploPubSubSubscription, ClientError) {
	// Apply the settings via Duplo.
	apiName := fmt.Sprintf("GCPTenantUpdatePubSubSubscription(%s, %s)", tenantID, duplo.Name)
	rp := DuploPubSubSubscription{}
	err := c.putAPI(apiName, fmt.Sprintf("v3/subscriptions/%s/google/subscription/%s", tenantID, duplo.Topic), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	return &rp, nil
}
