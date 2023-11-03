package duplosdk

import (
	"fmt"
	"k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sCronJob represents a kubernetes job in a Duplo tenant
type DuploK8sCronJob struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId string                `json:"-"` //nolint:govet
	Metadata metav1.ObjectMeta     `json:"metadata"`
	Spec     v1beta1.CronJobSpec   `json:"spec"`
	Status   v1beta1.CronJobStatus `json:"status"`
}

// K8sCronJobGetList retrieves a list of k8s jobs via the Duplo API.
func (c *Client) K8sCronJobGetList(tenantId string) (*[]DuploK8sCronJob, ClientError) {
	var rp []DuploK8sCronJob
	err := c.getAPI(
		fmt.Sprintf("k8sCronJobGetList(%s)", tenantId),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob", tenantId),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantId = tenantId
		}
	}

	return &rp, err
}

// K8sCronJobGet retrieves a k8s job via the Duplo API.
func (c *Client) K8sCronJobGet(tenantId, jobName string) (*DuploK8sCronJob, ClientError) {
	var rp DuploK8sCronJob
	err := c.getAPI(
		fmt.Sprintf("k8sCronJobGet(%s, %s)", tenantId, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob/%s", tenantId, jobName),
		&rp)

	if err != nil {
		return nil, newClientError(fmt.Sprintf("cronjob %s not found. %s", jobName, err))
	}
	// Add the tenant ID, then return the result.
	rp.TenantId = tenantId

	return &rp, err
}

// K8sCronJobCreate creates a k8s job via the Duplo API.
func (c *Client) K8sCronJobCreate(rq *DuploK8sCronJob) ClientError {
	rp := DuploK8sCronJob{}
	return c.postAPI(
		fmt.Sprintf("k8sCronJobCreate(%s, %s)", rq.TenantId, rq.Metadata.Name),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob", rq.TenantId),
		&rq,
		&rp,
	)
}

// K8sCronJobUpdate updates a k8s job via the Duplo API.
func (c *Client) K8sCronJobUpdate(tenantId string, jobName string, rq *DuploK8sCronJob) ClientError {
	rp := DuploK8sCronJob{}
	return c.putAPI(
		fmt.Sprintf("k8sCronJobUpdate(%s, %s)", tenantId, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob/%s", tenantId, jobName),
		&rq,
		&rp,
	)
}

// K8sCronJobDelete deletes a k8s job via the Duplo API.
func (c *Client) K8sCronJobDelete(tenantId, jobName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8sCronJobDelete(%s, %s)", tenantId, jobName),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob/%s", tenantId, jobName),
		nil)
}
