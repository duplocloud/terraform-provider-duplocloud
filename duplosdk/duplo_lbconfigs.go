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
	LBType                    int       `json:"LBType,omitempty"`
	Protocol                  string    `json:"Protocol,omitempty"`
	Port                      string    `json:"Port,omitempty"`
	ExternalPort              int       `json:"ExternalPort,omitempty"`
	HealthCheckURL            string    `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string    `json:"CertificateArn,omitempty"`
	ReplicationControllerName string    `json:"ReplicationControllerName"`
	IsNative                  bool      `json:"IsNative"`
	IsInternal                bool      `json:"IsInternal"`
	ExternalTrafficPolicy     string    `json:"ExternalTrafficPolicy"`
	HostNames                 *[]string `json:"HostNames"`
}

// DuploServiceLBConfigsGetList retrieves a list of services via the Duplo API.
func (c *Client) DuploServiceLBConfigsGetList(tenantID string) (*[]DuploServiceLBConfigs, ClientError) {
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
func (c *Client) DuploServiceLBConfigsGet(tenantID string, name string) (*DuploServiceLBConfigs, ClientError) {
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

func (c *Client) DuploServiceLBConfigsExists(tenantID string, name string) bool {
	rp := DuploServiceLBConfigs{}
	err := c.getAPI(fmt.Sprintf("DuploServiceLBConfigsGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name),
		&rp)
	if err == nil && (rp.LBConfigs == nil || len(*rp.LBConfigs) == 0) {
		return false
	}
	return true
}

// DuploServiceLBConfigsCreate creates a service's load balancer via the Duplo API.
func (c *Client) DuploServiceLBConfigsCreate(tenantID string, rq *DuploServiceLBConfigs) (*DuploServiceLBConfigs, ClientError) {
	return c.DuploServiceLBConfigsCreateOrUpdate(tenantID, rq, false)
}

// DuploServiceLBConfigsUpdate updates an service's load balancer via the Duplo API.
func (c *Client) DuploServiceLBConfigsUpdate(tenantID string, rq *DuploServiceLBConfigs) (*DuploServiceLBConfigs, ClientError) {
	return c.DuploServiceLBConfigsCreateOrUpdate(tenantID, rq, true)
}

// DuploServiceLBConfigsCreateOrUpdate updates an service's load balancer  via the Duplo API.
func (c *Client) DuploServiceLBConfigsCreateOrUpdate(tenantID string, rq *DuploServiceLBConfigs, updating bool) (*DuploServiceLBConfigs, ClientError) {

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
func (c *Client) DuploServiceLBConfigsDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploServiceLBConfigsDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name),
		nil)
}
