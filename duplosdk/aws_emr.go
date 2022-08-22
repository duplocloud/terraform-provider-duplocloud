package duplosdk

import (
	"fmt"
	"time"
)

// DuploEmrClusterRequest is a Duplo SDK object that represents a emr cluster
type DuploEmrClusterRequest struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID     string `json:"-"`
	Arn          string `json:"Arn,omitempty"`
	Name         string `json:"Name,omitempty"`
	ReleaseLabel string `json:"ReleaseLabel,omitempty"`
	Status       string `json:"Status,omitempty"`
	LogUri       string `json:"LogUri,omitempty"`

	JobFlowId    string `json:"JobFlowId,omitempty"`
	ResourceType int    `json:"ResourceType,omitempty"`
	CustomAmiId  string `json:"CustomAmiId,omitempty"`

	EbsRootVolumeSize           int    `json:"EbsRootVolumeSize,omitempty"`
	StepConcurrencyLevel        int    `json:"StepConcurrencyLevel,omitempty"`
	ScaleDownBehavior           string `json:"ScaleDownBehavior,omitempty"`
	TerminationProtection       bool   `json:"TerminationProtection,omitempty"`
	KeepJobFlowAliveWhenNoSteps bool   `json:"KeepJobFlowAliveWhenNoSteps,omitempty"`
	VisibleToAllUsers           bool   `json:"VisibleToAllUsers,omitempty"`

	//ec2
	MasterInstanceType string `json:"MasterInstanceType,omitempty"`
	SlaveInstanceType  string `json:"SlaveInstanceType,omitempty"`
	InstanceCount      int    `json:"InstanceCount,omitempty"`
	//can we use this for subnetid selection
	Zone int `json:"Zone,omitempty"`

	//JSON str
	Applications           string `json:"Applications,omitempty"`
	Steps                  string `json:"Steps,omitempty"`
	Configurations         string `json:"Configurations,omitempty"`
	BootstrapActions       string `json:"BootstrapActions,omitempty"`
	JobFlowInstancesConfig string `json:"JobFlowInstancesConfig,omitempty"`
	//JSON str
	AdditionalInfo       string `json:"AdditionalInfo,omitempty"`
	ManagedScalingPolicy string `json:"ManagedScalingPolicy,omitempty"`
	InstanceGroups       string `json:"InstanceGroups,omitempty"`
	InstanceFleets       string `json:"InstanceFleets,omitempty"`

	//== debug ec2-attributes
	MetaData string `json:"MetaData,omitempty"`
	State    string `json:"State,omitempty"`
}

type DuploEmrClusterSummary struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID  string `json:"-"`
	Name      string `json:"Name,omitempty"`
	Arn       string `json:"Arn,omitempty"`
	Status    string `json:"Status,omitempty"`
	JobFlowId string `json:"JobFlowId,omitempty"`
}

