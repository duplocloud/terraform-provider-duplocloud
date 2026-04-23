package duplosdk

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sDaemonSet represents a kubernetes DaemonSet in a Duplo tenant.
type DuploK8sDaemonSet struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId      string                 `json:"-"` //nolint:govet
	Metadata      metav1.ObjectMeta      `json:"metadata"`
	Spec          appsv1.DaemonSetSpec   `json:"spec"`
	Status        appsv1.DaemonSetStatus `json:"status,omitempty"`
	IsTenantLocal bool                   `json:"IsTenantLocal"`
}

// K8sDaemonSetGetList retrieves a list of k8s DaemonSets via the Duplo API.
func (c *Client) K8sDaemonSetGetList(tenantId string) (*[]DuploK8sDaemonSet, ClientError) {
	var rp []DuploK8sDaemonSet
	err := c.getAPI(
		fmt.Sprintf("k8sDaemonSetGetList(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet", tenantId),
		&rp)

	if err == nil {
		for i := range rp {
			rp[i].TenantId = tenantId
		}
	}

	return &rp, err
}

// K8sDaemonSetGet retrieves a k8s DaemonSet via the Duplo API.
func (c *Client) K8sDaemonSetGet(tenantId, name string) (*DuploK8sDaemonSet, ClientError) {
	var rp DuploK8sDaemonSet
	err := c.getAPI(
		fmt.Sprintf("k8sDaemonSetGet(%s, %s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantId, name),
		&rp)

	if err != nil {
		if err.Status() == 404 {
			return nil, nil
		}
		return nil, newClientError(fmt.Sprintf("daemonset %s not found. %s", name, err))
	}
	rp.TenantId = tenantId

	return &rp, nil
}

// K8sDaemonSetCreate creates a k8s DaemonSet via the Duplo API.
func (c *Client) K8sDaemonSetCreate(rq *DuploK8sDaemonSet) ClientError {
	rp := DuploK8sDaemonSet{}
	return c.postAPI(
		fmt.Sprintf("k8sDaemonSetCreate(%s, %s)", rq.TenantId, rq.Metadata.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet", rq.TenantId),
		&rq,
		&rp,
	)
}

// K8sDaemonSetUpdate updates a k8s DaemonSet via the Duplo API.
func (c *Client) K8sDaemonSetUpdate(tenantId, name string, rq *DuploK8sDaemonSet) ClientError {
	rp := DuploK8sDaemonSet{}
	return c.putAPI(
		fmt.Sprintf("k8sDaemonSetUpdate(%s, %s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantId, name),
		&rq,
		&rp,
	)
}

// K8sDaemonSetDelete deletes a k8s DaemonSet via the Duplo API.
func (c *Client) K8sDaemonSetDelete(tenantId, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8sDaemonSetDelete(%s, %s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantId, name),
		nil)
}
