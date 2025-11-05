package duplosdk

import (
	"fmt"
	"log"
)

// DuploEcacheInstance is a Duplo SDK object that represents an ECache instance
type DuploEcacheInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                        string   `json:"Identifier"`
	Arn                               string   `json:"Arn"`
	Endpoint                          string   `json:"Endpoint,omitempty"`
	ConfigurationEndpoint             string   `json:"ConfigurationEndpoint,omitempty"`
	CacheType                         int      `json:"CacheType,omitempty"`
	EngineVersion                     string   `json:"EngineVersion,omitempty"`
	Size                              string   `json:"Size,omitempty"`
	Replicas                          int      `json:"Replicas,omitempty"`
	EncryptionAtRest                  bool     `json:"EnableEncryptionAtRest,omitempty"`
	EncryptionInTransit               bool     `json:"EnableEncryptionAtTransit,omitempty"`
	KMSKeyID                          string   `json:"KmsKeyId,omitempty"`
	AuthToken                         string   `json:"AuthToken,omitempty"`
	ParameterGroupName                string   `json:"ParameterGroupName,omitempty"`
	InstanceStatus                    string   `json:"InstanceStatus,omitempty"`
	EnableClusterMode                 bool     `json:"ClusteringEnabled,omitempty"`
	AutomaticFailoverEnabled          bool     `json:"AutomaticFailoverEnabled,omitempty"`
	NumberOfShards                    int      `json:"NoOfShards,omitempty"`
	SnapshotName                      string   `json:"SnapshotName,omitempty"`
	SnapshotArns                      []string `json:"SnapshotArns,omitempty"`
	SnapshotRetentionLimit            int      `json:"SnapshotRetentionLimit,omitempty"`
	SnapshotWindow                    string   `json:"SnapshotWindow,omitempty"`
	IsGlobal                          bool     `json:"IsGlobal"`
	SecondaryTenantId                 string   `json:"SecondaryTenantId,omitempty"`
	GlobalReplicationGroupDescription string   `json:"GlobalReplicationGroupDescription,omitempty"`
	GlobalReplicationGroupId          string   `json:"GlobalReplicationGroupId,omitempty"`
	IsPrimary                         bool     `json:"IsPrimary,omitempty"`
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

type DuplocloudEcacheSnapshotRetentionLimitUpdateRequest struct {
	Identifier             string `json:"Identifier"`
	SnapshotRetentionLimit string `json:"SnapshotRetentionLimit"`
}

func (c *Client) EcacheInstanceUpdateSnapshotRetentionLimit(tenantID, name string, rq DuplocloudEcacheSnapshotRetentionLimitUpdateRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("EcacheInstanceUpdateSnapshotRetentionLimit(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/ECacheInstanceUpdateRetentionLimit", tenantID),
		&rq, nil)
}

func (c *Client) EcacheInstanceGet(tenantID, name string) (*DuploEcacheInstance, ClientError) {

	// Call the API.
	rp := DuploEcacheInstance{}
	conf := NewRetryConf()
	err := c.getAPIWithRetry(
		fmt.Sprintf("EcacheInstanceGet(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/duplo-%s", tenantID, name),
		&rp, &conf)
	if err != nil || rp.Identifier == "" {
		return nil, err
	}

	// Fill in the tenant ID and the name and return the object
	rp.TenantID = tenantID
	rp.Name = name
	return &rp, nil
}

type DuploEcacheGlobalPrimaryInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier               string `json:"Identifier"`
	CacheType                int    `json:"CacheType,omitempty"`
	EngineVersion            string `json:"EngineVersion,omitempty"`
	Size                     string `json:"Size,omitempty"`
	Replicas                 int    `json:"Replicas,omitempty"`
	ParameterGroupName       string `json:"ParameterGroupName,omitempty"`
	EnableClusterMode        bool   `json:"ClusteringEnabled,omitempty"`
	AutomaticFailoverEnabled bool   `json:"AutomaticFailoverEnabled,omitempty"`
	NumberOfShards           int    `json:"NoOfShards,omitempty"`
	IsGlobal                 bool   `json:"IsGlobal"`
	IsPrimary                bool   `json:"IsPrimary,omitempty"`
}

