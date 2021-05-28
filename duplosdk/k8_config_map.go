package duplosdk

import (
	"fmt"
)

// DuploK8sConfigMap represents a kubernetes config map in a Duplo tenant
type DuploK8sConfigMap struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"` //nolint:govet

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"-,omitempty"` //nolint:govet

	Data     map[string]interface{} `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// K8ConfigMapGetList retrieves a list of k8s config maps via the Duplo API.
func (c *Client) K8ConfigMapGetList(tenantID string) (*[]DuploK8sConfigMap, ClientError) {
	rp := []DuploK8sConfigMap{}
	err := c.getAPI(
		fmt.Sprintf("K8ConfigMapGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2", tenantID),
		&rp)

	// Add the tenant Id and name, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
			if name, ok := rp[i].Metadata["name"]; ok {
				rp[i].Name = name.(string)
			}
		}
	}
	return &rp, err
}

// K8ConfigMapGet retrieves a k8s config map via the Duplo API.
func (c *Client) K8ConfigMapGet(tenantID, name string) (*DuploK8sConfigMap, ClientError) {
	rp := DuploK8sConfigMap{}
	err := c.getAPI(
		fmt.Sprintf("K8ConfigMapGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2/%s", tenantID, name),
		&rp)

	// Add the tenant Id and name, then return the result.
	if err == nil {
		rp.TenantID = tenantID
		if name, ok := rp.Metadata["name"]; ok {
			rp.Name = name.(string)
		}
	}
	return &rp, err
}

// K8ConfigMapCreate creates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapCreate(tenantID string, rq *DuploK8sConfigMap) (*DuploK8sConfigMap, ClientError) {
	return c.K8ConfigMapCreateOrUpdate(tenantID, rq, false)
}

// K8ConfigMapUpdate updates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapUpdate(tenantID string, rq *DuploK8sConfigMap) (*DuploK8sConfigMap, ClientError) {
	return c.K8ConfigMapCreateOrUpdate(tenantID, rq, true)
}

// K8ConfigMapCreateOrUpdate creates or updates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapCreateOrUpdate(tenantID string, rq *DuploK8sConfigMap, updating bool) (*DuploK8sConfigMap, ClientError) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploK8sConfigMap{}
	err := c.doAPIWithRequestBody(
		verb,
		fmt.Sprintf("K8ConfigMapCreateOrUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// K8ConfigMapDelete deletes a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8ConfigMapDelete(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2/%s", tenantID, name),
		nil)
}
