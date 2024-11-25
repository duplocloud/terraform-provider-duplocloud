package duplosdk

import (
	"fmt"
	"log"
	"time"
)

const (
	// ResourceTypeS3Bucket represents an S3 bucket
	ResourceTypeS3Bucket int = 1

	// ResourceTypeDynamoDBTable represents an DynamoDB table
	ResourceTypeDynamoDBTable int = 2

	// ResourceTypeKafkaCluster represents a Kafka cluster
	ResourceTypeKafkaCluster int = 14

	// ResourceTypeApplicationLB represents an AWS application LB
	ResourceTypeApplicationLB int = 16

	// ResourceTypeApiGatewayRestAPI represents an AWS Api gateway REST API
	ResourceTypeApiGatewayRestAPI int = 8

	ResourceTypeSQSQueue int = 3

	ResourceTypeSNSTopic int = 4

	ResourceTypeLambdaFunction int = 7

	ResourceTypeElasticSearch int = 24
)

type CustomComponentType int

const (
	NETWORK CustomComponentType = iota
	REPLICATIONCONTROLLER
	MINION
	ASG
	AGENTPOOL
)

type CustomDataUpdate struct {
	ComponentId   string              `json:"ComponentId,omitempty"`
	ComponentType CustomComponentType `json:"ComponentType,omitempty"`
	State         string              `json:"State,omitempty"`
	Key           string              `json:"Key,omitempty"`
	Value         string              `json:"Value,omitempty"`
}

// DuploAwsCloudResource represents a generic AWS cloud resource for a Duplo tenant
type DuploAwsCloudResource struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Type     int    `json:"ResourceType,omitempty"`
	Name     string `json:"Name,omitempty"`
	Arn      string `json:"Arn,omitempty"`
	MetaData string `json:"MetaData,omitempty"`

	// S3 bucket and load balancer
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`

	// Only S3 bucket
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`

	// Only Load balancer
	IsInternal bool   `json:"IsInternal,omitempty"`
	WebACLID   string `json:"WebACLID,omitempty"`
}

// DuploS3Bucket represents an S3 bucket resource for a Duplo tenant
type DuploS3Bucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name              string                 `json:"Name,omitempty"`
	DomainName        string                 `json:"DomainName,omitempty"`
	Region            string                 `json:"Region,omitempty"`
	Location          string                 `json:"Location,omitempty"`
	Arn               string                 `json:"Arn,omitempty"`
	MetaData          string                 `json:"MetaData,omitempty"`
	EnableVersioning  bool                   `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool                   `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool                   `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string                 `json:"DefaultEncryption,omitempty"`
	Policies          []string               `json:"Policies,omitempty"`
	Tags              *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

type DuploGCPBucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                  string            `json:"Name,omitempty"`
	DomainName            string            `json:"SelfLink,omitempty"`
	Location              string            `json:"Location,omitempty"`
	EnableVersioning      bool              `json:"EnableVersioning,omitempty"`
	AllowPublicAccess     bool              `json:"AllowPublicAccess,omitempty"`
	DefaultEncryptionType int               `json:"DefaultEncryptionType,omitempty"`
	Labels                map[string]string `json:"Labels,omitempty"`
}

// DuploApplicationLB represents an AWS application load balancer resource for a Duplo tenant
type DuploApplicationLB struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name             string                 `json:"Name,omitempty"`
	Arn              string                 `json:"Arn,omitempty"`
	DNSName          string                 `json:"MetaData,omitempty"`
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	IsInternal       bool                   `json:"IsInternal,omitempty"`
	WebACLID         string                 `json:"WebACLID,omitempty"`
	LbType           *DuploStringValue      `json:"LbType,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploAwsLBConfiguration represents a request to create an AWS application load balancer resource
type DuploAwsLBConfiguration struct {
	Name             string `json:"Name"`
	State            string `json:"State,omitempty"`
	IsInternal       bool   `json:"IsInternal,omitempty"`
	EnableAccessLogs bool   `json:"EnableAccessLogs,omitempty"`
	RequestedLbType  string `json:"RequestedLbType,omitempty"`
}

type DuploAwsLbState struct {
	Code *DuploStringValue `json:"Code,omitempty"`
}

type DuploAwsLbAvailabilityZone struct {
	SubnetID string `json:"SubnetId,omitempty"`
	ZoneName string `json:"ZoneName,omitempty"`
}

