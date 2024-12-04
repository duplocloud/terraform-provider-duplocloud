package duplosdk

import "fmt"

type DuploAvailabilitySet struct {
	Name string `json:"Name"`
	Sku  struct {
		Name string `json:"Name"`
	} `json:"Sku"`
	PlatformUpdateDomainCount int `json:"PlatformUpdateDomainCount"`
	PlatformFaultDomainCount  int `json:"PlatformFaultDomainCount"`
}

type DuploAvailabilitySetResponse struct {
	Name string `json:"name"`
	Sku  struct {
		Name string `json:"name"`
	} `json:"sku"`
	PlatformUpdateDomainCount int                    `json:"properties.platformUpdateDomainCount"`
	PlatformFaultDomainCount  int                    `json:"properties.platformFaultDomainCount"`
	Tags                      map[string]interface{} `json:"tags"`
	Type                      string                 `json:"type"`
	Location                  string                 `json:"location"`
	AvailabilitySetId         string                 `json:"id"`
	VirtualMachines           []VMIds                `json:"properties.virtualMachines"`
}

type VMIds struct {
	Id string `json:"id,omitempty"`
}

func (c *Client) AzureAvailabilitySetCreate(tenantID string, rq *DuploAvailabilitySet) ClientError {
	err := c.postAPI(
		fmt.Sprintf("AzureAvailabilitySetCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset", tenantID),
		&rq,
		nil,
	)
	return err
}

func (c *Client) AzureAvailabilitySetList(tenantID string) (*[]DuploAvailabilitySet, ClientError) {
	rp := []DuploAvailabilitySet{}
	err := c.getAPI(
		fmt.Sprintf("AzureAvailabilitySetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureAvailabilitySetGet(tenantID, name string) (*DuploAvailabilitySetResponse, ClientError) {
	rp := DuploAvailabilitySetResponse{}
	err := c.getAPI(
		fmt.Sprintf("AzureAvailabilitySetGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureAvailabilitySetDelete(tenantID, name string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("AzureAvailabilitySetDelete(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset/%s", tenantID, name),
		nil,
	)
	return err
}
