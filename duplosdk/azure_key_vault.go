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

// Tenant level Key Vault API's

type DuploAzureTenantKeyVaultRequest struct {
	Name       string `json:"name"`
	Properties struct {
		Sku struct {
			Name   string `json:"name"`
			Family string `json:"family,omitempty"`
		} `json:"sku"`
		SoftDeleteRetentionInDays int  `json:"softDeleteRetentionInDays,omitempty"`
		EnablePurgeProtection     bool `json:"enablePurgeProtection,omitempty"`
	} `json:"properties"`
}

type DuploAzureTenantKeyVault struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	Location              string `json:"location"`
	EnablePurgeProtection string `json:"EnablePurgeProtection"`
	Properties            struct {
		TenantID string `json:"tenantId,omitempty"`
		Sku      struct {
			Name   string `json:"name"`
			Family string `json:"family"`
		} `json:"sku"`
		AccessPolicies               []interface{} `json:"accessPolicies,omitempty"`
		VaultURI                     string        `json:"vaultUri,omitempty"`
		EnabledForDeployment         bool          `json:"enabledForDeployment,omitempty"`
		EnabledForDiskEncryption     bool          `json:"enabledForDiskEncryption,omitempty"`
		EnabledForTemplateDeployment bool          `json:"enabledForTemplateDeployment,omitempty"`
		EnableSoftDelete             bool          `json:"enableSoftDelete,omitempty"`
		SoftDeleteRetentionInDays    int           `json:"softDeleteRetentionInDays,omitempty"`
		EnableRbacAuthorization      bool          `json:"enableRbacAuthorization,omitempty"`
		EnablePurgeProtection        bool          `json:"enablePurgeProtection,omitempty"`
	} `json:"properties"`
}

func (c *Client) TenantKeyVaultCreate(tenantID string, rq *DuploAzureTenantKeyVaultRequest) ClientError {
	resp := &DuploAzureTenantKeyVaultRequest{}
	return c.postAPI(
		fmt.Sprintf("TenantKeyVaultCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault", tenantID),
		&rq,
		&resp,
	)
}

func (c *Client) TenantKeyVaultGet(tenantID, secretName string) (*DuploAzureTenantKeyVault, ClientError) {

	list, err := c.TenantKeyVaultList(tenantID)

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, vault := range *list {
			if vault.Name == secretName {
				return &vault, nil
			}
		}
	}
	return nil, nil

}

func (c *Client) TenantKeyVaultList(tenantID string) (*[]DuploAzureTenantKeyVault, ClientError) {
	resp := []DuploAzureTenantKeyVault{}
	err := c.getAPI(
		fmt.Sprintf("TenantKeyVaultList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault", tenantID),
		&resp,
	)
	return &resp, err
}

func (c *Client) TenantKeyVaultDelete(tenantID string, secretName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("TenantKeyVaultDelete(%s, %s)", tenantID, secretName),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault/%s", tenantID, secretName),
		nil,
	)
}

// Tenant level Key Vault Secret API's

type DuploAzureTenantKeyVaultSecretRequest struct {
	VaultName   string `json:"VaultName"`
	SecretName  string `json:"SecretName"`
	SecretValue string `json:"SecretValue"`
	ContentType string `json:"ContentType,omitempty"`
}

type DuploAzureTenantKeyVaultSecret struct {
	SecretIdentifier struct {
		BaseIdentifier     string `json:"BaseIdentifier"`
		Identifier         string `json:"Identifier"`
		Name               string `json:"Name"`
		Vault              string `json:"Vault"`
		VaultWithoutScheme string `json:"VaultWithoutScheme"`
		Version            string `json:"Version"`
	} `json:"SecretIdentifier"`
	Value       string `json:"value"`
	ID          string `json:"id"`
	ContentType string `json:"contentType,omitempty"`
	Attributes  struct {
		RecoveryLevel string `json:"recoveryLevel"`
		Enabled       bool   `json:"enabled"`
		Created       int    `json:"created"`
		Updated       int    `json:"updated"`
	} `json:"attributes"`
}

func (c *Client) TenantKeyVaultSecretCreate(tenantID string, rq *DuploAzureTenantKeyVaultSecretRequest) ClientError {
	resp := &DuploAzureTenantKeyVaultSecret{}
	return c.postAPI(
		fmt.Sprintf("TenantKeyVaultSecretCreate(%s, %s, %s)", tenantID, rq.VaultName, rq.SecretName),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault/%s/secret", tenantID, rq.VaultName),
		&rq,
		&resp,
	)
}

func (c *Client) TenantKeyVaultSecretGet(tenantID, vaultName, secretName string) (*DuploAzureTenantKeyVaultSecret, ClientError) {

	list, err := c.TenantKeyVaultSecretList(tenantID, vaultName)

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, secret := range *list {
			if secret.SecretIdentifier.Name == secretName {
				return &secret, nil
			}
		}
	}
	return nil, nil

}

func (c *Client) TenantKeyVaultSecretList(tenantID, vaultName string) (*[]DuploAzureTenantKeyVaultSecret, ClientError) {
	resp := []DuploAzureTenantKeyVaultSecret{}
	err := c.getAPI(
		fmt.Sprintf("TenantKeyVaultSecretList(%s, %s)", tenantID, vaultName),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault/%s/secret", tenantID, vaultName),
		&resp,
	)
	return &resp, err
}

func (c *Client) TenantKeyVaultSecretDelete(tenantID string, vaultName string, secretName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("TenantKeyVaultSecretDelete(%s, %s, %s)", tenantID, vaultName, secretName),
		fmt.Sprintf("v3/subscriptions/%s/azure/keyvault/%s/secret/%s", tenantID, vaultName, secretName),
		nil,
	)
}