func (c *Client) DuploPrimaryEcacheCreate(tenantID string, rq *DuploEcacheInstance) ClientError {
	rp := map[string]interface{}{}
	return c.postAPI(
		fmt.Sprintf("DuploPrimaryEcacheCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore/createprimary", tenantID),
		&rq,
		&rp,
	)
}

type DuploEcacheGlobalDatastore struct {
	Description              string `json:"GlobalReplicationGroupDescription"`
	GlobalReplicationGroupId string `json:"GlobalReplicationGroupIdSuffix"`
	PrimaryInstance          string `json:"PrimaryReplicationGroupId"`
}

type DuploEcacheGlobalDatastoreResponse struct {
	GlobalReplicationGroup struct {
		GlobalReplicationGroupId string `json:"GlobalReplicationGroupId"`
	} `json:"GlobalReplicationGroup,omitempty"`
	GlobalReplicationGroupId          string                         `json:"GlobalReplicationGroupId"`
	GlobalReplicationGroupDescription string                         `json:"GlobalReplicationGroupDescription"`
	Members                           []DuploEcacheReplicationMember `json:"Members,omitempty"`
	Status                            string                         `json:"Status"`
	ReplicationGroup                  struct {
		ReplicationGroupId string `json:"ReplicationGroupId"`
	} `json:"ReplicationGroup,omitempty"`
	ReplicationGroupId string `json:"ReplicationGroupId"`
}

type DuploEcacheReplicationMember struct {
	ReplicationGroupId     string `json:"ReplicationGroupId"`
	Role                   string `json:"Role"`
	Status                 string `json:"Status"`
	TenantId               string `json:"TenantId"`
	ReplicationGroupRegion string `json:"ReplicationGroupRegion"`
	KmsKeyId               string `json:"KmsKeyId"`
	AuthToken              string `json:"AuthToken"`
}

