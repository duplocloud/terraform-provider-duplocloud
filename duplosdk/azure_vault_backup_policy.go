package duplosdk

import (
	"fmt"
	"time"
)

type DayOfWeek int

func DayOfWeekString(e int) string {
	switch e {
	case 0:
		return "Sunday"
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

func DayOfWeekIndex(e string) int {
	switch e {
	case "Sunday":
		return 0
	case "Monday":
		return 1
	case "Tuesday":
		return 2
	case "Wednesday":
		return 3
	case "Thursday":
		return 4
	case "Friday":
		return 5
	case "Saturday":
		return 6
	default:
		return 0
	}
}

func WeekOfMonthString(e int) string {
	switch e {
	case 0:
		return "First"
	case 1:
		return "Second"
	case 2:
		return "Third"
	case 3:
		return "Fourth"
	case 4:
		return "Last"
	case 5:
		return "Invalid"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

func WeekOfMonthIndex(e string) int {
	switch e {
	case "First":
		return 0
	case "Second":
		return 1
	case "Third":
		return 2
	case "Fourth":
		return 3
	case "Last":
		return 4
	case "Invalid":
		return 5
	default:
		return 5
	}
}

func MonthOfYearString(e int) string {
	switch e {
	case 0:
		return "Invalid"
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

func MonthOfYearIndex(e string) int {
	switch e {
	case "Invalid":
		return 0
	case "January":
		return 1
	case "February":
		return 2
	case "March":
		return 3
	case "April":
		return 4
	case "May":
		return 5
	case "June":
		return 6
	case "July":
		return 7
	case "August":
		return 8
	case "September":
		return 9
	case "October":
		return 10
	case "November":
		return 11
	case "December":
		return 12
	default:
		return 0
	}
}

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
	ScheduleRunDays         *[]int                               `json:"scheduleRunDays,omitempty"`
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
	ScheduleRunDays   *[]int                                  `json:"scheduleRunDays,omitempty"`
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
	Properties *DuploAzureVaultBackupPolicyProperties `json:"properties"`
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
		fmt.Sprintf("adminproxy/DeleteVaultBackupPolicy/%s", infraName),
		&rq,
		nil,
	)
}
