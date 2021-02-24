package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
	EBSEnabled bool             `json:"EBSEnabled,omitempty"`
	IOPS       int              `json:"Iops,omitempty"`
	VolumeSize int              `json:"VolumeSize,omitempty"`
	VolumeType DuploStringValue `json:"VolumeType,omitempty"`
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
	DedicatedMasterCount   int              `json:"DedicatedMasterCount,omitempty"`
	DedicatedMasterEnabled bool             `json:"DedicatedMasterEnabled,omitempty"`
	DedicatedMasterType    DuploStringValue `json:"DedicatedMasterType,omitempty"`
	InstanceCount          int              `json:"InstanceCount,omitempty"`
	InstanceType           DuploStringValue `json:"InstanceType,omitempty"`
}

// DuploElasticSearchDomainSnapshotOptions represents an AWS ElasticSearch domain's endpoint options for a Duplo tenant
type DuploElasticSearchDomainSnapshotOptions struct {
	AutomatedSnapshotStartHour int `json:"AutomatedSnapshotStartHour,omitempty"`
}

// DuploElasticSearchDomain represents an AWS ElasticSearch domain for a Duplo tenant
type DuploElasticSearchDomain struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

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
	KmsKeyID                    string                                       `json:"KmsKeyId,omitempty"`
	Endpoints                   map[string]interface{}                       `json:"Endpoints,omitempty"`
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
	Name          string                                `json:"Name,omitempty"`
	Version       string                                `json:"Version,omitempty"`
	KmsKeyID      string                                `json:"KmsKeyId,omitempty"`
	ClusterConfig DuploElasticSearchDomainClusterConfig `json:"ClusterConfig,omitempty"`
	EBSOptions    DuploElasticSearchDomainEBSOptions    `json:"EbsOptions,omitempty"`
	VPCOptions    DuploElasticSearchDomainVPCOptions    `json:"VPCOptions,omitempty"`
}

// TenantListElasticSearchDomains retrieves a list of AWS ElasticSearch domains.
func (c *Client) TenantListElasticSearchDomains(tenantID string) (*[]DuploElasticSearchDomain, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetElasticSearchDomains", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantListElasticSearchDomains 1 ********: %s ", url)

	// Get the list from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantListElasticSearchDomains 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantListElasticSearchDomains 3 ********: %s", bodyString)

	// Return it as an object.
	duploObjects := make([]DuploElasticSearchDomain, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantListElasticSearchDomains 4 ********")
	for i := range duploObjects {
		parts := strings.SplitN(duploObjects[i].DomainName, "-", 3)
		name := duploObjects[i].DomainName
		if len(parts) == 3 {
			name = parts[2]
		}
		duploObjects[i].TenantID = tenantID
		duploObjects[i].Name = name
	}
	return &duploObjects, nil
}

// TenantGetElasticSearchDomain retrieves a single AWS ElasticSearch domains from the list returned by Duplo.
func (c *Client) TenantGetElasticSearchDomain(tenantID string, name string) (*DuploElasticSearchDomain, error) {
	log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 1 ********")

	// Get the list from Duplo
	duploObjects, err := c.TenantListElasticSearchDomains(tenantID)
	if err != nil {
		return nil, err
	}

	// Return the matching object
	for _, duploObject := range *duploObjects {
		if duploObject.Name == name {
			log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 2 ********: %s", duploObject.DomainName)
			return &duploObject, nil
		}
	}

	// Nothing was found
	log.Printf("[TRACE] duplo-TenantGetElasticSearchDomain 3 ********: MISSING")
	return nil, nil
}

// TenantUpdateElasticSearchDomain creates or updates a single AWS ElasticSearch domain via Duplo.
func (c *Client) TenantUpdateElasticSearchDomain(tenantID string, duplo *DuploElasticSearchDomainRequest) error {
	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantUpdateElasticSearchDomain 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/ElasticSearchDomainUpdate", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantUpdateElasticSearchDomain 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantUpdateElasticSearchDomain 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantUpdateElasticSearchDomain 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to create ElasticSearch domain %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to create ElasticSearch domain %s: '%s'", tenantID, duplo.Name, bodyString)
}
