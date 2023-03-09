package duplosdk

import (
	"fmt"
)

// DuploAwsSecret represents a AWS secretsmanager secret for a Duplo tenant
type DuploAwsSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	SecretId               string                 `json:"SecretId"`
	Arn                    string                 `json:"ARN"`
	CreatedDate            string                 `json:"CreatedDate,omitempty"`
	DeletedDate            string                 `json:"DeletedDate,omitempty"`
	LastAccessedDate       string                 `json:"LastAccessedDate,omitempty"`
	LastChangedDate        string                 `json:"LastChangedDate,omitempty"`
	LastRotatedDate        string                 `json:"LastRotatedDate,omitempty"`
	Name                   string                 `json:"Name"`
	RotationEnabled        bool                   `json:"RotationEnabled,omitempty"`
	SecretVersionsToStages map[string][]string    `json:"SecretVersionsToStages,omitempty"`
	Tags                   *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

type DuploAwsSecretValue struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	SecretId      string    `json:"SecretId"`
	SecretString  string    `json:"SecretString"`
	Arn           string    `json:"ARN"`
	CreatedDate   string    `json:"CreatedDate,omitempty"`
	Name          string    `json:"Name"`
	VersionId     string    `json:"VersionId,omitempty"`
	VersionStages *[]string `json:"VersionStages,omitempty"`
}

// DuploAwsSecretCreatedResponse represents a AWS secretsmanager secret for a Duplo tenant
type DuploAwsSecretUpdatedResponse struct {
	Arn       string `json:"ARN"`
	Name      string `json:"Name"`
	VersionId string `json:"VersionId"`
}

// DuploAwsSecretCreateRequest represents a request to create a secret
type DuploAwsSecretCreateRequest struct {
	Name         string `json:"Name"`
	SecretString string `json:"SecretString"`
}

// DuploAwsSecretUpdateRequest represents a request to update a secret
type DuploAwsSecretUpdateRequest struct {
	SecretId        string `json:"SecretId"`
	SecretValueType string `json:"SecretValueType"`
	SecretString    string `json:"SecretString"`
}

// TenantListAwsSecrets retrieves a list of managed secrets via the Duplo API
func (c *Client) TenantListAwsSecrets(tenantID string) (*[]DuploAwsSecret, ClientError) {
	list := []DuploAwsSecret{}

	// Get the list from Duplo
	err := c.getAPI(
		fmt.Sprintf("TenantListAwsSecrets(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/secret", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID to each element and return the list.
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// TenantGetSecretByName retrieves a managed secret via the Duplo API
func (c *Client) TenantGetAwsSecret(tenantID string, name string) (*DuploAwsSecret, ClientError) {

	// Retrieve the list of secrets
	list, err := c.TenantListAwsSecrets(tenantID)
	if err != nil || list == nil {
		return nil, err
	}

	// Return the secret, if it exists.
	for _, secret := range *list {
		if secret.Name == name {
			return &secret, nil
		}
	}

	return nil, nil
}

// TenantGetAwsSecretValue retrieves a managed secret value via the Duplo API
func (c *Client) TenantGetAwsSecretValue(tenantID string, name string) (*DuploAwsSecretValue, ClientError) {
	rp := DuploAwsSecretValue{}
	err := c.getAPI(
		fmt.Sprintf("TenantGetAwsSecret(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/secret/%s", tenantID, name),
		&rp)
	return &rp, err
}

// TenantCreateAwsSecret creates an AWS secretsmanager secret via Duplo.
func (c *Client) TenantCreateAwsSecret(tenantID string, rq *DuploAwsSecretCreateRequest) (*DuploAwsSecretUpdatedResponse, ClientError) {
	rp := DuploAwsSecretUpdatedResponse{}
	err := c.postAPI(
		fmt.Sprintf("TenantCreateAwsSecret(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/secret", tenantID),
		&rq,
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// TenantUpdateAwsSecret updates an AWS secretsmanager secret via Duplo.
func (c *Client) TenantUpdateAwsSecret(tenantID, name string, rq *DuploAwsSecretUpdateRequest) (*DuploAwsSecretUpdatedResponse, ClientError) {
	rp := DuploAwsSecretUpdatedResponse{}
	err := c.putAPI(
		fmt.Sprintf("TenantCreateAwsSecret(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/secret/%s", tenantID, name),
		&rq,
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// TenantDeleteAwsSecret deletes a tenant secret via Duplo.
func (c *Client) TenantDeleteAwsSecret(tenantID string, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("TenantDeleteAwsSecret(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/secret/%s", tenantID, name),
		nil)
}
