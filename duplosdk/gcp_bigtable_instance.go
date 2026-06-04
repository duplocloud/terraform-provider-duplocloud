package duplosdk

import (
	"fmt"
	"log"
)

// Bigtable enum values mirror Google.Cloud.Bigtable.Admin.V2 protos, which the
// Duplo backend serializes as integers.
//
// Instance.Types.Type:        TYPE_UNSPECIFIED=0, PRODUCTION=1, DEVELOPMENT=2
// Instance.Types.State:       STATE_NOT_KNOWN=0, READY=1, CREATING=2
// StorageType:                STORAGE_TYPE_UNSPECIFIED=0, SSD=1, HDD=2
// Cluster.Types.State:        NOT_KNOWN=0, READY=1, CREATING=2, RESIZING=3, DISABLED=4
const (
	BigtableTypeUnspecified = 0
	BigtableTypeProduction  = 1
	BigtableTypeDevelopment = 2

	BigtableStorageUnspecified = 0
	BigtableStorageSSD         = 1
	BigtableStorageHDD         = 2

	BigtableStateReady = 1
)

// DuploBigtableAutoscalingLimits represents the min/max serve node bounds for
// a cluster's autoscaling configuration.
type DuploBigtableAutoscalingLimits struct {
	MinServeNodes int `json:"minServeNodes"`
	MaxServeNodes int `json:"maxServeNodes"`
}

// DuploBigtableAutoscalingTargets represents the utilization targets that drive
// autoscaling for a cluster.
type DuploBigtableAutoscalingTargets struct {
	CpuUtilizationPercent        int `json:"cpuUtilizationPercent,omitempty"`
	StorageUtilizationGibPerNode int `json:"storageUtilizationGibPerNode,omitempty"`
}

// DuploBigtableClusterAutoscalingConfig holds the autoscaling limits and targets.
type DuploBigtableClusterAutoscalingConfig struct {
	AutoscalingLimits  DuploBigtableAutoscalingLimits  `json:"autoscalingLimits"`
	AutoscalingTargets DuploBigtableAutoscalingTargets `json:"autoscalingTargets"`
}

// DuploBigtableClusterConfig wraps the autoscaling config (the "config" oneof on
// the Bigtable Cluster proto). When absent, the cluster uses manual node scaling.
type DuploBigtableClusterConfig struct {
	ClusterAutoscalingConfig *DuploBigtableClusterAutoscalingConfig `json:"clusterAutoscalingConfig,omitempty"`
}

// DuploBigtableCluster represents a Bigtable cluster (Google.Cloud.Bigtable.Admin.V2.Cluster).
type DuploBigtableCluster struct {
	Name               string                      `json:"name,omitempty"`
	Location           string                      `json:"location,omitempty"`
	State              int                         `json:"state,omitempty"`
	ServeNodes         int                         `json:"serveNodes,omitempty"`
	DefaultStorageType int                         `json:"defaultStorageType,omitempty"`
	ClusterConfig      *DuploBigtableClusterConfig `json:"clusterConfig,omitempty"`
}

