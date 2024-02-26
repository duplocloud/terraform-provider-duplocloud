package duplosdk

import (
	"fmt"
)

// DuploK8sSecret represents a kubernetes secret in a Duplo tenant
type DuploK8sSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

	SecretName        string                 `json:"SecretName"`
	SecretType        string                 `json:"SecretType"`
	SecretVersion     string                 `json:"SecretVersion,omitempty"`
	SecretData        map[string]interface{} `json:"SecretData"`
	SecretAnnotations map[string]string      `json:"SecretAnnotations,omitempty"`
	SecretLabels      map[string]string      `json:"SecretLabels,omitempty"`
}

// K8SecretGetList retrieves a list of k8s secrets via the Duplo API.
func (c *Client) K8SecretGetList(tenantID string) (*[]DuploK8sSecret, ClientError) {
	rp := []DuploK8sSecret{}
	err := c.getAPI(
		fmt.Sprintf("K8SecretGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAllK8Secrets", tenantID),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
		}
	}

	return &rp, err
}

// K8SecretGet retrieves a k8s secret via the Duplo API.
func (c *Client) K8SecretGet(tenantID, secretName string) (*DuploK8sSecret, ClientError) {
	// Retrieve the list of secrets
	list, err := c.K8SecretGetList(tenantID)
	if err != nil {
		return nil, newClientError(fmt.Sprintf("failed to get secret list: %s", err))
	}

	if list == nil {
		return nil, newClientError("secret list is nil")
	}

	// Return the secret, if it exists.
	for _, secret := range *list {
		if secret.SecretName == secretName {
			return &secret, nil
		}
	}

	return nil, newClientError(fmt.Sprintf("secret %s not found", secretName))
}

// K8SecretCreate creates a k8s secret via the Duplo API.
func (c *Client) K8SecretCreate(tenantID string, rq *DuploK8sSecret) ClientError {
	return c.K8SecretCreateOrUpdate(tenantID, rq, false)
}

// K8SecretUpdate updates a k8s secret via the Duplo API.
func (c *Client) K8SecretUpdate(tenantID string, rq *DuploK8sSecret) ClientError {
	return c.K8SecretCreateOrUpdate(tenantID, rq, true)
}

// K8SecretCreateOrUpdate creates or updates a k8s secret via the Duplo API.
func (c *Client) K8SecretCreateOrUpdate(tenantID string, rq *DuploK8sSecret, updating bool) ClientError {
	return c.postAPI(
		fmt.Sprintf("K8SecretCreateOrUpdate(%s, %s)", tenantID, rq.SecretName),
		fmt.Sprintf("subscriptions/%s/CreateOrUpdateK8Secret", tenantID),
		&rq,
		nil,
	)
}

// K8SecretDelete deletes a k8s secret via the Duplo API.
func (c *Client) K8SecretDelete(tenantID, secretName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8SecretDelete(%s, %s)", tenantID, secretName),
		fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantID, secretName),
		nil)
}
