package duplosdk

import (
	"fmt"
)

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

	TableName            string                            `json:"TableName"`
	TableId              string                            `json:"TableId"`
	TableArn             string                            `json:"TableArn,omitempty"`
	KeySchema            *[]DuploDynamoDBKeySchema         `json:"KeySchema,omitempty"`
	AttributeDefinitions *[]DuploDynamoDBAttributeDefinion `json:"AttributeDefinitions,omitempty"`
	TableStatus          *DuploStringValue                 `json:"TableStatus,omitempty"`
	TableSizeBytes       int                               `json:"TableSizeBytes,omitempty"`
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

type DuploDynamoDBTableRequestV2 struct {
	TableName             string                              `json:"TableName"`
	BillingMode           string                              `json:"BillingMode,omitempty"`
	Tags                  *[]DuploKeyStringValue              `json:"Tags,omitempty"`
	KeySchema             *[]DuploDynamoDBKeySchemaV2         `json:"KeySchema,omitempty"`
	AttributeDefinitions  *[]DuploDynamoDBAttributeDefinionV2 `json:"AttributeDefinitions,omitempty"`
	ProvisionedThroughput *DuploDynamoDBProvisionedThroughput `json:"ProvisionedThroughput,omitempty"`
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
