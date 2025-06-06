package duplosdk

import (
	"fmt"
	"log"
)

type DuploGCPK8NodePool struct {
	AutoRepair           bool                   `json:"AutoRepair"`
	AutoUpgrade          bool                   `json:"AutoUpgrade" default:"true"`
	AllocationTags       string                 `json:"AllocationTags"`
	DiscSizeGb           int                    `json:"DiscSizeGb" default:"100"`
	DiscType             string                 `json:"DiscType" default:"pd-standard"`
	ImageType            string                 `json:"ImageType" default:"cos_containerd"`
	InitialNodeCount     int                    `json:"InitialNodeCount"`
	IsAutoScalingEnabled bool                   `json:"IsAutoscalingEnabled"`
	Labels               map[string]string      `json:"Labels"`
	LinuxNodeConfig      *GCPLinuxNodeConfig    `json:"LinuxNodeConfig"`
	LocationPolicy       string                 `json:"LocationPolicy" default:"BALANCED"`
	LoggingConfig        *GCPLoggingConfig      `json:"LoggingConfig"`
	OauthScopes          []string               `json:"OauthScopes"`
	MachineType          string                 `json:"MachineType" default:"e2-medium"`
	MaxNodeCount         *int                   `json:"MaxNodeCount"`
	MinNodeCount         *int                   `json:"MinNodeCount"`
	Name                 string                 `json:"Name" valid:"required"`
	FullName             string                 `json:"-"`
	Spot                 bool                   `json:"Spot"`
	Tags                 []string               `json:"Tags"`
	Taints               []GCPNodeTaints        `json:"Taints"`
	UpgradeSettings      *GCPNodeUpgradeSetting `json:"UpgradeSettings"`
	Zones                []string               `json:"Zones"`
	Metadata             map[string]string      `json:"Metadata"`
	TotalMaxNodeCount    *int                   `json:"TotalMaxNodeCount"`
	TotalMinNodeCount    *int                   `json:"TotalMinNodeCount"`
	Accelerator          *Accelerator           `json:"Accelerator,omitempty"`
	ResourceLabels       map[string]string      `json:"ResourceLabels"`
	Status               string                 `json:"Status"`
}

type Accelerator struct {
	AcceleratorCount            int                         `json:"acceleratorCount"`
	AcceleratorType             string                      `json:"acceleratorType"`
	GPUPartitionSize            string                      `json:"gpuPartitionSize,omitempty"`
	GPUSharingConfig            GPUSharingConfig            `json:"gpuSharingConfig"`
	GPUDriverInstallationConfig GPUDriverInstallationConfig `json:"gpuDriverInstallationConfig"`
}

type GPUDriverInstallationConfig struct {
	GPUDriverVersion string `json:"gpuDriverVersion"`
}

type GPUSharingConfig struct {
	GPUSharingStrategy    string `json:"gpuSharingStrategy"`
	MaxSharedClientPerGPU int    `json:"maxSharedClientsPerGpu"`
}
type GCPNodeTaints struct {
	Key    string `json:"Key"`
	Value  string `json:"Value"`
	Effect string `json:"Effect"`
}

type GCPLinuxNodeConfig struct {
	CGroupMode string      `json:"cgroupMode" default:"CGROUP_MODE_UNSPECIFIED"`
	SysCtls    interface{} `json:"sysctls"`
}

type GCPLoggingConfig struct {
	VariantConfig *VariantConfig `json:"variantConfig"`
}

type VariantConfig struct {
	Variant string `json:"variant" default:"DEFAULT"`
}

type GCPNodeUpgradeSetting struct {
	Strategy          string             `json:"strategy" default:"NODE_POOL_UPDATE_STRATEGY_UNSPECIFIED"`
	MaxSurge          int                `json:"maxSurge"`
	MaxUnavailable    int                `json:"maxUnavailable"`
	BlueGreenSettings *BlueGreenSettings `json:"blueGreenSettings,omitempty"`
}

type BlueGreenSettings struct {
	StandardRolloutPolicy *StandardRolloutPolicy `json:"standardRolloutPolicy,omitempty"`
	NodePoolSoakDuration  string                 `json:"nodePoolSoakDuration,omitempty"`
}

type StandardRolloutPolicy struct {
	BatchPercentage   float32 `json:"batchPercentage,omitempty"`
	BatchNodeCount    int     `json:"batchNodeCount,omitempty"`
	BatchSoakDuration string  `json:"batchSoakDuration"`
}

func (c *Client) GCPK8NodePoolCreate(tenantID string, rq *DuploGCPK8NodePool) (*DuploGCPK8NodePool, ClientError) {
	log.Printf("[TRACE] \nNode pool request \n\n ******%+v\n*******", rq)
	resp := DuploGCPK8NodePool{}
	err := c.postAPI(
		fmt.Sprintf("GCPK8NodePoolCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/nodePools", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) GCPK8NodePoolGet(tenantID string, name string) (*DuploGCPK8NodePool, ClientError) {
	rp := DuploGCPK8NodePool{}
	err := c.getAPI(
		fmt.Sprintf("GCPK8NodePoolGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/nodePools/%s", tenantID, name),
		&rp,
	)

	return &rp, err
}

func (c *Client) GCPK8NodePoolList(tenantID string) (*[]DuploGCPK8NodePool, ClientError) {
	rp := []DuploGCPK8NodePool{}
	err := c.getAPI(
		fmt.Sprintf("GCPK8NodePoolList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/nodePools", tenantID),
		&rp,
	)

	return &rp, err
}

func (c *Client) GCPK8NodePoolDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GCPK8NodePoolDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/nodePools/%s", tenantID, name),
		nil)
}

func (c *Client) GCPK8NodePoolUpdate(tenantID, name, updateAttribute string, rq *DuploGCPK8NodePool) (*DuploGCPK8NodePool, ClientError) {
	rp := DuploGCPK8NodePool{}
	err := c.putAPI(
		fmt.Sprintf("GCPK8NodePoolUpdate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/nodePools/%s%s", tenantID, name, updateAttribute),
		&rq,
		&rp,
	)
	return &rp, err
}
