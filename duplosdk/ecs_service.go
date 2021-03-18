package duplosdk

import (
	"fmt"
	"strings"
)

// DuploEcsServiceLbConfig is a Duplo SDK object that represents load balancer configuration for an ECS service
type DuploEcsServiceLbConfig struct {
	ReplicationControllerName string `json:"ReplicationControllerName"`
	Port                      string `json:"Port,omitempty"`
	BackendProtocol           string `json:"BeProtocolVersion,omitempty"`
	ExternalPort              int    `json:"ExternalPort,omitempty"`
	Protocol                  string `json:"Protocol,omitempty"`
	IsInternal                bool   `json:"IsInternal,omitempty"`
	HealthCheckURL            string `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string `json:"CertificateArn,omitempty"`
	LbType                    int    `json:"LbType,omitempty"`
}

// DuploEcsService is a Duplo SDK object that represents an ECS service
type DuploEcsService struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name                          string                     `json:"Name"`
	TaskDefinition                string                     `json:"TaskDefinition,omitempty"`
	Replicas                      int                        `json:"Replicas,omitempty"`
	HealthCheckGracePeriodSeconds int                        `json:"HealthCheckGracePeriodSeconds,omitempty"`
	OldTaskDefinitionBufferSize   int                        `json:"OldTaskDefinitionBufferSize,omitempty"`
	IsTargetGroupOnly             bool                       `json:"IsTargetGroupOnly,omitempty"`
	DNSPrfx                       string                     `json:"DnsPrfx,omitempty"`
	LBConfigurations              *[]DuploEcsServiceLbConfig `json:"LBConfigurations,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcsServiceCreate creates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, error) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, false)
}

// EcsServiceUpdate updates an ECS service via the Duplo API.
func (c *Client) EcsServiceUpdate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, error) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, true)
}

// EcsServiceCreateOrUpdate creates or updates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreateOrUpdate(tenantID string, rq *DuploEcsService, updating bool) (*DuploEcsService, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploEcsService{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("EcsServiceCreateOrUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	rp.TenantID = tenantID
	return &rp, err
}

// EcsServiceDelete deletes an ECS service via the Duplo API.
func (c *Client) EcsServiceDelete(id string) error {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Delete the ECS service
	return c.deleteAPI(
		fmt.Sprintf("EcsServiceDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantID, name),
		nil)
}

// EcsServiceGet retrieves an ECS service via the Duplo API.
func (c *Client) EcsServiceGet(id string) (*DuploEcsService, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Retrieve the object.
	duploObject := DuploEcsService{}
	err := c.getAPI(
		fmt.Sprintf("EcsServiceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantID, name),
		&duploObject)
	if err != nil || duploObject.Name == "" {
		return nil, err
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}
