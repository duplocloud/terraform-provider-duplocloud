package duplosdk

import (
	"fmt"
	"net/url"
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
	DeletionProtectionEnabled bool                                        `json:"DeletionProtectionEnabled,omitempty"`
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
	TtlAttributeName          string                                      `json:"TtlAttributeName,omitempty"`
	TtlStatus                 string                                      `json:"TtlStatus,omitempty"`
}

type DuploDynamoDBTableV2Old struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	TableName                 string                                      `json:"TableName"`
	TableId                   string                                      `json:"TableId"`
	TableArn                  string                                      `json:"TableArn,omitempty"`
	DeletionProtectionEnabled bool                                        `json:"DeletionProtectionEnabled,omitempty"`
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

type DuploDynamoDBTableV2Response struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	TableName                 string                                              `json:"TableName"`
	TableId                   string                                              `json:"TableId"`
	TableArn                  string                                              `json:"TableArn,omitempty"`
	DeletionProtectionEnabled bool                                                `json:"DeletionProtectionEnabled,omitempty"`
	PointInTimeRecoveryStatus string                                              `json:"PointInTimeRecoveryStatus,omitempty"`
	KeySchema                 *[]DuploDynamoDBKeySchemaResponse                   `json:"KeySchema,omitempty"`
	AttributeDefinitions      *[]DuploDynamoDBAttributeDefinion                   `json:"AttributeDefinitions,omitempty"`
	TableStatus               *DuploStringValue                                   `json:"TableStatus,omitempty"`
	TableSizeBytes            int                                                 `json:"TableSizeBytes,omitempty"`
	LocalSecondaryIndexes     *[]DuploDynamoDBTableV2LocalSecondaryIndexResponse  `json:"LocalSecondaryIndexes,omitempty"`
	GlobalSecondaryIndexes    *[]DuploDynamoDBTableV2GlobalSecondaryIndexResponse `json:"GlobalSecondaryIndexes,omitempty"`
	LatestStreamArn           string                                              `json:"LatestStreamArn,omitempty"`
	LatestStreamLabel         string                                              `json:"LatestStreamLabel,omitempty"`
	ProvisionedThroughput     *DuploDynamoDBProvisionedThroughput                 `json:"ProvisionedThroughput,omitempty"`
	SSEDescription            *DuploDynamoDBTableV2SSESpecificationResponse       `json:"SSEDescription,omitempty"`
	StreamSpecification       *DuploDynamoDBTableV2StreamSpecification            `json:"StreamSpecification,omitempty"`
	BillingModeSummary        *DuploDynamoDBTableV2BillingModeSummary             `json:"BillingModeSummary,omitempty"`
	TtlAttributeName          string                                              `json:"TtlAttributeName,omitempty"`
	TtlStatus                 string                                              `json:"TtlStatus,omitempty"`
}

type DuploDynamoDBTableV2TimeInRecovery struct {
	IsPointInTimeRecovery bool `json:"IsPointInTimeRecovery,omitempty"`
}

