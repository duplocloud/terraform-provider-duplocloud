package duplosdk

import (
	"fmt"
	"strings"
)

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                  string `json:"Identifier"`
	Arn                         string `json:"Arn"`
	Endpoint                    string `json:"Endpoint,omitempty"`
	MasterUsername              string `json:"MasterUsername,omitempty"`
	MasterPassword              string `json:"MasterPassword,omitempty"`
	Engine                      int    `json:"Engine,omitempty"`
	EngineVersion               string `json:"EngineVersion,omitempty"`
	SnapshotID                  string `json:"SnapshotId,omitempty"`
	DBParameterGroupName        string `json:"DBParameterGroupName,omitempty"`
	StoreDetailsInSecretManager bool   `json:"StoreDetailsInSecretManager,omitempty"`
	Cloud                       int    `json:"Cloud,omitempty"`
	SizeEx                      string `json:"SizeEx,omitempty"`
	EncryptStorage              bool   `json:"EncryptStorage,omitempty"`
	InstanceStatus              string `json:"InstanceStatus,omitempty"`
}

// DuploRdsInstancePasswordChange is a Duplo SDK object that represents an RDS instance password change
type DuploRdsInstancePasswordChange struct {
	Identifier     string `json:"Identifier"`
	MasterPassword string `json:"MasterPassword"`
	StorePassword  bool   `json:"StorePassword,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// RdsInstanceCreate creates an ECS service via the Duplo API.
func (c *Client) RdsInstanceCreate(tenantID string, duploObject *DuploRdsInstance) (*DuploRdsInstance, ClientError) {
	return c.RdsInstanceCreateOrUpdate(tenantID, duploObject, false)
}

// RdsInstanceUpdate updates an ECS service via the Duplo API.
func (c *Client) RdsInstanceUpdate(tenantID string, duploObject *DuploRdsInstance) (*DuploRdsInstance, ClientError) {
	return c.RdsInstanceCreateOrUpdate(tenantID, duploObject, true)
}

// RdsInstanceCreateOrUpdate creates or updates an RDS instance via the Duplo API.
func (c *Client) RdsInstanceCreateOrUpdate(tenantID string, duploObject *DuploRdsInstance, updating bool) (*DuploRdsInstance, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploRdsInstance{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("RdsInstanceCreateOrUpdate(%s, duplo%s)", tenantID, duploObject.Name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance", tenantID),
		&duploObject,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// RdsInstanceDelete deletes an RDS instance via the Duplo API.
func (c *Client) RdsInstanceDelete(id string) (*DuploRdsInstance, ClientError) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	err := c.deleteAPI(
		fmt.Sprintf("RdsInstanceDelete(%s, duplo%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/duplo%s", tenantID, name),
		nil)
	if err != nil {
		return nil, err
	}

	// Return a placeholder - since the API does not return responses.
	return &DuploRdsInstance{TenantID: tenantID, Name: name}, nil
}

// RdsInstanceGet retrieves an RDS instance via the Duplo API.
func (c *Client) RdsInstanceGet(id string) (*DuploRdsInstance, ClientError) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	duploObject := DuploRdsInstance{}
	err := c.getAPI(
		fmt.Sprintf("RdsInstanceGet(%s, duplo%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/duplo%s", tenantID, name),
		&duploObject)
	if err != nil || duploObject.Identifier == "" {
		return nil, err
	}

	// Fill in the tenant ID and the name and return the object
	duploObject.TenantID = tenantID
	duploObject.Name = name
	return &duploObject, nil
}

// RdsInstanceChangePassword creates or updates an RDS instance via the Duplo API.
func (c *Client) RdsInstanceChangePassword(tenantID string, duploObject DuploRdsInstancePasswordChange) ClientError {
	// Call the API.
	return c.postAPI(
		fmt.Sprintf("RdsInstanceChangePassword(%s, %s)", tenantID, duploObject.Identifier),
		fmt.Sprintf("subscriptions/%s/RDSInstancePasswordChange", tenantID),
		&duploObject,
		nil,
	)
}
