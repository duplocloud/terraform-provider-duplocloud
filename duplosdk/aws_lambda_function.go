package duplosdk

import (
	"fmt"
)

// DuploLambdaFunction is a Duplo SDK object that represents a lambda function.
type DuploLambdaFunction struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	CodeSha256    string                    `json:"CodeSha256"`
	CodeSize      int                       `json:"CodeSize"`
	Code          *DuploLambdaCode          `json:"Code"`
	Description   string                    `json:"Description,omitempty"`
	Environment   *DuploLambdaEnvironment   `json:"Environment,omitempty"`
	FunctionArn   string                    `json:"FunctionArn,omitempty"`
	FunctionName  string                    `json:"FunctionName,omitempty"`
	Handler       string                    `json:"Handler,omitempty"`
	LastModified  string                    `json:"LastModified,omitempty"`
	Role          string                    `json:"Role,omitempty"`
	MemorySize    int                       `json:"MemorySize"`
	Runtime       *DuploStringValue         `json:"Runtime,omitempty"`
	Timeout       int                       `json:"Timeout,omitempty"`
	Version       string                    `json:"Version,omitempty"`
	TracingConfig *DuploLambdaTracingConfig `json:"TracingConfig,omitempty"`
	VpcConfig     *DuploLambdaVpcConfig     `json:"VpcConfig,omitempty"`
}

// DuploLambdaCode is a Duplo SDK object that represents a lambda function's code.
type DuploLambdaCode struct {
	S3Bucket string `json:"S3Bucket,omitempty"`
	S3Key    string `json:"S3Key,omitempty"`
}

// DuploLambdaEnvironment is a Duplo SDK object that represents a lambda function's tracing config.
type DuploLambdaEnvironment struct {
	Variables map[string]string `json:"Variables,omitempty"`
}

// DuploLambdaTracingConfig is a Duplo SDK object that represents a lambda function's tracing config.
type DuploLambdaTracingConfig struct {
	Mode DuploStringValue `json:"Mode,omitempty"`
}

// DuploLambdaVpcConfig is a Duplo SDK object that represents a lambda function's tracing config.
type DuploLambdaVpcConfig struct {
	SecurityGroupIDs []string `json:"SecurityGroupIds,omitempty"`
	SubnetIDs        []string `json:"SubnetIds,omitempty"`
	VpcID            string   `json:"VpcId,omitempty"`
}

// DuploLambdaUpdateRequest is a Duplo SDK object that represents a request to update a lambda function's code.
type DuploLambdaUpdateRequest struct {
	FunctionName string `json:"FunctionName,omitempty"`
	S3Bucket     string `json:"S3Bucket,omitempty"`
	S3Key        string `json:"S3Key,omitempty"`
}

// DuploLambdaConfigurationRequest is a Duplo SDK object that represents a request to update a lambda function's configuration.
type DuploLambdaConfigurationRequest struct {
	FunctionName string                  `json:"FunctionName,omitempty"`
	Handler      string                  `json:"Handler,omitempty"`
	Runtime      *DuploStringValue       `json:"Runtime,omitempty"`
	Description  string                  `json:"Description,omitempty"`
	Timeout      int                     `json:"Timeout,omitempty"`
	MemorySize   int                     `json:"MemorySize"`
	Environment  *DuploLambdaEnvironment `json:"Environment,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// LambdaFunctionCreate creates a lambda function via the Duplo API.
func (c *Client) LambdaFunctionCreate(tenantID string, rq *DuploLambdaFunction) error {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateLambdaFunction", tenantID),
		&rq,
		nil,
	)
}

// LambdaFunctionUpdate updates a lambda function via the Duplo API.
func (c *Client) LambdaFunctionUpdate(tenantID string, rq *DuploLambdaUpdateRequest) error {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionUpdate(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("subscriptions/%s/UpdateLambdaFunction", tenantID),
		&rq,
		nil,
	)
}

// LambdaFunctionUpdateConfiguration updates a lambda function's configuration via the Duplo API.
func (c *Client) LambdaFunctionUpdateConfiguration(tenantID string, rq *DuploLambdaConfigurationRequest) error {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionUpdateConfigurationg(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("subscriptions/%s/UpdateLambdaFunctionConfiguration", tenantID),
		&rq,
		nil,
	)
}

// LambdaFunctionDelete deletes a lambda function via the Duplo API.
func (c *Client) LambdaFunctionDelete(tenantID, identifier string) error {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionDelete(%s, duplo%s)", tenantID, identifier),
		fmt.Sprintf("subscriptions/%s/DeleteLambdaFunction/%s", tenantID, identifier),
		nil,
		nil)
}

// LambdaFunctionGetList gets a list of lambda functions via the Duplo API.
func (c *Client) LambdaFunctionGetList(tenantID string) (*[]DuploLambdaFunction, error) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploLambdaFunction{}
	err = c.getAPI(
		fmt.Sprintf("LambdaFunctionGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetLambdaFunctions", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID and name to each element and return the list.
	for i := range list {
		list[i].TenantID = tenantID
		list[i].Name, _ = UnprefixName(prefix, list[i].FunctionName)
	}
	return &list, nil
}

// LambdaFunctionGet gets a lambda function via the Duplo API.
func (c *Client) LambdaFunctionGet(tenantID string, name string) (*DuploLambdaFunction, error) {

	// Get the list from Duplo
	list, err := c.LambdaFunctionGetList(tenantID)
	if err != nil {
		return nil, err
	}

	// Return the matching object
	for _, item := range *list {
		if item.Name == name {
			return &item, nil
		}
	}

	// Nothing was found
	return nil, nil
}
