package duplosdk

import (
	"fmt"
	"log"
	"time"
)

type DuploGcpInfraMaintenanceWindow struct {
	DailyMaintenanceStartTime *time.Time   `json:"DailyMaintenanceStartTime,omitempty"`
	Exclusions                *[]Exclusion `json:"Exclusions,omitempty"`
	RecurringWindow           *Recurring   `json:"RecurringWindow,omitempty"`
}

type Exclusion struct {
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	Scope     string    `json:"Scope"`
}

type Recurring struct {
	StartTime  time.Time `json:"StartTime"`
	EndTime    time.Time `json:"EndTime"`
	Recurrence string    `json:"Recurrence"`
}

func (c *Client) CreateGCPInfraMaintenanceWindow(infraName string, rq *DuploGcpInfraMaintenanceWindow) ClientError {
	log.Printf("[TRACE] GCP Infra Maintenance Window request \n\n ******%+v\n*******", rq)
	err := c.postAPI(
		fmt.Sprintf("CreateGCPInfraMaintenanceWindow(%s)", infraName),
		fmt.Sprintf("v3/google/cluster/%s/maintenance", infraName),
		&rq,
		nil,
	)
	return err
}

func (c *Client) GetGCPInfraMaintenanceWindow(infraName string) (*DuploGcpInfraMaintenanceWindow, ClientError) {
	rp := DuploGcpInfraMaintenanceWindow{}
	err := c.getAPI(
		fmt.Sprintf("GetGCPInfraMaintenanceWindow(%s)", infraName),
		fmt.Sprintf("v3/google/cluster/%s/maintenance", infraName),
		&rp,
	)

	return &rp, err
}
