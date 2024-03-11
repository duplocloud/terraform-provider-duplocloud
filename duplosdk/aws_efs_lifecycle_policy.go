package duplosdk

import (
	"fmt"
)

type LifecyclePolicy struct {
	TransitionToArchive             *DuploStringValue `json:"TransitionToArchive"`
	TransitionToIA                  *DuploStringValue `json:"TransitionToIA"`
	TransitionToPrimaryStorageClass *DuploStringValue `json:"TransitionToPrimaryStorageClass"`
}

type PutLifecycleConfigurationInput struct {
	FileSystemId      string             `json:"FileSystemId"`
	LifecyclePolicies []*LifecyclePolicy `json:"LifecyclePolicies"`
}

/*************************************************
 * API CALLS to duplo
 */

func (c *Client) DuploAwsLifecyclePolicyUpdate(tenantID string, rq *PutLifecycleConfigurationInput) ClientError {
	err := c.putAPI(
		fmt.Sprintf("DuploEFSUpdate(%s, %s)", tenantID, rq.FileSystemId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs/%s", tenantID, rq.FileSystemId),
		&rq,
		nil,
	)
	return err
}

func (c *Client) DuploAWsLifecyclePolicyGet(tenantID string, efsId string) (*[]LifecyclePolicy, ClientError) {
	rp := []LifecyclePolicy{}
	err := c.getAPI(
		fmt.Sprintf("DuploEFSGet(%s, %s)", tenantID, efsId),
		fmt.Sprintf("v3/subscriptions/%s/aws/efs/%s/lifecyclePolicies", tenantID, efsId),
		&rp,
	)
	return &rp, err
}
