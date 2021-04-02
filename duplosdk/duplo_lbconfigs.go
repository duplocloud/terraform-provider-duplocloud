package duplosdk

import (
	"fmt"
)

// DuploServiceLBConfigs represents an service's load balancer in the Duplo SDK
type DuploServiceLBConfigs struct {
	ReplicationControllerName string                  `json:"ReplicationControllerName"`
	TenantID                  string                  `json:"TenantId,omitempty"`
	LBConfigs                 *[]DuploLBConfiguration `json:"LBConfigs,omitempty"`
	Arn                       string                  `json:"Arn,omitempty"`
	Status                    string                  `json:"Status,omitempty"`
}

// DuploLBConfiguration represents an load balancer's configuration in the Duplo SDK
type DuploLBConfiguration struct {
	LBType                    int    `LBType:"LBType,omitempty"`
	Protocol                  string `json:"Protocol,omitempty"`
	Port                      string `Port:"Port,omitempty"`
	ExternalPort              int    `ExternalPort:"ExternalPort,omitempty"`
	HealthCheckURL            string `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string `json:"CertificateArn,omitempty"`
	ReplicationControllerName string `json:"ReplicationControllerName"`
	IsNative                  bool   `json:"IsNative"`
	IsInternal                bool   `json:"IsInternal"`
}

// DuploServiceLBConfigsGetList retrieves a list of services via the Duplo API.
func (c *Client) DuploServiceLBConfigsGetList(tenantID string) (*[]DuploServiceLBConfigs, error) {
	rp := []DuploServiceLBConfigs{}
	err := c.getAPI(fmt.Sprintf("DuploServiceLBConfigsGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2", tenantID),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// DuploServiceLBConfigsGet retrieves a service's load balancer configs via the Duplo API.
func (c *Client) DuploServiceLBConfigsGet(tenantID string, name string) (*DuploServiceLBConfigs, error) {
	rp := DuploServiceLBConfigs{}
	err := c.getAPI(fmt.Sprintf("DuploServiceLBConfigsGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name),
		&rp)
	if err != nil || rp.ReplicationControllerName == "" {
		return nil, err
	}
	rp.TenantID = tenantID
	return &rp, err
}

// DuploServiceLBConfigsCreate creates a service's load balancer via the Duplo API.
func (c *Client) DuploServiceLBConfigsCreate(tenantID string, rq *DuploServiceLBConfigs) (*DuploServiceLBConfigs, error) {
	return c.DuploServiceLBConfigsCreateOrUpdate(tenantID, rq, false)
}

// DuploServiceLBConfigsUpdate updates an service's load balancer via the Duplo API.
func (c *Client) DuploServiceLBConfigsUpdate(tenantID string, rq *DuploServiceLBConfigs) (*DuploServiceLBConfigs, error) {
	return c.DuploServiceLBConfigsCreateOrUpdate(tenantID, rq, true)
}

// DuploServiceLBConfigsCreateOrUpdate updates an service's load balancer  via the Duplo API.
func (c *Client) DuploServiceLBConfigsCreateOrUpdate(tenantID string, rq *DuploServiceLBConfigs, updating bool) (*DuploServiceLBConfigs, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploServiceLBConfigs{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("DuploServiceLBConfigsCreateOrUpdate(%s, %s)", tenantID, rq.ReplicationControllerName),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// DuploServiceLBConfigsDelete deletes a service's load balancer via the Duplo API.
func (c *Client) DuploServiceLBConfigsDelete(tenantID, name string) error {
	return c.deleteAPI(
		fmt.Sprintf("DuploServiceLBConfigsDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name),
		nil)
}
