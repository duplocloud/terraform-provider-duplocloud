package duplosdk

import (
	"fmt"
)

type DuploAwsEcrRepositoryRequest struct {
	KmsEncryption         string `json:"KmsEncryption,omitempty"`
	EnableTagImmutability bool   `json:"EnableTagImmutability,omitempty"`
	EnableScanImageOnPush bool   `json:"EnableScanImageOnPush,omitempty"`
	Name                  string `json:"Name"`
}

type DuploAwsEcrRepository struct {
	KmsEncryption         string `json:"KmsEncryption,omitempty"`
	KmsEncryptionAlias    string `json:"KmsEncryptionAlias,omitempty"`
	EnableTagImmutability bool   `json:"EnableTagImmutability,omitempty"`
	EnableScanImageOnPush bool   `json:"EnableScanImageOnPush,omitempty"`
	Arn                   string `json:"Arn"`
	ResourceType          int    `json:"ResourceType,omitempty"`
	Name                  string `json:"Name"`
	RegistryId            string `json:"RegistryId,omitempty"`
	RepositoryUri         string `json:"RepositoryUri,omitempty"`
}

func (c *Client) AwsEcrRepositoryCreate(tenantID string, rq *DuploAwsEcrRepositoryRequest) ClientError {
	rp := DuploAwsEcrRepository{}
	return c.postAPI(
		fmt.Sprintf("AwsEcrRepositoryCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecrRepository", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) AwsEcrRepositoryGet(tenantID string, name string) (*DuploAwsEcrRepository, ClientError) {
	rp := DuploAwsEcrRepository{}
	err := c.getAPI(
		fmt.Sprintf("AwsEcrRepositoryGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecrRepository/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsEcrRepositoryList(tenantID string) (*[]DuploAwsEcrRepository, ClientError) {
	rp := []DuploAwsEcrRepository{}
	err := c.getAPI(
		fmt.Sprintf("AwsEcrRepositoryList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecrRepository", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AwsEcrRepositoryExists(tenantID, name string) (bool, ClientError) {
	list, err := c.AwsEcrRepositoryList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) AwsEcrRepositoryDelete(tenantID string, name string, forceDelete bool) ClientError {
	forceDeletePrefix := ""

	if forceDelete {
		forceDeletePrefix = "force/"
	}
	return c.deleteAPI(
		fmt.Sprintf("AwsEcrRepositoryDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecrRepository/%s%s", tenantID, forceDeletePrefix, name),
		nil,
	)
}
