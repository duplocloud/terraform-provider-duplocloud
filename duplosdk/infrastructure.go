package duplosdk

import (
	"fmt"
	"strings"
)

// DuploEksCredentials represents just-in-time EKS credentials in Duplo
type DuploEksCredentials struct {
	// NOTE: The PlanID field does not come from the backend - we synthesize it
	PlanID string `json:"-"`

	Name                           string `json:"Name"`
	APIServer                      string `json:"ApiServer"`
	Token                          string `json:"Token"`
	AwsRegion                      string `json:"AwsRegion"`
	K8sProvider                    int    `json:"K8Provider,omitempty"`
	K8sVersion                     string `json:"K8sVersion,omitempty"`
	CertificateAuthorityDataBase64 string `json:"CertificateAuthorityDataBase64,omitempty"`
	DefaultNamespace               string `json:"DefaultNamespace,omitempty"`
}

// DuploInfrastructureCreateRequest represents a Duplo infrastructure creation request
type DuploInfrastructureCreateRequest struct {
	Name                    string                   `json:"Name"`
	AccountId               string                   `json:"AccountId"`
	Cloud                   int                      `json:"Cloud"`
	Region                  string                   `json:"Region"`
	AzCount                 int                      `json:"AzCount"`
	EnableK8Cluster         bool                     `json:"EnableK8Cluster"`
	EnableECSCluster        bool                     `json:"EnableECSCluster"`
	EnableContainerInsights bool                     `json:"EnableContainerInsights"`
	Vnet                    *DuploInfrastructureVnet `json:"Vnet"`
	CustomData              *[]DuploKeyStringValue   `json:"CustomData,omitempty"`
	OnPrem                  *DuploOnPrem             `json:"OnPremConfig,omitempty"`
}

// DuploInfrastructure represents a Duplo infrastructure
type DuploInfrastructure struct {
	Name                    string                 `json:"Name"`
	AccountId               string                 `json:"AccountId"`
	Cloud                   int                    `json:"Cloud"`
	Region                  string                 `json:"Region"`
	AzCount                 int                    `json:"AzCount"`
	EnableK8Cluster         bool                   `json:"EnableK8Cluster"`
	EnableECSCluster        bool                   `json:"EnableECSCluster"`
	EnableContainerInsights bool                   `json:"EnableContainerInsights"`
	AddressPrefix           string                 `json:"AddressPrefix"`
	SubnetCidr              int                    `json:"SubnetCidr"`
	ProvisioningStatus      string                 `json:"ProvisioningStatus"`
	CustomData              *[]DuploKeyStringValue `json:"CustomData,omitempty"`
}

type DuploOnPremEKSConfig struct {
	PrivateSubnets          []string `json:"PrivateSubnets"`
	PublicSubnets           []string `json:"PublicSubnets"`
	VpcId                   string   `json:"VpcId"`
	IngressSecurityGroupIds []string `json:"IngressSecurityGroupIds"`
}
type DuploOnPremK8Config struct {
	Name                           string                `json:"Name"`
	Vendor                         int                   `json:"Vendor"`
	ClusterEndpoint                string                `json:"ClusterEndpoint"`
	ApiToken                       string                `json:"ApiToken"`
	CertificateAuthorityDataBase64 string                `json:"CertificateAuthorityDataBase64"`
	OnPremEKSConfig                *DuploOnPremEKSConfig `json:"OnPremEKSConfig,omitempty"`
}
type DuploOnPrem struct {
	OnPremK8Config *DuploOnPremK8Config `json:"OnPremK8Config,omitempty"`
	DataCenter     string               `json:"DataCenter"`
}

// DuploInfrastructureVnetSubnet represents a Duplo infrastructure VNET subnet
type DuploInfrastructureVnetSubnet struct {
	// Only used by write APIs
	State              string `json:"State,omitempty"`
	InfrastructureName string `json:"InfrastructureName,omitempty"`

	// Only used by read APIs
	ID string `json:"Id"`

	// Used by both read and write APIs
	AddressPrefix    string                 `json:"AddressPrefix"`
	Name             string                 `json:"NameEx"`
	Zone             string                 `json:"Zone"`
	SubnetType       string                 `json:"SubnetType"`
	ServiceEndpoints []string               `json:"ServiceEndpoints,omitempty"`
	IsolatedNetwork  bool                   `json:"IsolatedNetwork,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags"`
}

