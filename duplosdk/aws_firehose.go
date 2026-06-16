package duplosdk

import "fmt"

// DuploFirehoseRequest maps to the AWS SDK CreateDeliveryStreamRequest (used for Create).
// Field names are the exact C# property names from Amazon.KinesisFirehose.Model.CreateDeliveryStreamRequest.
type DuploFirehoseRequest struct {
	DeliveryStreamName                                 string            `json:"DeliveryStreamName"`
	DeliveryStreamType                                 string            `json:"DeliveryStreamType,omitempty"`
	ExtendedS3DestinationConfiguration                 interface{}       `json:"ExtendedS3DestinationConfiguration,omitempty"`
	RedshiftDestinationConfiguration                   interface{}       `json:"RedshiftDestinationConfiguration,omitempty"`
	ElasticsearchDestinationConfiguration              interface{}       `json:"ElasticsearchDestinationConfiguration,omitempty"`
	AmazonopensearchserviceDestinationConfiguration    interface{}       `json:"AmazonopensearchserviceDestinationConfiguration,omitempty"`
	AmazonOpenSearchServerlessDestinationConfiguration interface{}       `json:"AmazonOpenSearchServerlessDestinationConfiguration,omitempty"`
	SplunkDestinationConfiguration                     interface{}       `json:"SplunkDestinationConfiguration,omitempty"`
	HttpEndpointDestinationConfiguration               interface{}       `json:"HttpEndpointDestinationConfiguration,omitempty"`
	SnowflakeDestinationConfiguration                  interface{}       `json:"SnowflakeDestinationConfiguration,omitempty"`
	IcebergDestinationConfiguration                    interface{}       `json:"IcebergDestinationConfiguration,omitempty"`
	KinesisStreamSourceConfiguration                   interface{}       `json:"KinesisStreamSourceConfiguration,omitempty"`
	MSKSourceConfiguration                             interface{}       `json:"MSKSourceConfiguration,omitempty"`
	DeliveryStreamEncryptionConfigurationInput         interface{}       `json:"DeliveryStreamEncryptionConfigurationInput,omitempty"`
	Tags                                               map[string]string `json:"Tags,omitempty"`
}

// DuploFirehoseUpdateRequest maps to the AWS SDK UpdateDestinationRequest.
// Uses *DestinationUpdate field names (not *DestinationConfiguration) as required by the UpdateDestination API.
type DuploFirehoseUpdateRequest struct {
	DeliveryStreamName                          string      `json:"DeliveryStreamName"`
	CurrentDeliveryStreamVersionId              string      `json:"CurrentDeliveryStreamVersionId"`
	DestinationId                               string      `json:"DestinationId"`
	ExtendedS3DestinationUpdate                 interface{} `json:"ExtendedS3DestinationUpdate,omitempty"`
	RedshiftDestinationUpdate                   interface{} `json:"RedshiftDestinationUpdate,omitempty"`
	ElasticsearchDestinationUpdate              interface{} `json:"ElasticsearchDestinationUpdate,omitempty"`
	AmazonopensearchserviceDestinationUpdate    interface{} `json:"AmazonopensearchserviceDestinationUpdate,omitempty"`
	AmazonOpenSearchServerlessDestinationUpdate interface{} `json:"AmazonOpenSearchServerlessDestinationUpdate,omitempty"`
	SplunkDestinationUpdate                     interface{} `json:"SplunkDestinationUpdate,omitempty"`
	HttpEndpointDestinationUpdate               interface{} `json:"HttpEndpointDestinationUpdate,omitempty"`
	SnowflakeDestinationUpdate                  interface{} `json:"SnowflakeDestinationUpdate,omitempty"`
	IcebergDestinationUpdate                    interface{} `json:"IcebergDestinationUpdate,omitempty"`
}

// DuploFirehoseDeliveryStream is the response model for a Firehose delivery stream.
// DeliveryStreamType and DeliveryStreamStatus are C# ConstantClass objects serialized as {"Value":"..."}.
type DuploFirehoseDeliveryStream struct {
	DeliveryStreamName   string      `json:"DeliveryStreamName"`
	DeliveryStreamARN    string      `json:"DeliveryStreamARN,omitempty"`
	DeliveryStreamStatus interface{} `json:"DeliveryStreamStatus,omitempty"`
	DeliveryStreamType   interface{} `json:"DeliveryStreamType,omitempty"`
	CreateTimestamp      string      `json:"CreateTimestamp,omitempty"`
	Destinations         interface{} `json:"Destinations,omitempty"`
	HasMoreDestinations  bool        `json:"HasMoreDestinations,omitempty"`
	Source               interface{} `json:"Source,omitempty"`
	VersionId            string      `json:"VersionId,omitempty"`
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
// Uses PUT /aws/firehose/{name}/destination — maps to AWS UpdateDestination.
// Fetches CurrentDeliveryStreamVersionId and DestinationId via describe, then builds
// a DuploFirehoseUpdateRequest using *DestinationUpdate field names (not *DestinationConfiguration).
func (c *Client) DuploFirehoseUpdate(tenantID string, name string, rq *DuploFirehoseRequest) ClientError {
	// Fetch current stream to get VersionId and DestinationId required by UpdateDestination.
	stream, err := c.DuploFirehoseGet(tenantID, name)
	if err != nil {
		return err
	}
	if stream == nil {
		return newClientError(fmt.Sprintf("firehose stream '%s' not found", name))
	}

	urq := &DuploFirehoseUpdateRequest{
		DeliveryStreamName:             rq.DeliveryStreamName,
		CurrentDeliveryStreamVersionId: stream.VersionId,
	}

	// Extract DestinationId from Destinations[0].
	if dests, ok := stream.Destinations.([]interface{}); ok && len(dests) > 0 {
		if destMap, ok := dests[0].(map[string]interface{}); ok {
			if id, ok := destMap["DestinationId"].(string); ok {
				urq.DestinationId = id
			}
		}
	}

	// Map *DestinationConfiguration → *DestinationUpdate (AWS UpdateDestination field names).
	urq.ExtendedS3DestinationUpdate = rq.ExtendedS3DestinationConfiguration
	urq.RedshiftDestinationUpdate = rq.RedshiftDestinationConfiguration
	urq.ElasticsearchDestinationUpdate = rq.ElasticsearchDestinationConfiguration
	urq.AmazonopensearchserviceDestinationUpdate = rq.AmazonopensearchserviceDestinationConfiguration
	urq.AmazonOpenSearchServerlessDestinationUpdate = rq.AmazonOpenSearchServerlessDestinationConfiguration
	urq.SplunkDestinationUpdate = rq.SplunkDestinationConfiguration
	urq.HttpEndpointDestinationUpdate = rq.HttpEndpointDestinationConfiguration
	urq.SnowflakeDestinationUpdate = rq.SnowflakeDestinationConfiguration
	urq.IcebergDestinationUpdate = rq.IcebergDestinationConfiguration

	//rp := map[string]interface{}{}
	return c.putAPI(
		fmt.Sprintf("DuploFirehoseUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/firehose/%s/destination", tenantID, name),
		urq,
		nil,
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
