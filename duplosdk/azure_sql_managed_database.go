package duplosdk

import "fmt"

type DuploAzureSqlManagedDatabaseSku struct {
	Tier   string `json:"tier"`
	Name   string `json:"name"`
	Family string `json:"family"`
}

type DuploAzureSqlManagedDatabaseRequest struct {
	NameEx                               string                           `json:"NameEx"`
	PropertiesAdministratorLogin         string                           `json:"properties.administratorLogin"`
	PropertiesAdministratorLoginPassword string                           `json:"properties.administratorLoginPassword"`
	PropertiesSubnetID                   string                           `json:"properties.subnetId"`
	PropertiesVCores                     string                           `json:"properties.vCores"`
	PropertiesStorageSizeInGB            int                              `json:"properties.storageSizeInGB"`
	Sku                                  *DuploAzureSqlManagedDatabaseSku `json:"sku"`
}

type DuploAzureSqlManagedDatabaseInstance struct {
	Identity struct {
		PrincipalID string `json:"principalId"`
		Type        string `json:"type"`
		TenantID    string `json:"tenantId"`
	} `json:"identity"`
	Sku struct {
		Name     string `json:"name"`
		Tier     string `json:"tier"`
		Family   string `json:"family"`
		Capacity int    `json:"capacity"`
	} `json:"sku"`
	PropertiesProvisioningState         string                 `json:"properties.provisioningState"`
	PropertiesFullyQualifiedDomainName  string                 `json:"properties.fullyQualifiedDomainName"`
	PropertiesAdministratorLogin        string                 `json:"properties.administratorLogin"`
	PropertiesSubnetID                  string                 `json:"properties.subnetId"`
	PropertiesState                     string                 `json:"properties.state"`
	PropertiesLicenseType               string                 `json:"properties.licenseType"`
	PropertiesVCores                    int                    `json:"properties.vCores"`
	PropertiesStorageSizeInGB           int                    `json:"properties.storageSizeInGB"`
	PropertiesCollation                 string                 `json:"properties.collation"`
	PropertiesPublicDataEndpointEnabled bool                   `json:"properties.publicDataEndpointEnabled"`
	PropertiesProxyOverride             string                 `json:"properties.proxyOverride"`
	PropertiesTimezoneID                string                 `json:"properties.timezoneId"`
	PropertiesMinimalTLSVersion         string                 `json:"properties.minimalTlsVersion"`
	PropertiesStorageAccountType        string                 `json:"properties.storageAccountType"`
	Location                            string                 `json:"location"`
	Tags                                map[string]interface{} `json:"tags"`
	ID                                  string                 `json:"id"`
	Name                                string                 `json:"name"`
	Type                                string                 `json:"type"`
}

func (c *Client) SqlManagedDatabaseCreate(tenantID string, rq *DuploAzureSqlManagedDatabaseRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("SqlManagedDatabaseCreate(%s, %s)", tenantID, rq.NameEx),
		fmt.Sprintf("subscriptions/%s/CreateSqlManagedInstance", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) SqlManagedDatabaseGet(tenantID string, name string) (*DuploAzureSqlManagedDatabaseInstance, ClientError) {
	list, err := c.SqlManagedDatabaseList(tenantID)
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

func (c *Client) SqlManagedDatabaseList(tenantID string) (*[]DuploAzureSqlManagedDatabaseInstance, ClientError) {
	rp := []DuploAzureSqlManagedDatabaseInstance{}
	err := c.getAPI(
		fmt.Sprintf("SqlManagedDatabaseList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListSqlManagedInstances", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) SqlManagedDatabaseExists(tenantID, name string) (bool, ClientError) {
	list, err := c.SqlManagedDatabaseList(tenantID)
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

func (c *Client) SqlManagedDatabaseDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("SqlManagedDatabaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteSqlManagedInstance/%s", tenantID, name),
		nil,
		nil,
	)
}