type DuploInfrastructureVnetSecurityGroups struct {
	SystemId string                           `json:"SystemId,omitempty"`
	ReadOnly bool                             `json:"ReadOnly"`
	SgType   string                           `json:"SgType"`
	Name     string                           `json:"Name"`
	Rules    *[]DuploInfrastructureVnetSGRule `json:"Rules"`
}

type DuploInfrastructureVnetSGRule struct {
	Name                 string `json:"Name"`
	SrcRuleType          int    `json:"SrcRuleType"`
	SrcAddressPrefix     string `json:"SrcAddressPrefix"`
	SourcePortRange      string `json:"SourcePortRange"`
	Protocol             string `json:"Protocol"`
	Direction            string `json:"Direction"`
	RuleAction           string `json:"RuleAction"`
	Priority             int    `json:"Priority"`
	DstRuleType          int    `json:"DstRuleType"`
	DestinationPortRange string `json:"DestinationPortRange,omitempty"`
	DstAddressPrefix     string `json:"DstAddressPrefix,omitempty"`
}

// DuploInfrastructureVnet represents a Duplo infrastructure VNET
type DuploInfrastructureVnet struct {
	ID                 string                                   `json:"Id,omitempty"`
	Name               string                                   `json:"Name,omitempty"`
	AddressPrefix      string                                   `json:"AddressPrefix"`
	SubnetCidr         int                                      `json:"SubnetCidr"`
	Subnets            *[]DuploInfrastructureVnetSubnet         `json:"Subnets,omitempty"`
	ProvisioningStatus string                                   `json:"ProvisioningStatus,omitempty"`
	SecurityGroups     *[]DuploInfrastructureVnetSecurityGroups `json:"SecurityGroups,omitempty"`
}

// DuploInfrastructureConfig represents extended information about a Duplo infrastructure
type DuploInfrastructureConfig struct {
	Name                    string                   `json:"Name"`
	AccountId               string                   `json:"AccountId"`
	Cloud                   int                      `json:"Cloud"`
	Region                  string                   `json:"Region"`
	AzCount                 int                      `json:"AzCount"`
	EnableK8Cluster         bool                     `json:"EnableK8Cluster"`
	EnableECSCluster        bool                     `json:"EnableECSCluster"`
	EnableContainerInsights bool                     `json:"EnableContainerInsights"`
	IsServerlessKubernetes  bool                     `json:"IsServerlessKubernetes"`
	Vnet                    *DuploInfrastructureVnet `json:"Vnet"`
	ProvisioningStatus      string                   `json:"ProvisioningStatus"`
	CustomData              *[]DuploKeyStringValue   `json:"CustomData,omitempty"`
	AksConfig               *AksConfig               `json:"AksConfig,omitempty"`
	ClusterIpv4Cidr         string                   `json:"ClusterIpv4Cidr,omitempty"`
	DisablePublicSubnet     *bool                    `json:"DisablePublicSubnet,omitempty"`
	OnPrem                  *DuploOnPrem             `json:"OnPremConfig,omitempty"`
}

type AksConfig struct {
	Name              string `json:"Name"`
	CreateAndManage   bool   `json:"CreateAndManage"`
	PrivateCluster    bool   `json:"PrivateCluster"`
	K8sVersion        string `json:"K8sVersion,omitempty"`
	VmSize            string `json:"VmSize,omitempty"`
	NetworkPlugin     string `json:"NetworkPlugin,omitempty"`
	OutboundType      string `json:"OutboundType,omitempty"`
	NodeResourceGroup string `json:"NodeResourceGroup"`
}

type DuploAzureLogAnalyticsWorkspace struct {
	PropertiesProvisioningState string `json:"properties.provisioningState"`
	PropertiesCustomerID        string `json:"properties.customerId"`
	PropertiesSku               struct {
		Name string `json:"name"`
	} `json:"properties.sku"`
	PropertiesRetentionInDays                 int    `json:"properties.retentionInDays"`
	PropertiesPublicNetworkAccessForIngestion string `json:"properties.publicNetworkAccessForIngestion"`
	PropertiesPublicNetworkAccessForQuery     string `json:"properties.publicNetworkAccessForQuery"`
	Location                                  string `json:"location"`
	ID                                        string `json:"id"`
	Name                                      string `json:"name"`
	Type                                      string `json:"type"`
}

