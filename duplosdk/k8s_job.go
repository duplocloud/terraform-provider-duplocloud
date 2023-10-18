package duplosdk

import (
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sJob represents a kubernetes secret in a Duplo tenant
type DuploK8sJob struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string            `json:"-"` //nolint:govet
	Name     string            `json:"name"`
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     batchv1.JobSpec   `json:"spec"`
	Status   batchv1.JobStatus `json:"status"`
}

// K8sJobGetList retrieves a list of k8s jobs via the Duplo API.
func (c *Client) K8sJobGetList(tenantID string) (*[]DuploK8sJob, ClientError) {
	var rp []DuploK8sJob
	err := c.getAPI(
		fmt.Sprintf("k8sJobGetList(%s)", tenantID),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job", tenantID),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
		}
	}

	return &rp, err
}

// K8sJobGet retrieves a k8s job via the Duplo API.
func (c *Client) K8sJobGet(tenantID, jobName string) (*DuploK8sJob, ClientError) {
	var rp DuploK8sJob
	err := c.getAPI(
		fmt.Sprintf("k8sJobGet(%s, %s)", tenantID, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", tenantID, jobName),
		&rp)

	if err != nil {
		return nil, newClientError(fmt.Sprintf("job %s not found. %s", jobName, err))
	}
	// Add the tenant ID, then return the result.
	rp.TenantID = tenantID

	return &rp, err
}

// K8sJobCreate creates a k8s secret via the Duplo API.
func (c *Client) K8sJobCreate(tenantID string, rq *DuploK8sJob) ClientError {
	rp := DuploK8sJob{}
	return c.postAPI(
		fmt.Sprintf("k8sJobCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job", tenantID),
		&rq,
		&rp,
	)
}

// K8sJobUpdate updates a k8s job via the Duplo API.
func (c *Client) K8sJobUpdate(tenantID string, jobName string, rq *DuploK8sJob) ClientError {
	rp := DuploK8sJob{}
	return c.putAPI(
		fmt.Sprintf("k8sJobUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", tenantID, jobName),
		&rq,
		&rp,
	)
}

// K8sJobDelete deletes a k8s job via the Duplo API.
func (c *Client) K8sJobDelete(tenantID, jobName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8SecretDelete(%s, %s)", tenantID, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", tenantID, jobName),
		nil)
}
