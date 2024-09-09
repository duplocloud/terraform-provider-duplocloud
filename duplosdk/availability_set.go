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
	VirtualMachines           []string               `json:"properties.virtualMachines"`
}

func (c *Client) AzureAvailabilitySetCreate(tenantID string, rq *DuploAvailabilitySet) ClientError {
	err := c.postAPI(
		fmt.Sprintf("AzureAvailabilityZoneCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset", tenantID),
		&rq,
		nil,
	)
	return err
}

func (c *Client) AzureAvailabilitySetList(tenantID string) (*[]DuploAvailabilitySet, ClientError) {
	rp := []DuploAvailabilitySet{}
	err := c.getAPI(
		fmt.Sprintf("AzureAvailabilityZoneCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureAvailabilitySetGet(tenantID, name string) (*DuploAvailabilitySetResponse, ClientError) {
	rp := DuploAvailabilitySetResponse{}
	err := c.getAPI(
		fmt.Sprintf("AzureAvailabilityZoneGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureAvailabilitySetDelete(tenantID, name string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("AzureAvailabilityZoneDelete(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/host/availabilityset/%s", tenantID, name),
		nil,
	)
	return err
}