// DuploAwsLbSettings represents an AWS application load balancer's details, via a Duplo Service
type DuploAwsLbDetailsInService struct {
	LoadBalancerName      string                       `json:"LoadBalancerName"`
	LoadBalancerArn       string                       `json:"LoadBalancerArn"`
	AvailabilityZones     []DuploAwsLbAvailabilityZone `json:"AvailabilityZones"`
	CanonicalHostedZoneId string                       `json:"CanonicalHostedZoneId"`
	CreatedTime           time.Time                    `json:"CreatedTime"`
	DNSName               string                       `json:"DNSName"`
	IPAddressType         *DuploStringValue            `json:"IPAddressType,omitempty"`
	Scheme                *DuploStringValue            `json:"Scheme,omitempty"`
	Type                  *DuploStringValue            `json:"Type,omitempty"`
	SecurityGroups        []string                     `json:"SecurityGroups"`
	State                 *DuploAwsLbState             `json:"State,omitempty"`
	VpcID                 string                       `json:"VpcId,omitempty"`
}

// DuploAwsLbListenerCertificate represents a AWS load balancer listener's SSL Certificate
type DuploAwsLbListenerCertificate struct {
	CertificateArn string `json:"CertificateArn"`
	IsDefault      bool   `json:"IsDefault,omitempty"`
}

// DuploAwsLbListenerAction represents a AWS load balancer listener action
type DuploAwsLbListenerAction struct {
	Order          int               `json:"Order"`
	TargetGroupArn string            `json:"TargetGroupArn"`
	Type           *DuploStringValue `json:"Type,omitempty"`
}

type DuploAwsLbListenerActionCreate struct {
	TargetGroupArn string `json:"TargetGroupArn"`
	Type           string `json:"Type,omitempty"`
}

type DuploAwsLbListenerDeleteRequest struct {
	ListenerArn string `json:"ListenerArn"`
}

// DuploAwsLbListener represents an AWS application load balancer listener
type DuploAwsLbListener struct {
	LoadBalancerArn string                          `json:"LoadBalancerArn"`
	Certificates    []DuploAwsLbListenerCertificate `json:"Certificates"`
	ListenerArn     string                          `json:"ListenerArn"`
	SSLPolicy       string                          `json:"SslPolicy,omitempty"`
	Port            int                             `json:"Port"`
	Protocol        *DuploStringValue               `json:"Protocol,omitempty"`
	DefaultActions  []DuploAwsLbListenerAction      `json:"DefaultActions"`
}

type DuploAwsLbListenerCreate struct {
	Certificates   []DuploAwsLbListenerCertificate  `json:"Certificates"`
	Port           int                              `json:"Port"`
	Protocol       string                           `json:"Protocol,omitempty"`
	DefaultActions []DuploAwsLbListenerActionCreate `json:"DefaultActions"`
}

// DuploAwsTargetGroupMatcher represents an AWS lb target group matcher
type DuploAwsTargetGroupMatcher struct {
	HttpCode string `json:"HttpCode"`
	GrpcCode string `json:"GrpcCode"`
}

// DuploAwsLbTargetGroup represents an AWS lb target group
type DuploAwsLbTargetGroup struct {
	HealthCheckEnabled         bool                        `json:"HealthCheckEnabled"`
	HealthCheckIntervalSeconds int                         `json:"HealthCheckIntervalSeconds"`
	HealthCheckPath            string                      `json:"HealthCheckPath"`
	HealthCheckPort            string                      `json:"HealthCheckPort"`
	HealthCheckProtocol        *DuploStringValue           `json:"HealthCheckProtocol,omitempty"`
	HealthyThreshold           int                         `json:"HealthyThresholdCount"`
	HealthCheckTimeoutSeconds  int                         `json:"HealthCheckTimeoutSeconds"`
	LoadBalancerArns           []string                    `json:"LoadBalancerArns"`
	HealthMatcher              *DuploAwsTargetGroupMatcher `json:"Matcher,omitempty"`
	Protocol                   *DuploStringValue           `json:"Protocol,omitempty"`
	ProtocolVersion            string                      `json:"ProtocolVersion"`
	TargetGroupArn             string                      `json:"TargetGroupArn"`
	TargetGroupName            string                      `json:"TargetGroupName"`
	TargetType                 *DuploStringValue           `json:"TargetType,omitempty"`
	UnhealthyThreshold         int                         `json:"UnhealthThresholdCount"`
	VpcID                      string                      `json:"VpcId"`
}

// DuploS3BucketRequest represents a request to create an S3 bucket resource
type DuploS3BucketRequest struct {
	Type           int    `json:"ResourceType"`
	Name           string `json:"Name"`
	State          string `json:"State,omitempty"`
	InTenantRegion bool   `json:"InTenantRegion"`
}

// DuploS3BucketSettingsRequest represents a request to create an S3 bucket resource
type DuploS3BucketSettingsRequest struct {
	Name              string   `json:"Name,omitempty"`
	Region            string   `json:"Region,omitempty"`
	Location          string   `json:"Location,omitempty"`
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool     `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`
}

type DuploS3BucketReplication struct {
	Rule                    string `json:"Name"`
	DestinationBucket       string `json:"DestinationBucket"`
	SourceBucket            string `json:"SourceBucket"`
	Priority                int    `json:"Priority"`
	DeleteMarkerReplication bool   `json:"DeleteMarkerReplication"`
	StorageClass            string `json:"StorageClass"`
	DestinationBucketARN    string `json:"-"`
}

