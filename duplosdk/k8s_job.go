package duplosdk

import (
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sJob represents a kubernetes job in a Duplo tenant
type DuploK8sJob struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId string            `json:"-"` //nolint:govet
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     batchv1.JobSpec   `json:"spec"`
	Status   batchv1.JobStatus `json:"status"`
}

// K8sJobGetList retrieves a list of k8s jobs via the Duplo API.
func (c *Client) K8sJobGetList(tenantId string) (*[]DuploK8sJob, ClientError) {
	var rp []DuploK8sJob
	err := c.getAPI(
		fmt.Sprintf("k8sJobGetList(%s)", tenantId),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job", tenantId),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantId = tenantId
		}
	}

	return &rp, err
}

// K8sJobGet retrieves a k8s job via the Duplo API.
func (c *Client) K8sJobGet(tenantId, jobName string) (*DuploK8sJob, ClientError) {
	var rp DuploK8sJob
	err := c.getAPI(
		fmt.Sprintf("k8sJobGet(%s, %s)", tenantId, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", tenantId, jobName),
		&rp)

	if err != nil {
		return nil, newClientError(fmt.Sprintf("job %s not found. %s", jobName, err))
	}
	// Add the tenant ID, then return the result.
	rp.TenantId = tenantId

	return &rp, err
}

// K8sJobCreate creates a k8s job via the Duplo API.
func (c *Client) K8sJobCreate(rq *DuploK8sJob) ClientError {
	rp := DuploK8sJob{}
	return c.postAPI(
		fmt.Sprintf("k8sJobCreate(%s, %s)", rq.TenantId, rq.Metadata.Name),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job", rq.TenantId),
		&rq,
		&rp,
	)
}

// K8sJobUpdate updates a k8s job via the Duplo API.
func (c *Client) K8sJobUpdate(rq *DuploK8sJob) ClientError {
	rp := DuploK8sJob{}
	return c.putAPI(
		fmt.Sprintf("k8sJobUpdate(%s, %s)", rq.TenantId, rq.Metadata.Name),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", rq.TenantId, rq.Metadata.Name),
		&rq,
		&rp,
	)
}

// K8sJobDelete deletes a k8s job via the Duplo API.
func (c *Client) K8sJobDelete(tenantId, jobName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8sJobDelete(%s, %s)", tenantId, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/job/%s", tenantId, jobName),
		nil)
}
