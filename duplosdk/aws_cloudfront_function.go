package duplosdk

import (
	"fmt"
)

// CloudFrontFunction represents a CloudFront Function in Duplo.
type DuploCloudFrontFunction struct {
	Name     string                           `json:"Name"`
	Runtime  string                           `json:"FunctionConfigRuntime"`
	Code     string                           `json:"FunctionCode"`
	Comment  string                           `json:"FunctionConfigComment,omitempty"`
	Status   string                           `json:"Status,omitempty"`
	Metadata *DuploCloudFrontFunctionMetadata `json:"FunctionMetadata"`
}

// CloudFrontFunctionUpdateRequest represents a request to update a CloudFront Function.
type DuploCloudFrontFunctionMetadata struct {
	ARN string `json:"FunctionARN"`
}

// GetCloudFrontFunction retrieves a CloudFront Function by name.
func (c *Client) GetCloudFrontFunction(tenantID, name string) (*DuploCloudFrontFunction, ClientError) {
	rp := DuploCloudFrontFunction{}

	err := c.getAPI(fmt.Sprintf("GetCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name),
		&rp)

	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// CreateCloudFrontFunction creates a new CloudFront Function.
func (c *Client) CreateCloudFrontFunction(tenantID string, rq *DuploCloudFrontFunction) (*DuploCloudFrontFunction, ClientError) {
	rp := DuploCloudFrontFunction{}
	err := c.postAPI(fmt.Sprintf("CreateCloudFrontFunction(%s,%s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction", tenantID),
		rq,
		&rp)

	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// UpdateCloudFrontFunction updates an existing CloudFront Function.
//func (c *Client) UpdateCloudFrontFunction(tenantID, name string, rq *CloudFrontFunctionUpdateRequest) (*CloudFrontFunction, ClientError) {
//	rp := CloudFrontFunction{}
//	err := c.putAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name), rq, &rp)
//	if err != nil {
//		return nil, err
//	}
//	return &rp, nil
//}

// DeleteCloudFrontFunction deletes a CloudFront Function by name.
//func (c *Client) DeleteCloudFrontFunction(tenantID, name string) ClientError {
//	return c.deleteAPI(fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s", tenantID, name), nil)
//}

// PublishCloudFrontFunction publishes a CloudFront Function by name.
func (c *Client) PublishCloudFrontFunction(tenantID, name string) ClientError {
	rq := map[string]string{
		"Name": name,
	}
	var rp interface{}
	return c.putAPI(fmt.Sprintf("PublishCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/cloudfrontFunction/%s/publish", tenantID, name),
		rq,
		&rp)
}