func (c *Client) DuploEcacheGlobalDatastoreCreate(tenantID string, rq *DuploEcacheGlobalDatastore) (*DuploEcacheGlobalDatastoreResponse, ClientError) {
	rp := DuploEcacheGlobalDatastoreResponse{}
	err := c.postAPI(
		fmt.Sprintf("DuploEcacheGlobalDatastoreCreate(%s,%s)", tenantID, rq.GlobalReplicationGroupId),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploEcacheGlobalDatastoreGet(tenantID, name string) (*DuploEcacheGlobalDatastoreResponse, ClientError) {

	// Call the API.
	rp := []DuploEcacheGlobalDatastoreResponse{}
	err := c.getAPI(
		fmt.Sprintf("EcacheInstanceGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore", tenantID),
		&rp)

	for _, gds := range rp {
		log.Println("Respnse of DuploEcacheGlobalDatastoreGet \n ", gds)
		if gds.GlobalReplicationGroupId == name {
			log.Println("Name ", name, " Status ", gds.Status)
			return &gds, nil
		}
	}
	return nil, err
}

func (c *Client) DuploEcacheGlobalDatastoreDelete(tenantID, name string) ClientError {
	var rp interface{}
	return c.deleteAPI(
		fmt.Sprintf("DuploEcacheGlobalDatastoreDelete(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore/%s", tenantID, name),
		&rp,
	)

}

type DuploEcacheReplicationGroup struct {
	Description              string `json:"GlobalReplicationGroupDescription"`
	GlobalReplicationGroupId string `json:"GlobalReplicationGroupId"`
	SecondaryTenantId        string `json:"TenantId"`
	ReplicationGroupId       string `json:"ReplicationGroupId"`
	KmsKeyId                 string `json:"KmsKeyId,omitempty"`
	AuthToken                string `json:"AuthToken,omitempty"`
}

func (c *Client) DuploEcacheReplicationGroupCreate(tenantID string, rq *DuploEcacheReplicationGroup) (*DuploEcacheGlobalDatastoreResponse, ClientError) {
	rp := DuploEcacheGlobalDatastoreResponse{}
	err := c.postAPI(
		fmt.Sprintf("DuploEcacheReplicationGroupCreate(%s,%s)", tenantID, rq.ReplicationGroupId),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore/associate", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploEcacheReplicationGroupGet(tenantID, gDatastore, scTenantId, name string) (*DuploEcacheGlobalDatastoreResponse, *DuploEcacheReplicationMember, ClientError) {

	// Call the API.
	rp := []DuploEcacheGlobalDatastoreResponse{}
	err := c.getAPI(
		fmt.Sprintf("DuploEcacheReplicationGroupGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore", tenantID),
		&rp)

	for _, gds := range rp {
		if gds.GlobalReplicationGroupId == gDatastore {
			for _, m := range gds.Members {
				if m.ReplicationGroupId == name && scTenantId == m.TenantId {
					return &gds, &m, nil
				}
			}
		}
	}
	return nil, nil, err
}

func (c *Client) DuploEcacheReplicationGroupDisassociate(tenantID, secTenantId, glbDatastore, name string) ClientError {
	var rp interface{}
	req := map[string]interface{}{
		"GlobalReplicationGroupId": glbDatastore,
		"ReplicationGroupId":       name,
		"Region":                   secTenantId,
	}
	return c.postAPI(
		fmt.Sprintf("DuploEcacheReplicationGroupDisassociate(%s,%s,%s)", tenantID, glbDatastore, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecache/globaldatastore/disassociate", tenantID),
		&req,
		&rp,
	)

}

// serverless valkey

type DuploValkeyServerless struct {
	Name                   string `json:"ServerlessCacheName"`
	Description            string `json:"Description"`
	Engine                 string `json:"Engine"`
	KMSKeyId               string `json:"KmsKeyId,omitempty"`
	EngineVersion          string `json:"MajorEngineVersion"`
	SnapshotRetentionLimit int    `json:"SnapshotRetentionLimit,omitempty"`
}

type DuploValkeyServerlessResponse struct {
	ARN                    string                         `json:"ARN"`
	CreateTime             string                         `json:"CreateTime"`
	DailySnapshotTime      string                         `json:"DailySnapshotTime"`
	Description            string                         `json:"Description"`
	Endpoint               *DuploValkeyServerlessEndpoint `json:"Endpoint"`
	Engine                 string                         `json:"Engine"`
	FullEngineVersion      string                         `json:"FullEngineVersion"`
	KmsKeyId               string                         `json:"KmsKeyId"`
	MajorEngineVersion     string                         `json:"MajorEngineVersion"`
	SecurityGroupIds       []string                       `json:"SecurityGroupIds"`
	ServerlessCacheName    string                         `json:"ServerlessCacheName"`
	SnapshotRetentionLimit int                            `json:"SnapshotRetentionLimit"`
	SubnetIds              []string                       `json:"SubnetIds"`
	Arn                    string                         `json:"Arn"`
	Status                 string                         `json:"Status"`
	ResourceType           int                            `json:"ResourceType"`
	Name                   string                         `json:"Name"`
}

type DuploValkeyServerlessEndpoint struct {
	Address string `json:"Address"`
	Port    int    `json:"Port"`
}

func (c *Client) DuploValkeyServerlessCreate(tenantID string, rq *DuploValkeyServerless) (*DuploValkeyServerlessResponse, ClientError) {
	rp := DuploValkeyServerlessResponse{}
	err := c.postAPI(
		fmt.Sprintf("DuploValkeyServerlessCreate(%s,%s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/valkey", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploValkeyServerlessGet(tenantID, name string) (*DuploValkeyServerlessResponse, ClientError) {

	// Call the API.
	rp := DuploValkeyServerlessResponse{}
	err := c.getAPI(
		fmt.Sprintf("DuploValkeyServerlessGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/valkey/%s", tenantID, name),
		&rp)

	return nil, err
}

func (c *Client) DuploValkeyServerlessDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploValkeyServerlessDelete(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/valkey/%s", tenantID, name),
		nil)
}
