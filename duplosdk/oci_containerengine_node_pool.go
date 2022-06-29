package duplosdk

import (
	"fmt"
)

type NodeLifecycleStateEnum string

// Set of constants representing the allowable values for NodeLifecycleStateEnum
const (
	NodeLifecycleStateCreating NodeLifecycleStateEnum = "CREATING"
	NodeLifecycleStateActive   NodeLifecycleStateEnum = "ACTIVE"
	NodeLifecycleStateUpdating NodeLifecycleStateEnum = "UPDATING"
	NodeLifecycleStateDeleting NodeLifecycleStateEnum = "DELETING"
	NodeLifecycleStateDeleted  NodeLifecycleStateEnum = "DELETED"
	NodeLifecycleStateFailing  NodeLifecycleStateEnum = "FAILING"
	NodeLifecycleStateInactive NodeLifecycleStateEnum = "INACTIVE"
)

type DuploOciNodePoolKeyValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type DuploOciNodePoolDetailsCreateReq struct {
	CompartmentId     string                            `json:"compartmentId,omitempty"`
	ClusterId         string                            `json:"clusterId,omitempty"`
	Name              string                            `json:"name,omitempty"`
	KubernetesVersion string                            `json:"kubernetesVersion,omitempty"`
	NodeShape         string                            `json:"nodeShape,omitempty"`
	NodeShapeConfig   *NodeShapeConfig                  `json:"nodeShapeConfig,omitempty"`
	InitialNodeLabels *[]DuploOciNodePoolKeyValue       `json:"initialNodeLabels,omitempty"`
	SshPublicKey      string                            `json:"sshPublicKey,omitempty"`
	QuantityPerSubnet int                               `json:"quantityPerSubnet,omitempty"`
	SubnetIds         []string                          `json:"subnetIds,omitempty"`
	NodeConfigDetails *NodePoolNodeConfigDetails        `json:"nodeConfigDetails,omitempty"`
	FreeformTags      map[string]string                 `json:"freeformTags,omitempty"`
	DefinedTags       map[string]map[string]interface{} `json:"definedTags,omitempty"`
	NodeImageName     string                            `json:"nodeImageName,omitempty"`
	NodeImageId       string                            `json:"nodeImageId,omitempty"`
	NodeMetadata      map[string]string                 `json:"nodeMetadata,omitempty"`
	NodeSourceDetails *NodeSourceDetails                `json:"nodeSourceDetails,omitempty"`
}

type DuploOciNodePool struct {
	Id                string                            `json:"id,omitempty"`
	CompartmentId     string                            `json:"compartmentId,omitempty"`
	ClusterId         string                            `json:"clusterId,omitempty"`
	Name              string                            `json:"name,omitempty"`
	KubernetesVersion string                            `json:"kubernetesVersion,omitempty"`
	NodeMetadata      map[string]string                 `json:"nodeMetadata,omitempty"`
	NodeImageId       string                            `json:"nodeImageId,omitempty"`
	NodeImageName     string                            `json:"nodeImageName,omitempty"`
	NodeShapeConfig   *NodeShapeConfig                  `json:"nodeShapeConfig,omitempty"`
	NodeSource        *NodeSource                       `json:"nodeSource,omitempty"`
	NodeSourceDetails *NodeSourceDetails                `json:"nodeSourceDetails,omitempty"`
	NodeShape         string                            `json:"nodeShape,omitempty"`
	InitialNodeLabels *[]DuploOciNodePoolKeyValue       `json:"initialNodeLabels,omitempty"`
	SshPublicKey      string                            `json:"sshPublicKey,omitempty"`
	QuantityPerSubnet int                               `json:"quantityPerSubnet,omitempty"`
	SubnetIds         []string                          `json:"subnetIds,omitempty"`
	Nodes             *[]Node                           `json:"nodes,omitempty"`
	NodeConfigDetails *NodePoolNodeConfigDetails        `json:"nodeConfigDetails,omitempty"`
	FreeformTags      map[string]string                 `json:"freeformTags,omitempty"`
	DefinedTags       map[string]map[string]interface{} `json:"definedTags,omitempty"`
	SystemTags        map[string]map[string]interface{} `json:"systemTags,omitempty"`
}

type NodeShapeConfig struct {
	Ocpus       float32 `json:"ocpus,omitempty"`
	MemoryInGBs float32 `json:"memoryInGBs,omitempty"`
}

