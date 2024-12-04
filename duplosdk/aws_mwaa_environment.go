package duplosdk

import (
	"fmt"
)

type WebserverAccessMode string

const (
	WebserverAccessModePublic  WebserverAccessMode = "PUBLIC_ONLY"
	WebserverAccessModePrivate WebserverAccessMode = "PRIVATE_ONLY"
)

type DuploMwaaAirflowSummary struct {
	Name string `json:"Name,omitempty"`
}

type DuploMwaaLastUpdateError struct {
	ErrorCode    string `json:"ErrorCode,omitempty"`
	ErrorMessage string `json:"ErrorMessage,omitempty"`
}

type DuploMwaaLastUpdate struct {
	CreatedAt string                    `json:"CreatedAt,omitempty"`
	Status    *DuploStringValue         `json:"Status,omitempty"`
	Source    string                    `json:"Source,omitempty"`
	Error     *DuploMwaaLastUpdateError `json:"Error,omitempty"`
}

type DuploLoggingConfigurationInput struct {
	Enabled  bool              `json:"Enabled"`
	LogLevel *DuploStringValue `json:"LogLevel,omitempty"`
}

type DuploMwaaLoggingConfiguration struct {
	DagProcessingLogs *DuploLoggingConfigurationInput `json:"DagProcessingLogs,omitempty"`
	SchedulerLogs     *DuploLoggingConfigurationInput `json:"SchedulerLogs,omitempty"`
	TaskLogs          *DuploLoggingConfigurationInput `json:"TaskLogs,omitempty"`
	WebserverLogs     *DuploLoggingConfigurationInput `json:"WebserverLogs,omitempty"`
	WorkerLogs        *DuploLoggingConfigurationInput `json:"WorkerLogs,omitempty"`
}

type DuploMwaaAirflowDetail struct {
	Name                         string                         `json:"Name,omitempty"`
	CreatedAt                    string                         `json:"CreatedAt,omitempty"`
	AirflowVersion               string                         `json:"AirflowVersion,omitempty"`
	Arn                          string                         `json:"Arn,omitempty"`
	EnvironmentClass             string                         `json:"EnvironmentClass,omitempty"`
	ExecutionRoleArn             string                         `json:"ExecutionRoleArn,omitempty"`
	KmsKey                       string                         `json:"KmsKey,omitempty"`
	SourceBucketArn              string                         `json:"SourceBucketArn,omitempty"`
	DagS3Path                    string                         `json:"DagS3Path,omitempty"`
	MaxWorkers                   int                            `json:"MaxWorkers,omitempty"`
	MinWorkers                   int                            `json:"MinWorkers,omitempty"`
	Schedulers                   int                            `json:"Schedulers,omitempty"`
	ServiceRoleArn               string                         `json:"ServiceRoleArn,omitempty"`
	WebserverUrl                 string                         `json:"WebserverUrl,omitempty"`
	WeeklyMaintenanceWindowStart string                         `json:"WeeklyMaintenanceWindowStart,omitempty"`
	WebserverAccessMode          *DuploStringValue              `json:"WebserverAccessMode,omitempty"`
	Status                       *DuploStringValue              `json:"Status,omitempty"`
	LastUpdate                   *DuploMwaaLastUpdate           `json:"LastUpdate,omitempty"`
	Tags                         map[string]interface{}         `json:"Tags,omitempty"`
	LoggingConfiguration         *DuploMwaaLoggingConfiguration `json:"LoggingConfiguration,omitempty"`
	AirflowConfigurationOptions  map[string]string              `json:"AirflowConfigurationOptions,omitempty"`
	PluginsS3ObjectVersion       string                         `json:"PluginsS3ObjectVersion,omitempty"`
	PluginsS3Path                string                         `json:"PluginsS3Path,omitempty"`
	RequirementsS3ObjectVersion  string                         `json:"RequirementsS3ObjectVersion,omitempty"`
	RequirementsS3Path           string                         `json:"RequirementsS3Path,omitempty"`
	StartupScriptS3ObjectVersion string                         `json:"StartupScriptS3ObjectVersion,omitempty"`
	StartupScriptS3Path          string                         `json:"StartupScriptS3Path,omitempty"`
}

