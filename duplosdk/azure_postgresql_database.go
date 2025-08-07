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

func (c *Client) PostgresqlFlexibleDatabaseUpdate(tenantID string, rq *DuploAzurePostgresqlFlexibleRequest) (map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer/%s", tenantID, rq.Name),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseDelete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/postgres/flexiServer/%s", tenantID, name),
		nil,
	)
}

//***************** v2 *****************//

type DuploAzurePostgresqlFlexibleV2Request struct {
	Name string `json:"Name"`
	Sku  struct {
		Tier string `json:"Tier"`
		Name string `json:"Name"`
	} `json:"Sku"`
	BackUp struct {
		RetentionDays      int    `json:"BackupRetentionDays"`
		GeoRedundantBackUp string `json:"GeoRedundantBackup"`
	} `json:"Backup"`
	HighAvailability struct {
		Mode string `json:"Mode"`
	} `json:"HighAvailability"`
	Storage struct {
		StorageSize int `json:"StorageSizeInGB"`
	} `json:"Storage"`
	Network struct {
		PublicNetworkAccess string `json:"PublicNetworkAccess"`
		Subnet              struct {
			ResourceId string `json:"resourceId"`
		} `json:"DelegatedSubnetResourceId"`
		PrivateDnsZone struct {
			ResourceId string `json:"resourceId"`
		} `json:"PrivateDnsZoneArmResourceId"`
	} `json:"Network"`
	AdminUserName      string   `json:"AdministratorLogin"`
	AdminLoginPassword string   `json:"AdministratorLoginPassword"`
	Version            string   `json:"Version"`
	MinorVersion       string   `json:"MinorVersion"`
	AuthConfig         struct { //new
		ActiveDirectoryAuth string `json:"ActiveDirectoryAuth"`
		PasswordAuth        string `json:"PasswordAuth"`
	} `json:"AuthConfig"`
	AvailabilityZone string `json:"AvailabilityZone,omitempty"`
}

type DuploAzurePostgresqlFlexibleV2ADConfig struct {
	ADPrincipalName string `json:"PrincipalName"`
	ADTenantId      string `json:"TenantId"`
	ADPrincipalType string `json:"PrincipalType"`
	ObjectId        string `json:"-"`
}

func (c *Client) PostgresqlFlexibleDatabaseUpdateADConfig(tenantID, dbName, objId string, rq *DuploAzurePostgresqlFlexibleV2ADConfig) ClientError {
	var rp interface{}
	err := c.postAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseUpdateADConfig(%s, %s)", tenantID, dbName),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer/%s/authentication/%s", tenantID, dbName, objId),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) PostgresqlFlexibleDatabaseADDelete(tenantID string, name, objectId string) ClientError {
	var rp interface{}
	return c.deleteAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer/%s/authentication/%s", tenantID, name, objectId),
		&rp,
	)
}

func (c *Client) PostgresqlFlexibleDatabaseADGet(tenantID string, name, objectId string) (*DuploAzurePostgresqlFlexibleV2ADConfig, ClientError) {
	rp := DuploAzurePostgresqlFlexibleV2ADConfig{}
	err := c.getAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer/%s/authentication/%s", tenantID, name, objectId),
		&rp,
	)
	return &rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseV2Create(tenantID string, rq *DuploAzurePostgresqlFlexibleV2Request) (map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseV2Create(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseV2Update(tenantID string, rq *DuploAzurePostgresqlFlexibleV2Request) (map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseV2Update(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) PostgresqlFlexibleDatabaseV2Delete(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseV2Delete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer/%s", tenantID, name),
		nil,
	)
}

func (c *Client) PostgresqlFlexibleDatabaseV2Get(tenantID, name string) (*DuploAzurePostgresqlFlexibleV2, ClientError) {
	rp := DuploAzurePostgresqlFlexibleV2{}
	err := c.getAPI(
		fmt.Sprintf("PostgresqlFlexibleDatabaseV2Get(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/postgres/flexiServer/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

type DuploAzurePostgresqlFlexibleV2 struct {
	Name string `json:"Name"`
	Sku  struct {
		Tier string `json:"Tier"`
		Name string `json:"Name"`
	} `json:"Sku"`
	BackUp struct {
		RetentionDays      int    `json:"BackupRetentionDays"`
		GeoRedundantBackUp string `json:"GeoRedundantBackup"`
	} `json:"Backup"`
	HighAvailability struct {
		Mode string `json:"Mode"`
	} `json:"HighAvailability"`
	Storage struct {
		StorageSize int `json:"StorageSizeInGB"`
	} `json:"Storage"`

	Network struct {
		PublicNetworkAccess string `json:"PublicNetworkAccess"`
		Subnet              string `json:"DelegatedSubnetResourceId"`
		PrivateDnsZone      string `json:"PrivateDnsZoneArmResourceId"`
	} `json:"Network"`

	AuthConfig *struct {
		ActiveDirectoryAuth string `json:"ActiveDirectoryAuth"`
		PasswordAuth        string `json:"PasswordAuth"`
		TenantId            string `json:"TenantId"`
	} `json:"AuthConfig"`

	AdminUserName string `json:"AdministratorLogin"`
	Location      struct {
		Name string `json:"Name"`
	} `json:"Location"`

	Version                    string                 `json:"Version"`
	State                      string                 `json:"State"`
	Tags                       map[string]interface{} `json:"tags"`
	AzureResourceId            string                 `json:"Id"`
	FullyQualifiedDomainName   string                 `json:"FullyQualifiedDomainName"`
	PrivateEndpointConnections []string               `json:"PrivateEndpointConnections"`
	AvailabilityZone           string                 `json:"AvailabilityZone"`
}
