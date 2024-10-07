package duplosdk

import (
	"fmt"
	"time"
)

type DuploAzureStorageAccountBlob struct {
	ServiceAccount DuploAzureStorageAccountResourceServiceAccount `json:"ServiceClient"`
	URI            string                                         `json:"Uri"`
	Name           string                                         `json:"Name"`
	Metadata       DuploAzureStorageAccountResourceMetadata       `json:"Metadata"`
	Properties     DuploAzureStorageAccountResourceProperties     `json:"Properties"`
	StorageURI     DuploAzureStorageAccountResourceStorageURI     `json:"StorageUri"`
}

type DuploAzureStorageAccountQueue struct {
	ServiceAccount DuploAzureStorageAccountResourceServiceAccount `json:"ServiceClient"`
	URI            string                                         `json:"Uri"`
	StorageURI     DuploAzureStorageAccountResourceStorageURI     `json:"StorageUri"`
	Name           string                                         `json:"Name"`
	Metadata       DuploAzureStorageAccountResourceMetadata       `json:"Metadata"`
	EncodeMessage  bool                                           `json:"EncodeMessage"`
}

type DuploAzureStorageAccountTable struct {
	ServiceAccount DuploAzureStorageAccountResourceServiceAccount `json:"ServiceClient"`
	URI            string                                         `json:"Uri"`
	StorageURI     DuploAzureStorageAccountResourceStorageURI     `json:"StorageUri"`
	Name           string                                         `json:"Name"`
}
type DuploAzureStorageAccountResourceMetadata struct {
}

type DuploAzureStorageAccountResourceProperties struct {
	ETag          string    `json:"ETag"`
	LastModified  time.Time `json:"LastModified"`
	Quota         int       `json:"Quota"`
	LeaseStatus   int       `json:"LeaseStatus"`
	LeaseState    int       `json:"LeaseState"`
	LeaseDuration int       `json:"LeaseDuration"`
	PublicAccess  int       `json:"PublicAccess"`
}

type DuploAzureStorageAccountResourceServiceAccount struct {
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
	BaseURI               string                                     `json:"BaseUri"`
	StorageURI            DuploAzureStorageAccountResourceStorageURI `json:"StorageUri"`
	DefaultRequestOptions struct {
		RetryPolicy struct {
		} `json:"RetryPolicy"`
		LocationMode                     int         `json:"LocationMode"`
		RequireEncryption                bool        `json:"RequireEncryption"`
		ServerTimeout                    interface{} `json:"ServerTimeout"`
		MaximumExecutionTime             interface{} `json:"MaximumExecutionTime"`
		ParallelOperationThreadCount     int         `json:"ParallelOperationThreadCount"`
		UseTransactionalMD5              interface{} `json:"UseTransactionalMD5"`
		StoreFileContentMD5              interface{} `json:"StoreFileContentMD5"`
		DisableContentMD5Validation      interface{} `json:"DisableContentMD5Validation"`
		DefaultDelimiter                 string      `json:"DefaultDelimiter"`
		SingleBlobUploadThresholdInBytes int         `json:"SingleBlobUploadThresholdInBytes,omitempty"`
	} `json:"DefaultRequestOptions"`
}

type DuploAzureStorageAccountResourceStorageURI struct {
	PrimaryURI   string `json:"PrimaryUri"`
	SecondaryURI string `json:"SecondaryUri"`
}

func (c *Client) AzureStorageAccountBlobList(tenantID, storageAccName string) (*[]DuploAzureStorageAccountBlob, ClientError) {
	rp := []DuploAzureStorageAccountBlob{}
	err := c.getAPI(
		fmt.Sprintf("AzureStorageAccountBlobList(%s, %s)", tenantID, storageAccName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/blob", tenantID, storageAccName),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureStorageAccountBlobGet(tenantID, storageAccName, name string) (*DuploAzureStorageAccountBlob, ClientError) {
	blobs, err := c.AzureStorageAccountBlobList(tenantID, storageAccName)
	if err != nil {
		return nil, err
	}

	if blobs != nil {
		for _, blob := range *blobs {
			if blob.Name == name {
				return &blob, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureStorageAccountQueueList(tenantID, storageAccName string) (*[]DuploAzureStorageAccountQueue, ClientError) {
	rp := []DuploAzureStorageAccountQueue{}
	err := c.getAPI(
		fmt.Sprintf("AzureStorageAccountQueueList(%s, %s)", tenantID, storageAccName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/queue", tenantID, storageAccName),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureStorageAccountQueueGet(tenantID, storageAccName, name string) (*DuploAzureStorageAccountQueue, ClientError) {
	queues, err := c.AzureStorageAccountQueueList(tenantID, storageAccName)
	if err != nil {
		return nil, err
	}

	if queues != nil {
		for _, queue := range *queues {
			if queue.Name == name {
				return &queue, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureStorageAccountTableList(tenantID, storageAccName string) (*[]DuploAzureStorageAccountTable, ClientError) {
	rp := []DuploAzureStorageAccountTable{}
	err := c.getAPI(
		fmt.Sprintf("AzureStorageAccountTableList(%s, %s)", tenantID, storageAccName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/table", tenantID, storageAccName),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureStorageAccountTableGet(tenantID, storageAccName, name string) (*DuploAzureStorageAccountTable, ClientError) {
	queues, err := c.AzureStorageAccountTableList(tenantID, storageAccName)
	if err != nil {
		return nil, err
	}

	if queues != nil {
		for _, queue := range *queues {
			if queue.Name == name {
				return &queue, nil
			}
		}
	}
	return nil, nil
}

type DuploAzureStorageResource struct {
	Name string `json:"Name"`
}

func (c *Client) AzureStorageAccountBlobCreate(tenantID string, storageAccountName string, name string) ClientError {
	rp := &DuploAzureStorageAccountBlob{}
	return c.postAPI(
		fmt.Sprintf("AzureStorageAccountBlobCreate(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/blob", tenantID, storageAccountName),
		DuploAzureStorageResource{
			Name: name,
		},
		rp,
	)
}

func (c *Client) AzureStorageAccountQueueCreate(tenantID string, storageAccountName string, name string) ClientError {
	rp := &DuploAzureStorageAccountBlob{}

	return c.postAPI(
		fmt.Sprintf("AzureStorageAccountQueueCreate(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/queue", tenantID, storageAccountName),
		DuploAzureStorageResource{
			Name: name,
		},
		rp,
	)
}

func (c *Client) AzureStorageAccountTableCreate(tenantID string, storageAccountName string, name string) ClientError {
	rp := &DuploAzureStorageAccountBlob{}

	return c.postAPI(
		fmt.Sprintf("AzureStorageAccountTableCreate(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/table", tenantID, storageAccountName),
		DuploAzureStorageResource{
			Name: name,
		},
		rp,
	)
}

func (c *Client) AzureStorageAccountBlobDelete(tenantID string, storageAccountName string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AzureStorageAccountBlobDelete(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/blob/%s", tenantID, storageAccountName, name),
		nil,
	)
}

func (c *Client) AzureStorageAccountQueueDelete(tenantID string, storageAccountName string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AzureStorageAccountQueueDelete(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/queue/%s", tenantID, storageAccountName, name),
		nil,
	)
}

func (c *Client) AzureStorageAccountTableDelete(tenantID string, storageAccountName string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("AzureStorageAccountTableDelete(%s, %s)", tenantID, storageAccountName),
		fmt.Sprintf("v3/subscriptions/%s/azure/storageaccount/%s/table/%s", tenantID, storageAccountName, name),
		nil,
	)
}
