package duplosdk

import (
	"fmt"
)

// DuploK8sSecret represents a kubernetes secret in a Duplo tenant
type DuploK8sSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

	IsDuploManaged    bool                   `json:"IsDuploManaged"`
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
		fmt.Sprintf("v3/subscriptions/%s/k8s/secret", tenantID),
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
	rp := DuploK8sSecret{}
	err := c.getAPI(
		fmt.Sprintf("K8SecretGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secret/%s", tenantID, secretName),
		&rp)

	if err != nil {
		return &rp, nil
	}

	// Add the tenant ID, then return the result.
	rp.TenantID = tenantID
	return &rp, nil
}

// K8SecretCreate creates a k8s secret via the Duplo API.
func (c *Client) K8SecretCreate(tenantID string, rq *DuploK8sSecret) (*DuploK8sSecret, ClientError) {
	rp := DuploK8sSecret{}
	err := c.postAPI(
		fmt.Sprintf("K8SecretCreate(%s, %s)", tenantID, rq.SecretName),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secret/%s", tenantID, rq.SecretName),
		&rq,
		&rp,
	)

	if err != nil {
		return &rp, nil
	}

	// Add the tenant ID, then return the result.
	rp.TenantID = tenantID
	return &rp, nil
}

// K8SecretUpdate updates a k8s secret via the Duplo API.
func (c *Client) K8SecretUpdate(tenantID string, rq *DuploK8sSecret) (*DuploK8sSecret, ClientError) {
	rp := DuploK8sSecret{}
	err := c.postAPI(
		fmt.Sprintf("K8SecretUpdate(%s, %s)", tenantID, rq.SecretName),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secret/%s", tenantID, rq.SecretName),
		&rq,
		&rp,
	)

	if err != nil {
		return &rp, nil
	}

	// Add the tenant ID, then return the result.
	rp.TenantID = tenantID
	return &rp, nil
}

// K8SecretDelete deletes a k8s secret via the Duplo API.
func (c *Client) K8SecretDelete(tenantID, secretName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8SecretDelete(%s, %s)", tenantID, secretName),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secret/%s", tenantID, secretName),
		nil)
}
