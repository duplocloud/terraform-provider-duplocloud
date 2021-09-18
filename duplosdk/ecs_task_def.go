package duplosdk

import (
	"fmt"
)

// DuploEcsTaskDefPlacementConstraint represents an ECS placement constraint in the Duplo SDK
type DuploEcsTaskDefPlacementConstraint struct {
	Type       string `json:"Type"`
	Expression string `json:"Expression"`
}

// DuploEcsTaskDefProxyConfig represents an ECS proxy configuration in the Duplo SDK
type DuploEcsTaskDefProxyConfig struct {
	ContainerName string                  `json:"ContainerName"`
	Properties    *[]DuploNameStringValue `json:"Properties"`
	Type          string                  `json:"Type"`
}

// DuploEcsTaskDefInferenceAccelerator represents an inference accelerator in the Duplo SDK
type DuploEcsTaskDefInferenceAccelerator struct {
	DeviceName string `json:"DeviceName"`
	DeviceType string `json:"DeviceType"`
}

// DuploEcsTaskDef represents an ECS task definition in the Duplo SDK
type DuploEcsTaskDef struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Family                  string                                 `json:"Family"`
	Revision                int                                    `json:"Revision,omitempty"`
	Arn                     string                                 `json:"TaskDefinitionArn,omitempty"`
	ContainerDefinitions    []map[string]interface{}               `json:"ContainerDefinitions,omitempty"`
	CPU                     string                                 `json:"Cpu,omitempty"`
	TaskRoleArn             string                                 `json:"TaskRoleArn,omitempty"`
	ExecutionRoleArn        string                                 `json:"ExecutionRoleArn,omitempty"`
	Memory                  string                                 `json:"Memory,omitempty"`
	IpcMode                 string                                 `json:"IpcMode,omitempty"`
	PidMode                 string                                 `json:"PidMode,omitempty"`
	NetworkMode             *DuploStringValue                      `json:"NetworkMode,omitempty"`
	PlacementConstraints    *[]DuploEcsTaskDefPlacementConstraint  `json:"PlacementConstraints,omitempty"`
	ProxyConfiguration      *DuploEcsTaskDefProxyConfig            `json:"ProxyConfiguration,omitempty"`
	RequiresAttributes      *[]DuploName                           `json:"RequiresAttributes,omitempty"`
	RequiresCompatibilities []string                               `json:"RequiresCompatibilities,omitempty"`
	Tags                    *[]DuploKeyStringValue                 `json:"Tags,omitempty"`
	InferenceAccelerators   *[]DuploEcsTaskDefInferenceAccelerator `json:"InferenceAccelerators,omitempty"`
	Status                  *DuploStringValue                      `json:"Status,omitempty"`
	Volumes                 []map[string]interface{}               `json:"Volumes,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcsTaskDefinitionCreate creates an ECS task definition via the Duplo API.
func (c *Client) EcsTaskDefinitionCreate(tenantID string, rq *DuploEcsTaskDef) (string, ClientError) {
	var arn string

	err := c.postAPI(
		fmt.Sprintf("EcsTaskDefinitionCreate(%s, %s)", tenantID, rq.Family),
		fmt.Sprintf("subscriptions/%s/UpdateEcsTaskDefinition", tenantID),
		rq,
		&arn,
	)
	if err == nil && arn == "" {
		return "", newClientError(fmt.Sprintf("failed to create ECS task def: %s", rq.Family))
	}
	return arn, err
}

// EcsTaskDefinitionGet retrieves an ECS task definition via the Duplo API.
func (c *Client) EcsTaskDefinitionGet(tenantID, arn string) (*DuploEcsTaskDef, ClientError) {
	rq := map[string]interface{}{"Arn": arn}
	rp := DuploEcsTaskDef{}

	err := c.postAPI(
		fmt.Sprintf("EcsTaskDefinitionGet(%s, %s)", tenantID, arn),
		fmt.Sprintf("v2/subscriptions/%s/FindEcsTaskDefinition", tenantID),
		rq,
		&rp,
	)

	// Fill in the tenant ID and return the object
	rp.TenantID = tenantID
	return &rp, err
}

// EcsTaskDefinitionDelete deletes an ECS task definition via the Duplo API.
func (c *Client) EcsTaskDefinitionDelete(tenantID, arn string) ClientError {
	rq := map[string]interface{}{"Arn": arn}
	rp := DuploEcsTaskDef{}

	err := c.postAPI(
		fmt.Sprintf("EcsTaskDefinitionDelete(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/RemoveEcsTaskDefinition", tenantID),
		rq,
		&rp,
	)
	return err
}

// EcsTaskDefinitionExists checks if an ECS task definition is exists via the Duplo API.
func (c *Client) EcsTaskDefinitionExists(tenantID, arn string) (bool, ClientError) {
	rp := []string{}

	err := c.getAPI(
		fmt.Sprintf("EcsTaskDefinitionExists(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/GetEcsTaskDefinitionArns", tenantID),
		&rp,
	)
	if err != nil {
		return false, err
	}
	// Check if the host exists
	for _, taskArn := range rp {
		if taskArn == arn {
			return true, nil
		}
	}
	return false, nil
}
