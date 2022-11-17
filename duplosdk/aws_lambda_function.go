package duplosdk

import (
	"fmt"
)

// DuploLambdaFunction is a Duplo SDK object that represents a lambda function.
type DuploLambdaFunction struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Code          DuploLambdaCode          `json:"Code"`
	Configuration DuploLambdaConfiguration `json:"Configuration"`
	Tags          map[string]string        `json:"Tags,omitempty"`
}

type DuploLambdaLayerGet struct {
	Arn      string `json:"ARN"`
	CodeSize int64  `json:"CodeSize"`
}

// DuploLambdaConfiguration is a Duplo SDK object that represents a lambda function's configuration.
type DuploLambdaConfiguration struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	CodeSha256    string                    `json:"CodeSha256"`
	CodeSize      int                       `json:"CodeSize"`
	Description   string                    `json:"Description,omitempty"`
	Environment   *DuploLambdaEnvironment   `json:"Environment,omitempty"`
	FunctionArn   string                    `json:"FunctionArn,omitempty"`
	FunctionName  string                    `json:"FunctionName,omitempty"`
	Handler       string                    `json:"Handler,omitempty"`
	LastModified  string                    `json:"LastModified,omitempty"`
	MemorySize    int                       `json:"MemorySize"`
	Role          string                    `json:"Role,omitempty"`
	PackageType   *DuploStringValue         `json:"PackageType,omitempty"`
	Runtime       *DuploStringValue         `json:"Runtime,omitempty"`
	Timeout       int                       `json:"Timeout,omitempty"`
	TracingConfig *DuploLambdaTracingConfig `json:"TracingConfig,omitempty"`
	Version       string                    `json:"Version,omitempty"`
	VpcConfig     *DuploLambdaVpcConfig     `json:"VpcConfig,omitempty"`
	Layers        *[]DuploLambdaLayerGet    `json:"Layers,omitempty"`
}

// DuploLambdaCode is a Duplo SDK object that represents a lambda function's code.
type DuploLambdaCode struct {
	ImageURI string `json:"ImageUri,omitempty"`
	S3Bucket string `json:"S3Bucket,omitempty"`
	S3Key    string `json:"S3Key,omitempty"`
}

// DuploLambdaEnvironment is a Duplo SDK object that represents a lambda function's environment config.
type DuploLambdaEnvironment struct {
	Variables map[string]string `json:"Variables,omitempty"`
}

// DuploLambdaEphemeralStorage is a Duplo SDK object that represents a lambda function's ephemeral storage config.
type DuploLambdaEphemeralStorage struct {
	Size int  `json:"Size"`
}


// DuploLambdaTracingConfig is a Duplo SDK object that represents a lambda function's tracing config.
type DuploLambdaTracingConfig struct {
	Mode DuploStringValue `json:"Mode,omitempty"`
}

// DuploLambdaVpcConfig is a Duplo SDK object that represents a lambda function's vpn config.
type DuploLambdaVpcConfig struct {
	SecurityGroupIDs []string `json:"SecurityGroupIds,omitempty"`
	SubnetIDs        []string `json:"SubnetIds,omitempty"`
	VpcID            string   `json:"VpcId,omitempty"`
}

// DuploLambdaCreateRequest is a Duplo SDK object that represents a request to create a lambda function.
type DuploLambdaCreateRequest struct {
	FunctionName string                  `json:"FunctionName"`
	Code         DuploLambdaCode         `json:"Code"`
	Handler      string                  `json:"Handler,omitempty"`
	Description  string                  `json:"Description,omitempty"`
	Timeout      int                     `json:"Timeout,omitempty"`
	MemorySize   int                     `json:"MemorySize"`
	PackageType  *DuploStringValue       `json:"PackageType,omitempty"`
	Runtime      *DuploStringValue       `json:"Runtime,omitempty"`
	Environment  *DuploLambdaEnvironment `json:"Environment,omitempty"`
	Tags         map[string]string       `json:"Tags,omitempty"`
	Layers       *[]string               `json:"Layers,omitempty"`
}

// DuploLambdaUpdateRequest is a Duplo SDK object that represents a request to update a lambda function's code.
type DuploLambdaUpdateRequest struct {
	FunctionName string `json:"FunctionName,omitempty"`
	ImageURI     string `json:"ImageUri,omitempty"`
	S3Bucket     string `json:"S3Bucket,omitempty"`
	S3Key        string `json:"S3Key,omitempty"`
}

