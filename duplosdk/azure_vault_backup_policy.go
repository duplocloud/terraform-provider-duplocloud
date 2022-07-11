package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureVaultBackupHourlySchedule struct {
	Interval                int        `json:"interval,omitempty"`
	ScheduleWindowDuration  int        `json:"scheduleWindowDuration,omitempty"`
	ScheduleWindowStartTime *time.Time `json:"scheduleWindowStartTime,omitempty"`
}

type DuploAzureVaultBackupSchedulePolicy struct {
	SchedulePolicyType      string                               `json:"schedulePolicyType,omitempty"`
	ScheduleRunFrequency    string                               `json:"scheduleRunFrequency,omitempty"`
	ScheduleRunTimes        *[]time.Time                         `json:"scheduleRunTimes,omitempty"`
	ScheduleWeeklyFrequency int                                  `json:"scheduleWeeklyFrequency,omitempty"`
	ScheduleRunDays         *[]string                            `json:"scheduleRunDays,omitempty"`
	HourlySchedule          *DuploAzureVaultBackupHourlySchedule `json:"hourlySchedule,omitempty"`
	DailySchedule           *DuploAzureVaultBackupDailySchedule  `json:"dailySchedule,omitempty"`
	WeeklySchedule          *DuploAzureVaultBackupWeeklySchedule `json:"weeklySchedule,omitempty"`
}

type DuploAzureVaultBackupRetentionPolicy struct {
	RetentionPolicyType string                                `json:"retentionPolicyType,omitempty"`
	DailySchedule       *DuploAzureVaultBackupDailySchedule   `json:"dailySchedule,omitempty"`
	MonthlySchedule     *DuploAzureVaultBackupMonthlySchedule `json:"monthlySchedule,omitempty"`
	WeeklySchedule      *DuploAzureVaultBackupWeeklySchedule  `json:"weeklySchedule,omitempty"`
	YearlySchedule      *DuploAzureVaultBackupYearlySchedule  `json:"yearlySchedule,omitempty"`
}

type DuploAzureVaultBackupRetentionDuration struct {
	Count        int    `json:"count,omitempty"`
	DurationType string `json:"durationType,omitempty"`
}

type DuploAzureVaultBackupDailySchedule struct {
	RetentionDuration *DuploAzureVaultBackupRetentionDuration `json:"retentionDuration,omitempty"`
	RetentionTimes    *[]time.Time                            `json:"retentionTimes,omitempty"`
	ScheduleRunTimes  *[]time.Time                            `json:"scheduleRunTimes,omitempty"`
}

type DuploAzureVaultBackupRetentionScheduleWeekly struct {
	DaysOfTheWeek   *[]string `json:"daysOfTheWeek,omitempty"`
	WeeksOfTheMonth *[]string `json:"weeksOfTheMonth,omitempty"`
}

type DuploAzureVaultBackupMonthlySchedule struct {
	RetentionDuration           *DuploAzureVaultBackupRetentionDuration       `json:"retentionDuration,omitempty"`
	RetentionScheduleFormatType string                                        `json:"retentionScheduleFormatType,omitempty"`
	RetentionScheduleWeekly     *DuploAzureVaultBackupRetentionScheduleWeekly `json:"retentionScheduleWeekly,omitempty"`
	RetentionTimes              *[]time.Time                                  `json:"retentionTimes,omitempty"`
}

type DuploAzureVaultBackupWeeklySchedule struct {
	DaysOfTheWeek     *[]string                               `json:"daysOfTheWeek,omitempty"`
	RetentionDuration *DuploAzureVaultBackupRetentionDuration `json:"retentionDuration,omitempty"`
	RetentionTimes    *[]time.Time                            `json:"retentionTimes,omitempty"`
	ScheduleRunDays   *[]string                               `json:"scheduleRunDays,omitempty"`
	ScheduleRunTimes  *[]time.Time                            `json:"scheduleRunTimes,omitempty"`
}
type DuploAzureVaultBackupYearlySchedule struct {
	MonthsOfYear                *[]string                                     `json:"monthsOfYear,omitempty"`
	RetentionDuration           *DuploAzureVaultBackupRetentionDuration       `json:"retentionDuration,omitempty"`
	RetentionTimes              *[]time.Time                                  `json:"retentionTimes,omitempty"`
	RetentionScheduleWeekly     *DuploAzureVaultBackupRetentionScheduleWeekly `json:"retentionScheduleWeekly,omitempty"`
	RetentionScheduleFormatType string                                        `json:"retentionScheduleFormatType,omitempty"`
}

type DuploAzureVaultBackupPolicyProperties struct {
	SchedulePolicy                *DuploAzureVaultBackupSchedulePolicy  `json:"schedulePolicy"`
	RetentionPolicy               *DuploAzureVaultBackupRetentionPolicy `json:"retentionPolicy"`
	InstantRpRetentionRangeInDays int                                   `json:"instantRpRetentionRangeInDays,omitempty"`
	TimeZone                      string                                `json:"timeZone,omitempty"`
	BackupManagementType          string                                `json:"backupManagementType,omitempty"`
	PolicyType                    string                                `json:"policyType,omitempty"`
	ProtectedItemsCount           int                                   `json:"protectedItemsCount,omitempty"`
}

type DuploAzureVaultBackupPolicy struct {
	Properties *DuploAzureVaultBackupPolicyProperties `json:"properties,omitempty"`
	ID         string                                 `json:"id,omitempty"`
	Name       string                                 `json:"name"`
	Type       string                                 `json:"type,omitempty"`
}

func (c *Client) VaultBackupPolicyCreate(infraName string, rq *DuploAzureVaultBackupPolicy) ClientError {
	return c.postAPI(
		fmt.Sprintf("VaultBackupPolicyCreate(%s, %s)", infraName, rq.Name),
		fmt.Sprintf("adminproxy/CreateAzureVaultBackupPolicy/%s", infraName),
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
		fmt.Sprintf("adminproxy/GetAzureVaultBackupPolicies/%s", infraName),
		&rp,
	)
	return &rp, err
}

func (c *Client) VaultBackupPolicyDelete(infraName string, rq *DuploAzureVaultBackupPolicy) ClientError {
	return c.postAPI(
		fmt.Sprintf("VaultBackupPolicyDelete(%s, %s)", infraName, rq.Name),
		fmt.Sprintf("adminproxy/DeleteAzureVaultBackupPolicy/%s", infraName),
		&rq,
		nil,
	)
}
