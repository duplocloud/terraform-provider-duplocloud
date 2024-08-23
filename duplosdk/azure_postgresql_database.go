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

type DuploAzurePostgresqlFlexibleRequest struct {
	Name string `json:"name"`
	Sku  struct {
		Tier string `json:"tier"`
		Name string `json:"name"`
	} `json:"sku"`
	BackUp struct {
		RetentionDays      int    `json:"backupRetentionDays"`
		GeoRedundantBackUp string `json:"geoRedundantBackup"`
	} `json:"properties.backup"`
	HighAvailability struct {
		Mode string `json:"mode"`
	} `json:"properties.highAvailability"`
	Storage struct {
		StorageSize int `json:"storageSizeGB"`
	} `json:"properties.storage"`
	Network struct {
		Subnet string `json:"delegatedSubnetResourceId"`
	} `json:"properties.network"`
	RequestMode        string `json:"properties.createMode"`
	AdminUserName      string `json:"properties.administratorLogin"`
	AdminLoginPassword string `json:"properties.administratorLoginPassword"`
	Version            string `json:"properties.version"`
}

type DuploAzurePostgresqlFlexible struct {
	Name string `json:"name"`
	Sku  struct {
		Tier string `json:"tier"`
		Name string `json:"name"`
	} `json:"sku"`
	BackUp struct {
		RetentionDays      int    `json:"backupRetentionDays"`
		GeoRedundantBackUp string `json:"geoRedundantBackup"`
	} `json:"properties.backup"`
	HighAvailability struct {
		Mode string `json:"mode"`
	} `json:"properties.highAvailability"`
	Storage struct {
		StorageSize int `json:"storageSizeGB"`
	} `json:"properties.storage"`

	Subnet string `json:"id"`

	AdminUserName string                 `json:"properties.administratorLogin"`
	Location      string                 `json:"location"`
	Version       string                 `json:"properties.version"`
	State         string                 `json:"properties.state"`
	Tags          map[string]interface{} `json:"tags"`
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

func (c *Client) PostgresqlFlexibleDatabaseCreate(tenantID string, rq *DuploAzurePostgresqlFlexibleRequest) (map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseGet(tenantID, name string) (*DuploAzurePostgresqlFlexible, ClientError) {
	rp := DuploAzurePostgresqlFlexible{}
	err := c.getAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseUpdate(tenantID string, rq *DuploAzurePostgresqlFlexibleRequest) ClientError {
	return c.putAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer/%s", tenantID, rq.Name),
		&rq,
		nil,
	)
}

func (c *Client) PostgresqlFlexibleDatabaseDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer/%s", tenantID, name),
		nil,
	)
}
