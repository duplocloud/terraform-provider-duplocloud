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

	Identifier               string   `json:"Identifier"`
	Arn                      string   `json:"Arn"`
	Endpoint                 string   `json:"Endpoint,omitempty"`
	CacheType                int      `json:"CacheType,omitempty"`
	EngineVersion            string   `json:"EngineVersion,omitempty"`
	Size                     string   `json:"Size,omitempty"`
	Replicas                 int      `json:"Replicas,omitempty"`
	EncryptionAtRest         bool     `json:"EnableEncryptionAtRest,omitempty"`
	EncryptionInTransit      bool     `json:"EnableEncryptionAtTransit,omitempty"`
	KMSKeyID                 string   `json:"KmsKeyId,omitempty"`
	AuthToken                string   `json:"AuthToken,omitempty"`
	ParameterGroupName       string   `json:"ParameterGroupName,omitempty"`
	InstanceStatus           string   `json:"InstanceStatus,omitempty"`
	EnableClusterMode        bool     `json:"ClusteringEnabled,omitempty"`
	AutomaticFailoverEnabled bool     `json:"AutomaticFailoverEnabled,omitempty"`
	NumberOfShards           int      `json:"NoOfShards,omitempty"`
	SnapshotName             string   `json:"SnapshotName,omitempty"`
	SnapshotArns             []string `json:"SnapshotArns,omitempty"`
	SnapshotRetentionLimit   int      `json:"SnapshotRetentionLimit,omitempty"`
	SnapshotWindow           string   `json:"SnapshotWindow,omitempty"`
}

type AddDuploEcacheInstanceRequest struct {
	DuploEcacheInstance                                          // Embedded struct
	LogDeliveryConfigurations *[]LogDeliveryConfigurationRequest `json:"LogDeliveryConfigurations,omitempty"`
}

// Modeled after the class of the same name in 'RDSConfiguration.cs
// Represents an item in the response array from v2/subscriptions/{tenantID}/ECacheDBInstance
type DuploEcacheInstanceDetails struct {
	TenantID string `json:"TenantId"`

	Identifier                string `json:"Identifier"`
	Arn                       string `json:"Arn"`
	Endpoint                  string `json:"Endpoint"`
	CacheType                 int64  `json:"CacheType"`
	EngineVersion             string `json:"EngineVersion"`
	Size                      string `json:"Size"`
	Replicas                  int64  `json:"Replicas"`
	EnableEncryptionAtREST    bool   `json:"EnableEncryptionAtRest"`
	EnableEncryptionAtTransit bool   `json:"EnableEncryptionAtTransit"`

	InstanceStatus string `json:"InstanceStatus"`

	ClusteringEnabled bool  `json:"ClusteringEnabled"`
	NoOfShards        int64 `json:"NoOfShards"`

	Version                string        `json:"Version"`
	ClusterIdentifier      string        `json:"ClusterIdentifier"`
	SnapshotArns           []interface{} `json:"SnapshotArns"`
	SnapshotRetentionLimit int64         `json:"SnapshotRetentionLimit"`
}

type LogDeliveryConfigurationRequest struct {
	DestinationType    string              `json:"DestinationType,omitempty"`
	LogFormat          string              `json:"LogFormat,omitempty"`
	LogType            string              `json:"LogType,omitempty"`
	DestinationDetails *DestinationDetails `json:"DestinationDetails,omitempty"`
}

type DestinationDetails struct {
	CloudWatchLogsDetails  *CloudWatchLogsDestinationDetails `json:"CloudWatchLogsDetails,omitempty"`
	KinesisFirehoseDetails *KinesisFirehoseDetails           `json:"KinesisFirehoseDetails,omitempty"`
}

type CloudWatchLogsDestinationDetails struct {
	LogGroup string `json:"LogGroup,omitempty"`
}

type KinesisFirehoseDetails struct {
	DeliveryStream string `json:"DeliveryStream,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcacheInstanceCreate creates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreate(tenantID string, duploObject *AddDuploEcacheInstanceRequest) (*DuploEcacheInstance, ClientError) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, false)
}

// EcacheInstanceUpdate updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceUpdate(tenantID string, duploObject *AddDuploEcacheInstanceRequest) (*DuploEcacheInstance, ClientError) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, true)
}

// EcacheInstanceCreateOrUpdate creates or updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreateOrUpdate(tenantID string, duploObject *AddDuploEcacheInstanceRequest, updating bool) (*DuploEcacheInstance, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	var rp DuploEcacheInstance
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("ECacheInstanceUpdate(%s, duplo-%s)", tenantID, duploObject.Name),
		fmt.Sprintf("subscriptions/%s/ECacheInstanceUpdate", tenantID),
		&duploObject,
		&rp,
	)

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
