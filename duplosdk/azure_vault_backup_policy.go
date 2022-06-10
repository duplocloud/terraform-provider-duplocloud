package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureVaultBackupPolicyPostReq struct {
	Name string `json:"name"`
}

type DuploAzureVaultBackupPolicy struct {
	Properties struct {
		InstantRPDetails struct {
		} `json:"instantRPDetails"`
		SchedulePolicy struct {
			ScheduleRunFrequency    string      `json:"scheduleRunFrequency"`
			ScheduleRunTimes        []time.Time `json:"scheduleRunTimes"`
			ScheduleWeeklyFrequency int         `json:"scheduleWeeklyFrequency"`
		} `json:"schedulePolicy"`
		RetentionPolicy struct {
			DailySchedule struct {
				RetentionTimes    []time.Time `json:"retentionTimes"`
				RetentionDuration struct {
					Count        int    `json:"count"`
					DurationType string `json:"durationType"`
				} `json:"retentionDuration"`
			} `json:"dailySchedule"`
		} `json:"retentionPolicy"`
		InstantRpRetentionRangeInDays int    `json:"instantRpRetentionRangeInDays"`
		TimeZone                      string `json:"timeZone"`
		ProtectedItemsCount           int    `json:"protectedItemsCount"`
	} `json:"properties"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *Client) VaultBackupPolicyCreate(infraName string, rq *DuploAzureVaultBackupPolicyPostReq) ClientError {
	return c.postAPI(
		fmt.Sprintf("VaultBackupPolicyCreate(%s, %s)", infraName, rq.Name),
		fmt.Sprintf("admin/CreateAzureVaultBackupPolicy/%s", infraName),
		&rq,
		nil,
	)
}

func (c *Client) VaultBackupPolicyGet(infraName, policyName string) (*DuploAzureVaultBackupPolicy, ClientError) {

	list, err := c.VaultBackupPolicyList(infraName)

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, policy := range *list {
			if policy.Name == policyName {
				return &policy, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) VaultBackupPolicyList(infraName string) (*[]DuploAzureVaultBackupPolicy, ClientError) {
	rp := []DuploAzureVaultBackupPolicy{}
	err := c.getAPI(
		fmt.Sprintf("VaultBackupPolicyList(%s)", infraName),
		fmt.Sprintf("admin/GetAzureVaultBackupPolicies/%s", infraName),
		&rp,
	)
	return &rp, err
}

func (c *Client) VaultBackupPolicyDelete(infraName string, rq *DuploAzureVaultBackupPolicyPostReq) ClientError {
	return c.postAPI(
		fmt.Sprintf("VaultBackupPolicyDelete(%s, %s)", infraName, rq.Name),
		fmt.Sprintf("admin/DeleteVaultBackupPolicy/%s", infraName),
		&rq,
		nil,
	)
}
