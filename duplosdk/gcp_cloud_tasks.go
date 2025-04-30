package duplosdk

import (
	"fmt"
)

type DuploGCPCloudTasks struct {
	TaskName    string            `json:"Name"`
	TaskType    int               `json:"TaskType"`
	Url         string            `json:"Url,omitempty"`
	RelativeUri string            `json:"RelativeUri,omitempty"`
	Method      string            `json:"HttpMethod"`
	Headers     map[string]string `json:"Headers"`
	Body        string            `json:"Body"`
}

type DuploGCPCloudTaskQueue struct {
	QueueName string `json:"Name"`
	Location  string `json:"Location"`
}

func (c *Client) GcpCloudTasksQueueCreate(tenantID string, rq *DuploGCPCloudTaskQueue) ClientError {
	var resp interface{}
	err := c.postAPI(
		fmt.Sprintf("GcpCloudTasksQueueCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/queues", tenantID),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) GcpCloudTasksCreate(tenantID string, queue string, rq *DuploGCPCloudTasks) ClientError {
	var resp interface{}
	err := c.postAPI(
		fmt.Sprintf("GcpCloudTasksQueueCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/queues/%s/tasks", tenantID, queue),
		&rq,
		&resp,
	)
	return err
}
func (c *Client) GCPCloudTasksQueueGet(tenantID string, name string) (*DuploGCPCloudTaskQueue, ClientError) {
	rp := DuploGCPCloudTaskQueue{}
	err := c.getAPI(
		fmt.Sprintf("GCPCloudTasksQueueGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/queues/%s", tenantID, name),
		&rp,
	)

	return &rp, err
}

func (c *Client) GCPCloudTasksGet(tenantID, queue, task string) (*DuploGCPCloudTasks, ClientError) {
	rp := DuploGCPCloudTasks{}
	err := c.getAPI(
		fmt.Sprintf("GCPCloudTasksGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/queues/%s/tasks/%s", tenantID, queue, task),
		&rp,
	)

	return &rp, err
}

func (c *Client) GCPCloudTasksDelete(tenantID, queue, task string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GCPCloudTasksDelete(%s, %s,%s)", tenantID, queue, task),
		fmt.Sprintf("v3/subscriptions/%s/google/queues/%s/tasks/%s", tenantID, queue, task),
		nil)
}

func (c *Client) GCPCloudTasksQueueDelete(tenantID, queue string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GCPCloudTasksQueueDelete(%s, %s)", tenantID, queue),
		fmt.Sprintf("v3/subscriptions/%s/google/queues/%s", tenantID, queue),
		nil)
}
