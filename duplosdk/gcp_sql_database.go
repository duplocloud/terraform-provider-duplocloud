package duplosdk

import (
	"fmt"
)

const (
	DuploGCPDatabaseInstanceResourceType = 27
)

type DuploGCPSqlDBInstance struct {
	Name            string            `json:"Name"`
	DatabaseVersion string            `json:"DatabaseVersion"`
	Tier            string            `json:"Tier"`
	DataDiskSizeGb  int               `json:"DataDiskSizeGb"`
	Status          string            `json:"Status,omitempty"`
	ResourceType    int               `json:"ResourceType,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	SelfLink        string            `json:"SelfLink,omitempty"`
	RootPassword    string            `json:"RootPassword,omitempty"`
	IPAddress       []string          `json:"IpAddress,omitempty"`
	ConnectionName  string            `json:"ConnectionName,omitempty"`
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

func (c *Client) GCPSqlDBInstanceDelete(tenantID string, name string, backup bool) ClientError {
	uri := fmt.Sprintf("v3/subscriptions/%s/google/database/%s", tenantID, name)
	if backup {
		uri = fmt.Sprintf("v3/subscriptions/%s/google/database/%s?needBackup=%t", tenantID, name, backup)
	}

	return c.deleteAPI(
		fmt.Sprintf("GCPSqlDBInstanceDelete(%s, %s)", tenantID, name),
		uri,
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

func (c *Client) GCPSqlDBInstanceVersionsList(tenantID string) ([]string, ClientError) {
	versions := []string{}
	err := c.getAPI(
		fmt.Sprintf("GCPSqlDBInstanceVersionsList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/databaseVersions", tenantID),
		&versions,
	)
	return versions, err
}

func (c *Client) GCPSqlDBInstanceVersionsRequiringPasswordList(tenantID string) ([]string, ClientError) {
	versions := []string{}
	err := c.getAPI(
		fmt.Sprintf("GCPSqlDBInstanceVersionsRequiringPasswordList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/databaseVersionsRequiringPassword", tenantID),
		&versions,
	)
	return versions, err
}
