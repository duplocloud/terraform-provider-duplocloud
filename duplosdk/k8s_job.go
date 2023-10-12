package duplosdk

import (
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sJob represents a kubernetes secret in a Duplo tenant
type DuploK8sJob struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

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

//// k8sJobCreate creates a k8s secret via the Duplo API.
//func (c *Client) k8sJobCreate(tenantID string, rq *DuploK8sSecret) ClientError {
//	return c.K8SecretCreateOrUpdate(tenantID, rq, false)
//}
//
//// k8sJobUpdate updates a k8s secret via the Duplo API.
//func (c *Client) k8sJobUpdate(tenantID string, rq *DuploK8sSecret) ClientError {
//	return c.K8SecretCreateOrUpdate(tenantID, rq, true)
//}
//
//// k8sJobCreateOrUpdate creates or updates a k8s secret via the Duplo API.
//func (c *Client) k8sJobCreateOrUpdate(tenantID string, rq *DuploK8sSecret, updating bool) ClientError {
//	return c.postAPI(
//		fmt.Sprintf("K8SecretCreateOrUpdate(%s, %s)", tenantID, rq.SecretName),
//		fmt.Sprintf("subscriptions/%s/CreateOrUpdateK8Secret", tenantID),
//		&rq,
//		nil,
//	)
//}

//// k8sJobDelete deletes a k8s secret via the Duplo API.
//func (c *Client) k8sJobDelete(tenantID, secretName string) ClientError {
//	return c.deleteAPI(
//		fmt.Sprintf("K8SecretDelete(%s, %s)", tenantID, secretName),
//		fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantID, secretName),
//		nil)
//}
