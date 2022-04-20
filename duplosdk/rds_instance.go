package duplosdk

import (
	"fmt"
	"strings"
)

const (
	DUPLO_RDS_ENGINE_MYSQL                        = 0
	DUPLO_RDS_ENGINE_POSTGRESQL                   = 1
	DUPLO_RDS_ENGINE_MSSQL_EXPRESS                = 2
	DUPLO_RDS_ENGINE_MSSQL_STANDARD               = 3
	DUPLO_RDS_ENGINE_AURORA_MYSQL                 = 8
	DUPLO_RDS_ENGINE_AURORA_POSTGRESQL            = 9
	DUPLO_RDS_ENGINE_MSSQL_WEB                    = 10
	DUPLO_RDS_ENGINE_AURORA_SERVERLESS_MYSQL      = 11
	DUPLO_RDS_ENGINE_AURORA_SERVERLESS_POSTGRESQL = 12
	DUPLO_RDS_ENGINE_DOCUMENTDB                   = 13
)

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

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
	EnableLogging               bool   `json:"EnableLogging,omitempty"`
	MultiAZ                     bool   `json:"MultiAZ,omitempty"`
	InstanceStatus              string `json:"InstanceStatus,omitempty"`
	ClusterIdentifier           string `json:"ClusterIdentifier,omitempty"`
	Version                     string `json:"Version,omitempty"`
}

// DuploRdsInstancePasswordChange is a Duplo SDK object that represents an RDS instance password change
type DuploRdsInstancePasswordChange struct {
	Identifier     string `json:"Identifier"`
	MasterPassword string `json:"MasterPassword"`
	StorePassword  bool   `json:"StorePassword,omitempty"`
}

type DuploRdsInstanceDeleteProtection struct {
	DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
	DeletionProtection   *bool  `json:"DeletionProtection,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// RdsInstanceGet retrieves an RDS instance via the Duplo API.
func (c *Client) RdsInstanceList(tenantID string) (*[]DuploRdsInstance, ClientError) {
	rp := []DuploRdsInstance{}
	err := c.getAPI(
		fmt.Sprintf("RdsInstanceList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetRDSInstances", tenantID),
		&rp)
	if err != nil {
		return nil, err
	}

	// Fill in the tenant ID and the name.
	for i := range rp {
		rp[i].TenantID = tenantID
		rp[i].Name = strings.TrimPrefix(rp[i].Identifier, "duplo")
	}
	return &rp, nil
}

// RdsInstanceGet retrieves a single RDS instance.
func (c *Client) RdsInstanceGet(tenantID, name string) (*DuploRdsInstance, ClientError) {
	allResources, err := c.RdsInstanceList(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the resource with the specific type and name.
	for _, resource := range *allResources {
		if resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

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

// RdsInstanceGetV2 retrieves an RDS instance via the Duplo API.
func (c *Client) RdsInstanceGetV2(id string) (*DuploRdsInstance, ClientError) {
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

func (c *Client) RdsInstanceChangeDeleteProtection(tenantID string, duploObject DuploRdsInstanceDeleteProtection) ClientError {
	return c.postAPI(
		fmt.Sprintf("RdsInstanceChangeDeleteProtection(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
		fmt.Sprintf("subscriptions/%s/ModifyRDSDBInstance", tenantID),
		&duploObject,
		nil,
	)
}

func RdsIsAurora(engine int) bool {
	return engine == DUPLO_RDS_ENGINE_AURORA_MYSQL ||
		engine == DUPLO_RDS_ENGINE_AURORA_POSTGRESQL ||
		engine == DUPLO_RDS_ENGINE_AURORA_SERVERLESS_MYSQL ||
		engine == DUPLO_RDS_ENGINE_AURORA_SERVERLESS_POSTGRESQL
}

func RdsIsMsSQL(engine int) bool {
	return engine == DUPLO_RDS_ENGINE_MSSQL_EXPRESS ||
		engine == DUPLO_RDS_ENGINE_MSSQL_STANDARD ||
		engine == DUPLO_RDS_ENGINE_MSSQL_WEB
}
