package duplosdk

import (
	"fmt"
)

// DuploReplicationController represents a service in the Duplo SDK
type DuploReplicationController struct {
	Name                              string                 `json:"Name"`
	Replicas                          int                    `json:"Replicas"`
	ReplicasPrev                      int                    `json:"ReplicasPrev,omitempty"`
	ReplicasMatchingAsgName           string                 `json:"ReplicasMatchingAsgName,omitempty"`
	DnsPrfx                           string                 `json:"DnsPrfx"`
	ElbDnsName                        string                 `json:"ElbDnsName"`
	Fqdn                              string                 `json:"Fqdn"`
	ParentDomain                      string                 `json:"ParentDomain"`
	IsInfraDeployment                 bool                   `json:"IsInfraDeployment,omitempty"`
	IsDaemonset                       bool                   `json:"IsDaemonset,omitempty"`
	IsLBSyncedDeployment              bool                   `json:"IsLBSyncedDeployment,omitempty"`
	IsReplicaCollocationAllowed       bool                   `json:"IsReplicaCollocationAllowed,omitempty"`
	IsAnyHostAllowed                  bool                   `json:"IsAnyHostAllowed,omitempty"`
	IsCloudCredsFromK8sServiceAccount bool                   `json:"IsCloudCredsFromK8sServiceAccount,omitempty"`
	Volumes                           string                 `json:"Volumes,omitempty"`
	Template                          *DuploPodTemplate      `json:"Template,omitempty"`
	Tags                              *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploPodTemplate represents a pod template in the Duplo SDK
type DuploPodTemplate struct {
	Name                  string                           `json:"Name"`
	Containers            *[]DuploPodContainer             `json:"Containers,omitempty"`
	Interfaces            *[]DuploPodInterface             `json:"Interfaces,omitempty"`
	AgentPlatform         int                              `json:"AgentPlatform"`
	Cloud                 int                              `json:"Cloud"`
	Volumes               string                           `json:"Volumes,omitempty"`
	Commands              []string                         `json:"Commands"`
	ApplicationUrl        string                           `json:"ApplicationUrl,omitempty"`
	SecondaryTenant       string                           `json:"SecondaryTenant,omitempty"`
	ExtraConfig           string                           `json:"ExtraConfig,omitempty"`
	OtherDockerConfig     string                           `json:"OtherDockerConfig,omitempty"`
	OtherDockerHostConfig string                           `json:"OtherDockerHostConfig,omitempty"`
	DeviceIds             []string                         `json:"DeviceIds"`
	BaseVersion           string                           `json:"BaseVersion,omitempty"`
	LbConfigsVersion      string                           `json:"LbConfigsVersion,omitempty"`
	ImageUpdateTime       string                           `json:"ImageUpdateTime,omitempty"`
	AllocationTags        string                           `json:"AllocationTags,omitempty"`
	IsReadOnly            bool                             `json:"IsReadOnly,omitempty"`
	LBCCount              int                              `json:"LBCCount,omitempty"`
	LBConfigurations      map[string]*DuploLbConfiguration `json:"LBConfigurations,omitempty"`
}

// DuploPodContainer represents a container within a pod template in the Duplo SDK
type DuploPodContainer struct {
	Name       string `json:"Name"`
	Image      string `json:"Image"`
	InstanceId string `json:"InstanceId,omitempty"`
	DockerId   string `json:"DockerId,omitempty"`
}

// DuploPodContainer represents a network interface within a pod template in the Duplo SDK
type DuploPodInterface struct {
	NetworkId       string `json:"NetworkId"`
	IpAddress       string `json:"IpAddress,omitempty"`
	ExternalAddress string `json:"ExternalAddress,omitempty"`
}

// DuploPodLbConfiguration represents an LB configuration within a pod template in the Duplo SDK
type DuploLbConfiguration struct {
	TenantId                  string                    `json:"TenantId"`
	ReplicationControllerName string                    `json:"ReplicationControllerName"`
	LbType                    int                       `json:"LbType"`
	Protocol                  string                    `json:"Protocol"`
	Port                      string                    `json:"Port"`
	HostPort                  int                       `json:"HostPort"`
	ExternalPort              int                       `json:"ExternalPort"`
	TgCount                   int                       `json:"TgCount,omitempty"`
	IsInfraDeployment         bool                      `json:"IsInfraDeployment,omitempty"`
	DnsName                   string                    `json:"DnsName,omitempty"`
	CertificateArn            string                    `json:"CertificateArn,omitempty"`
	CloudName                 string                    `json:"CloudName,omitempty"`
	HealthCheckUrl            string                    `json:"HealthCheckUrl,omitempty"`
	ExternalTrafficPolicy     string                    `json:"ExternalTrafficPolicy,omitempty"`
	BeProtocolVersion         string                    `json:"BeProtocolVersion,omitempty"`
	FrontendIp                string                    `json:"FrontendIp,omitempty"`
	IsInternal                bool                      `json:"IsInternal,omitempty"`
	ForHealthCheck            bool                      `json:"ForHealthCheck,omitempty"`
	IsNative                  bool                      `json:"IsNative,omitempty"`
	HealthCheckConfig         *DuploLbHealthCheckConfig `json:"HealthCheckConfig,omitempty"`

	// TODO: DIPAddresses
}

type DuploLbHealthCheckConfig struct {
	HealthyThresholdCount           int    `json:"HealthyThresholdCount"`
	UnhealthyThresholdCount         int    `json:"UnhealthyThresholdCount"`
	HealthCheckTimeoutSeconds       int    `json:"HealthCheckTimeoutSeconds"`
	LbHealthCheckIntervalSecondsype int    `json:"HealthCheckIntervalSeconds"`
	HttpSuccessCode                 string `json:"HttpSuccessCode,omitempty"`
	GrpcSuccessCode                 string `json:"GrpcSuccessCode,omitempty"`
}

type DuploReplicationControllerCreateRequest struct {
	TenantId                          string                 `json:"TenantId"`
	Name                              string                 `json:"Name"`
	Image                             string                 `json:"DockerImage"`
	NetworkId                         string                 `json:"NetworkId"`
	Cloud                             int                    `json:"Cloud"`
	AgentPlatform                     int                    `json:"AgentPlatform"`
	Replicas                          int                    `json:"Replicas,omitempty"`
	ReplicasMatchingAsgName           string                 `json:"ReplicasMatchingAsgName,omitempty"`
	IsDaemonset                       bool                   `json:"IsDaemonset"`
	IsLBSyncedDeployment              bool                   `json:"IsLBSyncedDeployment"`
	IsReplicaCollocationAllowed       bool                   `json:"IsReplicaCollocationAllowed"`
	IsAnyHostAllowed                  bool                   `json:"IsAnyHostAllowed"`
	IsCloudCredsFromK8sServiceAccount bool                   `json:"IsCloudCredsFromK8sServiceAccount"`
	AllocationTags                    string                 `json:"AllocationTags,omitempty"`
	Volumes                           string                 `json:"Volumes,omitempty"`
	ExtraConfig                       string                 `json:"ExtraConfig,omitempty"`
	OtherDockerConfig                 string                 `json:"OtherDockerConfig,omitempty"`
	OtherDockerHostConfig             string                 `json:"OtherDockerHostConfig,omitempty"`
	Tags                              *[]DuploKeyStringValue `json:"Tags,omitempty"`

	// TODO: Test this field
	Commands string `json:"Commands,omitempty"`

	// TODO: DeviceIds
}

type DuploReplicationControllerUpdateRequest struct {
	Name                              string `json:"Name"`
	Image                             string `json:"Image"`
	AgentPlatform                     int    `json:"AgentPlatform"`
	Replicas                          int    `json:"Replicas,omitempty"`
	ReplicasMatchingAsgName           string `json:"ReplicasMatchingAsgName,omitempty"`
	IsDaemonset                       bool   `json:"IsDaemonset"`
	IsLBSyncedDeployment              bool   `json:"IsLBSyncedDeployment"`
	IsReplicaCollocationAllowed       bool   `json:"IsReplicaCollocationAllowed"`
	IsAnyHostAllowed                  bool   `json:"IsAnyHostAllowed"`
	IsCloudCredsFromK8sServiceAccount bool   `json:"IsCloudCredsFromK8sServiceAccount"`
	AllocationTags                    string `json:"AllocationTags,omitempty"`
	Volumes                           string `json:"Volumes,omitempty"`
	ExtraConfig                       string `json:"ExtraConfig,omitempty"`
	OtherDockerConfig                 string `json:"OtherDockerConfig,omitempty"`
	OtherDockerHostConfig             string `json:"OtherDockerHostConfig,omitempty"`
}

type DuploReplicationControllerDeleteRequest struct {
	TenantId      string `json:"TenantId,omitempty"`
	Name          string `json:"Name"`
	NetworkId     string `json:"NetworkId,omitempty"`
	AgentPlatform int    `json:"AgentPlatform,omitempty"`
	Image         string `json:"DockerImage,omitempty"`
	State         string `json:"State"`
}

// ReplicationControllerList retrieves a list of replication controllers via the Duplo API.
func (c *Client) ReplicationControllerList(tenantID string) (*[]DuploReplicationController, ClientError) {
	rp := []DuploReplicationController{}
	err := c.getAPI(fmt.Sprintf("ReplicationControllerList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetReplicationControllers", tenantID),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// ReplicationControllerGet retrieves a replication controller via the Duplo API.
func (c *Client) ReplicationControllerGet(tenantID, name string) (*DuploReplicationController, ClientError) {
	allResources, err := c.ReplicationControllerList(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the resource with the specific type and name.
	for _, resource := range *allResources {
		if resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// ReplicationControllerCreate creates a replication controller via the Duplo API.
func (c *Client) ReplicationControllerCreate(tenantID string, rq *DuploReplicationControllerCreateRequest) ClientError {
	if rq.NetworkId == "" {
		rq.NetworkId = "default"
	}
	return c.postAPI(
		fmt.Sprintf("ReplicationControllerCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/ReplicationControllerUpdate", tenantID),
		&rq,
		nil,
	)
}

// ReplicationControllerUpdate creates a replication controller via the Duplo API.
func (c *Client) ReplicationControllerUpdate(tenantID string, rq *DuploReplicationControllerUpdateRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("ReplicationControllerUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/ReplicationControllerChangeAll", tenantID),
		&rq,
		nil,
	)
}

// ReplicationControllerDelete deletes a replication controller via the Duplo API.
func (c *Client) ReplicationControllerDelete(tenantID string, rq *DuploReplicationControllerDeleteRequest) ClientError {
	rq.TenantId = tenantID
	rq.State = "delete"
	if rq.NetworkId == "" {
		rq.NetworkId = "default"
	}

	return c.postAPI(
		fmt.Sprintf("ReplicationControllerDelete(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/ReplicationControllerUpdate", tenantID),
		&rq,
		nil,
	)
}
