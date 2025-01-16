package duplosdk

import (
	"fmt"
	"strings"
)

const (
	RDS_TYPE_CLUSTER  string = "cluster"
	RDS_TYPE_INSTANCE string = "instance"
)

const (
	DUPLO_RDS_ENGINE_MYSQL             = 0
	DUPLO_RDS_ENGINE_POSTGRESQL        = 1
	DUPLO_RDS_ENGINE_MSSQL_EXPRESS     = 2
	DUPLO_RDS_ENGINE_MSSQL_STANDARD    = 3
	DUPLO_RDS_ENGINE_AURORA_MYSQL      = 8
	DUPLO_RDS_ENGINE_AURORA_POSTGRESQL = 9
	DUPLO_RDS_ENGINE_MSSQL_WEB         = 10
	DUPLO_RDS_ENGINE_DOCUMENTDB        = 13
)

const (
	REDIS_LOG_DELIVERYDIST_DEST_TYPE_CLOUDWATCH_LOGS  string = "cloudwatch-logs"
	REDIS_LOG_DELIVERYDIST_DEST_TYPE_KINESIS_FIREHOSE string = "kinesis-firehose"
	REDIS_LOG_DELIVERY_LOG_FORMAT_JSON                string = "json"
	REDIS_LOG_DELIVERY_LOG_FORMAT_TEXT                string = "text"
	REDIS_LOG_DELIVERY_LOG_TYPE_SLOW_LOG              string = "slow-log"
	REDIS_LOG_DELIVERY_LOG_TYPE_ENGINE_LOG            string = "engine-log"
)

//"slow-log", "engine-log"

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                         string                  `json:"Identifier"`
	ClusterIdentifier                  string                  `json:"ClusterIdentifier,omitempty"`
	ReplicationSourceIdentifier        string                  `json:"ReplicationSourceIdentifier,omitempty"`
	Arn                                string                  `json:"Arn"`
	Endpoint                           string                  `json:"Endpoint,omitempty"`
	MasterUsername                     string                  `json:"MasterUsername,omitempty"`
	MasterPassword                     string                  `json:"MasterPassword,omitempty"`
	Engine                             int                     `json:"Engine,omitempty"`
	EngineVersion                      string                  `json:"EngineVersion,omitempty"`
	SnapshotID                         string                  `json:"SnapshotId,omitempty"`
	DBParameterGroupName               string                  `json:"DBParameterGroupName,omitempty"`
	ClusterParameterGroupName          string                  `json:"ClusterParameterGroupName,omitempty"`
	StoreDetailsInSecretManager        bool                    `json:"StoreDetailsInSecretManager,omitempty"`
	Cloud                              int                     `json:"Cloud,omitempty"`
	SizeEx                             string                  `json:"SizeEx,omitempty"`
	EncryptStorage                     bool                    `json:"EncryptStorage,omitempty"`
	StorageType                        string                  `json:"StorageType,omitempty"`
	Iops                               int                     `json:"Iops,omitempty"`
	AllocatedStorage                   int                     `json:"AllocatedStorage,omitempty"`
	EncryptionKmsKeyId                 string                  `json:"EncryptionKmsKeyId,omitempty"`
	EnableLogging                      bool                    `json:"EnableLogging,omitempty"`
	BackupRetentionPeriod              int                     `json:"BackupRetentionPeriod,omitempty"`
	SkipFinalSnapshot                  bool                    `json:"SkipFinalSnapshot,omitempty"`
	MultiAZ                            bool                    `json:"MultiAZ,omitempty"`
	InstanceStatus                     string                  `json:"InstanceStatus,omitempty"`
	DBSubnetGroupName                  string                  `json:"DBSubnetGroupName,omitempty"`
	V2ScalingConfiguration             *V2ScalingConfiguration `json:"V2ScalingConfiguration,omitempty"`
	AvailabilityZone                   string                  `json:"AvailabilityZone,omitempty"`
	EnableIamAuth                      bool                    `json:"EnableIamAuth"`
	MonitoringInterval                 int                     `json:"MonitoringInterval"`
	DatabaseName                       string                  `json:"DatabaseName,omitempty"`
	EnablePerformanceInsights          bool                    `json:"EnablePerformanceInsights"`
	PerformanceInsightsRetentionPeriod int                     `json:"PerformanceInsightsRetentionPeriod,omitempty"`
	PerformanceInsightsKMSKeyId        string                  `json:"PerformanceInsightsKMSKeyId,omitempty"`
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