// DuploBigtableInstance represents a Bigtable instance (Google.Cloud.Bigtable.Admin.V2.Instance).
type DuploBigtableInstance struct {
	Name        string            `json:"name,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	State       int               `json:"state,omitempty"`
	Type        int               `json:"type,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// DuploBigtableCreateInstanceRequest is the body for creating a Bigtable instance
// along with its initial set of clusters (Google.Cloud.Bigtable.Admin.V2.CreateInstanceRequest).
type DuploBigtableCreateInstanceRequest struct {
	InstanceId string                          `json:"instanceId"`
	Instance   DuploBigtableInstance           `json:"instance"`
	Clusters   map[string]DuploBigtableCluster `json:"clusters"`
}

func bigtablePrefix(tenantID string) string {
	return fmt.Sprintf("v3/subscriptions/%s/google/bigtable", tenantID)
}

// GcpBigtableInstanceCreate creates a Bigtable instance with its initial clusters.
// The backend returns the name of the long-running create operation.
func (c *Client) GcpBigtableInstanceCreate(tenantID string, rq *DuploBigtableCreateInstanceRequest) (string, ClientError) {
	log.Printf("[TRACE] GcpBigtableInstanceCreate(%s, %s)", tenantID, rq.InstanceId)
	resp := ""
	err := c.postAPI(
		fmt.Sprintf("GcpBigtableInstanceCreate(%s, %s)", tenantID, rq.InstanceId),
		bigtablePrefix(tenantID),
		&rq,
		&resp,
	)
	return resp, err
}

// GcpBigtableInstanceGet retrieves a Bigtable instance by its instance ID. The
// backend returns a null body (zero-value struct) when the instance is missing.
func (c *Client) GcpBigtableInstanceGet(tenantID, instanceID string) (*DuploBigtableInstance, ClientError) {
	rp := DuploBigtableInstance{}
	err := c.getAPI(
		fmt.Sprintf("GcpBigtableInstanceGet(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("%s/%s", bigtablePrefix(tenantID), instanceID),
		&rp,
	)
	return &rp, err
}

// GcpBigtableInstanceList lists the Bigtable instances belonging to the tenant.
func (c *Client) GcpBigtableInstanceList(tenantID string) (*[]DuploBigtableInstance, ClientError) {
	rp := []DuploBigtableInstance{}
	err := c.getAPI(
		fmt.Sprintf("GcpBigtableInstanceList(%s)", tenantID),
		bigtablePrefix(tenantID),
		&rp,
	)
	return &rp, err
}

// GcpBigtableInstanceUpdate updates the instance-level fields (display name, type,
// labels). The backend polls the operation until it completes.
func (c *Client) GcpBigtableInstanceUpdate(tenantID, instanceID string, rq *DuploBigtableInstance) (*DuploBigtableInstance, ClientError) {
	rp := DuploBigtableInstance{}
	err := c.putAPI(
		fmt.Sprintf("GcpBigtableInstanceUpdate(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("%s/%s", bigtablePrefix(tenantID), instanceID),
		&rq,
		&rp,
	)
	return &rp, err
}

// GcpBigtableInstanceDelete deletes a Bigtable instance and all of its clusters.
// The backend responds with a plain confirmation string, so capture it instead
// of passing a nil response target (which would treat the body as unexpected).
func (c *Client) GcpBigtableInstanceDelete(tenantID, instanceID string) ClientError {
	resp := ""
	return c.deleteAPI(
		fmt.Sprintf("GcpBigtableInstanceDelete(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("%s/%s", bigtablePrefix(tenantID), instanceID),
		&resp,
	)
}

// GcpBigtableClusterCreate adds a cluster to an existing Bigtable instance. The
// backend returns the name of the long-running create operation.
func (c *Client) GcpBigtableClusterCreate(tenantID, instanceID, clusterID string, rq *DuploBigtableCluster) (string, ClientError) {
	log.Printf("[TRACE] GcpBigtableClusterCreate(%s, %s, %s)", tenantID, instanceID, clusterID)
	resp := ""
	err := c.postAPI(
		fmt.Sprintf("GcpBigtableClusterCreate(%s, %s, %s)", tenantID, instanceID, clusterID),
		fmt.Sprintf("%s/%s/clusters/%s", bigtablePrefix(tenantID), instanceID, clusterID),
		&rq,
		&resp,
	)
	return resp, err
}

// GcpBigtableClusterGet retrieves a single cluster of a Bigtable instance.
func (c *Client) GcpBigtableClusterGet(tenantID, instanceID, clusterID string) (*DuploBigtableCluster, ClientError) {
	rp := DuploBigtableCluster{}
	err := c.getAPI(
		fmt.Sprintf("GcpBigtableClusterGet(%s, %s, %s)", tenantID, instanceID, clusterID),
		fmt.Sprintf("%s/%s/clusters/%s", bigtablePrefix(tenantID), instanceID, clusterID),
		&rp,
	)
	return &rp, err
}

// GcpBigtableClusterList lists all clusters belonging to a Bigtable instance.
func (c *Client) GcpBigtableClusterList(tenantID, instanceID string) (*[]DuploBigtableCluster, ClientError) {
	rp := []DuploBigtableCluster{}
	err := c.getAPI(
		fmt.Sprintf("GcpBigtableClusterList(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("%s/%s/clusters", bigtablePrefix(tenantID), instanceID),
		&rp,
	)
	return &rp, err
}

// GcpBigtableClusterUpdate updates a cluster's serve node count and/or autoscaling
// configuration.
func (c *Client) GcpBigtableClusterUpdate(tenantID, instanceID, clusterID string, rq *DuploBigtableCluster) ClientError {
	rp := DuploBigtableCluster{}
	return c.putAPI(
		fmt.Sprintf("GcpBigtableClusterUpdate(%s, %s, %s)", tenantID, instanceID, clusterID),
		fmt.Sprintf("%s/%s/clusters/%s", bigtablePrefix(tenantID), instanceID, clusterID),
		&rq,
		&rp,
	)
}

// GcpBigtableClusterDelete removes a cluster from a Bigtable instance. The
// backend responds with a plain confirmation string, so capture it instead of
// passing a nil response target (which would treat the body as unexpected).
func (c *Client) GcpBigtableClusterDelete(tenantID, instanceID, clusterID string) ClientError {
	resp := ""
	return c.deleteAPI(
		fmt.Sprintf("GcpBigtableClusterDelete(%s, %s, %s)", tenantID, instanceID, clusterID),
		fmt.Sprintf("%s/%s/clusters/%s", bigtablePrefix(tenantID), instanceID, clusterID),
		&resp,
	)
}