type NodePoolNodeConfigDetails struct {
	Size                           int                               `json:"size,omitempty"`
	NsgIds                         []string                          `json:"nsgIds,omitempty"`
	KmsKeyId                       string                            `json:"kmsKeyId,omitempty"`
	IsPvEncryptionInTransitEnabled bool                              `json:"isPvEncryptionInTransitEnabled,omitempty"`
	FreeformTags                   map[string]string                 `json:"freeformTags,omitempty"`
	DefinedTags                    map[string]map[string]interface{} `json:"definedTags,omitempty"`
	PlacementConfigs               *[]NodePoolPlacementConfigDetails `json:"placementConfigs,omitempty"`
}

type NodePoolPlacementConfigDetails struct {
	AvailabilityDomain    string `json:"availabilityDomain,omitempty"`
	SubnetId              string `json:"subnetId,omitempty"`
	CapacityReservationId string `json:"capacityReservationId,omitempty"`
}

type NodeSourceDetails struct {
	SourceType          string `json:"sourceType,omitempty"`
	ImageId             string `json:"imageId,omitempty"`
	BootVolumeSizeInGBs int64  `json:"bootVolumeSizeInGBs,omitempty"`
}

type NodeSource struct {
	SourceType string `json:"sourceType,omitempty"`
	SourceName string `json:"sourceName,omitempty"`
	ImageId    string `json:"imageId,omitempty"`
}

type Node struct {
	Id                 string                            `json:"id,omitempty"`
	Name               string                            `json:"name,omitempty"`
	KubernetesVersion  string                            `json:"kubernetesVersion,omitempty"`
	AvailabilityDomain string                            `json:"availabilityDomain,omitempty"`
	SubnetId           string                            `json:"subnetId,omitempty"`
	NodePoolId         string                            `json:"nodePoolId,omitempty"`
	FaultDomain        string                            `json:"faultDomain,omitempty"`
	PrivateIp          string                            `json:"privateIp,omitempty"`
	PublicIp           string                            `json:"publicIp,omitempty"`
	FreeformTags       map[string]string                 `json:"freeformTags,omitempty"`
	DefinedTags        map[string]map[string]interface{} `json:"definedTags,omitempty"`
	SystemTags         map[string]map[string]interface{} `json:"systemTags,omitempty"`
	LifecycleState     NodeLifecycleStateEnum            `json:"lifecycleState,omitempty"`
	LifecycleDetails   string                            `json:"lifecycleDetails,omitempty"`
}

type DuploOciNodePoolDetailsCreateResponse struct {
	OpcWorkRequestId string `json:"OpcWorkRequestId,omitempty"`
	OpcRequestId     string `json:"OpcRequestId,omitempty"`
}

type DuploOciNodePoolGetResponse struct {
	NodePool     *DuploOciNodePool `json:"NodePool,omitempty"`
	OpcRequestId string            `json:"OpcRequestId,omitempty"`
	Etag         string            `json:"Etag,omitempty"`
}

func (c *Client) OciNodePoolCreate(tenantID string, name string, rq *DuploOciNodePoolDetailsCreateReq) ClientError {
	resp := DuploOciNodePoolDetailsCreateResponse{}
	return c.postAPI(
		fmt.Sprintf("NodePoolCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/oracle/nodepool", tenantID),
		&rq,
		&resp,
	)
}

func (c *Client) OciNodePoolGet(tenantID string, name string) (*DuploOciNodePool, ClientError) {
	list, err := c.OciNodePoolList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, nodePool := range *list {
			if nodePool.Name == name {
				return &nodePool, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) OciNodePoolGetWithNodes(tenantID string, id string) (*DuploOciNodePoolGetResponse, ClientError) {
	rp := DuploOciNodePoolGetResponse{}
	err := c.getAPI(
		fmt.Sprintf("NodePoolList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/oracle/nodepool/%s", tenantID, id),
		&rp,
	)
	return &rp, err
}

func (c *Client) OciNodePoolList(tenantID string) (*[]DuploOciNodePool, ClientError) {
	rp := []DuploOciNodePool{}
	err := c.getAPI(
		fmt.Sprintf("NodePoolList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/oracle/nodepool", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) OciNodePoolExists(tenantID, name string) (bool, ClientError) {
	list, err := c.OciNodePoolList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, nodePool := range *list {
			if nodePool.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) OciNodePoolDelete(tenantID string, id string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("NodePoolDelete(%s, %s)", tenantID, id),
		fmt.Sprintf("v3/subscriptions/%s/oracle/nodepool/%s", tenantID, id),
		nil,
	)
}