// DuploRdsUpdatePayload is a Duplo SDK object that represents an update payload for size and enabling/disabling logging
type DuploRdsUpdatePayload struct {
	EnableLogging             *bool  `json:"EnableLogging,omitempty"`
	SizeEx                    string `json:"SizeEx,omitempty"`
	DbParameterGroupName      string `json:"DbParameterGroupName,omitempty"`
	ClusterParameterGroupName string `json:"ClusterParameterGroupName,omitempty"`
}

type DuploRdsUpdateInstance struct {
	DBInstanceIdentifier  string `json:"DBInstanceIdentifier"`
	DeletionProtection    *bool  `json:"DeletionProtection,omitempty"`
	BackupRetentionPeriod int    `json:"BackupRetentionPeriod,omitempty"`
	SkipFinalSnapshot     bool   `json:"SkipFinalSnapshot"`
}

type DuploRdsUpdatePerformanceInsights struct {
	DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
	Enable               *PerformanceInsightEnable
	Disable              *PerformanceInsightDisable
}
type PerformanceInsightEnable struct {
	EnablePerformanceInsights          bool   `json:"EnablePerformanceInsights"`
	PerformanceInsightsRetentionPeriod int    `json:"PerformanceInsightsRetentionPeriod,omitempty"`
	PerformanceInsightsKMSKeyId        string `json:"PerformanceInsightsKMSKeyId,omitempty"`
	ApplyImmediately                   bool   `json:"ApplyImmediately"`
}
type PerformanceInsightDisable struct {
	EnablePerformanceInsights bool `json:"EnablePerformanceInsights"`
}
type DuploRdsUpdateCluster struct {
	DBClusterIdentifier                string `json:"DBClusterIdentifier"`
	ApplyImmediately                   bool   `json:"ApplyImmediately"`
	DeletionProtection                 *bool  `json:"DeletionProtection,omitempty"`
	BackupRetentionPeriod              int    `json:"BackupRetentionPeriod,omitempty"`
	SkipFinalSnapshot                  bool   `json:"SkipFinalSnapshot"`
	EnablePerformanceInsights          bool   `json:"EnablePerformanceInsights,omitempty"`
	PerformanceInsightsRetentionPeriod int    `json:"PerformanceInsightsRetentionPeriod,omitempty"`
	PerformanceInsightsKMSKeyId        string `json:"PerformanceInsightsKMSKeyId,omitempty"`
}

type DuploRdsModifyAuroraV2ServerlessInstanceSize struct {
	Identifier             string                  `json:"Identifier"`
	ClusterIdentifier      string                  `json:"ClusterIdentifier"`
	ApplyImmediately       bool                    `json:"ApplyImmediately"`
	SizeEx                 string                  `json:"SizeEx,omitempty"`
	V2ScalingConfiguration *V2ScalingConfiguration `json:"V2ScalingConfiguration,omitempty"`
}

type DuploRDSTag struct {
	ResourceType string `json:"ResourceType"`
	ResourceId   string `json:"ResourceId"`
	Key          string `json:"Key"`
	Value        string `json:"Value"`
}

