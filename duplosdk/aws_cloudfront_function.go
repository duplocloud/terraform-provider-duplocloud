package duplosdk

import (
	"fmt"
)

// CloudFrontFunction represents a CloudFront Function in Duplo.
type CloudFrontFunctionRequest struct {
	Name    string `json:"Name"`
	Runtime string `json:"FunctionConfigRuntime"`
	Code    string `json:"FunctionCode"`
	Comment string `json:"FunctionConfigComment,omitempty"`
}

// CloudFrontFunctionUpdateRequest represents a request to update a CloudFront Function.
type CloudFrontFunctionUpdateRequest struct {
	Code  string `json:"code"`
	Stage string `json:"stage"`
}

// GetCloudFrontFunction retrieves a CloudFront Function by name.
func (c *Client) GetCloudFrontFunction(tenantID, name string) (*CloudFrontFunctionRequest, error) {
	rp := CloudFrontFunctionRequest{}
	err := c.getAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name), &rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// CreateCloudFrontFunction creates a new CloudFront Function.
func (c *Client) CreateCloudFrontFunction(tenantID string, rq *CloudFrontFunctionCreateRequest) (*CloudFrontFunction, error) {
	rp := CloudFrontFunction{}
	err := c.postAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction", tenantID), rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// UpdateCloudFrontFunction updates an existing CloudFront Function.
func (c *Client) UpdateCloudFrontFunction(tenantID, name string, rq *CloudFrontFunctionUpdateRequest) (*CloudFrontFunction, error) {
	rp := CloudFrontFunction{}
	err := c.putAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name), rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// DeleteCloudFrontFunction deletes a CloudFront Function by name.
func (c *Client) DeleteCloudFrontFunction(tenantID, name string) error {
	return c.deleteAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name), nil)
}
