package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
func (c *Client) TenantListSecrets(tenantID string) (*[]DuploTenantSecret, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/ListTenantSecrets", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantListSecrets 1 ********: %s ", url)

	// Get the secrets from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantListSecrets 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantListSecrets 3 ********: %s", bodyString)

	// Return it as an object.
	duploObjects := make([]DuploTenantSecret, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetAwsCredentials 4 ********")
	for i := range duploObjects {
		duploObjects[i].TenantID = tenantID
	}
	return &duploObjects, nil
}

// TenantGetSecretByName retrieves a managed secret via the Duplo API
func (c *Client) TenantGetSecretByName(tenantID string, name string) (*DuploTenantSecret, error) {
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
func (c *Client) TenantGetSecretByNameSuffix(tenantID string, nameSuffix string) (*DuploTenantSecret, error) {
	allSecrets, err := c.TenantListSecrets(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific name.
	for _, secret := range *allSecrets {
		nameParts := strings.SplitN(secret.Name, "-", 3)
		if len(nameParts) == 3 && nameParts[2] == nameSuffix {
			return &secret, nil
		}
	}

	// No secret was found.
	return nil, nil
}

// TenantCreateSecret creates a tenant secret via Duplo.
func (c *Client) TenantCreateSecret(tenantID string, duplo *DuploTenantSecretRequest) error {
	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantCreateSecret 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/CreateTenantSecret", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantCreateSecret 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantCreateSecret 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantCreateSecret 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to create secret %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to create secret %s: '%s'", tenantID, duplo.Name, bodyString)
}

// TenantDeleteSecret deletes a tenant secret via Duplo.
func (c *Client) TenantDeleteSecret(tenantID string, name string) error {
	// Build the request
	rqBody, err := json.Marshal(map[string]interface{}{
		"ForceDeleteWithoutRecovery": true,
		"SecretId":                   name,
	})
	if err != nil {
		log.Printf("[TRACE] TenantDeleteSecret 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/deleteTenantSecret", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantDeleteSecret 2 : %s", url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantDeleteSecret 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantDeleteSecret 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to delete secret %s: '%s'", tenantID, name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to delete secret %s: '%s'", tenantID, name, bodyString)
}
