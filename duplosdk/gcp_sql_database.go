package duplosdk

import (
	"fmt"
)

const (
	DuploGCPDatabaseInstanceResourceType = 27
)

const (
	MYSQL_5_6 = 0
	MYSQL_5_7 = 1
	MYSQL_8_0 = 2

	POSTGRES_10  = 3
	POSTGRES_11  = 4
	POSTGRES_12  = 5
	POSTGRES_13  = 6
	POSTGRES_14  = 7
	POSTGRES_15  = 8
	POSTGRES_9_6 = 9

	SQLSERVER_2017_STANDARD   = 10
	SQLSERVER_2017_ENTERPRISE = 11
	SQLSERVER_2017_EXPRESS    = 12
	SQLSERVER_2017_WEB        = 13

	SQLSERVER_2019_STANDARD   = 14
	SQLSERVER_2019_ENTERPRISE = 15
	SQLSERVER_2019_EXPRESS    = 16
	SQLSERVER_2019_WEB        = 17

	SQLSERVER_2022_STANDARD   = 18
	SQLSERVER_2022_ENTERPRISE = 19
	SQLSERVER_2022_EXPRESS    = 20
	SQLSERVER_2022_WEB        = 21
)

var DuploGCPSqlDBInstanceVersionMappings = map[string]int{
	"MYSQL_5_6": MYSQL_5_6,
	"MYSQL_5_7": MYSQL_5_7,
	"MYSQL_8_0": MYSQL_8_0,

	"POSTGRES_10":  POSTGRES_10,
	"POSTGRES_11":  POSTGRES_11,
	"POSTGRES_12":  POSTGRES_12,
	"POSTGRES_13":  POSTGRES_13,
	"POSTGRES_14":  POSTGRES_14,
	"POSTGRES_15":  POSTGRES_15,
	"POSTGRES_9_6": POSTGRES_9_6,

	"SQLSERVER_2017_STANDARD":   SQLSERVER_2017_STANDARD,
	"SQLSERVER_2017_ENTERPRISE": SQLSERVER_2017_ENTERPRISE,
	"SQLSERVER_2017_EXPRESS":    SQLSERVER_2017_EXPRESS,
	"SQLSERVER_2017_WEB":        SQLSERVER_2017_WEB,

	"SQLSERVER_2019_STANDARD":   SQLSERVER_2019_STANDARD,
	"SQLSERVER_2019_ENTERPRISE": SQLSERVER_2019_ENTERPRISE,
	"SQLSERVER_2019_EXPRESS":    SQLSERVER_2019_EXPRESS,
	"SQLSERVER_2019_WEB":        SQLSERVER_2019_WEB,

	"SQLSERVER_2022_STANDARD":   SQLSERVER_2022_STANDARD,
	"SQLSERVER_2022_ENTERPRISE": SQLSERVER_2022_ENTERPRISE,
	"SQLSERVER_2022_EXPRESS":    SQLSERVER_2022_EXPRESS,
	"SQLSERVER_2022_WEB":        SQLSERVER_2022_WEB,
}

type DuploGCPSqlDBInstance struct {
	Name            string            `json:"Name"`
	DatabaseVersion int               `json:"DatabaseVersion"`
	Tier            string            `json:"Tier"`
	DataDiskSizeGb  int               `json:"DataDiskSizeGb"`
	Status          string            `json:"Status,omitempty"`
	ResourceType    int               `json:"ResourceType,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	SelfLink        string            `json:"SelfLink,omitempty"`
	RootPassword    string            `json:"RootPassword,omitempty"`
}

func (c *Client) GCPSqlDBInstanceCreate(tenantID string, rq *DuploGCPSqlDBInstance) (*DuploGCPSqlDBInstance, ClientError) {
	resp := DuploGCPSqlDBInstance{}
	clientErr := c.postAPI(
		fmt.Sprintf("GCPSqlDBInstanceCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/database", tenantID),
		&rq,
		&resp,
	)
	return &resp, clientErr
}

func (c *Client) GCPSqlDBInstanceUpdate(tenantID string, rq *DuploGCPSqlDBInstance) (*DuploGCPSqlDBInstance, ClientError) {
	resp := DuploGCPSqlDBInstance{}
	clientErr := c.putAPI(
		fmt.Sprintf("GCPSqlDBInstanceUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/database/%s", tenantID, rq.Name),
		&rq,
		&resp,
	)
	return &resp, clientErr
}

func (c *Client) GCPSqlDBInstanceGet(tenantID string, name string) (*DuploGCPSqlDBInstance, ClientError) {
	rp := DuploGCPSqlDBInstance{}
	err := c.getAPI(
		fmt.Sprintf("GCPSqlDBInstanceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/database/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) GCPSqlDBInstanceDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GCPSqlDBInstanceDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/database/%s", tenantID, name),
		nil,
	)
}

func (c *Client) GCPSqlDBInstanceList(tenantID string) (*[]DuploGCPSqlDBInstance, ClientError) {
	rp := []DuploGCPSqlDBInstance{}
	err := c.getAPI(
		fmt.Sprintf("GCPSqlDBInstanceList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/database", tenantID),
		&rp,
	)
	return &rp, err
}
