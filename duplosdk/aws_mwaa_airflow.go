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
	Name string `json:"name,omitempty"`
}
type DuploMwaaAirflowDetail struct {
	Id                           string                  `json:"Name,omitempty"`
	Name                         string                  `json:"name,omitempty"`
	CreatedAt                    string                  `json:"CreatedAt,omitempty"`
	AirflowVersion               string                  `json:"AirflowVersion,omitempty"`
	Arn                          string                  `json:"Arn,omitempty"`
	EnvironmentClass             string                  `json:"EnvironmentClass,omitempty"`
	ExecutionRoleArn             string                  `json:"ExecutionRoleArn,omitempty"`
	KmsKey                       string                  `json:"KmsKey,omitempty"`
	SourceBucketArn              string                  `json:"SourceBucketArn,omitempty"`
	DagS3Path                    string                  `json:"DagS3Path,omitempty"`
	MaxWorkers                   int                     `json:"MaxWorkers,omitempty"`
	MinWorkers                   int                     `json:"MinWorkers,omitempty"`
	Schedulers                   int                     `json:"Schedulers,omitempty"`
	ServiceRoleArn               string                  `json:"ServiceRoleArn,omitempty"`
	WebserverUrl                 string                  `json:"WebserverUrl,omitempty"`
	WeeklyMaintenanceWindowStart string                  `json:"WeeklyMaintenanceWindowStart,omitempty"`
	WebserverAccessMode          *map[string]interface{} `json:"WebserverAccessMode,omitempty"`
	Status                       *map[string]interface{} `json:"Status,omitempty"`
	LastUpdate                   *map[string]interface{} `json:"LastUpdate,omitempty"`
	Tags                         *map[string]interface{} `json:"Tags,omitempty"`
	LoggingConfiguration         *map[string]interface{} `json:"LoggingConfiguration,omitempty"`
	AirflowConfigurationOptions  *map[string]interface{} `json:"AirflowConfigurationOptions,omitempty"`
}

type DuploMwaaAirflowCreateRequest struct {
	Name                         string                  `json:"name,omitempty"`
	AirflowVersion               string                  `json:"AirflowVersion,omitempty"`
	EnvironmentClass             string                  `json:"EnvironmentClass,omitempty"`
	KmsKey                       string                  `json:"KmsKey,omitempty"`
	SourceBucketArn              string                  `json:"SourceBucketArn,omitempty"`
	DagS3Path                    string                  `json:"DagS3Path,omitempty"`
	MaxWorkers                   int                     `json:"MaxWorkers,omitempty"`
	MinWorkers                   int                     `json:"MinWorkers,omitempty"`
	Schedulers                   int                     `json:"Schedulers,omitempty"`
	WebserverAccessMode          string                  `json:"WebserverAccessMode,omitempty"`
	WeeklyMaintenanceWindowStart string                  `json:"WeeklyMaintenanceWindowStart,omitempty"`
	LoggingConfiguration         *map[string]interface{} `json:"LoggingConfiguration,omitempty"`
	AirflowConfigurationOptions  *map[string]interface{} `json:"AirflowConfigurationOptions,omitempty"`
}
type DuploMwaaAirflowUpdateRequest struct {
	Name string `json:"name,omitempty"`
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

func (c *Client) MwaaAirflowExists(tenantID, name string) (bool, ClientError) {
	list, err := c.MwaaAirflowList(tenantID)
	if err != nil {
		return false, err
	}
	if list != nil {
		for _, duploMwaaAirflowSummary := range *list {
			if duploMwaaAirflowSummary.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) MwaaAirflowDelete(tenantID string, id string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("MwaaAirflowDelete(%s, %s)", tenantID, id),
		fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow/%s", tenantID, id),
		nil,
	)
}

func (c *Client) MwaaAirflowUpdate(tenantID string, id string, rq *DuploMwaaAirflowUpdateRequest) (string, ClientError) {
	verb := "POST"
	msg := fmt.Sprintf("MwaaAirflowScale(%s, %s)", tenantID, id)
	api := fmt.Sprintf("v3/subscriptions/%s/aws/mwaaairflow", tenantID)

	rp := ""
	err := c.doAPIWithRequestBody(verb, msg, api, &rq, &rp)
	if err != nil {
		return rp, err
	}
	return rp, nil
}
