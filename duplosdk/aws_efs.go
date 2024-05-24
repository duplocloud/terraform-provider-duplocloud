package duplosdk

import (
	"fmt"
)

type DuploEFSCreateReq struct {
	Name                         string  `json:"Name"`
	PerformanceMode              string  `json:"PerformanceMode,omitempty"`
	ThroughputMode               string  `json:"ThroughputMode,omitempty"`
	Backup                       bool    `json:"Backup"`
	Encrypted                    bool    `json:"Encrypted"`
	CreationToken                string  `json:"CreationToken,omitempty"`
	ProvisionedThroughputInMibps float64 `json:"ProvisionedThroughputInMibps,omitempty"`
}

type DuploEFSUpdateReq struct {
	FileSystemId                 string            `json:"FileSystemId"`
	ThroughputMode               *DuploStringValue `json:"ThroughputMode,omitempty"`
	ProvisionedThroughputInMibps float64           `json:"ProvisionedThroughputInMibps,omitempty"`
}

type EFSSizeInBytes struct {
	Timestamp       string `json:"Timestamp,omitempty"`
	Value           int    `json:"Value,omitempty"`
	ValueInIA       int    `json:"ValueInIA,omitempty"`
	ValueInStandard int    `json:"ValueInStandard,omitempty"`
}
type DuploEFSGetResp struct {
	CreationTime                 string                 `json:"CreationTime,omitempty"`
	CreationToken                string                 `json:"CreationToken,omitempty"`
	Encrypted                    bool                   `json:"Encrypted,omitempty"`
	FileSystemArn                string                 `json:"FileSystemArn"`
	FileSystemID                 string                 `json:"FileSystemId"`
	LifeCycleState               *DuploStringValue      `json:"LifeCycleState,omitempty"`
	Name                         string                 `json:"Name"`
	NumberOfMountTargets         int                    `json:"NumberOfMountTargets,omitempty"`
	OwnerID                      string                 `json:"OwnerId,omitempty"`
	PerformanceMode              *DuploStringValue      `json:"PerformanceMode,omitempty"`
	ProvisionedThroughputInMibps float64                `json:"ProvisionedThroughputInMibps,omitempty"`
	SizeInBytes                  *EFSSizeInBytes        `json:"SizeInBytes,omitempty"`
	Tags                         *[]DuploKeyStringValue `json:"Tags,omitempty"`
	ThroughputMode               *DuploStringValue      `json:"ThroughputMode,omitempty"`
	MountTarget                  *[]MountTarget         `json:"-"`
}

type MountTarget struct {
	IP               string         `json:"IpAddress"`
	AvailabilityZone string         `json:"AvailabilityZoneName"`
	SubnetId         string         `json:"SubnetId"`
	LifeCycleState   LifeCycleState `json:"LifeCycleState"`
	MountTargetId    string         `json:"MountTargetId"`
}

/*************************************************
 * API CALLS to duplo
 */

func (c *Client) DuploEFSCreate(tenantID string, rq *DuploEFSCreateReq) (*DuploEFSGetResp, ClientError) {
	rp := DuploEFSGetResp{}
	err := c.postAPI(
		fmt.Sprintf("DuploEFSCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploEFSUpdate(tenantID string, rq *DuploEFSUpdateReq) (*DuploEFSGetResp, ClientError) {
	rp := DuploEFSGetResp{}
	err := c.putAPI(
		fmt.Sprintf("DuploEFSUpdate(%s, %s)", tenantID, rq.FileSystemId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploEFSGet(tenantID string, efsId string) (*DuploEFSGetResp, ClientError) {
	rp := DuploEFSGetResp{}
	err := c.getAPI(
		fmt.Sprintf("DuploEFSGet(%s, %s)", tenantID, efsId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs/%s", tenantID, efsId),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploEFSDelete(tenantID string, efsId string) (*DuploEFSGetResp, ClientError) {
	rp := DuploEFSGetResp{}
	err := c.deleteAPI(
		fmt.Sprintf("DuploEFSDelete(%s, %s)", tenantID, efsId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs/%s", tenantID, efsId),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAWsMountTargetGet(tenantID string, efsId string) (*[]MountTarget, ClientError) {
	rp := []MountTarget{}
	err := c.getAPI(
		fmt.Sprintf("DuploEFSGet(%s, %s)", tenantID, efsId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs_mount_targets/%s", tenantID, efsId),
		&rp,
	)
	return &rp, err
}
