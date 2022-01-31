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