type DuploS3BucketReplicationRule struct {
	Rule              string           `json:"Id"`
	Priority          int              `json:"Priority"`
	Status            DuploStringValue `json:"Status"`
	DestinationBucket struct {
		BucketArn    string            `json:"BucketArn"`
		StorageClass *DuploStringValue `json:"StorageClass,omitempty"`
	} `json:"Destination"`
	DeleteMarkerReplication struct {
		Status DuploStringValue `json:"Status"`
	} `json:"DeleteMarkerReplication"`
}

type DuploS3BucketReplicationResponse struct {
	Rule []DuploS3BucketReplicationRule `json:"Rules,omitempty"`
}

// DuploKafkaEbsStorageInfo represents a Kafka cluster's EBS storage info
type DuploKafkaEbsStorageInfo struct {
	VolumeSize int `json:"VolumeSize"`
}

// DuploKafkaBrokerStorageInfo represents a Kafka cluster's broker storage info
type DuploKafkaBrokerStorageInfo struct {
	EbsStorageInfo DuploKafkaEbsStorageInfo `json:"EbsStorageInfo"`
}

// DuploKafkaBrokerSoftwareInfo represents a Kafka cluster's broker software info
type DuploKafkaBrokerSoftwareInfo struct {
	ConfigurationArn      string `json:"ConfigurationArn,omitempty"`
	ConfigurationRevision int    `json:"ConfigurationRevision,omitempty"`
	KafkaVersion          string `json:"KafkaVersion,omitempty"`
}

// DuploKafkaClusterPrometheusExporter represents a Kafka cluster's prometheus exporter info
type DuploKafkaClusterPrometheusExporter struct {
	EnabledInBroker bool `json:"EnabledInBroker,omitempty"`
}

// DuploKafkaClusterPrometheus represents a Kafka cluster's prometheus info
type DuploKafkaClusterPrometheus struct {
	JmxExporter  *DuploKafkaClusterPrometheusExporter `json:"JmxExporter,omitempty"`
	NodeExporter *DuploKafkaClusterPrometheusExporter `json:"NodeExporter,omitempty"`
}

// DuploKafkaClusterOpenMonitoring represents a Kafka cluster's open monitoring info
type DuploKafkaClusterOpenMonitoring struct {
	Prometheus *DuploKafkaClusterPrometheus `json:"Prometheus,omitempty"`
}

// DuploKafkaClusterEncryptionAtRest represents a Kafka cluster's encryption-at-rest info
type DuploKafkaClusterEncryptionAtRest struct {
	KmsKeyID string `json:"DataVolumeKMSKeyId,omitempty"`
}

// DuploKafkaClusterEncryptionInTransit represents a Kafka cluster's encryption-in-transit info
type DuploKafkaClusterEncryptionInTransit struct {
	ClientBroker *DuploStringValue `json:"ClientBroker,omitempty"`
	InCluster    bool              `json:"InCluster,omitempty"`
}

// DuploKafkaClusterEncryptionInfo represents a Kafka cluster's encryption info
type DuploKafkaClusterEncryptionInfo struct {
	AtRest    *DuploKafkaClusterEncryptionAtRest    `json:"EncryptionAtRest,omitempty"`
	InTransit *DuploKafkaClusterEncryptionInTransit `json:"EncryptionInTransit,omitempty"`
}

// DuploKafkaBrokerNodeGroupInfo represents a Kafka cluster's broker node group info
type DuploKafkaBrokerNodeGroupInfo struct {
	InstanceType   string                      `json:"InstanceType,omitempty"`
	Subnets        *[]string                   `json:"ClientSubnets,omitempty"`
	SecurityGroups *[]string                   `json:"SecurityGroups,omitempty"`
	AZDistribution *DuploStringValue           `json:"BrokerAZDistribution,omitempty"`
	StorageInfo    DuploKafkaBrokerStorageInfo `json:"StorageInfo"`
}

// DuploKafkaConfigurationInfo represents a Kafka cluster's configuration
type DuploKafkaConfigurationInfo struct {
	Arn      string `json:"Arn,omitempty"`
	Revision int64  `json:"Revision,omitempty"`
}

// DuploKafkaClusterRequest represents a request to create a Kafka Cluster
type DuploKafkaClusterRequest struct {
	Name              string                           `json:"ClusterName,omitempty"`
	Arn               string                           `json:"ClusterArn,omitempty"`
	KafkaVersion      string                           `json:"KafkaVersion,omitempty"`
	BrokerNodeGroup   *DuploKafkaBrokerNodeGroupInfo   `json:"BrokerNodeGroupInfo,omitempty"`
	ConfigurationInfo *DuploKafkaConfigurationInfo     `json:"ConfigurationInfo,omitempty"`
	State             string                           `json:"State,omitempty"`
	EncryptionInfo    *DuploKafkaClusterEncryptionInfo `json:"EncryptionInfo,omitempty"`
}

