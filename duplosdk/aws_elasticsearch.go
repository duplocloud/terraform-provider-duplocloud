package duplosdk

import (
	"fmt"
	"log"
)

// DuploElasticSearchDomainVPCOptions represents an AWS ElasticSearch domain's VPC options for a Duplo tenant
type DuploElasticSearchDomainVPCOptions struct {
	VpcID             string   `json:"VPCId,omitempty"`
	AvailabilityZones []string `json:"AvailabilityZones,omitempty"`
	SecurityGroupIDs  []string `json:"SecurityGroupIds,omitempty"`
	SubnetIDs         []string `json:"SubnetIds,omitempty"`
}

// DuploElasticSearchDomainEBSOptions represents an AWS ElasticSearch domain's EBS options for a Duplo tenant
type DuploElasticSearchDomainEBSOptions struct {
	EBSEnabled bool              `json:"EBSEnabled,omitempty"`
	IOPS       int               `json:"Iops,omitempty"`
	VolumeSize int               `json:"VolumeSize,omitempty"`
	VolumeType *DuploStringValue `json:"VolumeType,omitempty"`
}

// DuploElasticSearchDomainEncryptAtRestOptions represents an AWS ElasticSearch domain's endpoint options for a Duplo tenant
type DuploElasticSearchDomainEncryptAtRestOptions struct {
	Enabled  bool   `json:"Enabled,omitempty"`
	KmsKeyID string `json:"KmsKeyId,omitempty"`
}

// DuploElasticSearchDomainEndpointOptions represents an AWS ElasticSearch domain's endpoint options for a Duplo tenant
type DuploElasticSearchDomainEndpointOptions struct {
	EnforceHTTPS      bool             `json:"EnforceHTTPS,omitempty"`
	TLSSecurityPolicy DuploStringValue `json:"TLSSecurityPolicy,omitempty"`
}

// DuploElasticSearchDomainClusterConfig represents an AWS ElasticSearch domain's endpoint options for a Duplo tenant
type DuploElasticSearchDomainClusterConfig struct {
	DedicatedMasterCount   int                                        `json:"DedicatedMasterCount,omitempty"`
	DedicatedMasterEnabled bool                                       `json:"DedicatedMasterEnabled,omitempty"`
	DedicatedMasterType    DuploStringValue                           `json:"DedicatedMasterType,omitempty"`
	InstanceCount          int                                        `json:"InstanceCount,omitempty"`
	InstanceType           DuploStringValue                           `json:"InstanceType,omitempty"`
	WarmCount              int                                        `json:"WarmCount,omitempty"`
	WarmEnabled            bool                                       `json:"WarmEnabled,omitempty"`
	WarmType               DuploStringValue                           `json:"WarmType,omitempty"`
	ColdStorageOptions     DuploElasticSearchDomainColdStorageOptions `json:"ColdStorageOptions,omitempty"`
}

type DuploElasticSearchDomainColdStorageOptions struct {
	Enabled bool `json:"Enabled,omitempty"`
}

// DuploElasticSearchDomainSnapshotOptions represents an AWS ElasticSearch domain's endpoint options for a Duplo tenant
type DuploElasticSearchDomainSnapshotOptions struct {
	AutomatedSnapshotStartHour int `json:"AutomatedSnapshotStartHour,omitempty"`
}

// DuploElasticSearchDomain represents an AWS ElasticSearch domain for a Duplo tenant
type DuploElasticSearchDomain struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name       string `json:"Name,omitempty"` // synthesized on retrieval
	Arn        string `json:"ARN,omitempty"`
	DomainID   string `json:"DomainId,omitempty"`
	DomainName string `json:"DomainName,omitempty"`

	AccessPolicies              string                                       `json:"AccessPolicies,omitempty"`
	AdvancedOptions             map[string]string                            `json:"AdvancedOptions,omitempty"`
	CognitoOptions              DuploEnabled                                 `json:"CognitoOptions,omitempty"`
	DomainEndpointOptions       DuploElasticSearchDomainEndpointOptions      `json:"DomainEndpointOptions,omitempty"`
	EBSOptions                  DuploElasticSearchDomainEBSOptions           `json:"EBSOptions,omitempty"`
	ClusterConfig               DuploElasticSearchDomainClusterConfig        `json:"ElasticsearchClusterConfig,omitempty"`
	NodeToNodeEncryptionOptions DuploEnabled                                 `json:"NodeToNodeEncryptionOptions,omitempty"`
	ElasticSearchVersion        string                                       `json:"ElasticsearchVersion,omitempty"`
	EncryptionAtRestOptions     DuploElasticSearchDomainEncryptAtRestOptions `json:"EncryptionAtRestOptions,omitempty"`
	Endpoints                   map[string]string                            `json:"Endpoints,omitempty"`
	LogPublishingOptions        map[string]interface{}                       `json:"LogPublishingOptions,omitempty"`
	SnapshotOptions             DuploElasticSearchDomainSnapshotOptions      `json:"SnapshotOptions,omitempty"`
	VPCOptions                  DuploElasticSearchDomainVPCOptions           `json:"VPCOptions,omitempty"`

	Created           bool `json:"Created,omitempty"`
	Deleted           bool `json:"Deleted,omitempty"`
	Processing        bool `json:"Processing,omitempty"`
	UpgradeProcessing bool `json:"UpgradeProcessing,omitempty"`
}

