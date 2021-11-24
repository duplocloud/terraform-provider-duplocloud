package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureMySqlRequest struct {
	BackupRetentionDays int    `json:"BackupRetentionDays"`
	GeoRedundantBackup  string `json:"GeoRedundantBackup"`
	Version             string `json:"Version"`
	Name                string `json:"Name"`
	AdminUsername       string `json:"AdminUsername"`
	AdminPassword       string `json:"AdminPassword"`
	StorageMB           int    `json:"StorageMB"`
	Size                string `json:"Size"`
}

type DuploAzureMySqlServer struct {
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

type DuploAzureMySqlDatabase struct {
	PropertiesCharset   string `json:"properties.charset"`
	PropertiesCollation string `json:"properties.collation"`
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
}

func (c *Client) MySqlDatabaseCreate(tenantID string, rq *DuploAzureMySqlRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("MySqlDatabaseCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateMySql", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) MySqlServerGet(tenantID string, name string) (*DuploAzureMySqlServer, ClientError) {
	list, err := c.MySqlServerList(tenantID)
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

func (c *Client) MySqlServerList(tenantID string) (*[]DuploAzureMySqlServer, ClientError) {
	rp := []DuploAzureMySqlServer{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListMySqls", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) MySqlDatabaseList(tenantID, name string) (*[]DuploAzureMySqlDatabase, ClientError) {
	rp := []DuploAzureMySqlDatabase{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListMySqlDatabases/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) MySqlServerExists(tenantID, name string) (bool, ClientError) {
	list, err := c.MySqlServerList(tenantID)
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

func (c *Client) MySqlDatabaseDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("StorageAccountDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteMySql/%s", tenantID, name),
		nil,
		nil,
	)
}
