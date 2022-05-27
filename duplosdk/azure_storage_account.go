package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureStorageAccount struct {
	Sku struct {
		Name string `json:"name,omitempty"`
		Tier string `json:"tier,omitempty"`
	} `json:"sku,omitempty"`
	Kind                        string `json:"kind,omitempty"`
	PropertiesProvisioningState string `json:"properties.provisioningState,omitempty"`
	PropertiesPrimaryEndpoints  struct {
		Blob  string `json:"blob,omitempty"`
		Queue string `json:"queue,omitempty"`
		Table string `json:"table,omitempty"`
		File  string `json:"file,omitempty"`
		Web   string `json:"web,omitempty"`
		Dfs   string `json:"dfs,omitempty"`
	} `json:"properties.primaryEndpoints,omitempty"`
	PropertiesPrimaryLocation string    `json:"properties.primaryLocation,omitempty"`
	PropertiesStatusOfPrimary string    `json:"properties.statusOfPrimary,omitempty"`
	PropertiesCreationTime    time.Time `json:"properties.creationTime,omitempty"`
	PropertiesEncryption      struct {
		Services struct {
			Blob struct {
				Enabled         bool      `json:"enabled,omitempty"`
				LastEnabledTime time.Time `json:"lastEnabledTime,omitempty"`
				KeyType         string    `json:"keyType,omitempty"`
			} `json:"blob,omitempty"`
			File struct {
				Enabled         bool      `json:"enabled,omitempty"`
				LastEnabledTime time.Time `json:"lastEnabledTime,omitempty"`
				KeyType         string    `json:"keyType,omitempty"`
			} `json:"file,omitempty"`
		} `json:"services,omitempty"`
		KeySource string `json:"keySource,omitempty"`
	} `json:"properties.encryption,omitempty"`
	PropertiesAccessTier               string `json:"properties.accessTier,omitempty"`
	PropertiesSupportsHTTPSTrafficOnly bool   `json:"properties.supportsHttpsTrafficOnly,omitempty"`
	PropertiesNetworkAcls              struct {
		Bypass              string        `json:"bypass,omitempty"`
		VirtualNetworkRules []interface{} `json:"virtualNetworkRules,omitempty"`
		IPRules             []interface{} `json:"ipRules,omitempty"`
		DefaultAction       string        `json:"defaultAction,omitempty"`
	} `json:"properties.networkAcls,omitempty"`
	PropertiesPrivateEndpointConnections []interface{} `json:"properties.privateEndpointConnections,omitempty"`
	Tags                                 struct {
	} `json:"tags,omitempty"`
	Location string `json:"location,omitempty"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
}

func (c *Client) StorageAccountCreate(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("StorageAccountCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/CreateStorageAccount/%s", tenantID, name),
		nil,
		nil,
	)
}

func (c *Client) StorageAccountGet(tenantID, name string) (*DuploAzureStorageAccount, ClientError) {
	rp := DuploAzureStorageAccount{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountGet(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/GetStorageAccountDetails/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) StorageAccountList(tenantID string) (*[]DuploAzureStorageAccount, ClientError) {
	rp := []DuploAzureStorageAccount{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListStorageAccounts", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) StorageAccountExists(tenantID, name string) (bool, ClientError) {
	list, err := c.StorageAccountList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, sa := range *list {
			if sa.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) StorageAccountDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("StorageAccountDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteStorageAccount/%s", tenantID, name),
		nil,
		nil,
	)
}

func (c *Client) StorageAccountGetKey(tenantID, name string) (*DuploKeyStringValue, ClientError) {
	rp := DuploKeyStringValue{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountGetKey(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/GetStorageAccountKeys/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}
