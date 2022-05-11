package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureMsSqlRequest struct {
	PropertiesMinimalTLSVersion          string `json:"properties.minimalTlsVersion"`
	PropertiesPublicNetworkAccess        string `json:"properties.publicNetworkAccess"`
	PropertiesVersion                    string `json:"properties.version"`
	Name                                 string `json:"name"`
	PropertiesAdministratorLogin         string `json:"properties.administratorLogin"`
	PropertiesAdministratorLoginPassword string `json:"properties.administratorLoginPassword"`
}

type DuploAzureMsSqlDatabaseRequest struct {
	Name                    string                      `json:"name"`
	Sku                     *DuploAzureMsSqlDatabaseSku `json:"sku"`
	PropertiesCollation     string                      `json:"properties.collation,omitempty"`
	PropertiesElasticPoolId string                      `json:"properties.elasticPoolId,omitempty"`
}

type DuploAzureMsSqlElasticPoolRequest struct {
	Name string                      `json:"name"`
	Sku  *DuploAzureMsSqlDatabaseSku `json:"sku"`
}

type DuploAzureMsSqlDatabaseDeleteRequest struct {
	DbName     string `json:"DbName"`
	ServerName string `json:"ServerName"`
}

type DuploAzureMsSqlEPDeleteRequest struct {
	ElasticPoolName string `json:"ElasticPoolName"`
	ServerName      string `json:"ServerName"`
}

type DuploAzureMsSqlDatabaseSku struct {
	Name     string `json:"name"`
	Tier     string `json:"tier,omitempty"`
	Capacity int    `json:"capacity"`
}

type DuploAzureMsSqlServer struct {
	Kind                                 string                 `json:"kind"`
	PropertiesAdministratorLogin         string                 `json:"properties.administratorLogin"`
	PropertiesVersion                    string                 `json:"properties.version"`
	PropertiesState                      string                 `json:"properties.state"`
	PropertiesFullyQualifiedDomainName   string                 `json:"properties.fullyQualifiedDomainName"`
	PropertiesPrivateEndpointConnections []interface{}          `json:"properties.privateEndpointConnections"`
	PropertiesMinimalTLSVersion          string                 `json:"properties.minimalTlsVersion"`
	PropertiesPublicNetworkAccess        string                 `json:"properties.publicNetworkAccess"`
	Tags                                 map[string]interface{} `json:"tags"`
	Location                             string                 `json:"location"`
	ID                                   string                 `json:"id"`
	Name                                 string                 `json:"name"`
	Type                                 string                 `json:"type"`
}

type DuploAzureMsSqlDatabase struct {
	Kind                                    string    `json:"kind"`
	ManagedBy                               string    `json:"managedBy"`
	PropertiesCollation                     string    `json:"properties.collation"`
	PropertiesMaxSizeBytes                  int64     `json:"properties.maxSizeBytes"`
	PropertiesStatus                        string    `json:"properties.status"`
	PropertiesDatabaseID                    string    `json:"properties.databaseId"`
	PropertiesCreationDate                  time.Time `json:"properties.creationDate"`
	PropertiesCurrentServiceObjectiveName   string    `json:"properties.currentServiceObjectiveName"`
	PropertiesRequestedServiceObjectiveName string    `json:"properties.requestedServiceObjectiveName"`
	PropertiesDefaultSecondaryLocation      string    `json:"properties.defaultSecondaryLocation"`
	PropertiesCatalogCollation              string    `json:"properties.catalogCollation"`
	PropertiesZoneRedundant                 bool      `json:"properties.zoneRedundant"`
	PropertiesReadScale                     string    `json:"properties.readScale"`
	PropertiesReadReplicaCount              int       `json:"properties.readReplicaCount"`
	PropertiesElasticPoolId                 string    `json:"properties.elasticPoolId,omitempty"`

	PropertiesCurrentSku *DuploAzureMsSqlDatabaseSku `json:"properties.currentSku"`
	Tags                 map[string]interface{}      `json:"tags"`
	Location             string                      `json:"location"`
	ID                   string                      `json:"id"`
	Name                 string                      `json:"name"`
	Type                 string                      `json:"type"`
}