type DuploAzureLogAnalyticsWorkspaceRq struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type DuploAzureRecoveryServicesVault struct {
	Properties struct {
		ProvisioningState                   string `json:"provisioningState"`
		PrivateEndpointStateForBackup       string `json:"privateEndpointStateForBackup"`
		PrivateEndpointStateForSiteRecovery string `json:"privateEndpointStateForSiteRecovery"`
	} `json:"properties"`
	Sku struct {
		Name string `json:"name"`
	} `json:"sku"`
	Location string `json:"location"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	ETag     string `json:"eTag"`
}

type DuploAzureRecoveryServicesVaultRq struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type InfrastructureSgUpdate struct {
	Name          string                           `json:"Name"`
	SgName        string                           `json:"SgName"`
	RulesToAdd    *[]DuploInfrastructureVnetSGRule `json:"RulesToAdd,omitempty"`
	RulesToRemove []string                         `json:"RulesToRemove,omitempty"`
	State         string                           `json:"State,omitempty"`
}

// DuploInfrastructureSetting represents a Duplo infrastruture's settings
type DuploInfrastructureSetting struct {
	InfraName string                 `json:"InfraName,omitempty"`
	Setting   *[]DuploKeyStringValue `json:"Setting,omitempty"`
}

// DuploInfrastructureSettingUpdateRequest represents a request to update a Duplo tenant's configuration
type DuploInfrastructureSettingUpdateRequest struct {
	Key   string `json:"Key,omitempty"`
	State string `json:"State,omitempty"`
	Value string `json:"Value,omitempty"`
}

// DuploInfrastructureECSConfigUpdate represents a request to update a Duplo infrastructure's ECS cluster
type DuploInfrastructureECSConfigUpdate struct {
	EnableECSCluster        bool `json:"EnableECS"`
	EnableContainerInsights bool `json:"EnableContainerInsights"`
}

// InfrastructureList retrieves a list of infrastructures via the Duplo API.
func (c *Client) InfrastructureList() (*[]DuploInfrastructureConfig, ClientError) {
	var list []DuploInfrastructureConfig
	err := c.getAPI("InfrastructureList()", "adminproxy/GetInfrastructureConfigs", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// InfrastructureGetList retrieves a list of infrastructures via the Duplo API.
func (c *Client) InfrastructureGetList() (*[]DuploInfrastructure, ClientError) {
	var list []DuploInfrastructure
	err := c.getAPI("InfrastructureGetList()", "v2/admin/InfrastructureV2", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// InfrastructureGet retrieves an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureGet(name string) (*DuploInfrastructureConfig, ClientError) {
	allInfras, err := c.InfrastructureList()
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific type and name.
	for _, infra := range *allInfras {
		if infra.Name == name {
			return &infra, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// InfrastructureGetConfig retrieves extended infrastructure configuration by name via the Duplo API.
func (c *Client) InfrastructureGetConfig(name string) (*DuploInfrastructureConfig, ClientError) {
	rp := DuploInfrastructureConfig{}
	err := c.getAPI(fmt.Sprintf("InfrastructureGetConfig(%s)", name), fmt.Sprintf("adminproxy/GetInfrastructureConfig/%s", name), &rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, nil
}

// InfrastructureGetSetting retrieves tenant configuration metadata via the Duplo API.
func (c *Client) InfrastructureGetSetting(infraName string) (*DuploInfrastructureSetting, ClientError) {
	config, err := c.InfrastructureGetConfig(infraName)
	if config == nil || err != nil {
		return nil, err
	}
	return &DuploInfrastructureSetting{InfraName: config.Name, Setting: config.CustomData}, nil
}

// InfrastructureReplaceSetting replaces tenant configuration metadata via the Duplo API.
func (c *Client) InfrastructureReplaceSetting(setting DuploInfrastructureSetting) ClientError {
	existing, err := c.InfrastructureGetSetting(setting.InfraName)
	if err != nil {
		return err
	}
	return c.InfrastructureChangeSetting(setting.InfraName, existing.Setting, setting.Setting)
}

// InfrastructureChangeSetting changes tenant configuration metadata via the Duplo API, using the supplied
// oldConfig and newConfig, for the given tenantID.
func (c *Client) InfrastructureChangeSetting(infraName string, oldSetting, newSetting *[]DuploKeyStringValue) ClientError {

	// Next, update all keys that are present, keeping a record of each one that is present
	present := map[string]struct{}{}
	if newSetting != nil {
		for _, kv := range *newSetting {
			if err := c.InfrastructureSetSettingKey(infraName, kv.Key, kv.Value); err != nil {
				return err
			}
			present[kv.Key] = struct{}{}
		}
	}

	// Finally, delete any keys that are no longer present.
	if oldSetting != nil {
		for _, kv := range *oldSetting {
			if _, ok := present[kv.Key]; !ok {
				if err := c.InfrastructureDeleteSettingKey(infraName, kv.Key); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// InfrastructureDeleteSettingKey deletes a specific configuration key for a tenant via the Duplo API.
func (c *Client) InfrastructureDeleteSettingKey(infraName, key string) ClientError {
	rq := DuploInfrastructureSettingUpdateRequest{State: "delete", Key: key}
	return c.postAPI(
		fmt.Sprintf("InfrastructureDeleteSettingKey(%s, %s)", infraName, key),
		fmt.Sprintf("adminproxy/UpdateInfrastructureCustomData/%s", infraName),
		&rq,
		nil)
}

// InfrastructureSetSettingKey set a specific configuration key for a tenant via the Duplo API.
func (c *Client) InfrastructureSetSettingKey(infraName, key, value string) ClientError {
	rq := DuploInfrastructureSettingUpdateRequest{Key: key, Value: value}
	return c.postAPI(
		fmt.Sprintf("InfrastructureSetSettingKey(%s, %s)", infraName, key),
		fmt.Sprintf("adminproxy/UpdateInfrastructureCustomData/%s", infraName),
		&rq,
		nil)
}

// InfrastructureGetSubnet retrieves a specific infrastructure subnet via the Duplo API.
func (c *Client) InfrastructureGetSubnet(infraName string, subnetName string, subnetCidr string) (*DuploInfrastructureVnetSubnet, ClientError) {

	// Get the entire infra config, since there is no limited API to call.
	config, err := c.InfrastructureGetConfig(infraName)
	if config == nil || err != nil {
		return nil, err
	}

	// Return the subnet, if it exists.
	for _, subnet := range *config.Vnet.Subnets {
		if subnet.Name == subnetName && subnet.AddressPrefix == subnetCidr {

			// Interpret the subnet type (FIXME - the backend needs to return this)
			if subnet.SubnetType == "" {
				if strings.Contains(strings.ToLower(subnet.Name), "public") {
					subnet.SubnetType = "public"
				} else {
					subnet.SubnetType = "private"
				}
			}

			return &subnet, nil
		}
	}

	// Nothing was found.
	return nil, nil
}

// InfrastructureCreateOrUpdateSubnet creates or updates an infrastructure subnet via the Duplo API.
func (c *Client) InfrastructureCreateOrUpdateSubnet(rq DuploInfrastructureVnetSubnet) ClientError {
	return c.postAPI(
		fmt.Sprintf("InfrastructureCreateOrUpdateSubnet(%s, %s)", rq.InfrastructureName, rq.Name),
		"adminproxy/UpdateInfrastructureSubnet",
		&rq,
		nil)
}

// InfrastructureDeleteSubnet deletes an infrastructure subnet via the Duplo API.
func (c *Client) InfrastructureDeleteSubnet(infraName, subnetName, subnetCidr string) ClientError {
	rq := DuploInfrastructureVnetSubnet{
		State:              "delete",
		InfrastructureName: infraName,
		Name:               subnetName,
		AddressPrefix:      subnetCidr,
	}
	return c.postAPI(
		fmt.Sprintf("InfrastructureDeletSubnet(%s, %s)", infraName, subnetName),
		"adminproxy/UpdateInfrastructureSubnet",
		&rq,
		nil)
}

// InfrastructureCreate creates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureCreate(rq DuploInfrastructureConfig) ClientError {
	return c.postAPI(
		fmt.Sprintf("InfrastructureCreate(%s)", rq.Name),
		"adminproxy/CreateInfrastructureConfig",
		&rq,
		nil)
}

// InfrastructureUpdateECSConfig creates or updates an infrastructure's ECS cluster via the Duplo API.
func (c *Client) InfrastructureUpdateECSConfig(infraName string, rq DuploInfrastructureECSConfigUpdate) ClientError {
	return c.postAPI(
		fmt.Sprintf("InfrastructureUpdateECSConfig(%s)", infraName),
		fmt.Sprintf("adminproxy/UpdateInfrastructureECS/%s", infraName),
		&rq,
		nil)
}

// InfrastructureDelete deletes an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureDelete(infraName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("InfrastructureDelete(%s)", infraName),
		fmt.Sprintf("adminproxy/DeleteInfrastructureConfig/%s", infraName),
		nil,
		nil)
}

// GetEksCredentials retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetK8sCredentials(planID string) (*DuploEksCredentials, ClientError) {
	creds := DuploEksCredentials{}
	err := c.getAPI(fmt.Sprintf("GetK8sCredentials(%s)", planID), fmt.Sprintf("adminproxy/%s/GetEksClusterByInfra", planID), &creds)
	if err != nil {
		return nil, err
	}
	creds.PlanID = planID
	return &creds, nil
}

func (c *Client) AzureLogAnalyticsWorkspaceCreate(infraName string, rq DuploAzureLogAnalyticsWorkspaceRq) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureLogAnalyticsWorkspaceCreate(%s,%s)", infraName, rq.Name),
		fmt.Sprintf("adminproxy/SetInfrastructureLogAnalyticsConfig/%s", infraName),
		&rq,
		nil)
}

func (c *Client) AzureLogAnalyticsWorkspaceGet(infraName string) (*DuploAzureLogAnalyticsWorkspace, ClientError) {
	rp := DuploAzureLogAnalyticsWorkspace{}
	err := c.getAPI(
		fmt.Sprintf("AzureLogAnalyticsWorkspaceGet(%s)", infraName),
		fmt.Sprintf("adminproxy/GetInfrastructureLogAnalyticsWorkspace/%s", infraName),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (c *Client) AzureRecoveryServicesVaultCreate(infraName string, rq DuploAzureRecoveryServicesVaultRq) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureRecoveryServicesVaultCreate(%s,%s)", infraName, rq.Name),
		fmt.Sprintf("adminproxy/SetInfrastructureRecoveryServicesVaultConfig/%s", infraName),
		&rq,
		nil)
}

func (c *Client) AzureRecoveryServicesVaultGet(infraName string) (*DuploAzureRecoveryServicesVault, ClientError) {
	rp := DuploAzureRecoveryServicesVault{}
	err := c.getAPI(
		fmt.Sprintf("AzureRecoveryServicesVaultGet(%s)", infraName),
		fmt.Sprintf("adminproxy/GetInfrastructureRecoveryServicesVault/%s", infraName),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (c *Client) NetworkSgRuleCreateOrDelete(rq *InfrastructureSgUpdate) ClientError {
	return c.postAPI(
		fmt.Sprintf("NetworkSgRuleCreate(%s)", rq.Name),
		"adminproxy/UpdateInfrastructureSg",
		&rq,
		nil,
	)
}

func (c *Client) NetworkSgRuleGet(infraName, sgName, ruleName string) (*DuploInfrastructureVnetSGRule, ClientError) {
	config, err := c.InfrastructureGet(infraName)
	if err != nil {
		return nil, err
	}
	if config != nil && config.Vnet != nil {
		sgList := config.Vnet.SecurityGroups
		if sgList != nil {
			for _, sg := range *sgList {
				if sg.Name == sgName {
					rules := sg.Rules
					for _, rule := range *rules {
						if rule.Name == ruleName {
							return &rule, nil
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureK8sClusterCreate(infraName string, rq *AksConfig) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureK8sClusterCreate(%s,%s)", infraName, rq.Name),
		fmt.Sprintf("adminproxy/UpdateInfrastructureAksConfig/%s", infraName),
		&rq,
		nil)
}