type DuploEmrClusterGetResponse struct {
	TenantID       string `json:"-"`
	InstanceGroups []struct {
		Configurations        []interface{} `json:"Configurations"`
		ConfigurationsVersion int           `json:"ConfigurationsVersion"`
		EbsBlockDevices       []interface{} `json:"EbsBlockDevices"`
		EbsOptimized          bool          `json:"EbsOptimized"`
		ID                    string        `json:"Id"`
		InstanceGroupType     struct {
			Value string `json:"Value"`
		} `json:"InstanceGroupType"`
		InstanceType                                 string        `json:"InstanceType"`
		LastSuccessfullyAppliedConfigurations        []interface{} `json:"LastSuccessfullyAppliedConfigurations"`
		LastSuccessfullyAppliedConfigurationsVersion int           `json:"LastSuccessfullyAppliedConfigurationsVersion"`
		Market                                       struct {
			Value string `json:"Value"`
		} `json:"Market"`
		Name                   string `json:"Name"`
		RequestedInstanceCount int    `json:"RequestedInstanceCount"`
		RunningInstanceCount   int    `json:"RunningInstanceCount"`
		ShrinkPolicy           struct {
			DecommissionTimeout int `json:"DecommissionTimeout"`
		} `json:"ShrinkPolicy"`
		Status struct {
			State struct {
				Value string `json:"Value"`
			} `json:"State"`
			StateChangeReason struct {
				Message string `json:"Message"`
			} `json:"StateChangeReason"`
			Timeline struct {
				CreationDateTime time.Time `json:"CreationDateTime"`
				EndDateTime      string    `json:"EndDateTime"`
				ReadyDateTime    string    `json:"ReadyDateTime"`
			} `json:"Timeline"`
		} `json:"Status"`
	} `json:"InstanceGroups"`
	Instances []struct {
		EbsVolumes      []interface{} `json:"EbsVolumes"`
		Ec2InstanceID   string        `json:"Ec2InstanceId"`
		ID              string        `json:"Id"`
		InstanceGroupID string        `json:"InstanceGroupId"`
		InstanceType    string        `json:"InstanceType"`
		Market          struct {
			Value string `json:"Value"`
		} `json:"Market"`
		PrivateDNSName   string `json:"PrivateDnsName"`
		PrivateIPAddress string `json:"PrivateIpAddress"`
		PublicDNSName    string `json:"PublicDnsName"`
		Status           struct {
			State struct {
				Value string `json:"Value"`
			} `json:"State"`
			StateChangeReason struct {
			} `json:"StateChangeReason"`
			Timeline struct {
				CreationDateTime time.Time `json:"CreationDateTime"`
				EndDateTime      string    `json:"EndDateTime"`
				ReadyDateTime    string    `json:"ReadyDateTime"`
			} `json:"Timeline"`
		} `json:"Status"`
	} `json:"Instances"`
	ReleaseLabel                string `json:"ReleaseLabel"`
	AutoTerminate               bool   `json:"AutoTerminate"`
	MasterPublicDNSName         string `json:"MasterPublicDnsName"`
	NormalizedInstanceHours     int    `json:"NormalizedInstanceHours"`
	LogURI                      string `json:"LogUri"`
	StepConcurrencyLevel        int    `json:"StepConcurrencyLevel"`
	TerminationProtection       bool   `json:"TerminationProtection"`
	KeepJobFlowAliveWhenNoSteps bool   `json:"KeepJobFlowAliveWhenNoSteps"`
	VisibleToAllUsers           bool   `json:"VisibleToAllUsers"`
	Applications                []struct {
		AdditionalInfo struct {
		} `json:"AdditionalInfo"`
		Args    []interface{} `json:"Args,omitempty"`
		Name    string        `json:"Name,omitempty"`
		Version string        `json:"Version,omitempty"`
	} `json:"Applications"`
	KerberosAttributes struct {
	} `json:"KerberosAttributes"`
	InstanceCount     int    `json:"InstanceCount"`
	EbsRootVolumeSize int    `json:"EbsRootVolumeSize"`
	ScaleDownBehavior string `json:"ScaleDownBehavior"`
	Zone              int    `json:"Zone"`
	JobFlowID         string `json:"JobFlowId"`
	ClusterTerminated bool   `json:"ClusterTerminated"`
	Arn               string `json:"Arn"`
	Status            string `json:"Status"`
	ResourceType      int    `json:"ResourceType"`
	MetaDataObject    struct {
		AdditionalMasterSecurityGroups []string      `json:"AdditionalMasterSecurityGroups"`
		AdditionalSlaveSecurityGroups  []string      `json:"AdditionalSlaveSecurityGroups"`
		Ec2AvailabilityZone            string        `json:"Ec2AvailabilityZone"`
		Ec2KeyName                     string        `json:"Ec2KeyName"`
		Ec2SubnetID                    string        `json:"Ec2SubnetId"`
		EmrManagedMasterSecurityGroup  string        `json:"EmrManagedMasterSecurityGroup"`
		EmrManagedSlaveSecurityGroup   string        `json:"EmrManagedSlaveSecurityGroup"`
		IamInstanceProfile             string        `json:"IamInstanceProfile"`
		RequestedEc2AvailabilityZones  []interface{} `json:"RequestedEc2AvailabilityZones"`
		RequestedEc2SubnetIds          []string      `json:"RequestedEc2SubnetIds"`
		ServiceAccessSecurityGroup     string        `json:"ServiceAccessSecurityGroup"`
	} `json:"MetaDataObject"`
	Name                 string           `json:"Name"`
	CustomAmiId          string           `json:"CustomAmiId"`
	MasterInstanceType   string           `json:"MasterInstanceType,omitempty"`
	SlaveInstanceType    string           `json:"SlaveInstanceType,omitempty"`
	BootstrapActions     *BootstrapAction `json:"BootstrapActions,omitempty"`
	Configurations       interface{}      `json:"Configurations,omitempty"`
	Steps                interface{}      `json:"Steps,omitempty"`
	AdditionalInfo       string           `json:"AdditionalInfo,omitempty"`
	ManagedScalingPolicy interface{}      `json:"ManagedScalingPolicy,omitempty"`
	InstanceFleets       interface{}      `json:"InstanceFleets,omitempty"`
}

type BootstrapAction struct {
	Name       string   `json:"Name,omitempty"`
	ScriptPath string   `json:"ScriptPath,omitempty"`
	Args       []string `json:"Args,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// DuploEmrClusterCreate creates an emr cluster via the Duplo API.
func (c *Client) DuploEmrClusterCreate(tenantID string, rq *DuploEmrClusterRequest) (*DuploEmrClusterRequest, ClientError) {
	rp := DuploEmrClusterRequest{}
	err := c.postAPI(
		fmt.Sprintf("DuploEmrClusterCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster", tenantID),
		&rq,
		&rp,
	)
	rp.TenantID = tenantID
	return &rp, err
}

// DuploEmrClusterDelete deletes an emr cluster via the Duplo API.
func (c *Client) DuploEmrClusterDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploEmrClusterDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster/%s", tenantID, name),
		nil)
}

// DuploEmrClusterGet retrieves an emr cluster via the Duplo API
func (c *Client) DuploEmrClusterGet(tenantID string, name string) (*DuploEmrClusterGetResponse, ClientError) {
	rp := DuploEmrClusterGetResponse{}
	err := c.getAPI(
		fmt.Sprintf("DuploEmrClusterGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	return &rp, err
}

// DuploEmrClusterGetList retrieves a emr cluster via the Duplo API
func (c *Client) DuploEmrClusterGetList(tenantID string) (*[]DuploEmrClusterSummary, ClientError) {
	// todo: not tested data
	rp := []DuploEmrClusterSummary{}
	err := c.getAPI(
		fmt.Sprintf("DuploEmrClusterGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/emrCluster", tenantID),
		&rp)
	return &rp, err
}
