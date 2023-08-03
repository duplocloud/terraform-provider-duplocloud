package duplosdk

import (
	"fmt"
	"strings"
)

const (
	DuploRdsEngineMysql                      = 0
	DuploRdsEnginePostgresSql                = 1
	DuploRdsEngineMssqlExpress               = 2
	DuploRdsEngineMssqlStandard              = 3
	DuploRdsEngineAuroraMysql                = 8
	DuploRdsEngineAuroraPostgresql           = 9
	DuploRdsEngineMssqlWeb                   = 10
	DuploRdsEngineAuroraServerlessMysql      = 11
	DuploRdsEngineAuroraServerlessPostgresql = 12
	DuploRdsEngineDocumentdb                 = 13
)

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                  string                  `json:"Identifier"`
	ClusterIdentifier           string                  `json:"ClusterIdentifier,omitempty"`
	ReplicationSourceIdentifier string                  `json:"ReplicationSourceIdentifier,omitempty"`
	Arn                         string                  `json:"Arn"`
	Endpoint                    string                  `json:"Endpoint,omitempty"`
	MasterUsername              string                  `json:"MasterUsername,omitempty"`
	MasterPassword              string                  `json:"MasterPassword,omitempty"`
	Engine                      int                     `json:"Engine,omitempty"`
	EngineVersion               string                  `json:"EngineVersion,omitempty"`
	SnapshotID                  string                  `json:"SnapshotId,omitempty"`
	DBParameterGroupName        string                  `json:"DBParameterGroupName,omitempty"`
	StoreDetailsInSecretManager bool                    `json:"StoreDetailsInSecretManager,omitempty"`
	Cloud                       int                     `json:"Cloud,omitempty"`
	SizeEx                      string                  `json:"SizeEx,omitempty"`
	EncryptStorage              bool                    `json:"EncryptStorage,omitempty"`
	AllocatedStorage            int                     `json:"AllocatedStorage,omitempty"`
	EncryptionKmsKeyId          string                  `json:"EncryptionKmsKeyId,omitempty"`
	EnableLogging               bool                    `json:"EnableLogging,omitempty"`
	MultiAZ                     bool                    `json:"MultiAZ,omitempty"`
	InstanceStatus              string                  `json:"InstanceStatus,omitempty"`
	DBSubnetGroupName           string                  `json:"DBSubnetGroupName,omitempty"`
	V2ScalingConfiguration      *V2ScalingConfiguration `json:"V2ScalingConfiguration,omitempty"`
	AvailabilityZone            string                  `json:"AvailabilityZone,omitempty"`
}

type V2ScalingConfiguration struct {
	MinCapacity float64 `json:"MinCapacity,omitempty"`
	MaxCapacity float64 `json:"MaxCapacity,omitempty"`
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

type DuploRdsClusterDeleteProtection struct {
	DBClusterIdentifier string `json:"DBClusterIdentifier"`
	ApplyImmediately    bool   `json:"ApplyImmediately"`
	DeletionProtection  *bool  `json:"DeletionProtection,omitempty"`
}

type DuploRdsModifyAuroraV2ServerlessInstanceSize struct {
	Identifier             string                  `json:"Identifier"`
	ClusterIdentifier      string                  `json:"ClusterIdentifier"`
	ApplyImmediately       bool                    `json:"ApplyImmediately"`
	SizeEx                 string                  `json:"SizeEx,omitempty"`
	V2ScalingConfiguration *V2ScalingConfiguration `json:"V2ScalingConfiguration,omitempty"`
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

func (c *Client) RdsInstanceGetByName(tenantID, name string) (*DuploRdsInstance, ClientError) {

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

func (c *Client) RdsClusterChangeDeleteProtection(tenantID string, duploObject DuploRdsClusterDeleteProtection) ClientError {
	return c.postAPI(
		fmt.Sprintf("RdsClusterChangeDeleteProtection(%s, %s)", tenantID, duploObject.DBClusterIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/modifyRdsCluster", tenantID),
		&duploObject,
		nil,
	)
}

func (c *Client) RdsModifyAuroraV2ServerlessInstanceSize(tenantID string, duploObject DuploRdsModifyAuroraV2ServerlessInstanceSize) ClientError {
	return c.postAPI(
		fmt.Sprintf("RdsModifyAuroraV2ServerlessInstanceSize(%s, %s)", tenantID, duploObject.ClusterIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/modifyAuroraToV2Serverless", tenantID),
		&duploObject,
		nil,
	)
}

func RdsIsAurora(engine int) bool {
	return engine == DuploRdsEngineAuroraMysql ||
		engine == DuploRdsEngineAuroraPostgresql ||
		engine == DuploRdsEngineAuroraServerlessMysql ||
		engine == DuploRdsEngineAuroraServerlessPostgresql
}

func RdsIsMsSQL(engine int) bool {
	return engine == DuploRdsEngineMssqlExpress ||
		engine == DuploRdsEngineMssqlStandard ||
		engine == DuploRdsEngineMssqlWeb
}