// DuploKafkaCluster represents an AWS kafka cluster resource for a Duplo tenant
type DuploKafkaCluster struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name string `json:"Name,omitempty"`
	Arn  string `json:"Arn,omitempty"`
}

// DuploKafkaClusterInfo represents a non-cached view of an AWS kafka cluster for a Duplo tenant
type DuploKafkaClusterInfo struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                      string                           `json:"ClusterName,omitempty"`
	Arn                       string                           `json:"ClusterArn,omitempty"`
	CreationTime              time.Time                        `json:"CreationTime,omitempty"`
	CurrentVersion            string                           `json:"CurrentVersion,omitempty"`
	BrokerNodeGroup           *DuploKafkaBrokerNodeGroupInfo   `json:"BrokerNodeGroupInfo,omitempty"`
	CurrentSoftware           *DuploKafkaBrokerSoftwareInfo    `json:"CurrentBrokerSoftwareInfo,omitempty"`
	NumberOfBrokerNodes       int                              `json:"NumberOfBrokerNodes,omitempty"`
	EnhancedMonitoring        *DuploStringValue                `json:"EnhancedMonitoring,omitempty"`
	OpenMonitoring            *DuploKafkaClusterOpenMonitoring `json:"OpenMonitoring,omitempty"`
	State                     *DuploStringValue                `json:"State,omitempty"`
	Tags                      map[string]interface{}           `json:"Tags,omitempty"`
	ZookeeperConnectString    string                           `json:"ZookeeperConnectString,omitempty"`
	ZookeeperConnectStringTls string                           `json:"ZookeeperConnectStringTls,omitempty"`
	EncryptionInfo            *DuploKafkaClusterEncryptionInfo `json:"EncryptionInfo,omitempty"`
}

// DuploKafkaBootstrapBrokers represents a non-cached view of an AWS kafka cluster's bootstrap brokers for a Duplo tenant
type DuploKafkaBootstrapBrokers struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name,omitempty"`

	BootstrapBrokerString    string `json:"BootstrapBrokerString,omitempty"`
	BootstrapBrokerStringTls string `json:"BootstrapBrokerStringTls,omitempty"`
}

type DuploApiGatewayRequest struct {
	Name           string `json:"Name"`
	LambdaFunction string `json:"LambdaFunction,omitempty"`
	State          string `json:"State,omitempty"`
}

type DuploApiGatewayResource struct {
	Name         string `json:"Name"`
	MetaData     string `json:"MetaData,omitempty"`
	ResourceType int    `json:"ResourceType,omitempty"`
}

type DuploMinion struct {
	Name             string                 `json:"Name"`
	ConnectionURL    string                 `json:"ConnectionUrl"`
	NetworkAgentURL  string                 `json:"NetworkAgentUrl,omitempty"`
	ConnectionStatus string                 `json:"ConnectionStatus,omitempty"`
	Subnet           string                 `json:"Subnet,omitempty"`
	DirectAddress    string                 `json:"DirectAddress"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`
	ExternalAddress  string                 `json:"ExternalAddress,omitempty"`
	Tier             string                 `json:"Tier"`
	UserAccount      string                 `json:"UserAccount,omitempty"`
	Tunnel           int                    `json:"Tunnel"`
	AgentPlatform    int                    `json:"AgentPlatform"`
	Cloud            int                    `json:"Cloud"`
	Taints           []DuploMinionTaint     `json:"Taints"`
}

type DuploMinionDeleteReq struct {
	Name          string `json:"Name"`
	DirectAddress string `json:"DirectAddress"`
	State         string `json:"State"`
}

type DuploHostOOBData struct {
	IPAddress   string                 `json:"IpAddress"`
	InstanceId  string                 `json:"InstanceId"`
	Cloud       int                    `json:"Cloud"`
	Credentials *[]DuploHostCredential `json:"Credentials"`
}

type DuploHostCredential struct {
	Username   string `json:"Username"`
	Password   string `json:"Password,omitempty"`
	Privatekey string `json:"Privatekey,omitempty"`
}

