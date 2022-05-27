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

type DuploAzureStorageAccountShareFileCreateReq struct {
	Name string `json:"Name"`
}

type DuploAzureStorageAccountShareFileGetReq struct {
	ServiceClient struct {
		AuthenticationScheme int         `json:"AuthenticationScheme"`
		BufferManager        interface{} `json:"BufferManager"`
		Credentials          struct {
			SASToken     interface{} `json:"SASToken"`
			AccountName  string      `json:"AccountName"`
			KeyName      interface{} `json:"KeyName"`
			IsAnonymous  bool        `json:"IsAnonymous"`
			IsSAS        bool        `json:"IsSAS"`
			IsSharedKey  bool        `json:"IsSharedKey"`
			SASSignature interface{} `json:"SASSignature"`
		} `json:"Credentials"`
		BaseURI    string `json:"BaseUri"`
		StorageURI struct {
			PrimaryURI   string `json:"PrimaryUri"`
			SecondaryURI string `json:"SecondaryUri"`
		} `json:"StorageUri"`
		DefaultRequestOptions struct {
			RetryPolicy struct {
			} `json:"RetryPolicy"`
			LocationMode                 int         `json:"LocationMode"`
			RequireEncryption            bool        `json:"RequireEncryption"`
			ServerTimeout                interface{} `json:"ServerTimeout"`
			MaximumExecutionTime         interface{} `json:"MaximumExecutionTime"`
			ParallelOperationThreadCount int         `json:"ParallelOperationThreadCount"`
			UseTransactionalMD5          interface{} `json:"UseTransactionalMD5"`
			StoreFileContentMD5          interface{} `json:"StoreFileContentMD5"`
			DisableContentMD5Validation  interface{} `json:"DisableContentMD5Validation"`
		} `json:"DefaultRequestOptions"`
	} `json:"ServiceClient"`
	URI        string `json:"Uri"`
	StorageURI struct {
		PrimaryURI   string `json:"PrimaryUri"`
		SecondaryURI string `json:"SecondaryUri"`
	} `json:"StorageUri"`
	SnapshotTime                interface{} `json:"SnapshotTime"`
	IsSnapshot                  bool        `json:"IsSnapshot"`
	SnapshotQualifiedURI        string      `json:"SnapshotQualifiedUri"`
	SnapshotQualifiedStorageURI struct {
		PrimaryURI   string `json:"PrimaryUri"`
		SecondaryURI string `json:"SecondaryUri"`
	} `json:"SnapshotQualifiedStorageUri"`
	Name     string `json:"Name"`
	Metadata struct {
	} `json:"Metadata"`
	Properties struct {
		ETag         string    `json:"ETag"`
		LastModified time.Time `json:"LastModified"`
		Quota        int       `json:"Quota"`
	} `json:"Properties"`
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

func (c *Client) StorageShareFileCreate(tenantID string, storageAccountName string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("StorageShareFileCreate(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("subscriptions/%s/CreateFileShare/%s", tenantID, storageAccountName),
		DuploAzureStorageAccountShareFileCreateReq{
			Name: name,
		},
		nil,
	)
}

func (c *Client) StorageAccountShareFileList(tenantID, storageAccName string) (*[]DuploAzureStorageAccountShareFileGetReq, ClientError) {
	rp := []DuploAzureStorageAccountShareFileGetReq{}
	err := c.getAPI(
		fmt.Sprintf("StorageAccountShareFileList(%s, %s)", tenantID, storageAccName),
		fmt.Sprintf("subscriptions/%s/ListFileShares/%s", tenantID, storageAccName),
		&rp,
	)
	return &rp, err
}

func (c *Client) StorageShareFileGet(tenantID, storageAccName, name string) (*DuploAzureStorageAccountShareFileGetReq, ClientError) {
	list, err := c.StorageAccountShareFileList(tenantID, storageAccName)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, ssf := range *list {
			if ssf.Name == name {
				return &ssf, nil
			}
		}
	}
	return nil, nil
}
