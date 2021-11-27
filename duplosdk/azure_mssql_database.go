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
	Sku struct {
		Name     string `json:"name"`
		Tier     string `json:"tier"`
		Capacity int    `json:"capacity"`
	} `json:"sku"`
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
	PropertiesCurrentSku                    struct {
		Name     string `json:"name"`
		Tier     string `json:"tier"`
		Capacity int    `json:"capacity"`
	} `json:"properties.currentSku"`
	Tags     map[string]interface{} `json:"tags"`
	Location string                 `json:"location"`
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
}

func (c *Client) MsSqlDatabaseCreate(tenantID string, rq *DuploAzureMsSqlRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlDatabaseCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateSqlServer", tenantID),
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

func (c *Client) MsSqlDatabaseDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("MsSqlDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteSqlServer/%s", tenantID, name),
		nil,
		nil,
	)
}
