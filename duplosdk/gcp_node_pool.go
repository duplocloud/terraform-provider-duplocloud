package duplosdk

import "fmt"

type DuploGKENodePoolRequest struct {
	Name        string             `json:"name"`
	MachineType string             `json:"machineType"`
	NumNodes    int                `json:"numNodes"`
	DiskSizeGb  int                `json:"diskSizeGb,omitempty"`
	Preemptible bool               `json:"preemptible,omitempty"`
	AutoScaling *AutoScalingConfig `json:"autoscaling,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Taints      []string           `json:"taints,omitempty"`
}

type AutoScalingConfig struct {
	MinNodes int `json:"minNodes"`
	MaxNodes int `json:"maxNodes"`
}

func (c *Client) GCPGKENodePoolCreate(tenantID string, rq *DuploGKENodePoolRequest) (*string, ClientError) {
	resp := ""
	err := c.postAPI(
		fmt.Sprintf("GCPGKENodePoolCreate(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/UpdateGCPAgentPool", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}
