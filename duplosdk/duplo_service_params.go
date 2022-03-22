package duplosdk

import (
	"fmt"
)

// DuploServiceParams represents a service's parameters in the Duplo SDK
type DuploServiceParams struct {
	ReplicationControllerName string `json:"ReplicationControllerName"`
	TenantID                  string `json:"TenantId,omitempty"`
	WebACLId                  string `json:"WebACLId,omitempty"`
	DNSPrfx                   string `json:"DnsPrfx,omitempty"`
}

// DuploServiceParamsGetList retrieves a list of service load balancers via the Duplo API.
func (c *Client) DuploServiceParamsGetList(tenantID string) (*[]DuploServiceParams, ClientError) {

	// Retrieve the objects.
	list := []DuploServiceParams{}
	err := c.getAPI(
		fmt.Sprintf("DuploServiceParamsGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Assign the tenant ID to every object and return the list.
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// DuploServiceParamsDelete deletes a service's load balancer via the Duplo API.
func (c *Client) DuploServiceParamsDelete(tenantID string, name string) ClientError {
	// Delete the service parameters
	return c.deleteAPI(
		fmt.Sprintf("DuploServiceParamsDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenantID, name),
		nil)
}