type DuploMonitoringInterval struct {
	DBInstanceIdentifier string `json:"DBInstanceIdentifier"`
	ApplyImmediately     bool   `json:"ApplyImmediately"`
	MonitoringInterval   int    `json:"MonitoringInterval"`
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

	// call update request
	if updating {
		rp := DuploRdsInstance{}
		err := c.doAPIWithRequestBody(
			"PUT",
			fmt.Sprintf("RdsInstanceCreateOrUpdate(%s, duplo%s)", tenantID, duploObject.Name),
			fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, duploObject.Identifier),
			&duploObject,
			&rp,
		)
		if err != nil {
			return nil, err
		}
		return &rp, err
	}

	// Call create API.
	rp := DuploRdsInstance{}
	err := c.doAPIWithRequestBody(
		"POST",
		fmt.Sprintf("RdsInstanceCreateOrUpdate(%s, duplo%s)", tenantID, duploObject.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance", tenantID),
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
	identifier := EnsureDuploPrefixInRdsIdentifier(name)
	// Call the API.
	err := c.deleteAPI(
		fmt.Sprintf("RdsInstanceDelete(%s, %s)", tenantID, identifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, identifier),
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
	identifier := EnsureDuploPrefixInRdsIdentifier(name)
	// Call the API.
	duploObject := DuploRdsInstance{}
	err := c.getAPI(
		fmt.Sprintf("RdsInstanceGet(%s, %s)", tenantID, identifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, identifier),
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
	identifier := EnsureDuploPrefixInRdsIdentifier(name)
	// Call the API.
	duploObject := DuploRdsInstance{}
	err := c.getAPI(
		fmt.Sprintf("RdsInstanceGet(%s, %s)", tenantID, identifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, identifier),
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
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s/changePassword", tenantID, duploObject.Identifier),
		&duploObject,
		nil,
	)
}

// RdsInstanceChangeSizeOrEnableLogging changes the size of an RDS instance or enables logging via the Duplo API.
// DuploRdsUpdatePayload, despite the name, is only used for size and logging changes.
func (c *Client) RdsInstanceChangeSizeOrEnableLogging(tenantID string, instanceId string, rdsUpdate *DuploRdsUpdatePayload) error {
	return c.putAPI(
		fmt.Sprintf("RdsInstanceChangeSizeOrEnableLogging(%s, %s, %+v)", tenantID, instanceId, rdsUpdate),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s/updatePayload", tenantID, instanceId),
		&rdsUpdate,
		nil,
	)
}

func (c *Client) UpdateRDSDBInstance(tenantID string, duploObject DuploRdsUpdateInstance) ClientError {
	return c.putAPI(
		fmt.Sprintf("UpdateRDSDBInstance(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, duploObject.DBInstanceIdentifier),
		&duploObject,
		nil,
	)
}

func (c *Client) UpdateDBInstancePerformanceInsight(tenantID string, duploObject DuploRdsUpdatePerformanceInsights) ClientError {
	if duploObject.Enable != nil {
		err := c.putAPI(
			fmt.Sprintf("UpdateDBInstancePerformanceInsight(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
			fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, duploObject.DBInstanceIdentifier),
			duploObject.Enable,
			nil,
		)
		return err
	}
	err := c.putAPI(
		fmt.Sprintf("UpdateDBInstancePerformanceInsight(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s", tenantID, duploObject.DBInstanceIdentifier),
		duploObject.Disable,
		nil,
	)
	return err
}

func (c *Client) UpdateDBClusterPerformanceInsight(tenantID string, duploObject DuploRdsUpdatePerformanceInsights) ClientError {
	if duploObject.Enable != nil {
		err := c.putAPI(
			fmt.Sprintf("UpdateDBClusterPerformanceInsight(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
			fmt.Sprintf("v3/subscriptions/%s/aws/rds/cluster/%s", tenantID, duploObject.DBInstanceIdentifier),
			duploObject.Enable,
			nil,
		)
		return err
	}
	err := c.putAPI(
		fmt.Sprintf("UpdateDBClusterPerformanceInsight(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/cluster/%s", tenantID, duploObject.DBInstanceIdentifier),
		duploObject.Disable,
		nil,
	)
	return err
}

func (c *Client) UpdateRdsCluster(tenantID string, duploObject DuploRdsUpdateCluster) ClientError {
	return c.putAPI(
		fmt.Sprintf("UpdateRdsCluster(%s, %s)", tenantID, duploObject.DBClusterIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/cluster/%s", tenantID, duploObject.DBClusterIdentifier),
		&duploObject,
		nil,
	)
}

func (c *Client) RdsModifyAuroraV2ServerlessInstanceSize(tenantID string, duploObject DuploRdsModifyAuroraV2ServerlessInstanceSize) ClientError {
	return c.postAPI(
		fmt.Sprintf("RdsModifyAuroraV2ServerlessInstanceSize(%s, %s)", tenantID, duploObject.ClusterIdentifier),
		fmt.Sprintf("v3/subscriptions/%s/aws/modifyAuroraV2Serverless", tenantID),
		&duploObject,
		nil,
	)
}

func (c *Client) RdsUpdateMonitoringInterval(tenantID string, duploObject DuploMonitoringInterval) ClientError {
	return c.postAPI(
		fmt.Sprintf("RdsUpdateMonitoringInterval(%s, %s)", tenantID, duploObject.DBInstanceIdentifier),
		fmt.Sprintf("/subscriptions/%s/ModifyRDSDBInstance", tenantID),
		&duploObject,
		nil,
	)
}

func RdsIsAurora(engine int) bool {
	return engine == DUPLO_RDS_ENGINE_AURORA_MYSQL ||
		engine == DUPLO_RDS_ENGINE_AURORA_POSTGRESQL
}

func RdsIsMsSQL(engine int) bool {
	return engine == DUPLO_RDS_ENGINE_MSSQL_EXPRESS ||
		engine == DUPLO_RDS_ENGINE_MSSQL_STANDARD ||
		engine == DUPLO_RDS_ENGINE_MSSQL_WEB
}

/*************************************************
 * API Calls for RDS Tags
 */

func (c *Client) RdsTagCreateV3(tenantID string, tag DuploRDSTag) ClientError {
	resp := &DuploKeyStringValue{}
	return c.postAPI(
		fmt.Sprintf("RdsTagCreateV3(%s, %s)", tenantID, tag.ResourceId),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/%s/%s/tag", tenantID, tag.ResourceType, tag.ResourceId),
		&DuploKeyStringValue{
			Key:   tag.Key,
			Value: tag.Value,
		},
		&resp,
	)
}

func (c *Client) RdsTagUpdateV3(tenantID string, tag DuploRDSTag) ClientError {
	resp := &DuploKeyStringValue{}
	return c.putAPI(
		fmt.Sprintf("RdsTagUpdateV3(%s, %s)", tenantID, tag.ResourceId),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/%s/%s/tag/%s", tenantID, tag.ResourceType, tag.ResourceId, urlSafeBase64Encode(tag.Key)),
		&DuploKeyStringValue{
			Key:   tag.Key,
			Value: tag.Value,
		},
		&resp,
	)
}

func (c *Client) RdsTagListV3(tenantID, resourceType, resourceId string) (*[]DuploKeyStringValue, ClientError) {
	tags := []DuploKeyStringValue{}
	err := c.getAPI(
		fmt.Sprintf("RdsTagListV3(%s, %s)", tenantID, resourceId),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/%s/%s/tag", tenantID, resourceType, resourceId),
		&tags,
	)
	return &tags, err
}

func (c *Client) RdsTagGetV3(tenantID string, tag DuploRDSTag) (*DuploKeyStringValue, ClientError) {
	tags := DuploKeyStringValue{}
	err := c.getAPI(
		fmt.Sprintf("RdsTagGetV3(%s, %s)", tenantID, tag.ResourceId),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/%s/%s/tag/%s", tenantID, tag.ResourceType, tag.ResourceId, urlSafeBase64Encode(tag.Key)),
		&tags,
	)
	return &tags, err
}

func (c *Client) RdsTagDeleteV3(tenantID string, tag DuploRDSTag) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("RdsTagDeleteV3(%s, %s)", tenantID, tag.ResourceId),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/%s/%s/tag/%s", tenantID, tag.ResourceType, tag.ResourceId, urlSafeBase64Encode(tag.Key)),
		nil,
	)
}

const (
	AWSRdsPrefix = "duplo"
)

func EnsureDuploPrefixInRdsIdentifier(name string) string {
	identifier := strings.TrimSpace(name)
	for strings.HasPrefix(identifier, AWSRdsPrefix) {
		identifier = strings.TrimPrefix(identifier, AWSRdsPrefix)
		identifier = strings.TrimSpace(identifier)
	}
	return AWSRdsPrefix + identifier
}

func ValidateRdsNoDoubleDuploPrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	trimmedValue := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmedValue, AWSRdsPrefix) {
		return
	}
	// Check for multiple consecutive 'duplo' prefixes
	count := 0
	for strings.HasPrefix(trimmedValue, AWSRdsPrefix) {
		trimmedValue = strings.TrimPrefix(trimmedValue, AWSRdsPrefix)
		trimmedValue = strings.TrimSpace(trimmedValue)
		count++
	}
	if count > 1 {
		errors = append(errors, fmt.Errorf("%q cannot contain multiple consecutive '%s' prefixes", k, AWSRdsPrefix))
	}
	return
}

func (c *Client) RdsInstanceUpdateParameterGroupName(tenantID string, instanceId string, rdsUpdate *DuploRdsUpdatePayload) error {
	return c.putAPI(
		fmt.Sprintf("RdsInstanceUpdateParameterGroupName(%s, %s, %+v)", tenantID, instanceId, rdsUpdate),
		fmt.Sprintf("v3/subscriptions/%s/aws/rds/instance/%s/updatePayload", tenantID, instanceId),
		&rdsUpdate,
		nil,
	)
}