type DuploAzureMsSqlElasticPool struct {
	Kind                    string                      `json:"kind"`
	PropertiesMaxSizeBytes  int64                       `json:"properties.maxSizeBytes"`
	PropertiesState         string                      `json:"properties.state"`
	PropertiesCreationDate  time.Time                   `json:"properties.creationDate"`
	PropertiesZoneRedundant bool                        `json:"properties.zoneRedundant"`
	PropertiesReadScale     string                      `json:"properties.readScale"`
	Sku                     *DuploAzureMsSqlDatabaseSku `json:"sku"`
	Tags                    map[string]interface{}      `json:"tags"`
	Location                string                      `json:"location"`
	ID                      string                      `json:"id"`
	Name                    string                      `json:"name"`
	Type                    string                      `json:"type"`
}

func (c *Client) MsSqlServerCreate(tenantID string, rq *DuploAzureMsSqlRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlServerCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateSqlServer", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) MsSqlDatabaseCreate(tenantID, serverName string, rq *DuploAzureMsSqlDatabaseRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlDatabaseCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateSqlDatabase/%s", tenantID, serverName),
		&rq,
		nil,
	)
}

func (c *Client) MsSqlElasticPoolCreate(tenantID, serverName string, rq *DuploAzureMsSqlElasticPoolRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlElasticPoolCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateSqlElasticPool/%s", tenantID, serverName),
		&rq,
		nil,
	)
}

func (c *Client) MsSqlServerGet(tenantID string, name string) (*DuploAzureMsSqlServer, ClientError) {
	list, err := c.MsSqlServerList(tenantID)
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

func (c *Client) MsSqlDatabaseGet(tenantID, serverName, dbName string) (*DuploAzureMsSqlDatabase, ClientError) {
	list, err := c.MsSqlDatabaseList(tenantID, serverName)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, db := range *list {
			if db.Name == dbName {
				return &db, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) MsSqlElasticPoolGet(tenantID, serverName, epName string) (*DuploAzureMsSqlElasticPool, ClientError) {
	list, err := c.MsSqlElasticPoolList(tenantID, serverName)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, ep := range *list {
			if ep.Name == epName {
				return &ep, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) MsSqlServerList(tenantID string) (*[]DuploAzureMsSqlServer, ClientError) {
	rp := []DuploAzureMsSqlServer{}
	err := c.getAPI(
		fmt.Sprintf("MsSqlServerList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListSqlServers", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) MsSqlDatabaseList(tenantID, name string) (*[]DuploAzureMsSqlDatabase, ClientError) {
	rp := []DuploAzureMsSqlDatabase{}
	err := c.getAPI(
		fmt.Sprintf("MsSqlDatabaseList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListSqlDatabases/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) MsSqlElasticPoolList(tenantID, name string) (*[]DuploAzureMsSqlElasticPool, ClientError) {
	rp := []DuploAzureMsSqlElasticPool{}
	err := c.getAPI(
		fmt.Sprintf("MsSqlElasticPoolList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListSqlElasticPools/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) MsSqlServerExists(tenantID, name string) (bool, ClientError) {
	list, err := c.MsSqlServerList(tenantID)
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

func (c *Client) MsSqlDatabaseExists(tenantID, serverName, dbName string) (bool, ClientError) {
	list, err := c.MsSqlDatabaseList(tenantID, serverName)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, db := range *list {
			if db.Name == dbName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) MsSqlElasticPoolExists(tenantID, serverName, epName string) (bool, ClientError) {
	list, err := c.MsSqlElasticPoolList(tenantID, serverName)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, ep := range *list {
			if ep.Name == epName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) MsSqlServerDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlServerDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteSqlServer/%s", tenantID, name),
		nil,
		nil,
	)
}
func (c *Client) MsSqlDatabaseDelete(tenantID, serverName, dbName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlDatabaseDelete(%s, %s, %s)", tenantID, serverName, dbName),
		fmt.Sprintf("subscriptions/%s/DeleteSqlDatabase", tenantID),
		&DuploAzureMsSqlDatabaseDeleteRequest{
			DbName:     dbName,
			ServerName: serverName,
		},
		nil,
	)
}

func (c *Client) MsSqlElasticPoolDelete(tenantID, serverName, epName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlElasticPoolDelete(%s, %s, %s)", tenantID, serverName, epName),
		fmt.Sprintf("subscriptions/%s/DeleteSqlElasticPool", tenantID),
		&DuploAzureMsSqlEPDeleteRequest{
			ElasticPoolName: epName,
			ServerName:      serverName,
		},
		nil,
	)
}
