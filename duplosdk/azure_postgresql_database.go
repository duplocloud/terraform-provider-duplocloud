package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzurePostgresqlRequest struct {
	BackupRetentionDays int    `json:"BackupRetentionDays"`
	GeoRedundantBackup  string `json:"GeoRedundantBackup"`
	Version             string `json:"Version"`
	Name                string `json:"Name"`
	AdminUsername       string `json:"AdminUsername"`
	AdminPassword       string `json:"AdminPassword"`
	StorageMB           int    `json:"StorageMB"`
	Size                string `json:"Size"`
}

type DuploAzurePostgresqlServer struct {
	Sku struct {
		Name     string `json:"name"`
		Tier     string `json:"tier"`
		Capacity int    `json:"capacity"`
		Family   string `json:"family"`
	} `json:"sku"`
	PropertiesAdministratorLogin       string    `json:"properties.administratorLogin"`
	PropertiesVersion                  string    `json:"properties.version"`
	PropertiesSslEnforcement           string    `json:"properties.sslEnforcement"`
	PropertiesMinimalTLSVersion        string    `json:"properties.minimalTlsVersion"`
	PropertiesByokEnforcement          string    `json:"properties.byokEnforcement"`
	PropertiesInfrastructureEncryption string    `json:"properties.infrastructureEncryption"`
	PropertiesUserVisibleState         string    `json:"properties.userVisibleState"`
	PropertiesFullyQualifiedDomainName string    `json:"properties.fullyQualifiedDomainName"`
	PropertiesEarliestRestoreDate      time.Time `json:"properties.earliestRestoreDate"`
	PropertiesStorageProfile           struct {
		BackupRetentionDays int    `json:"backupRetentionDays"`
		GeoRedundantBackup  string `json:"geoRedundantBackup"`
		StorageMB           int    `json:"storageMB"`
		StorageAutogrow     string `json:"storageAutogrow"`
	} `json:"properties.storageProfile"`
	PropertiesReplicationRole            string                 `json:"properties.replicationRole"`
	PropertiesMasterServerID             string                 `json:"properties.masterServerId"`
	PropertiesPublicNetworkAccess        string                 `json:"properties.publicNetworkAccess"`
	PropertiesPrivateEndpointConnections []interface{}          `json:"properties.privateEndpointConnections"`
	Tags                                 map[string]interface{} `json:"tags"`
	Location                             string                 `json:"location"`
	ID                                   string                 `json:"id"`
	Name                                 string                 `json:"name"`
	Type                                 string                 `json:"type"`
}

type DuploAzurePostgresqlDatabase struct {
	PropertiesCharset   string `json:"properties.charset"`
	PropertiesCollation string `json:"properties.collation"`
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
}

func (c *Client) PostgresqlDatabaseCreate(tenantID string, rq *DuploAzurePostgresqlRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("PostgresqlDatabaseCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreatePgSql", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) PostgresqlServerGet(tenantID string, name string) (*DuploAzurePostgresqlServer, ClientError) {
	list, err := c.PostgresqlServerList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, server := range *list {
			if server.Name == name {
				return &server, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) PostgresqlServerList(tenantID string) (*[]DuploAzurePostgresqlServer, ClientError) {
	rp := []DuploAzurePostgresqlServer{}
	err := c.getAPI(
		fmt.Sprintf("PostgresqlServerList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListPgSqls", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) PostgresqlDatabaseList(tenantID, name string) (*[]DuploAzurePostgresqlDatabase, ClientError) {
	rp := []DuploAzurePostgresqlDatabase{}
	err := c.getAPI(
		fmt.Sprintf("PostgresqlDatabaseList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListSqlDatabases/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) PostgresqlServerExists(tenantID, name string) (bool, ClientError) {
	list, err := c.PostgresqlServerList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, server := range *list {
			if server.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) PostgresqlDatabaseDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("PostgresqlDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeletePgSql/%s", tenantID, name),
		nil,
		nil,
	)
}
