package duplosdk

import (
	"fmt"
)

// DuploEcacheInstance is a Duplo SDK object that represents an ECache instance
type DuploEcacheInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier          string `json:"Identifier"`
	Arn                 string `json:"Arn"`
	Endpoint            string `json:"Endpoint,omitempty"`
	CacheType           int    `json:"CacheType,omitempty"`
	Size                string `json:"Size,omitempty"`
	Replicas            int    `json:"Replicas,omitempty"`
	EncryptionAtRest    bool   `json:"EnableEncryptionAtRest,omitempty"`
	EncryptionInTransit bool   `json:"EnableEncryptionAtTransit,omitempty"`
	KMSKeyID            string `json:"KmsKeyId,omitempty"`
	AuthToken           string `json:"AuthToken,omitempty"`
	InstanceStatus      string `json:"InstanceStatus,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcacheInstanceCreate creates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreate(tenantID string, duploObject *DuploEcacheInstance) (*DuploEcacheInstance, ClientError) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, false)
}

// EcacheInstanceUpdate updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceUpdate(tenantID string, duploObject *DuploEcacheInstance) (*DuploEcacheInstance, ClientError) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, true)
}

// EcacheInstanceCreateOrUpdate creates or updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreateOrUpdate(tenantID string, duploObject *DuploEcacheInstance, updating bool) (*DuploEcacheInstance, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploEcacheInstance{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("EcacheInstanceCreateOrUpdate(%s, duplo-%s)", tenantID, duploObject.Name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance", tenantID),
		&duploObject,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// EcacheInstanceDelete deletes an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("EcacheInstanceDelete(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/duplo-%s", tenantID, name),
		nil)
}

// EcacheInstanceGet retrieves an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceGet(tenantID, name string) (*DuploEcacheInstance, ClientError) {

	// Call the API.
	rp := DuploEcacheInstance{}
	err := c.getAPI(
		fmt.Sprintf("EcacheInstanceGet(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/duplo-%s", tenantID, name),
		&rp)
	if err != nil || rp.Identifier == "" {
		return nil, err
	}

	// Fill in the tenant ID and the name and return the object
	rp.TenantID = tenantID
	rp.Name = name
	return &rp, nil
}
