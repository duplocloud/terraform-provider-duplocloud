package duplosdk

import (
	"fmt"
)

type DuploAzureSecretItem struct {
	Identifier struct {
		BaseIdentifier     string `json:"BaseIdentifier"`
		Identifier         string `json:"Identifier"`
		Name               string `json:"Name"`
		Vault              string `json:"Vault"`
		VaultWithoutScheme string `json:"VaultWithoutScheme"`
		Version            string `json:"Version"`
	} `json:"Identifier"`
	ID         string `json:"id"`
	Attributes struct {
		RecoveryLevel string `json:"recoveryLevel"`
		Enabled       bool   `json:"enabled"`
		Created       int    `json:"created"`
		Updated       int    `json:"updated"`
	} `json:"attributes"`
	Tags struct {
		CreatedBy string `json:"CreatedBy"`
	} `json:"tags"`
	ContentType string `json:"contentType"`
}

type DuploAzureKeyVaultRequest struct {
	SecretName  string `json:"SecretName"`
	SecretValue string `json:"SecretValue,omitempty"`
	SecretType  string `json:"SecretType,omitempty"`
}

func (c *Client) KeyVaultSecretCreate(tenantID string, rq *DuploAzureKeyVaultRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("KeyVaultSecretCreate(%s, %s)", tenantID, rq.SecretName),
		fmt.Sprintf("subscriptions/%s/CreateKeyVaultSecret", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) KeyVaultSecretGet(tenantID, secretName string) (*DuploAzureSecretItem, ClientError) {

	list, err := c.KeyVaultSecretList(tenantID)

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, secret := range *list {
			if secret.Identifier.Name == secretName {
				return &secret, nil
			}
		}
	}
	return nil, nil

	// rp := ""
	// err := c.getAPI(
	// 	fmt.Sprintf("KeyVaultSecretGet(%s, %s)", tenantID, secretName),
	// 	fmt.Sprintf("subscriptions/%s/GetKeyVaultSecret/%s", tenantID, secretName),
	// 	&rp,
	// )
	// return rp, err
}

func (c *Client) KeyVaultSecretList(tenantID string) (*[]DuploAzureSecretItem, ClientError) {
	rp := []DuploAzureSecretItem{}
	err := c.getAPI(
		fmt.Sprintf("KeyVaultSecretList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListKeyVaultSecrets", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) KeyVaultSecretDelete(tenantID string, secretName string) ClientError {
	return c.postAPI(
		fmt.Sprintf("KeyVaultSecretDelete(%s, %s)", tenantID, secretName),
		fmt.Sprintf("subscriptions/%s/DeleteKeyVaultSecret/%s", tenantID, secretName),
		nil,
		nil,
	)
}
