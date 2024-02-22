package duplosdk

import (
	"fmt"
	"time"
)

const DynamoDBProvisionedThroughputMinValue = 1
const DynamoDBKeyTypeRange = "RANGE"
const DynamoDBKeyTypeHash = "HASH"
const DynamoDBBillingModeProvisioned = "PROVISIONED"
const DynamoDBBillingModePerRequest = "PAY_PER_REQUEST"

// DuploDynamoDBTable is a Duplo SDK object that represents a dynamodb table
type DuploDynamoDBTable struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                 string                            `json:"Name"`
	Arn                  string                            `json:"Arn"`
	Status               string                            `json:"TableStatus,omitempty"`
	KeySchema            *[]DuploDynamoDBKeySchema         `json:"KeySchema,omitempty"`
	AttributeDefinitions *[]DuploDynamoDBAttributeDefinion `json:"AttributeDefinitions,omitempty"`
}

type DuploDynamoDBTableV2 struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	TableName                 string                                      `json:"TableName"`
	TableId                   string                                      `json:"TableId"`
	TableArn                  string                                      `json:"TableArn,omitempty"`
	PointInTimeRecoveryStatus string                                      `json:"PointInTimeRecoveryStatus,omitempty"`
	KeySchema                 *[]DuploDynamoDBKeySchema                   `json:"KeySchema,omitempty"`
	AttributeDefinitions      *[]DuploDynamoDBAttributeDefinion           `json:"AttributeDefinitions,omitempty"`
	TableStatus               *DuploStringValue                           `json:"TableStatus,omitempty"`
	TableSizeBytes            int                                         `json:"TableSizeBytes,omitempty"`
	LocalSecondaryIndexes     *[]DuploDynamoDBTableV2LocalSecondaryIndex  `json:"LocalSecondaryIndexes,omitempty"`
	GlobalSecondaryIndexes    *[]DuploDynamoDBTableV2GlobalSecondaryIndex `json:"GlobalSecondaryIndexes,omitempty"`
	LatestStreamArn           string                                      `json:"LatestStreamArn,omitempty"`
	LatestStreamLabel         string                                      `json:"LatestStreamLabel,omitempty"`
	ProvisionedThroughput     *DuploDynamoDBProvisionedThroughput         `json:"ProvisionedThroughput,omitempty"`
	SSEDescription            *DuploDynamoDBTableV2SSESpecification       `json:"SSEDescription,omitempty"`
	StreamSpecification       *DuploDynamoDBTableV2StreamSpecification    `json:"StreamSpecification,omitempty"`
	BillingModeSummary        *DuploDynamoDBTableV2BillingModeSummary     `json:"BillingModeSummary,omitempty"`
}

type DuploDynamoDBTableV2TimeInRecovery struct {
	IsPointInTimeRecovery bool `json:"IsPointInTimeRecovery,omitempty"`
}

type DuploDynamoDBTableV2ContinuousBackupsDescription struct {
	LatestRestorableDateTime time.Time `json:"LatestRestorableDateTime,omitempty"`

	ContinuousBackupsStatus struct {
		Value string `json:"Value,omitempty"`
	} `json:"ContinuousBackupsStatus,omitempty"`

	PointInTimeRecoveryDescription struct {
		EarliestRestorableDateTime time.Time `json:"EarliestRestorableDateTime,omitempty"`
	} `json:"PointInTimeRecoveryDescription,omitempty"`

	PointInTimeRecoveryStatus struct {
		Value string `json:"Value,omitempty"`
	} `json:"PointInTimeRecoveryStatus,omitempty"`
}

type DuploDynamoDBProvisionedThroughput struct {
	TableSizNumberOfDecreasesTodayeBytes int `json:"NumberOfDecreasesToday,omitempty"`
	ReadCapacityUnits                    int `json:"ReadCapacityUnits,omitempty"`
	WriteCapacityUnits                   int `json:"WriteCapacityUnits,omitempty"`
}

// DuploDynamoDBKeySchema is a Duplo SDK object that represents a dynamodb key schema
type DuploDynamoDBKeySchema struct {
	AttributeName string            `json:"AttributeName"`
	KeyType       *DuploStringValue `json:"KeyType,omitempty"`
}

type DuploDynamoDBKeySchemaV2 struct {
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType,omitempty"`
}

// DuploDynamoDBAttributeDefinition is a Duplo SDK object that represents a dynamodb attribute definition
type DuploDynamoDBAttributeDefinion struct {
	AttributeName string            `json:"AttributeName"`
	AttributeType *DuploStringValue `json:"AttributeType,omitempty"`
}

type DuploDynamoDBAttributeDefinionV2 struct {
	AttributeName string `json:"AttributeName"`
	AttributeType string `json:"AttributeType,omitempty"`
}

// DuploDynamoDBTableRequest is a Duplo SDK object that represents a request to create a dynamodb table
type DuploDynamoDBTableRequest struct {
	Name           string `json:"Name"`
	State          string `json:"State,omitempty"`
	ResourceType   int    `json:"ResourceType,omitempty"`
	PrimaryKeyName string `json:"PrimaryKeyName,omitempty"`
	AttributeType  string `json:"AttributeType,omitempty"`
	KeyType        string `json:"KeyType,omitempty"`
}

