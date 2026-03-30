package duplosdk

import "fmt"

// DuploFirehoseRequest maps directly to the AWS SDK CreateDeliveryStreamRequest.
// The Duplo backend accepts this struct as-is (no custom DTO wrapper).
// Field names are the exact C# property names from Amazon.KinesisFirehose.Model.CreateDeliveryStreamRequest.
type DuploFirehoseRequest struct {
	DeliveryStreamName                              string            `json:"DeliveryStreamName"`
	DeliveryStreamType                              string            `json:"DeliveryStreamType,omitempty"`
	ExtendedS3DestinationConfiguration              interface{}       `json:"ExtendedS3DestinationConfiguration,omitempty"`
	RedshiftDestinationConfiguration                interface{}       `json:"RedshiftDestinationConfiguration,omitempty"`
	ElasticsearchDestinationConfiguration           interface{}       `json:"ElasticsearchDestinationConfiguration,omitempty"`
	AmazonopensearchserviceDestinationConfiguration interface{}       `json:"AmazonopensearchserviceDestinationConfiguration,omitempty"`
	AmazonOpenSearchServerlessDestinationConfiguration interface{}    `json:"AmazonOpenSearchServerlessDestinationConfiguration,omitempty"`
	SplunkDestinationConfiguration                  interface{}       `json:"SplunkDestinationConfiguration,omitempty"`
	HttpEndpointDestinationConfiguration            interface{}       `json:"HttpEndpointDestinationConfiguration,omitempty"`
	SnowflakeDestinationConfiguration               interface{}       `json:"SnowflakeDestinationConfiguration,omitempty"`
	IcebergDestinationConfiguration                 interface{}       `json:"IcebergDestinationConfiguration,omitempty"`
	KinesisStreamSourceConfiguration                interface{}       `json:"KinesisStreamSourceConfiguration,omitempty"`
	MSKSourceConfiguration                          interface{}       `json:"MSKSourceConfiguration,omitempty"`
	DeliveryStreamEncryptionConfigurationInput      interface{}       `json:"DeliveryStreamEncryptionConfigurationInput,omitempty"`
	Tags                                            map[string]string `json:"Tags,omitempty"`
}

// DuploFirehoseDeliveryStream is the response model for a Firehose delivery stream.
// DeliveryStreamType and DeliveryStreamStatus are C# ConstantClass objects serialized as {"Value":"..."}.
type DuploFirehoseDeliveryStream struct {
	DeliveryStreamName  string      `json:"DeliveryStreamName"`
	DeliveryStreamARN   string      `json:"DeliveryStreamARN,omitempty"`
	DeliveryStreamStatus interface{} `json:"DeliveryStreamStatus,omitempty"`
	DeliveryStreamType  interface{} `json:"DeliveryStreamType,omitempty"`
	CreateTimestamp     string      `json:"CreateTimestamp,omitempty"`
	Destinations        interface{} `json:"Destinations,omitempty"`
	HasMoreDestinations bool        `json:"HasMoreDestinations,omitempty"`
	Source              interface{} `json:"Source,omitempty"`
	VersionId           string      `json:"VersionId,omitempty"`
}

// FirehoseStringValue extracts a plain string from either a string or a ConstantClass {"Value":"..."} object.
func FirehoseStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if m, ok := v.(map[string]interface{}); ok {
		if s, ok := m["Value"].(string); ok {
			return s
		}
	}
	return ""
}

// DuploFirehoseCreate creates a Firehose delivery stream.
// The short name is provided; the backend prepends the tenant prefix automatically.
func (c *Client) DuploFirehoseCreate(tenantID string, rq *DuploFirehoseRequest) ClientError {
	rp := map[string]interface{}{}
	return c.postAPI(
		fmt.Sprintf("DuploFirehoseCreate(%s, %s)", tenantID, rq.DeliveryStreamName),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose", tenantID),
		rq,
		&rp,
	)
}

// DuploFirehoseGet retrieves a single Firehose delivery stream by short name.
func (c *Client) DuploFirehoseGet(tenantID string, name string) (*DuploFirehoseDeliveryStream, ClientError) {
	rp := DuploFirehoseDeliveryStream{}
	err := c.getAPI(
		fmt.Sprintf("DuploFirehoseGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose/%s", tenantID, name),
		&rp,
	)
	if err != nil {
		return nil, err
	}
	if rp.DeliveryStreamName == "" {
		return nil, nil
	}
	return &rp, nil
}

// DuploFirehoseList returns the short names of all Firehose delivery streams for a tenant.
func (c *Client) DuploFirehoseList(tenantID string) (*[]string, ClientError) {
	rp := []string{}
	err := c.getAPI(
		fmt.Sprintf("DuploFirehoseList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose", tenantID),
		&rp,
	)
	return &rp, err
}

// DuploFirehoseUpdate updates the destination configuration of a Firehose delivery stream.
// Uses PUT /aws/firehose (no /{name}/destination) — the backend accepts CreateDeliveryStreamRequest,
// fetches CurrentDeliveryStreamVersionId and DestinationId internally via a describe call.
func (c *Client) DuploFirehoseUpdate(tenantID string, name string, rq *DuploFirehoseRequest) ClientError {
	rp := map[string]interface{}{}
	return c.putAPI(
		fmt.Sprintf("DuploFirehoseUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose", tenantID),
		rq,
		&rp,
	)
}

// DuploFirehoseDelete deletes a Firehose delivery stream by short name.
func (c *Client) DuploFirehoseDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploFirehoseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose/%s", tenantID, name),
		nil,
	)
}