type DuploDynamoDBTableV2ContinuousBackupsDescription struct {
	//LatestRestorableDateTime time.Time `json:"LatestRestorableDateTime,omitempty"`

	ContinuousBackupsStatus struct {
		Value string `json:"Value,omitempty"`
	} `json:"ContinuousBackupsStatus,omitempty"`

	PointInTimeRecoveryDescription struct {
		EarliestRestorableDateTime string `json:"EarliestRestorableDateTime,omitempty"`
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
	AttributeName string `json:"AttributeName"`
	KeyType       string `json:"KeyType,omitempty"`
}

type DuploDynamoDBKeySchemaResponse struct {
	AttributeName string           `json:"AttributeName"`
	KeyType       DuploStringValue `json:"KeyType,omitempty"`
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

type DuploDynamoDBTableV2SSESpecificationResponse struct {
	Status          DuploStringValue  `json:"Status,omitempty"`
	KMSMasterKeyId  string            `json:"KMSMasterKeyId,omitempty"`
	SSEType         *DuploStringValue `json:"SSEType,omitempty"`
	KMSMasterKeyArn string            `json:"KMSMasterKeyArn,omitempty"`
}

type DuploDynamoDBTableV2Projection struct {
	NonKeyAttributes []string `json:"NonKeyAttributes,omitempty"`
	ProjectionType   string   `json:"ProjectionType,omitempty"`
}

type DuploDynamoDBTableV2ProjectionResponse struct {
	NonKeyAttributes []string          `json:"NonKeyAttributes,omitempty"`
	ProjectionType   *DuploStringValue `json:"ProjectionType,omitempty"`
}

type DuploDynamoDBTableV2LocalSecondaryIndex struct {
	IndexName  string                          `json:"IndexName"`
	Projection *DuploDynamoDBTableV2Projection `json:"Projection,omitempty"`
	KeySchema  *[]DuploDynamoDBKeySchema       `json:"KeySchema,omitempty"`
}

type DuploDynamoDBTableV2LocalSecondaryIndexResponse struct {
	IndexName  string                                  `json:"IndexName"`
	Projection *DuploDynamoDBTableV2ProjectionResponse `json:"Projection,omitempty"`
	KeySchema  *[]DuploDynamoDBKeySchemaResponse       `json:"KeySchema,omitempty"`
}

type DuploDynamoDBTableV2GlobalSecondaryIndex struct {
	IndexName             string                              `json:"IndexName"`
	Projection            *DuploDynamoDBTableV2Projection     `json:"Projection,omitempty"`
	KeySchema             *[]DuploDynamoDBKeySchema           `json:"KeySchema,omitempty"`
	ProvisionedThroughput *DuploDynamoDBProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
}

type DuploDynamoDBTableV2GlobalSecondaryIndexResponse struct {
	IndexName             string                                  `json:"IndexName"`
	Projection            *DuploDynamoDBTableV2ProjectionResponse `json:"Projection,omitempty"`
	KeySchema             *[]DuploDynamoDBKeySchemaResponse       `json:"KeySchema,omitempty"`
	ProvisionedThroughput *DuploDynamoDBProvisionedThroughput     `json:"ProvisionedThroughput,omitempty"`
	IndexStatus           *DuploStringValue                       `json:"IndexStatus,omitempty"`
}
type UpdateGSIReq struct {
	UpdateGSI UpdateGSI `json:"Update"`
}
type UpdateGSI struct {
	GlobalSecondaryIndexes *[]DuploDynamoDBTableV2GlobalSecondaryIndex
}
type DuploDynamoDBTableV2BillingModeSummary struct {
	BillingMode *DuploStringValue `json:"BillingMode,omitempty"`
}

type DuploDynamoDBTableV2TTl struct {
	AttributeName string `json:"AttributeName"`
	Enabled       bool   `json:"IsTtlEnabled"`
}

type DuploDynamoDBTableRequestV2 struct {
	TableName                 string                                      `json:"TableName"`
	BillingMode               string                                      `json:"BillingMode,omitempty"`
	DeletionProtectionEnabled *bool                                       `json:"DeletionProtectionEnabled,omitempty"`
	Tags                      *[]DyanmoDbV2Tag                            `json:"Tags,omitempty"`
	KeySchema                 *[]DuploDynamoDBKeySchemaV2                 `json:"KeySchema,omitempty"`
	AttributeDefinitions      *[]DuploDynamoDBAttributeDefinionV2         `json:"AttributeDefinitions,omitempty"`
	ProvisionedThroughput     *DuploDynamoDBProvisionedThroughput         `json:"ProvisionedThroughput,omitempty"`
	StreamSpecification       *DuploDynamoDBTableV2StreamSpecification    `json:"StreamSpecification,omitempty"`
	SSESpecification          *DuploDynamoDBTableV2SSESpecification       `json:"SSESpecification,omitempty"`
	LocalSecondaryIndexes     *[]DuploDynamoDBTableV2LocalSecondaryIndex  `json:"LocalSecondaryIndexes,omitempty"`
	GlobalSecondaryIndexes    *[]DuploDynamoDBTableV2GlobalSecondaryIndex `json:"GlobalSecondaryIndexes,omitempty"`
}

type DyanmoDbV2Tag struct {
	Key       string `json:"Key"`
	Value     string `json:"Value,omitempty"`
	DeleteTag bool   `json:"-"`
}
type DuploDynamoDBTagResource struct {
	ResourceArn string                 `json:"Arn,omitempty"`         // The ARN of the resource to tag
	Tags        *[]DuploKeyStringValue `json:"Tags,omitempty"`        // A list of tags to associate with the resource
	DeleteTags  []string               `json:"DeletedTags,omitempty"` //A list of tags to be deleted
}

type DuploDynamoDBTagResourceResponse struct {
	StatusCode int    `json:"statusCode"` // HTTP status code of the response
	Message    string `json:"message"`    // A message about the result of the operation, e.g., "Success"
	RequestID  string `json:"requestID"`  // The AWS request ID associated with the operation
}

/*************************************************
 * API CALLS to duplo
 */

// DynamoDBTableCreate creates a dynamodb table via the Duplo API.
func (c *Client) DynamoDBTableCreate(
	tenantID string,
	rq *DuploDynamoDBTableRequest,
) (*DuploDynamoDBTable, ClientError) {
	fmt.Println("calling DynamoDBTableCreate")
	rp := DuploDynamoDBTable{}
	conf := NewRetryConf()
	err := c.postAPIWithRetry(
		fmt.Sprintf("DynamoDBTableCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable", tenantID),
		&rq,
		&rp, &conf,
	)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableCreateV2(
	tenantID string,
	rq *DuploDynamoDBTableRequestV2,
) (*DuploDynamoDBTableV2Response, ClientError) {
	fmt.Println("calling DynamoDBTableCreateV2")
	conf := NewRetryConf()

	rp := DuploDynamoDBTableV2Response{}
	err := c.postAPIWithRetry(
		fmt.Sprintf("DynamoDBTableCreate(%s, %s)", tenantID, rq.TableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2", tenantID),
		&rq,
		&rp, &conf,
	)

	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableUpdateV2(
	tenantID string,
	rq *DuploDynamoDBTableRequestV2) (*DuploDynamoDBTableV2, ClientError) {
	rp := DuploDynamoDBTableV2{}
	conf := NewRetryConf()
	err := c.putAPIWithRetry(
		fmt.Sprintf("DynamoDBTableUpdate(%s, %s)", tenantID, rq.TableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, rq.TableName),
		&rq,
		&rp, &conf,
	)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableUpdateTagsV2(
	tenantId, name string,
	rq *DuploDynamoDBTagResource) (*DuploDynamoDBTagResource, ClientError) {
	rp := DuploDynamoDBTagResource{}
	err := c.postAPI(
		fmt.Sprintf("DynamoDBTableUpdateTags(%s, %s)", tenantId, rq.ResourceArn),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/manage", tenantId, name),
		&rq,
		nil,
	)
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
	conf := NewRetryConf()

	return c.deleteAPIWithRetry(
		fmt.Sprintf("DynamoDBTableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, name),
		nil, &conf)
}

// DynamoDBTableGet retrieves a dynamodb table via the Duplo API
func (c *Client) DynamoDBTableGet(tenantID string, name string) (*DuploDynamoDBTable, ClientError) {
	rp := DuploDynamoDBTable{}
	conf := NewRetryConf()
	err := c.getAPIWithRetry(
		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable/%s", tenantID, name),
		&rp, &conf)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableGetV2(tenantID string, name string) (*DuploDynamoDBTableV2Response, ClientError) {
	conf := NewRetryConf()
	rp := DuploDynamoDBTableV2Response{}
	err := c.getAPIWithRetry(
		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, name),
		&rp, &conf)
	rp.TenantID = tenantID
	return &rp, err
}

func (c *Client) DynamoDBTableV2PointInRecovery(tenantID, tableName string, isPointInRecovery bool) (*DuploDynamoDBTableV2ContinuousBackupsDescription, ClientError) {
	rp := DuploDynamoDBTableV2ContinuousBackupsDescription{}
	rq := DuploDynamoDBTableV2TimeInRecovery{
		IsPointInTimeRecovery: isPointInRecovery,
	}
	conf := NewRetryConf()

	err := c.putAPIWithRetry(
		fmt.Sprintf("DynamoDBTableV2PointInRecovery(%s, %s)", tenantID, tableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s/point-in-time-recovery", tenantID, tableName),
		&rq,
		&rp, &conf,
	)
	return &rp, err
}

func (c *Client) DynamoDBTableV2TTl(tenantID, tableName string, rq *DuploDynamoDBTableV2TTl) (*DuploDynamoDBTableV2TTl, ClientError) {
	rp := DuploDynamoDBTableV2TTl{}
	conf := NewRetryConf()

	err := c.putAPIWithRetry(
		fmt.Sprintf("DynamoDBTableV2TTl(%s, %s)", tenantID, tableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s/ttl", tenantID, tableName),
		&rq,
		&rp, &conf,
	)
	return &rp, err
}

type ModifyGSI struct {
	TableName                   string                             `json:"TableName"`
	AttributeDefinitions        []DuploDynamoDBAttributeDefinionV2 `json:"AttributeDefinitions,omitempty"`
	GlobalSecondaryIndexUpdates []GlobalSecondaryIndexUpdates
}
type GlobalSecondaryIndexUpdates struct {
	Create *DuploDynamoDBTableV2GlobalSecondaryIndex `json:"Create,omitempty"`
	Delete *Delete                                   `json:"Delete,omitempty"`
	Update *Update                                   `json:"Update,omitempty"`
}

type Delete struct {
	IndexName string `json:"IndexName"`
}

type Update struct {
	IndexName             string                             `json:"IndexName"`
	ProvisionedThroughput DuploDynamoDBProvisionedThroughput `json:"ProvisionedThroughput"`
}

// remove after july 2024 release updation
//func (c *Client) DynamoDBTableGetV2Old(tenantID string, name string) (*DuploDynamoDBTableV2Old, ClientError) {
//	rp := DuploDynamoDBTableV2Old{}
//	err := c.getAPI(
//		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, name),
//		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, name),
//		&rp)
//	rp.TenantID = tenantID
//	return &rp, err
//}

func (c *Client) DynamoDBTableUpdateGSIV2(
	tenantID string,
	rq *ModifyGSI) (*DuploDynamoDBTableV2, ClientError) {
	rp := DuploDynamoDBTableV2{}
	conf := NewRetryConf()

	err := c.putAPIWithRetry(
		fmt.Sprintf("DynamoDBTableUpdate(%s, %s)", tenantID, rq.TableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, rq.TableName),
		&rq,
		&rp, &conf,
	)
	rp.TenantID = tenantID
	return &rp, err
}

//func (c *Client) DynamoDBTableUpdateV21(
//	tenantID string,
//	rq *DuploDynamoDBTableRequestV2) (*DuploDynamoDBTableV2, ClientError) {
//	rp := DuploDynamoDBTableV2{}
//	//rs := &[]DuploDynamoDBTableV2GlobalSecondaryIndex{}
//	//rs = rq.GlobalSecondaryIndexes
//	err := c.putAPIWithRetry(
//		fmt.Sprintf("DynamoDBTableUpdate(%s, %s)", tenantID, rq.TableName),
//		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTableV2/%s", tenantID, rq.TableName),
//		rq,
//		&rp,
//	)
//	rp.TenantID = tenantID
//	return &rp, err
//}

//func (c *Client) DynamoDBTableExistsV2(tenantID string, name string) (bool, ClientError) {
//	_, err := c.DynamoDBTableGetV2(tenantID, name)
//	if err != nil {
//		return false, err
//	}
//	return true, nil
//}

// DuploDynamoDBTableV2UpdateSSESpecification updates the server side encryption
// settings on the provide DynamoDB table. Per the the AWS .NET SDK@3.7:
// "server side encryption modification must be the only operation in the request"
func (c *Client) DuploDynamoDBTableV2UpdateSSESpecification(
	tenantID string,
	rq *DuploDynamoDBTableRequestV2) (*DuploDynamoDBTableV2, ClientError) {

	r := DuploDynamoDBTableRequestV2{}

	r.TableName = rq.TableName
	r.SSESpecification = rq.SSESpecification

	return c.DynamoDBTableUpdateV2(tenantID, &r)
}

// DuploDynamoDBTableV2UpdateDeletionProtection updates the deletion protection
// settings on the provide DynamoDB table. Per the the AWS .NET SDK@3.7:
// "DeletionProtection modification must be the only operation in the request"
func (c *Client) DuploDynamoDBTableV2UpdateDeletionProtection(
	tenantID string,
	r *DuploDynamoDBTableRequestV2) (*DuploDynamoDBTableV2, ClientError) {

	return c.DynamoDBTableUpdateV2(tenantID, r)
}

func (c *Client) DynamoDBTableGetTags(tenantID string, arn string) ([]DuploKeyStringValue, ClientError) {
	rp := []DuploKeyStringValue{}
	err := c.getAPI(
		fmt.Sprintf("DynamoDBTableGet(%s, %s)", tenantID, arn), // triple encoding needed to fetch the data
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s", tenantID, url.PathEscape(url.PathEscape(url.PathEscape(arn)))),
		&rp)
	return rp, err
}
