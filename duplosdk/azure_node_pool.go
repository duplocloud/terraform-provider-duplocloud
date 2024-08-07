package duplosdk

import (
	"fmt"
)

type DuploAzureK8NodePoolRequest struct {
	MinSize                int                    `json:"MinSize"`
	MaxSize                int                    `json:"MaxSize"`
	DesiredCapacity        int                    `json:"DesiredCapacity"`
	EnableAutoScaling      bool                   `json:"EnableAutoScaling"`
	FriendlyName           string                 `json:"FriendlyName"`
	Capacity               string                 `json:"Capacity"`
	CustomDataTags         *[]DuploKeyStringValue `json:"CustomDataTags"`
	ScaleSetPriority       string                 `json:"ScaleSetPriority,omitempty"`
	ScaleSetEvictionPolicy string                 `json:"ScaleSetEvictionPolicy,omitempty"`
	SpotMaxPrice           float32                `json:"SpotMaxPrice,omitempty"`
}

type DuploAzureK8NodePoolDeleteRequest struct {
	FriendlyName string `json:"FriendlyName"`
	State        string `json:"State"`
}

type DuploAzureK8NodePool struct {
	CustomDataTags []struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"CustomDataTags"`
	FriendlyName           string        `json:"FriendlyName"`
	Capacity               string        `json:"Capacity"`
	IsMinion               bool          `json:"IsMinion"`
	Zone                   int           `json:"Zone"`
	Volumes                []interface{} `json:"Volumes"`
	Tags                   []interface{} `json:"Tags"`
	TagsEx                 []interface{} `json:"TagsEx"`
	AgentPlatform          int           `json:"AgentPlatform"`
	IsEbsOptimized         bool          `json:"IsEbsOptimized"`
	Cloud                  int           `json:"Cloud"`
	AllocatedPublicIP      bool          `json:"AllocatedPublicIp"`
	NetworkInterfaces      []interface{} `json:"NetworkInterfaces"`
	MetaData               []interface{} `json:"MetaData"`
	MinionTags             []interface{} `json:"MinionTags"`
	EncryptDisk            bool          `json:"EncryptDisk"`
	ProvisioningState      string        `json:"ProvisioningState"`
	DesiredCapacity        int           `json:"DesiredCapacity"`
	MaxSize                int           `json:"MaxSize"`
	MinSize                int           `json:"MinSize"`
	EnableAutoScaling      bool          `json:"EnableAutoScaling"`
	ScaleSetPriority       string        `json:"ScaleSetPriority,omitempty"`
	ScaleSetEvictionPolicy string        `json:"ScaleSetEvictionPolicy,omitempty"`
	SpotMaxPrice           float32       `json:"SpotMaxPrice,omitempty"`
}

func (c *Client) AzureK8NodePoolCreate(tenantID string, rq *DuploAzureK8NodePoolRequest) (*string, ClientError) {
	resp := ""
	err := c.postAPI(
		fmt.Sprintf("AzureK8NodePoolCreate(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/UpdateAzureAgentPool", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) AzureK8NodePoolGet(tenantID string, name string) (*DuploAzureK8NodePool, ClientError) {
	list, err := c.AzureK8NodePoolList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, ap := range *list {
			if ap.FriendlyName == name {
				return &ap, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureK8NodePoolList(tenantID string) (*[]DuploAzureK8NodePool, ClientError) {
	rp := []DuploAzureK8NodePool{}
	err := c.getAPI(
		fmt.Sprintf("AzureK8NodePoolList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAzureAgentPoolDetails", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureK8NodePoolExists(tenantID, name string) (bool, ClientError) {
	list, err := c.AzureK8NodePoolList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, ap := range *list {
			if ap.FriendlyName == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) AzureK8NodePoolDelete(tenantID string, name string) ClientError {
	resp := ""
	return c.postAPI(
		fmt.Sprintf("AzureK8NodePoolDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/UpdateAzureAgentPool", tenantID),
		&DuploAzureK8NodePoolDeleteRequest{
			FriendlyName: name,
			State:        "delete",
		},
		&resp,
	)
}
