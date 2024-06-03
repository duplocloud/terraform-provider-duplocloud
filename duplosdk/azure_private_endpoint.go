package duplosdk

import (
	"fmt"
)

type DuploAzurePrivateEndpointSubnetID struct {
	Id string `json:"id"`
}
type DuploAzurePrivateEndpoint struct {
	Name                          string                                     `json:"name"`
	PropertiesSubnetID            *DuploAzurePrivateEndpointSubnetID         `json:"properties.subnet"`
	PrivateLinkServiceConnections *[]DuploAzurePrivateLinkServiceConnections `json:"properties.privateLinkServiceConnections"`
}

type DuploAzurePrivateLinkServiceConnections struct {
	PrivateLinkServiceId string    `json:"properties.privateLinkServiceId"`
	GroupIds             *[]string `json:"properties.groupIds"`
	Name                 string    `json:"name"`
}

func (c *Client) PrivateEndpointCreate(tenantID string, rq *DuploAzurePrivateEndpoint) ClientError {
	resp := DuploAzurePrivateEndpoint{}
	return c.postAPI(
		fmt.Sprintf("PrivateEndpointCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/privateEndpoint", tenantID),
		&rq,
		&resp,
	)
}

func (c *Client) PrivateEndpointGet(tenantID string, name string) (*DuploAzurePrivateEndpoint, ClientError) {
	list, err := c.PrivateEndpointList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, pe := range *list {
			if pe.Name == name {
				return &pe, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) PrivateEndpointList(tenantID string) (*[]DuploAzurePrivateEndpoint, ClientError) {
	rp := []DuploAzurePrivateEndpoint{}
	err := c.getAPI(
		fmt.Sprintf("PrivateEndpointList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/privateEndpoint", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) PrivateEndpointExists(tenantID, name string) (bool, ClientError) {
	list, err := c.PrivateEndpointList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, pe := range *list {
			if pe.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) PrivateEndpointDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PrivateEndpointDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/privateEndpoint/%s", tenantID, name),
		nil,
	)
}
