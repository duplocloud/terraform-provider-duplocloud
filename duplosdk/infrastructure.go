package duplosdk

import (
	"fmt"
)

// DuploEksCredentials represents just-in-time EKS credentials in Duplo
type DuploEksCredentials struct {
	// NOTE: The PlanID field does not come from the backend - we synthesize it
	PlanID string `json:"-,omitempty"`

	Name        string `json:"Name"`
	APIServer   string `json:"ApiServer"`
	Token       string `json:"Token"`
	AwsRegion   string `json:"AwsRegion"`
	K8sProvider int    `json:"K8Provider,omitempty"`
}

// DuploInfrastructure represents a Duplo infrastructure
type DuploInfrastructure struct {
	Name               string `json:"Name"`
	AccountId          string `json:"AccountId"`
	Cloud              int    `json:"Cloud"`
	Region             string `json:"Region"`
	AzCount            int    `json:"AzCount"`
	EnableK8Cluster    bool   `json:"EnableK8Cluster"`
	AddressPrefix      string `json:"AddressPrefix"`
	SubnetCidr         int    `json:"SubnetCidr"`
	ProvisioningStatus string `json:"ProvisioningStatus"`
}

// InfrastructureGetList retrieves a list of infrastructures via the Duplo API.
func (c *Client) InfrastructureGetList() (*[]DuploInfrastructure, error) {
	list := []DuploInfrastructure{}
	err := c.getAPI("InfrastructureGetList()", "v2/admin/InfrastructureV2", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// InfrastructureGet retrieves an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureGet(name string) (*DuploInfrastructure, error) {
	rp := DuploInfrastructure{}
	err := c.getAPI(fmt.Sprintf("InfrastructureGet(%s)", name), fmt.Sprintf("v2/admin/InfrastructureV2/%s", name), &rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, nil
}

// InfrastructureCreate creates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureCreate(rq DuploInfrastructure) (*DuploInfrastructure, error) {
	return c.InfrastructureCreateOrUpdate(rq, false)
}

// InfrastructureUpdate updates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureUpdate(rq DuploInfrastructure) (*DuploInfrastructure, error) {
	return c.InfrastructureCreateOrUpdate(rq, true)
}

// InfrastructureCreateOrUpdate creates or updates an infrastructure by name via the Duplo API.
func (c *Client) InfrastructureCreateOrUpdate(rq DuploInfrastructure, updating bool) (*DuploInfrastructure, error) {

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
func (c *Client) InfrastructureDelete(name string) error {
	return c.deleteAPI(fmt.Sprintf("InfrastructureDelete(%s)", name), fmt.Sprintf("v2/admin/InfrastructureV2/%s", name), nil)
}

// GetEksCredentials retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetEksCredentials(planID string) (*DuploEksCredentials, error) {
	creds := DuploEksCredentials{}
	err := c.getAPI(fmt.Sprintf("GetEksCredentials(%s)", planID), fmt.Sprintf("adminproxy/%s/GetEksClusterByInfra", planID), &creds)
	if err != nil {
		return nil, err
	}
	creds.PlanID = planID
	return &creds, nil
}
