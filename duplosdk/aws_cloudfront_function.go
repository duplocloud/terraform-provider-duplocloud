package duplosdk

import (
	"fmt"
	"time"
)

// CloudFrontFunction represents a CloudFront Function in Duplo.
type DuploCloudFrontFunction struct {
	Name     string                           `json:"Name"`
	Runtime  string                           `json:"FunctionConfigRuntime"`
	Code     string                           `json:"FunctionCode"`
	Comment  string                           `json:"FunctionConfigComment,omitempty"`
	Status   string                           `json:"Status,omitempty"`
	Metadata *DuploCloudFrontFunctionMetadata `json:"FunctionMetadata,omitempty"`
}

// CloudFrontFunctionUpdateRequest represents a request to update a CloudFront Function.
type DuploCloudFrontFunctionMetadata struct {
	ARN string `json:"FunctionARN"`
}

// GetCloudFrontFunction retrieves a CloudFront Function by name.
func (c *Client) GetCloudFrontFunction(tenantID, name string) (*DuploCloudFrontFunctionResponse, ClientError) {
	rp := DuploCloudFrontFunctionResponse{}
	code := ""
	err := c.getAPI(fmt.Sprintf("GetCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudfront/function/%s", tenantID, name),
		&rp)

	if err != nil {
		return nil, err
	}

	err = c.getAPI(fmt.Sprintf("GetCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudfront/function/%s/code?stage=%s", tenantID, name, rp.FunctionSummary.FunctionMetadata.Stage.Value),
		&code)

	if err != nil {
		return nil, err
	}
	rp.FunctionCode = code

	return &rp, nil
}

// CreateCloudFrontFunction creates a new CloudFront Function.
func (c *Client) CreateCloudFrontFunction(tenantID string, rq *DuploCloudFrontFunction) (*DuploCloudFrontFunction, ClientError) {
	rp := DuploCloudFrontFunction{}
	err := c.postAPI(fmt.Sprintf("CreateCloudFrontFunction(%s,%s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFront/function", tenantID),
		rq,
		&rp)

	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// UpdateCloudFrontFunction updates an existing CloudFront Function.
func (c *Client) UpdateCloudFrontFunction(tenantID, name string, rq *DuploCloudFrontFunction) ClientError {
	rp := map[string]interface{}{}
	return c.putAPI(fmt.Sprintf("UpdateCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFront/function/%s", tenantID, name), rq, &rp)
}

// DeleteCloudFrontFunction deletes a CloudFront Function by name.
func (c *Client) DeleteCloudFrontFunction(tenantID, name string) ClientError {
	return c.deleteAPI(fmt.Sprintf("DeleteCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFront/function/%s", tenantID, name), nil)
}

func (c *Client) PublishCloudFrontFunction(tenantID, name string) ClientError {
	rq := map[string]string{
		"Name": name,
	}
	var rp interface{}
	return c.putAPI(fmt.Sprintf("PublishCloudFrontFunction(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/cloudFront/function/%s/publish", tenantID, name),
		rq,
		&rp)
}

type DuploCloudFrontFunctionResponse struct {
	ETag             string                                  `json:"ETag"`
	FunctionSummary  DuploCloudFrontFunctionFunctionSummary  `json:"FunctionSummary"`
	ResponseMetadata DuploCloudFrontFunctionResponseMetadata `json:"ResponseMetadata"`
	ContentLength    int                                     `json:"ContentLength"`
	HttpStatusCode   int                                     `json:"HttpStatusCode"`
	FunctionCode     string                                  `json:"-"`
}

type DuploCloudFrontFunctionFunctionSummary struct {
	FunctionConfig   DuploCloudFrontFunctionFunctionConfig    `json:"FunctionConfig"`
	FunctionMetadata *DuploCloudFrontFunctionFunctionMetadata `json:"FunctionMetadata,omitempty"`
	Name             string                                   `json:"Name"`
	Status           string                                   `json:"Status"`
}

type DuploCloudFrontFunctionFunctionConfig struct {
	Comment string                              `json:"Comment"`
	Runtime DuploCloudFrontFunctionRuntimeValue `json:"Runtime"`
}

type DuploCloudFrontFunctionRuntimeValue struct {
	Value string `json:"Value"`
}

type DuploCloudFrontFunctionFunctionMetadata struct {
	CreatedTime      time.Time `json:"CreatedTime"`
	FunctionARN      string    `json:"FunctionARN"`
	LastModifiedTime time.Time `json:"LastModifiedTime"`
	Stage            Stage     `json:"Stage"`
}

type Stage struct {
	Value string `json:"Value"`
}

type DuploCloudFrontFunctionResponseMetadata struct {
	RequestID                string `json:"requestIdField"`
	ChecksumAlgorithm        int    `json:"<ChecksumAlgorithm>k__BackingField"`
	ChecksumValidationStatus int    `json:"<ChecksumValidationStatus>k__BackingField"`
}