// TenantListAwsCloudResources retrieves a list of the generic AWS cloud resources for a tenant via the Duplo API.
func (c *Client) TenantListAwsCloudResources(tenantID string) (*[]DuploAwsCloudResource, ClientError) {
	apiName := fmt.Sprintf("TenantListAwsCloudResources(%s)", tenantID)
	list := []DuploAwsCloudResource{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetCloudResources", tenantID), &list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID to each element and return the list.
	log.Printf("[TRACE] %s: %d items", apiName, len(list))
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// TenantGetAwsCloudResource retrieves a cloud resource by type and name
func (c *Client) TenantCreateV3S3Bucket(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, ClientError) {

	resp := DuploS3Bucket{}

	// Create the bucket via Duplo.
	err := c.postAPI(
		fmt.Sprintf("TenantCreateV3S3Bucket(%s, %s)", tenantID, duplo.Name),
		//  fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket", tenantID),
		&duplo,
		&resp)

	if err != nil || resp.Name == "" {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) TenantGetAwsCloudResource(tenantID string, resourceType int, name string) (*DuploAwsCloudResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific type and name.
	for _, resource := range *allResources {
		if resource.Type == resourceType && resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// TenantGetApplicationLbFullName retrieves the full name of a pass-thru AWS application load balancer.
func (c *Client) TenantGetApplicationLbFullName(tenantID string, name string) (string, ClientError) {
	return c.GetResourceName("duplo3", tenantID, name, false)
}

// TenantGetS3Bucket retrieves a managed S3 bucket via the Duplo API
func (c *Client) TenantGetS3Bucket(tenantID string, name string) (*DuploS3Bucket, ClientError) {
	// Figure out the full resource name.
	features, _ := c.AdminGetSystemFeatures()
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeS3Bucket, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploS3Bucket{
		TenantID:          tenantID,
		Name:              resource.Name,
		Arn:               resource.Arn,
		MetaData:          resource.MetaData,
		EnableVersioning:  resource.EnableVersioning,
		AllowPublicAccess: resource.AllowPublicAccess,
		EnableAccessLogs:  resource.EnableAccessLogs,
		DefaultEncryption: resource.DefaultEncryption,
		Policies:          resource.Policies,
		Tags:              resource.Tags,
	}, nil
}

// TenantGetKafkaCluster retrieves a managed Kafka Cluster via the Duplo API
func (c *Client) TenantGetKafkaCluster(tenantID string, name string) (*DuploKafkaCluster, ClientError) {
	// Figure out the full resource name.
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeKafkaCluster, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploKafkaCluster{
		TenantID: tenantID,
		Name:     resource.Name,
		Arn:      resource.Arn,
	}, nil
}

// TenantGetApplicationLB retrieves an application load balancer via the Duplo API
func (c *Client) TenantGetApplicationLB(tenantID string, name string) (*DuploApplicationLB, ClientError) {
	// Figure out the full resource name.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeApplicationLB, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploApplicationLB{
		TenantID:         tenantID,
		Name:             resource.Name,
		Arn:              resource.Arn,
		DNSName:          resource.MetaData,
		IsInternal:       resource.IsInternal,
		EnableAccessLogs: resource.EnableAccessLogs,
		Tags:             resource.Tags,
	}, nil
}

// TenantCreateV3S3Bucket creates an S3 bucket resource via V3 Duplo Api.

func (c *Client) GCPTenantCreateV3StorageBucketV2(tenantID string, duplo DuploGCPBucket) (*DuploS3Bucket, ClientError) {

	resp := DuploS3Bucket{}

	// Create the bucket via Duplo.
	err := c.postAPI(
		fmt.Sprintf("TenantCreateV3S3Bucket(%s, %s)", tenantID, duplo.Name),
		//  fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket", tenantID),
		&duplo,
		&resp)

	if err != nil || resp.Name == "" {
		return nil, err
	}
	return &resp, nil
}

// TenantDeleteV3S3Bucket deletes an S3 bucket resource via V3 Duplo Api.
func (c *Client) TenantDeleteV3S3Bucket(tenantID string, fullname string) ClientError {
	// Delete the bucket via Duplo.
	return c.deleteAPI(
		fmt.Sprintf("TenantDeleteV3S3Bucket(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s", tenantID, fullname),
		nil)
}

// TenantGetV3S3Bucket gets a non-cached view of the  S3 buckets's settings V3 Duplo Api.
func (c *Client) TenantGetV3S3Bucket(tenantID string, name string) (*DuploS3Bucket, ClientError) {
	rp := DuploS3Bucket{}
	err := c.getAPI(fmt.Sprintf("TenantGetV3S3Bucket(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s", tenantID, name),
		&rp)
	if err != nil || rp.Arn == "" {
		return nil, err
	}
	return &rp, err
}

func (c *Client) GCPTenantGetV3StorageBucketV2(tenantID string, name string) (*DuploGCPBucket, ClientError) {
	rp := DuploGCPBucket{}
	err := c.getAPI(fmt.Sprintf("GCPTenantGetV3StorageBucketV2(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, name),
		&rp)
	if err != nil { //|| rp.Arn == "" {
		return nil, err
	}
	return &rp, err
}

// TenantUpdateV3S3Bucket applies settings to an S3 bucket resource V3 Duplo Api.
func (c *Client) TenantUpdateV3S3Bucket(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, ClientError) {
	// Apply the settings via Duplo.
	apiName := fmt.Sprintf("TenantUpdateV3S3Bucket(%s, %s)", tenantID, duplo.Name)
	rp := DuploS3Bucket{}
	err := c.putAPI(apiName, fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s", tenantID, duplo.Name), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	// Deal with a missing response.
	if rp.Name == "" {
		message := fmt.Sprintf("%s: unexpected missing response from backend", apiName)
		log.Printf("[TRACE] %s", message)
		return nil, newClientError(message)
	}

	// Return the response.
	rp.TenantID = tenantID
	return &rp, nil
}

func (c *Client) GCPTenantUpdateV3StorageBucketV2(tenantID string, duplo DuploGCPBucket) (*DuploGCPBucket, ClientError) {
	// Apply the settings via Duplo.
	apiName := fmt.Sprintf("TenantUpdateV3S3Bucket(%s, %s)", tenantID, duplo.Name)
	rp := DuploGCPBucket{}
	err := c.putAPI(apiName, fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, duplo.Name), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	// Deal with a missing response.
	if rp.Name == "" {
		message := fmt.Sprintf("%s: unexpected missing response from backend", apiName)
		log.Printf("[TRACE] %s", message)
		return nil, newClientError(message)
	}

	// Return the response.
	rp.TenantID = tenantID
	return &rp, nil
}

// TenantCreateS3Bucket creates an S3 bucket resource via Duplo.
func (c *Client) TenantCreateS3Bucket(tenantID string, duplo DuploS3BucketRequest) ClientError {
	duplo.Type = ResourceTypeS3Bucket

	// Create the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantCreateS3Bucket(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteS3Bucket deletes an S3 bucket resource via Duplo.
func (c *Client) TenantDeleteS3Bucket(tenantID string, name string) ClientError {

	// Get the full name of the S3 bucket
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	features, _ := c.AdminGetSystemFeatures()
	if features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}
	if err != nil {
		return err
	}

	// Delete the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantDeleteS3Bucket(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&DuploS3BucketRequest{Type: ResourceTypeS3Bucket, Name: fullName, State: "delete"},
		nil)
}

func (c *Client) GCPTenantDeleteStorageBucketV2(tenantID string, name, fullName string) ClientError {

	// Delete the bucket via Duplo.
	return c.deleteAPI(fmt.Sprintf("GCPTenantDeleteStorageBucketV2(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, fullName),
		nil)

}

// TenantGetS3BucketSettings gets a non-cached view of the  S3 buckets's settings via Duplo.
func (c *Client) TenantGetS3BucketSettings(tenantID string, name string) (*DuploS3Bucket, ClientError) {
	rp := DuploS3Bucket{}

	err := c.getAPI(fmt.Sprintf("TenantGetS3BucketSettings(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/GetS3BucketSettings/%s", tenantID, name),
		&rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, err
}

// TenantApplyS3BucketSettings applies settings to an S3 bucket resource via Duplo.
func (c *Client) TenantApplyS3BucketSettings(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, ClientError) {
	apiName := fmt.Sprintf("TenantApplyS3BucketSettings(%s, %s)", tenantID, duplo.Name)

	// Figure out the full resource name.
	features, _ := c.AdminGetSystemFeatures()
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, duplo.Name)
	if features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + duplo.Name
	}
	if err != nil {
		return nil, err
	}
	duplo.Name = fullName

	// Apply the settings via Duplo.
	rp := DuploS3Bucket{}
	err = c.postAPI(apiName, fmt.Sprintf("subscriptions/%s/ApplyS3BucketSettings", tenantID), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	// Deal with a missing response.
	if rp.Name == "" {
		message := fmt.Sprintf("%s: unexpected missing response from backend", apiName)
		log.Printf("[TRACE] %s", message)
		return nil, newClientError(message)
	}

	// Return the response.
	rp.TenantID = tenantID
	return &rp, nil
}

// TenantCreateKafkaCluster creates a kafka cluster resource via Duplo.
func (c *Client) TenantCreateKafkaCluster(tenantID string, duplo DuploKafkaClusterRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("TenantCreateKafkaCluster(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/KafkaClusterUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteKafkaCluster deletes a kafka cluster resource via Duplo.
func (c *Client) TenantDeleteKafkaCluster(tenantID, arn string) ClientError {
	return c.postAPI(
		fmt.Sprintf("TenantDeleteKafkaCluster(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/KafkaClusterUpdate", tenantID),
		&DuploKafkaClusterRequest{Arn: arn, State: "delete"},
		nil)
}

// TenantGetKafkaClusterInfo gets a non-cached view of the kafka cluster's info via Duplo.
func (c *Client) TenantGetKafkaClusterInfo(tenantID string, arn string) (*DuploKafkaClusterInfo, ClientError) {
	rp := DuploKafkaClusterInfo{}

	err := c.postAPI(fmt.Sprintf("TenantGetKafkaClusterInfo(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/FetchKafkaClusterInfo", tenantID),
		map[string]interface{}{"ClusterArn": arn},
		&rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, err
}

// TenantGetKafkaClusterBootstrapBrokers gets a non-cached view of the kafka cluster's info via Duplo.
func (c *Client) TenantGetKafkaClusterBootstrapBrokers(tenantID string, arn string) (*DuploKafkaBootstrapBrokers, ClientError) {
	rp := DuploKafkaBootstrapBrokers{}

	err := c.postAPI(fmt.Sprintf("TenantGetKafkaClusterBootstrapBrokers(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/FetchKafkaBootstrapBrokers", tenantID),
		map[string]interface{}{"ClusterArn": arn},
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// TenantGetLbDetailsInService retrieves application load balancer details via a Duplo service.
func (c *Client) TenantGetAlbDetailsInService(tenantID string, name string) (*DuploAwsLbDetailsInService, ClientError) {
	apiName := fmt.Sprintf("TenantGetAlbDetailsInService(%s, %s)", tenantID, name)
	details := DuploAwsLbDetailsInService{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetALbDetailsInService/%s", tenantID, name), &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

// TenantGetLbDetailsInService retrieves load balancer details via a Duplo service.
func (c *Client) TenantGetLbDetailsInService(tenantID string, name string) (*DuploAwsLbDetailsInService, ClientError) {
	apiName := fmt.Sprintf("TenantGetLbDetailsInService(%s, %s)", tenantID, name)
	details := DuploAwsLbDetailsInService{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetLbDetailsInService/%s", tenantID, name), &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func (c *Client) TenantGetLbDetailsInServiceNew(tenantID string, name string) (map[string]interface{}, ClientError) {
	apiName := fmt.Sprintf("TenantGetLbDetailsInService(%s, %s)", tenantID, name)
	details := make(map[string]interface{})

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("v3/subscriptions/%s/containers/replicationController/%s/lbDetails", tenantID, name), &details)
	if err != nil {
		return nil, err
	}

	return details, nil
}

// TenantCreateApplicationLB creates an application LB resource via Duplo.
func (c *Client) TenantCreateApplicationLB(tenantID string, duplo DuploAwsLBConfiguration) ClientError {
	return c.postAPI("TenantCreateApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteApplicationLB deletes an AWS application LB resource via Duplo.
func (c *Client) TenantDeleteApplicationLB(tenantID string, name string) ClientError {
	// Get the full name of the ALB.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Call the API.
	return c.postAPI("TenantDeleteApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&DuploAwsLBConfiguration{Name: fullName, State: "delete"},
		nil)
}

// TenantListApplicationLbTargetGroups retrieves a list of AWS LB target groups
func (c *Client) TenantListApplicationLbTargetGroups(tenantID string) (*[]DuploAwsLbTargetGroup, ClientError) {
	rp := []DuploAwsLbTargetGroup{}

	err := c.getAPI("TenantListApplicationLbTargetGroups",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbTargetGroups", tenantID),
		&rp)

	return &rp, err
}

// TenantListApplicationLbListeners retrieves a list of AWS LB listeners
func (c *Client) TenantListApplicationLbListeners(tenantID string, name string) (*[]DuploAwsLbListener, ClientError) {
	// Get the full name of the ALB.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	rp := []DuploAwsLbListener{}

	err = c.getAPI("TenantListApplicationLbListeners",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbListerner/%s", tenantID, fullName),
		&rp)

	return &rp, err
}

func (c *Client) TenantUpdateCustomData(tenantID string, customeData CustomDataUpdate) ClientError {
	return c.postAPI("TenantUpdateCustomData",
		fmt.Sprintf("subscriptions/%s/UpdateCustomData", tenantID),
		customeData,
		nil)
}

func (c *Client) TenantApplicationLbListenersByTargetGrpArn(tenantID string, fullName string, targetGrpArn string, port int) (*DuploAwsLbListener, ClientError) {
	rp := []DuploAwsLbListener{}

	err := c.getAPI("TenantListApplicationLbListeners",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbListerner/%s", tenantID, fullName),
		&rp)
	for _, item := range rp {
		for _, action := range item.DefaultActions {
			if action.TargetGroupArn == targetGrpArn && port == item.Port {
				return &item, nil
			}
		}
	}
	return nil, err
}

// TenantCreateApplicationLbListener creates a AWS LB listener
func (c *Client) TenantCreateApplicationLbListener(tenantID string, fullName string, duplo DuploAwsLbListenerCreate) ClientError {
	return c.postAPI("TenantCreateApplicationLB",
		fmt.Sprintf("subscriptions/%s/CreateApplicationLbListerner/%s", tenantID, fullName),
		&duplo,
		nil)
}

// TenantDeleteApplicationLbListener deletes an AWS application LB listener via Duplo.
func (c *Client) TenantDeleteApplicationLbListener(tenantID string, fullName string, listenerArn string) ClientError {
	// Call the API.
	return c.postAPI("TenantDeleteApplicationLB",
		fmt.Sprintf("subscriptions/%s/DeleteApplicationLbListerner/%s", tenantID, fullName),
		&DuploAwsLbListenerDeleteRequest{ListenerArn: listenerArn},
		nil)
}

func (c *Client) TenantCreateAPIGateway(tenantID string, duplo DuploApiGatewayRequest) ClientError {
	return c.postAPI("TenantCreateAPIGateway",
		fmt.Sprintf("subscriptions/%s/ApiGatewayRestApiUpdate", tenantID),
		&duplo,
		nil)
}

func (c *Client) TenantDeleteAPIGateway(tenantID, name string) ClientError {
	return c.postAPI("TenantCreateAPIGateway",
		fmt.Sprintf("subscriptions/%s/ApiGatewayRestApiUpdate", tenantID),
		&DuploApiGatewayRequest{Name: name, State: "delete"},
		nil)
}

func (c *Client) TenantGetAPIGateway(tenantID string, fullName string) (*DuploApiGatewayResource, ClientError) {
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeApiGatewayRestAPI, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploApiGatewayResource{
		Name:         resource.Name,
		MetaData:     resource.MetaData,
		ResourceType: resource.Type,
	}, nil
}

func (c *Client) TenantMinionUpdate(tenantID string, duplo DuploMinion) ClientError {
	return c.postAPI("TenantMinionUpdate",
		fmt.Sprintf("subscriptions/%s/MinionUpdate", tenantID),
		&duplo,
		nil)
}

func (c *Client) TenantMinionDelete(tenantID string, duplo DuploMinionDeleteReq) ClientError {
	return c.postAPI("TenantMinionDelete",
		fmt.Sprintf("subscriptions/%s/MinionUpdate", tenantID),
		&duplo,
		nil)
}

func (c *Client) TenantListMinions(tenantID string) (*[]DuploMinion, ClientError) {
	apiName := fmt.Sprintf("TenantListMinions(%s)", tenantID)
	list := []DuploMinion{}

	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetMinions", tenantID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) TenantGetByoh(tenantID, byohName string) (*DuploMinion, ClientError) {
	list, err := c.TenantListMinions(tenantID)
	if err != nil {
		return nil, err
	}
	for _, minion := range *list {
		if minion.Cloud == 4 && byohName == minion.Name {
			return &minion, nil
		}
	}
	return nil, nil
}

func (c *Client) TenantHostCredentialsUpdate(tenantID string, duplo DuploHostOOBData) ClientError {
	return c.postAPI("TenantHostCredentialsUpdate",
		fmt.Sprintf("subscriptions/%s/UpdateHostCredentialsInOOBData", tenantID),
		&duplo,
		nil)
}

func (c *Client) TenantHostCredentialsGet(tenantID string, duplo DuploHostOOBData) (*DuploHostCredential, ClientError) {
	resp := DuploHostCredential{}
	err := c.postAPI("TenantHostCredentialsGet",
		fmt.Sprintf("subscriptions/%s/FindHostCredentialsFromOOBData", tenantID),
		&duplo,
		&resp)
	if err != nil {
		return nil, err
	}
	return &resp, err
}

func (c *Client) TenantCreateV3S3BucketReplication(tenantID string, duplo DuploS3BucketReplication) ClientError {

	// Create the bucket via Duplo.
	err := c.postAPI(
		fmt.Sprintf("TenantCreateV3S3BucketReplication(%s, %s)", tenantID, duplo.SourceBucket),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s/replication", tenantID, duplo.SourceBucket),
		&duplo,
		nil)

	return err
}

func (c *Client) TenantGetV3S3BucketReplication(tenantID string, name string) (*DuploS3BucketReplicationResponse, ClientError) {
	rp := DuploS3BucketReplicationResponse{}
	err := c.getAPI(fmt.Sprintf("TenantGetV3S3BucketReplication(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s/replication", tenantID, name),
		&rp)

	return &rp, err
}

func (c *Client) TenantUpdateV3S3BucketReplication(tenantID, ruleFullName string, duplo DuploS3BucketReplication) ClientError {
	// Apply the settings via Duplo.
	apiName := fmt.Sprintf("TenantUpdateV3S3BucketReplication(%s, %s,%s)", tenantID, duplo.SourceBucket, ruleFullName)
	err := c.putAPI(apiName, fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s/replication/%s", tenantID, duplo.SourceBucket, ruleFullName), &duplo, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) TenantDeleteV3S3BucketReplication(tenantID, sourceBucket, ruleFullName string) ClientError {
	// Apply the settings via Duplo.
	return c.deleteAPI(fmt.Sprintf("TenantDeleteV3S3BucketReplication(%s, %s,%s)", tenantID, sourceBucket, ruleFullName),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3Bucket/%s/replication/%s", tenantID, sourceBucket, ruleFullName),
		nil)
}
