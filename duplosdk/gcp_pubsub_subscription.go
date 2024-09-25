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

type DuploPubSubExpirationPolicy struct {
	Ttl string `json:"ttl"`
}

type DuploPubSubDeadLetterPolicy struct {
	DeadLetterTopic     string `json:"deadLetterTopic"`
	MaxDeliveryAttempts int    `json:"maxDeliveryAttempts"`
}

type DuploPubSubRetryPolicy struct {
	MinimumBackoff string `json:"minimumBackoff"`
	MaximumBackoff string `json:"maximumBackoff"`
}

func (c *Client) GCPTenantCreatePubSubSubscription(tenantID string, duplo DuploPubSubSubscription) (*DuploS3Bucket, ClientError) {

	resp := DuploS3Bucket{}

	// Create the bucket via Duplo.
	err := c.postAPI(
		fmt.Sprintf("GCPTenantCreatePubSubSubscription(%s, %s)", tenantID, duplo.Name),
		//  fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket", tenantID),
		&duplo,
		&resp)

	if err != nil || resp.Name == "" {
		return nil, err
	}
	return &resp, nil
}
