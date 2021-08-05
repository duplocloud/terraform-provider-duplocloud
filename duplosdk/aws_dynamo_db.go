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

// DuploDynamoDBKeySchema is a Duplo SDK object that represents a dynamodb key schema
type DuploDynamoDBKeySchema struct {
	AttributeName string            `json:"AttributeName"`
	KeyType       *DuploStringValue `json:"KeyType,omitempty"`
}

// DuploDynamoDBAttributeDefinition is a Duplo SDK object that represents a dynamodb attribute definition
type DuploDynamoDBAttributeDefinion struct {
	AttributeName string            `json:"AttributeName"`
	AttributeType *DuploStringValue `json:"AttributeType,omitempty"`
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

// DynamoDBTableDelete deletes a dynamodb table via the Duplo API.
func (c *Client) DynamoDBTableDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DynamoDBTableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/dynamodbTable/%s", tenantID, name),
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