// DuploLambdaConfigurationRequest is a Duplo SDK object that represents a request to update a lambda function's configuration.
type DuploLambdaConfigurationRequest struct {
	FunctionName     string                       `json:"FunctionName,omitempty"`
	Handler          string                       `json:"Handler,omitempty"`
	Runtime          *DuploStringValue            `json:"Runtime,omitempty"`
	Description      string                       `json:"Description,omitempty"`
	Timeout          int                          `json:"Timeout,omitempty"`
	MemorySize       int                          `json:"MemorySize"`
	Environment      *DuploLambdaEnvironment      `json:"Environment,omitempty"`
	EphemeralStorage *DuploLambdaEphemeralStorage `json:"EphemeralStorage,omitempty"`
	Tags             map[string]string            `json:"Tags,omitempty"`
	Layers           *[]string                    `json:"Layers,omitempty"`
}

type DuploLambdaPermissionStatement struct {
	Sid       string                         `json:"Sid,omitempty"`
	Effect    string                         `json:"Effect,omitempty"`
	Principal DuploLambdaPermissionPrincipal `json:"Principal,omitempty"`
	Action    string                         `json:"Action,omitempty"`
	Resource  string                         `json:"Resource,omitempty"`
}

type DuploLambdaPermissionPrincipal struct {
	Service string `json:"Service,omitempty"`
}

type DuploLambdaPermissionRequest struct {
	Action           string `json:"Action,omitempty"`
	FunctionName     string `json:"FunctionName,omitempty"`
	Principal        string `json:"Principal,omitempty"`
	EventSourceToken string `json:"EventSourceToken,omitempty"`
	Qualifier        string `json:"Qualifier,omitempty"`
	SourceAccount    string `json:"SourceAccount,omitempty"`
	SourceArn        string `json:"SourceArn,omitempty"`
	StatementId      string `json:"StatementId,omitempty"`
	RevisionId       string `json:"RevisionId,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// LambdaFunctionCreate creates a lambda function via the Duplo API.
func (c *Client) LambdaFunctionCreate(tenantID string, rq *DuploLambdaCreateRequest) (*DuploLambdaConfiguration, ClientError) {
	rp := DuploLambdaConfiguration{}
	err := c.postAPI(
		fmt.Sprintf("LambdaFunctionCreate(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambda", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// LambdaFunctionUpdate updates a lambda function via the Duplo API.
func (c *Client) LambdaFunctionUpdate(tenantID string, rq *DuploLambdaUpdateRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionUpdate(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("subscriptions/%s/UpdateLambdaFunction", tenantID),
		&rq,
		nil,
	)
}

// LambdaFunctionUpdateConfiguration updates a lambda function's configuration via the Duplo API.
func (c *Client) LambdaFunctionUpdateConfiguration(tenantID string, rq *DuploLambdaConfigurationRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("LambdaFunctionUpdateConfigurationg(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("subscriptions/%s/UpdateLambdaFunctionConfiguration", tenantID),
		&rq,
		nil,
	)
}

// LambdaFunctionDelete deletes a lambda function via the Duplo API.
func (c *Client) LambdaFunctionDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("LambdaFunctionDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambda/%s", tenantID, name),
		nil)
}

// LambdaFunctionGetList gets a list of lambda functions via the Duplo API.
func (c *Client) LambdaFunctionGetList(tenantID string) (*[]DuploLambdaConfiguration, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}
	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploLambdaConfiguration{}
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
		list[i].Name, _ = UnwrapName(prefix, accountID, list[i].FunctionName, true)
	}
	return &list, nil
}

// LambdaFunctionGet gets a lambda function via the Duplo API.
func (c *Client) LambdaFunctionGet(tenantID string, name string) (*DuploLambdaFunction, ClientError) {

	// Get the list from Duplo
	rp := DuploLambdaFunction{}
	err := c.getAPI(
		fmt.Sprintf("LambdaFunctionGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambda/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	rp.Name = name
	rp.Configuration.TenantID = tenantID
	rp.Configuration.Name = name
	return &rp, err
}

func (c *Client) LambdaPermissionCreate(tenantID string, rq *DuploLambdaPermissionRequest) (*DuploLambdaPermissionRequest, ClientError) {
	rp := DuploLambdaPermissionRequest{}
	err := c.postAPI(
		fmt.Sprintf("LambdaPermissionCreate(%s, %s)", tenantID, rq.FunctionName),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambdapermission", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

func (c *Client) LambdaPermissionDelete(tenantID, functionName, statementId string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("LambdaPermissionDelete(%s, %s, %s)", tenantID, functionName, statementId),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambdapermission/%s/%s", tenantID, functionName, statementId), nil)
}

func (c *Client) LambdaPermissionGet(tenantID string, functionName string) (*[]DuploLambdaPermissionStatement, ClientError) {
	rp := []DuploLambdaPermissionStatement{}
	err := c.getAPI(
		fmt.Sprintf("LambdaPermissionGet(%s, %s)", tenantID, functionName),
		fmt.Sprintf("v3/subscriptions/%s/serverless/lambdapermission/%s", tenantID, functionName),
		&rp)
	if len(rp) == 0 {
		return nil, err
	}
	return &rp, err
}