type DuploMwaaAirflowCreateRequest struct {
	Name                         string                         `json:"Name"`
	AirflowVersion               string                         `json:"AirflowVersion,omitempty"`
	EnvironmentClass             string                         `json:"EnvironmentClass,omitempty"`
	KmsKey                       string                         `json:"KmsKey,omitempty"`
	SourceBucketArn              string                         `json:"SourceBucketArn,omitempty"`
	DagS3Path                    string                         `json:"DagS3Path,omitempty"`
	ExecutionRoleArn             string                         `json:"ExecutionRoleArn,omitempty"`
	MaxWorkers                   int                            `json:"MaxWorkers,omitempty"`
	MinWorkers                   int                            `json:"MinWorkers,omitempty"`
	Schedulers                   int                            `json:"Schedulers,omitempty"`
	WebserverAccessMode          *DuploStringValue              `json:"WebserverAccessMode,omitempty"`
	WeeklyMaintenanceWindowStart string                         `json:"WeeklyMaintenanceWindowStart,omitempty"`
	LoggingConfiguration         *DuploMwaaLoggingConfiguration `json:"LoggingConfiguration,omitempty"`
	AirflowConfigurationOptions  map[string]string              `json:"AirflowConfigurationOptions,omitempty"`
	PluginsS3ObjectVersion       string                         `json:"PluginsS3ObjectVersion,omitempty"`
	PluginsS3Path                string                         `json:"PluginsS3Path,omitempty"`
	RequirementsS3ObjectVersion  string                         `json:"RequirementsS3ObjectVersion,omitempty"`
	RequirementsS3Path           string                         `json:"RequirementsS3Path,omitempty"`
	StartupScriptS3ObjectVersion string                         `json:"StartupScriptS3ObjectVersion,omitempty"`
	StartupScriptS3Path          string                         `json:"StartupScriptS3Path,omitempty"`
}

type DuploMwaaAirflowCreateResponse struct {
	Name string `json:"Name,omitempty"`
	Arn  string `json:"Arn,omitempty"`
}

func (c *Client) MwaaAirflowCreate(tenantID string, name string, rq *DuploMwaaAirflowCreateRequest) ClientError {
	resp := DuploMwaaAirflowCreateResponse{}
	return c.postAPI(
		fmt.Sprintf("MwaaAirflowCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/mwaaAirflow", tenantID),
		&rq,
		&resp,
	)
}

func (c *Client) MwaaAirflowGet(tenantID string, name string) (*DuploMwaaAirflowSummary, ClientError) {
	list, err := c.MwaaAirflowList(tenantID)
	if err != nil {
		return nil, err
	}
	if list != nil {
		for _, duploMwaaAirflowSummary := range *list {
			if duploMwaaAirflowSummary.Name == name {
				return &duploMwaaAirflowSummary, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) MwaaAirflowDetailsGet(tenantID string, id string) (*DuploMwaaAirflowDetail, ClientError) {
	rp := DuploMwaaAirflowDetail{}
	err := c.getAPI(
		fmt.Sprintf("MwaaAirflowList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow/%s", tenantID, id),
		&rp,
	)
	return &rp, err
}

func (c *Client) MwaaAirflowList(tenantID string) (*[]DuploMwaaAirflowSummary, ClientError) {
	rp := []DuploMwaaAirflowSummary{}
	err := c.getAPI(
		fmt.Sprintf("MwaaAirflowList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) MwaaAirflowDelete(tenantID string, id string) ClientError {
	cloudResource := DuploAwsCloudResource{}
	return c.deleteAPI(
		fmt.Sprintf("MwaaAirflowDelete(%s, %s)", tenantID, id),
		fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow/%s", tenantID, id),
		&cloudResource,
	)
}

func (c *Client) MwaaAirflowUpdate(tenantID string, id string, rq *DuploMwaaAirflowCreateRequest) ClientError {
	verb := "PUT"
	msg := fmt.Sprintf("MwaaAirflowScale(%s, %s)", tenantID, id)
	api := fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow/%s", tenantID, id)
	rp := DuploMwaaAirflowCreateResponse{}
	err := c.doAPIWithRequestBody(verb, msg, api, &rq, &rp)
	if err != nil {
		return err
	}
	return nil
}
