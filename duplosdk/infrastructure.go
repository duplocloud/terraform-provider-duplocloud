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

// DuploInfrastructure represents a Duplo infrastructure
type DuploInfrastructure struct {
	Name               string                 `json:"Name"`
	AccountId          string                 `json:"AccountId"`
	Cloud              int                    `json:"Cloud"`
	Region             string                 `json:"Region"`
	AzCount            int                    `json:"AzCount"`
	EnableK8Cluster    bool                   `json:"EnableK8Cluster"`
	AddressPrefix      string                 `json:"AddressPrefix"`
	SubnetCidr         int                    `json:"SubnetCidr"`
	ProvisioningStatus string                 `json:"ProvisioningStatus"`
	CustomData         *[]DuploKeyStringValue `json:"CustomData,omitempty"`
}

// DuploInfrastructureVnet represents a Duplo infrastructure VNET subnet
type DuploInfrastructureVnetSubnet struct {
	// Only used by write APIs
	State              string `json:"State,omitempty"`
	InfrastructureName string `json:"InfrastructureName,omitempty"`

	// Only used by read APIs
	ID string `json:"Id"`

	// Used by both read and write APIs
	AddressPrefix string                 `json:"AddressPrefix"`
	Name          string                 `json:"NameEx"`
	Zone          string                 `json:"Zone"`
	SubnetType    string                 `json:"SubnetType"`
	Tags          *[]DuploKeyStringValue `json:"Tags"`
}

type DuploInfrastructureVnetSecurityGroups struct {
	SystemId string                           `json:"SystemId,omitempty"`
	ReadOnly bool                             `json:"ReadOnly"`
	SgType   string                           `json:"SgType"`
	Name     string                           `json:"Name"`
	Rules    *[]DuploInfrastructureVnetSGRule `json:"Rules"`
}

type DuploInfrastructureVnetSGRule struct {
	SrcRuleType      int    `json:"SrcRuleType"`
	SrcAddressPrefix string `json:"SrcAddressPrefix"`
	SourcePortRange  string `json:"SourcePortRange"`
	Protocol         string `json:"Protocol"`
	Direction        string `json:"Direction"`
	RuleAction       string `json:"RuleAction"`
	Priority         int    `json:"Priority"`
	DstRuleType      int    `json:"DstRuleType"`
}

// DuploInfrastructureVnet represents a Duplo infrastructure VNET
type DuploInfrastructureVnet struct {
	ID                 string                                   `json:"Id"`
	Name               string                                   `json:"Name"`
	AddressPrefix      string                                   `json:"AddressPrefix"`
	SubnetCidr         int                                      `json:"SubnetCidr"`
	Subnets            *[]DuploInfrastructureVnetSubnet         `json:"Subnets,omitempty"`
	ProvisioningStatus string                                   `json:"ProvisioningStatus"`
	SecurityGroups     *[]DuploInfrastructureVnetSecurityGroups `json:"SecurityGroups"`
}

// DuploInfrastructure represents extended information about a Duplo infrastructure
type DuploInfrastructureConfig struct {
	Name               string                   `json:"Name"`
	AccountId          string                   `json:"AccountId"`
	Cloud              int                      `json:"Cloud"`
	Region             string                   `json:"Region"`
	AzCount            int                      `json:"AzCount"`
	EnableK8Cluster    bool                     `json:"EnableK8Cluster"`
	Vnet               *DuploInfrastructureVnet `json:"Vnet"`
	ProvisioningStatus string                   `json:"ProvisioningStatus"`
}

// InfrastructureGetList retrieves a list of infrastructures via the Duplo API.
func (c *Client) InfrastructureGetList() (*[]DuploInfrastructure, ClientError) {
	list := []DuploInfrastructure{}
	err := c.getAPI("InfrastructureGetList()", "v2/admin/InfrastructureV2", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// InfrastructureGet retrieves an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureGet(name string) (*DuploInfrastructure, ClientError) {
	rp := DuploInfrastructure{}
	err := c.getAPI(fmt.Sprintf("InfrastructureGet(%s)", name), fmt.Sprintf("v2/admin/InfrastructureV2/%s", name), &rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, nil
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
func (c *Client) InfrastructureCreate(rq DuploInfrastructure) (*DuploInfrastructure, ClientError) {
	return c.InfrastructureCreateOrUpdate(rq, false)
}

// InfrastructureUpdate updates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureUpdate(rq DuploInfrastructure) (*DuploInfrastructure, ClientError) {
	return c.InfrastructureCreateOrUpdate(rq, true)
}

// InfrastructureCreateOrUpdate creates or updates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureCreateOrUpdate(rq DuploInfrastructure, updating bool) (*DuploInfrastructure, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploInfrastructure{}
	err := c.doAPIWithRequestBody(verb, fmt.Sprintf("InfrastructureCreateOrUpdate(%s)", rq.Name), "v2/admin/InfrastructureV2", &rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// InfrastructureDelete deletes an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureDelete(name string) ClientError {
	return c.deleteAPI(fmt.Sprintf("InfrastructureDelete(%s)", name), fmt.Sprintf("v2/admin/InfrastructureV2/%s", name), nil)
}

// GetEksCredentials retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetEksCredentials(planID string) (*DuploEksCredentials, ClientError) {
	creds := DuploEksCredentials{}
	err := c.getAPI(fmt.Sprintf("GetEksCredentials(%s)", planID), fmt.Sprintf("adminproxy/%s/GetEksClusterByInfra", planID), &creds)
	if err != nil {
		return nil, err
	}
	creds.PlanID = planID
	return &creds, nil
}