type DuploDynamoDBTableV2StreamSpecification struct {
	StreamEnabled  bool              `json:"StreamEnabled,omitempty"`
	StreamViewType *DuploStringValue `json:"StreamViewType,omitempty"`
}

type DuploDynamoDBTableV2SSESpecification struct {
	Enabled         bool              `json:"Enabled,omitempty"`
	KMSMasterKeyId  string            `json:"KMSMasterKeyId,omitempty"`
	SSEType         *DuploStringValue `json:"SSEType,omitempty"`
	KMSMasterKeyArn string            `json:"KMSMasterKeyArn,omitempty"`
}

type DuploDynamoDBTableV2Projection struct {
	NonKeyAttributes []string          `json:"NonKeyAttributes,omitempty"`
	ProjectionType   *DuploStringValue `json:"ProjectionType,omitempty"`
}

type DuploDynamoDBTableV2LocalSecondaryIndex struct {
	IndexName  string                          `json:"IndexName"`
	Projection *DuploDynamoDBTableV2Projection `json:"Projection,omitempty"`
	KeySchema  *[]DuploDynamoDBKeySchema       `json:"KeySchema,omitempty"`
}

type DuploDynamoDBTableV2GlobalSecondaryIndex struct {
	IndexName             string                              `json:"IndexName"`
	Projection            *DuploDynamoDBTableV2Projection     `json:"Projection,omitempty"`
	KeySchema             *[]DuploDynamoDBKeySchema           `json:"KeySchema,omitempty"`
	ProvisionedThroughput *DuploDynamoDBProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
}

type DuploDynamoDBTableV2BillingModeSummary struct {
	BillingMode *DuploStringValue `json:"BillingMode,omitempty"`
}

type DuploDynamoDBTableRequestV2 struct {
	TableName                 string                                      `json:"TableName"`
	BillingMode               string                                      `json:"BillingMode,omitempty"`
	DeletionProtectionEnabled bool                                        `json:"DeletionProtectionEnabled,omitempty"`
	Tags                      *[]DuploKeyStringValue                      `json:"Tags,omitempty"`
	KeySchema                 *[]DuploDynamoDBKeySchemaV2                 `json:"KeySchema,omitempty"`
	AttributeDefinitions      *[]DuploDynamoDBAttributeDefinionV2         `json:"AttributeDefinitions,omitempty"`
	ProvisionedThroughput     *DuploDynamoDBProvisionedThroughput         `json:"ProvisionedThroughput,omitempty"`
	StreamSpecification       *DuploDynamoDBTableV2StreamSpecification    `json:"StreamSpecification,omitempty"`
	SSESpecification          *DuploDynamoDBTableV2SSESpecification       `json:"SSESpecification,omitempty"`
	LocalSecondaryIndexes     *[]DuploDynamoDBTableV2LocalSecondaryIndex  `json:"LocalSecondaryIndexes,omitempty"`
	GlobalSecondaryIndexes    *[]DuploDynamoDBTableV2GlobalSecondaryIndex `json:"GlobalSecondaryIndexes,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// DynamoDBTableCreate creates a dynamodb table via the Duplo API.
func (c *Client) DynamoDBTableCreate(tenantID string, rq *DuploDynamoDBTableRequest) (*DuploDynamoDBTable, ClientError) {
	rp := DuploDynamoDBTable{}
	err := c.postAPI(
		fmt.Sprintf("DynamoDBTableCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable", tenantID),
		&rq,
		&rp,
	)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableCreateV2(tenantID string, rq *DuploDynamoDBTableRequestV2) (*DuploDynamoDBTableV2, ClientError) {
	rp := DuploDynamoDBTableV2{}
	err := c.postAPI(
		fmt.Sprintf("DynamoDBTableCreate(%s, %s)", tenantID, rq.TableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2", tenantID),
		&rq,
		&rp,
	)
	rp.TenantID = tenantID
	return &rp, err
}

// DynamoDBTableDelete deletes a dynamodb table via the Duplo API.
func (c *Client) DynamoDBTableDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DynamoDBTableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable/%s", tenantID, name),
		nil)
}

func (c *Client) DynamoDBTableDeleteV2(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DynamoDBTableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, name),
		nil)
}

// DynamoDBTableGet retrieves a dynamodb table via the Duplo API
func (c *Client) DynamoDBTableGet(tenantID string, name string) (*DuploDynamoDBTable, ClientError) {
	rp := DuploDynamoDBTable{}
	err := c.getAPI(
		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableGetV2(tenantID string, name string) (*DuploDynamoDBTableV2, ClientError) {
	rp := DuploDynamoDBTableV2{}
	err := c.getAPI(
		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableV2PointInRecovery(tenantID, tableName string, isPointInRecovery bool) (*DuploDynamoDBTableV2ContinuousBackupsDescription, ClientError) {
	rp := DuploDynamoDBTableV2ContinuousBackupsDescription{}
	rq := DuploDynamoDBTableV2TimeInRecovery{
		IsPointInTimeRecovery: isPointInRecovery,
	}
	err := c.putAPI(
		fmt.Sprintf("DynamoDBTableV2PointInRecovery(%s, %s)", tenantID, tableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s/point-in-time-recovery", tenantID, tableName),
		&rq,
		&rp,
	)
	return &rp, err
}
