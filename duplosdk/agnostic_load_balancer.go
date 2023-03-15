package duplosdk

import (
	"fmt"
)

// AgnosticLbSettings represents a load balancer's settings.
type AgnosticLbSettings struct {
	Cloud                     int     `json:"Cloud"`
	NativeId                  string  `json:"NativeId,omitempty"`
	LoadBalancerType          int     `json:"LoadBalancerType"`
	LoadBalancerId            string  `json:"LoadBalancerId"`
	EnableAccessLogs          *bool   `json:"EnableAccessLogs,omitempty"`
	EnableHttpToHttpsRedirect *bool   `json:"EnableHttpToHttpsRedirect,omitempty"`
	SecurityPolicyId          *string `json:"SecurityPolicyId,omitempty"`
	IsHttpPortUsed            bool    `json:"IsHttpPortUsed,omitempty"`
	Timeout                   int     `json:"Timeout"`

	Aws *AgnosticLbSettingsAws `json:"Aws,omitempty"`
	Gcp *AgnosticLbSettingsGcp `json:"Gcp,omitempty"`
}

// AgnosticLbSettingsAws represents a load balancer's AWS specific settings.
type AgnosticLbSettingsAws struct {
	DropInvalidHeaders             bool   `json:"DropInvalidHeaders,omitempty"`
	SessionAffinity                *bool  `json:"SessionAffinity,omitempty"`
	HttpRedirectListnerArn         string `json:"HttpRedirectListnerArn,omitempty"`
	ElbV1ConnectionDrainingTimeout int    `json:"ElbV1ConnectionDrainingTimeout,omitempty"`
}

// AgnosticLbSettingsGcp represents a load balancer's GCP specific settings.
type AgnosticLbSettingsGcp struct {
	SessionAffinity           *bool `json:"SessionAffinity,omitempty"`
	ConnectionDrainingTimeout int   `json:"ConnectionDrainingTimeout,omitempty"`
}

// TenantUpdateLbSettings updates a load balancer's settings via Duplo.
func (c *Client) TenantUpdateLbSettings(tenantID, loadBalancerID string, rq *AgnosticLbSettings) (*AgnosticLbSettings, ClientError) {
	rp := AgnosticLbSettings{}
	err := c.postAPI("TenantUpdateLbSettings",
		fmt.Sprintf("v3/subscriptions/%s/agnostic/loadBalancer/%s/setting", tenantID, loadBalancerID),
		&rq,
		&rp)
	return &rp, err
}

// TenantGetLbSettings retrieves a load balancer's settings via Duplo.
func (c *Client) TenantGetLbSettings(tenantID, loadBalancerID string) (*AgnosticLbSettings, ClientError) {
	rp := AgnosticLbSettings{}
	err := c.getAPI("TenantGetLbSettings",
		fmt.Sprintf("v3/subscriptions/%s/agnostic/loadBalancer/%s/setting", tenantID, loadBalancerID),
		&rp)
	return &rp, err
}
