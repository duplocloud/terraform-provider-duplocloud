package duplosdk

import (
	"fmt"
)

// DuploDynamoDBTable is a Duplo SDK object that represents a dynamodb table
type DuploDynamoDBTable struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name           string `json:"Name"`
	PrimaryKeyName string `json:"PrimaryKeyName"`
	AttributeType  string `json:"AttributeType"`
	KeyType        string `json:"KeyType"`
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
func (c *Client) DynamoDBTableCreate(tenantID string, rq *DuploDynamoDBTableRequest) error {
	rq.ResourceType = ResourceTypeDynamoDBTable

	return c.postAPI(
		fmt.Sprintf("DynamoDBTableCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/DynamoDBTableUpdate", tenantID),
		&rq,
		nil,
	)
}

// DynamoDBTableDelete deletes a dynamodb table via the Duplo API.
func (c *Client) DynamoDBTableDelete(tenantID, name string) error {
	rq := &DuploDynamoDBTableRequest{State: "delete", Name: name}

	return c.postAPI(
		fmt.Sprintf("DynamoDBTableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DynamoDBTableUpdate", tenantID),
		&rq,
		nil)
}

// DynamoDBTableGet retrieves a dynamodb table via the Duplo API
func (c *Client) DynamoDBTableGet(tenantID string, name string) (*DuploDynamoDBTable, error) {
	// Figure out the full resource name.
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeDynamoDBTable, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploDynamoDBTable{TenantID: tenantID, Name: fullName}, nil
}
