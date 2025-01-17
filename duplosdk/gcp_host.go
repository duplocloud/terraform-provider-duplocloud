package duplosdk

import (
	"fmt"
)

type DuploGcpHost struct {
	Capacity              string                 `json:"Capacity"`
	AgentPlatform         int                    `json:"AgentPlatform"`
	Tags                  []string               `json:"Tags"`
	ImageId               string                 `json:"ImageId,omitempty"`
	EnablePublicIpAddress bool                   `json:"EnablePublicIpAddress"`
	Zone                  string                 `json:"Zone"`
	Name                  string                 `json:"Name"`
	AcceleratorCount      int                    `json:"AcceleratorCount,omitempty"`
	AcceleratorType       string                 `json:"AcceleratorType,omitempty"`
	Labels                map[string]string      `json:"Labels,omitempty"`
	Metadata              map[string]interface{} `json:"Metadata,omitempty"`
	UserData              string                 `json:"UserData,omitempty"`
	/***/
	PrivateIpAddress string `json:"PrivateIpAddress,omitempty"`
	PublicIpAddress  string `json:"PublicIpAddress,omitempty"`
	Arch             string `json:"Arch,omitempty"`
	IdentityRole     string `json:"IdentityRole,omitempty"`
	InstanceId       string `json:"InstanceId,omitempty"`
	SelfLink         string `json:"SelfLink,omitempty"`
	Status           string `json:"Status,omitempty"`
	ResourceType     int    `json:"ResourceType default:29"`
}

func (c *Client) DuploGcpHostCreate(tenantID string, rq *DuploGcpHost) (*DuploGcpHost, ClientError) {
	rp := DuploGcpHost{}
	err := c.postAPI(
		fmt.Sprintf("DuploGcpHostCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/nativeHosts", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploGcpNativeHostGet(tenantID, instanceId string) (*DuploGcpHost, ClientError) {
	var resp DuploGcpHost
	err := c.getAPI(
		fmt.Sprintf("DuploGcpNativeHostGet(%s, %s)", tenantID, instanceId),
		fmt.Sprintf("v3/subscriptions/%s/google/nativeHosts/%s", tenantID, instanceId),
		&resp)
	return &resp, err
}

func (c *Client) DuploGcpHostUpdate(tenantId, instanceId string, rq *DuploGcpHost) (*DuploGcpHost, ClientError) {
	resp := DuploGcpHost{}
	err := c.putAPI(
		fmt.Sprintf("DuploGcpHostUpdate(%s, %s)", tenantId, instanceId),
		fmt.Sprintf("v3/subscriptions/%s/google/nativeHosts/%s", tenantId, instanceId),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploGcpHostDelete(tenantId string, instanceId string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploGcpHostDelete(%s, %s)", tenantId, instanceId),
		fmt.Sprintf("v3/subscriptions/%s/google/nativeHosts/%s", tenantId, instanceId),
		nil,
	)
}
