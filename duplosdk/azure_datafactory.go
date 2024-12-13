package duplosdk

import "fmt"

type DuplocloudAzureDataFactoryRequest struct {
	Name           string `json:"Name"`
	PublicEndPoint bool   `json:"PublicEndPoint"`
}

type DuplocloudAzureDataFactoryResponse struct {
	Name         string `json:"name"`
	PublicAccess string `json:"properties.publicNetworkAccess"`
	Version      string `json:"properties.version"`
	Type         string `json:"type"`
	Location     string `json:"location"`
	State        string `json:"properties.provisioningState"`
	ETag         string `json:"eTag"`
}

func (c *Client) CreateAzureDataFactory(tenantId string, rq DuplocloudAzureDataFactoryRequest) ClientError {
	rp := map[string]interface{}{}
	return c.postAPI(fmt.Sprintf("CreateAzureDataFactory(%s,%s)", tenantId, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/datafactory", tenantId), &rq, &rp)
}

func (c *Client) GetAzureDataFactory(tenantId string, name string) (*DuplocloudAzureDataFactoryResponse, ClientError) {
	rp := DuplocloudAzureDataFactoryResponse{}
	err := c.getAPI(fmt.Sprintf("GetAzureDataFactory(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/datafactory/%s", tenantId, name), &rp)
	return &rp, err
}

func (c *Client) DeleteAzureDataFactory(tenantId string, name string) ClientError {
	return c.deleteAPI(fmt.Sprintf("DeleteAzureDataFactory(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/datafactory/%s", tenantId, name), nil)

}
