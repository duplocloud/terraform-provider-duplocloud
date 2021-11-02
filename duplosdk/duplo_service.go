package duplosdk

import (
	"fmt"
)

// DuploService represents a service in the Duplo SDK
type DuploService struct {
	Name                    string                 `json:"Name"`
	TenantID                string                 `json:"TenantId,omitempty"`
	OtherDockerHostConfig   string                 `json:"OtherDockerHostConfig,omitempty"`
	OtherDockerConfig       string                 `json:"OtherDockerConfig,omitempty"`
	AllocationTags          string                 `json:"AllocationTags,omitempty"`
	ExtraConfig             string                 `json:"ExtraConfig,omitempty"`
	Commands                string                 `json:"Commands,omitempty"`
	Volumes                 string                 `json:"Volumes,omitempty"`
	DockerImage             string                 `json:"DockerImage"`
	ReplicasMatchingAsgName string                 `json:"ReplicasMatchingAsgName,omitempty"`
	Replicas                int                    `json:"Replicas"`
	AgentPlatform           int                    `json:"AgentPlatform"`
	Cloud                   int                    `json:"Cloud"`
	Tags                    *[]DuploKeyStringValue `json:"Tags,omitempty"`
	IsLBSyncedDeployment    bool                   `json:"IsLBSyncedDeployment,omitempty"`
}

// DuploServiceGetList retrieves a list of services via the Duplo API.
func (c *Client) DuploServiceList(tenantID string) (*[]DuploService, ClientError) {
	rp := []DuploService{}
	err := c.getAPI(fmt.Sprintf("DuploServiceList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2", tenantID),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// DuploServiceGet retrieves a service's load balancer via the Duplo API.
func (c *Client) DuploServiceGet(tenantID string, name string) (*DuploService, ClientError) {
	rp := DuploService{}
	err := c.getAPI(fmt.Sprintf("DuploServiceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenantID, name),
		&rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, err
}

func (c *Client) DuploServiceExist(tenantID string, name string) bool {
	rp := DuploService{}
	err := c.getAPI(fmt.Sprintf("DuploServiceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenantID, name),
		&rp)
	if err != nil || rp.Name == "" {
		return false
	}
	return true
}

// DuploServiceCreate creates a service via the Duplo API.
func (c *Client) DuploServiceCreate(tenantID string, rq *DuploService) (*DuploService, ClientError) {
	return c.DuploServiceCreateOrUpdate(tenantID, rq, false)
}

// DuploServiceUpdate updates a service via the Duplo API.
func (c *Client) DuploServiceUpdate(tenantID string, rq *DuploService) (*DuploService, ClientError) {
	return c.DuploServiceCreateOrUpdate(tenantID, rq, true)
}

// DuploServiceCreateOrUpdate creates or updates a service via the Duplo API.
func (c *Client) DuploServiceCreateOrUpdate(tenantID string, rq *DuploService, updating bool) (*DuploService, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploService{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("DuploServiceCreateOrUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// DuploServiceDelete deletes a service via the Duplo API.
func (c *Client) DuploServiceDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploServiceDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenantID, name),
		nil)
}
