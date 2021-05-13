package duplosdk

import (
	"fmt"
	"log"
)

// DuploTenantSecret represents a managed secret for a Duplo tenant
type DuploTenantSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Arn                    string                 `json:"ARN"`
	Name                   string                 `json:"Name"`
	RotationEnabled        bool                   `json:"RotationEnabled,omitempty"`
	SecretVersionsToStages map[string][]string    `json:"SecretVersionsToStages,omitempty"`
	Tags                   *[]DuploKeyStringValue `json:"Tags,omitempty"`
	CreatedDate            string                 `json:"CreatedDate,omitempty"`
	DeletedDate            string                 `json:"DeletedDate,omitempty"`
	LastAccessedDate       string                 `json:"LastAccessedDate,omitempty"`
	LastChangedDate        string                 `json:"LastChangedDate,omitempty"`
	LastRotatedDate        string                 `json:"LastRotatedDate,omitempty"`
}

// DuploTenantSecretRequest represents a request to create a secret
type DuploTenantSecretRequest struct {
	Name         string `json:"Name"`
	SecretString string `json:"SecretString"`
}

// TenantListSecrets retrieves a list of managed secrets via the Duplo API
func (c *Client) TenantListSecrets(tenantID string) (*[]DuploTenantSecret, ClientError) {
	apiName := fmt.Sprintf("TenantListSecrets(%s)", tenantID)
	list := []DuploTenantSecret{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/ListTenantSecrets", tenantID), &list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID to each element and return the list.
	log.Printf("[TRACE] %s: %d items", apiName, len(list))
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// TenantGetSecretByName retrieves a managed secret via the Duplo API
func (c *Client) TenantGetSecretByName(tenantID string, name string) (*DuploTenantSecret, ClientError) {
	allSecrets, err := c.TenantListSecrets(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific name.
	for _, secret := range *allSecrets {
		if secret.Name == name {
			return &secret, nil
		}
	}

	// No secret was found.
	return nil, nil
}

// TenantGetSecretByNameSuffix retrieves a managed secret via the Duplo API
func (c *Client) TenantGetSecretByNameSuffix(tenantID string, nameSuffix string) (*DuploTenantSecret, ClientError) {
	name, err := c.GetDuploServicesName(tenantID, nameSuffix)
	if err != nil {
		return nil, err
	}
	return c.TenantGetSecretByName(tenantID, name)
}

// TenantCreateSecret creates a tenant secret via Duplo.
func (c *Client) TenantCreateSecret(tenantID string, duplo *DuploTenantSecretRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("TenantCreateSecret(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/CreateTenantSecret", tenantID),
		&duplo,
		nil)
}

// TenantDeleteSecret deletes a tenant secret via Duplo.
func (c *Client) TenantDeleteSecret(tenantID string, name string) error {
	return c.postAPI(
		fmt.Sprintf("TenantDeleteSecret(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/deleteTenantSecret", tenantID),
		&map[string]interface{}{"ForceDeleteWithoutRecovery": true, "SecretId": name},
		nil)
}