// DuploElasticSearchDomainRequest represents a request to create an AWS ElasticSearch domain for a Duplo tenant
type DuploElasticSearchDomainRequest struct {
	Name                       string                                `json:"Name,omitempty"`
	State                      string                                `json:"State,omitempty"`
	Version                    string                                `json:"Version,omitempty"`
	KmsKeyID                   string                                `json:"KmsKeyId,omitempty"`
	ClusterConfig              DuploElasticSearchDomainClusterConfig `json:"ClusterConfig,omitempty"`
	EBSOptions                 DuploElasticSearchDomainEBSOptions    `json:"EbsOptions,omitempty"`
	VPCOptions                 DuploElasticSearchDomainVPCOptions    `json:"VPCOptions,omitempty"`
	EnableNodeToNodeEncryption bool                                  `json:"EnableNodeToNodeEncryption,omitempty"`
	RequireSSL                 bool                                  `json:"RequireSSL,omitempty"`
	UseLatestTLSCipher         bool                                  `json:"UseLatestTLSCipher,omitempty"`
}

// TenantListElasticSearchDomains retrieves a list of AWS ElasticSearch domains.
func (c *Client) TenantListElasticSearchDomains(tenantID string) (*[]DuploElasticSearchDomain, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}

	apiName := fmt.Sprintf("TenantListElasticSearchDomains(%s)", tenantID)
	list := []DuploElasticSearchDomain{}

	// Get the list from Duplo
	err = c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetElasticSearchDomains", tenantID), &list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID and name to each element and return the list.
	log.Printf("[TRACE] %s: %d items", apiName, len(list))
	for i := range list {
		list[i].TenantID = tenantID
		list[i].Name, _ = UnprefixName(prefix, list[i].DomainName)
	}
	return &list, nil
}

// TenantGetElasticSearchDomain retrieves a single AWS ElasticSearch domains from the list returned by Duplo.
func (c *Client) TenantGetElasticSearchDomain(tenantID string, name string, deleted bool) (*DuploElasticSearchDomain, ClientError) {
	log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 1 ********")

	// Get the list from Duplo
	duploObjects, err := c.TenantListElasticSearchDomains(tenantID)
	if err != nil {
		return nil, err
	}

	// Return the matching object
	for _, duploObject := range *duploObjects {
		if duploObject.Name == name && (deleted || !duploObject.Deleted) {
			log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 2 ********: %s", duploObject.DomainName)
			return &duploObject, nil
		}
	}

	// Nothing was found
	log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 3 ********: MISSING")
	return nil, nil
}

// TenantUpdateElasticSearchDomain creates a single AWS ElasticSearch domain via Duplo.
func (c *Client) TenantUpdateElasticSearchDomain(tenantID string, duplo *DuploElasticSearchDomainRequest) ClientError {
	// Create the ES domain via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantUpdateElasticSearchDomain(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/ElasticSearchDomainUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteElasticSearchDomain deletes a single AWS ElasticSearch domain via Duplo.
func (c *Client) TenantDeleteElasticSearchDomain(tenantID string, domainName string) ClientError {
	// Delete the ES domain via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantDeleteElasticSearchDomain(%s, %s)", tenantID, domainName),
		fmt.Sprintf("subscriptions/%s/ElasticSearchDomainUpdate", tenantID),
		&DuploElasticSearchDomainRequest{Name: domainName, State: "delete"},
		nil)
}
