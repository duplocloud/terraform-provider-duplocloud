package duplosdk

import (
	"fmt"
	"time"
)

type DuploAwsLaunchTemplateRequest struct {
	LaunchTemplateName string `json:"LaunchTemplateName"`
	//SourceVersion      string                   `json:"SourceVersion"`
	VersionDescription string                   `json:"VersionDescription,omitempty"`
	LaunchTemplateData *DuploLaunchTemplateData `json:"LaunchTemplateData,omitempty"`
	DefaultVersion     string                   `json:"DefaultVersion,omitempty"`
}

type DuploLaunchTemplateData struct {
	InstanceType DuploStringValue `json:"InstanceType,omitempty"`
	ImageId      string           `json:"ImageId,omitempty"`
}

func (c *Client) CreateAwsLaunchTemplate(tenantId string, rq *DuploAwsLaunchTemplateRequest) ClientError {
	rp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("CreateAwsLaunchTemplate(%s, %s)", tenantId, rq.LaunchTemplateName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/launchtemplateversion", tenantId),
		rq,
		&rp,
	)
	return err
}

func (c *Client) UpdateAwsLaunchTemplateVersion(tenantId string, rq *DuploAwsLaunchTemplateRequest) ClientError {
	rp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("UpdateAwsLaunchTemplateVersion(%s, %s)", tenantId, rq.LaunchTemplateName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/launchtemplate", tenantId),
		rq,
		&rp,
	)
	return err
}

func (c *Client) GetAwsLaunchTemplate(tenantID string, asgName string) (*[]DuploLaunchTemplateResponse, ClientError) {
	rp := []DuploLaunchTemplateResponse{}
	err := c.getAPI(
		fmt.Sprintf("GetAwsLaunchTemplate(%s, %s)", tenantID, asgName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/%s/launchtemplateversions", tenantID, asgName),
		&rp)
	return &rp, err
}

type DuploLaunchTemplateResponse struct {
	LaunchTemplateId   string                              `json:"LaunchTemplateId"`
	LaunchTemplateName string                              `json:"LaunchTemplateName"`
	VersionNumber      int64                               `json:"VersionNumber"`
	VersionDescription string                              `json:"VersionDescription"`
	CreateTime         time.Time                           `json:"CreateTime"`
	CreatedBy          string                              `json:"CreatedBy"`
	DefaultVersion     bool                                `json:"DefaultVersion"`
	LaunchTemplateData DuploLaunchTemplateDataResponse     `json:"LaunchTemplateData"`
	Operator           DuploLaunchTemplateOperatorResponse `json:"Operator,omitempty"`
}

type DuploLaunchTemplateDataResponse struct {
	KernelId                          string                                              `json:"KernelId,omitempty"`
	EbsOptimized                      bool                                                `json:"EbsOptimized"`
	IamInstanceProfile                DuploLaunchTemplateIamInstanceProfileSpecification  `json:"IamInstanceProfile"`
	BlockDeviceMappings               []DuploLaunchTemplateBlockDeviceMapping             `json:"BlockDeviceMappings"`
	NetworkInterfaces                 []DuploLaunchTemplateNetworkInterfaceSpecification  `json:"NetworkInterfaces"`
	ImageId                           string                                              `json:"ImageId"`
	InstanceType                      DuploKeyStringValue                                 `json:"InstanceType"`
	KeyName                           string                                              `json:"KeyName"`
	Monitoring                        DuploLaunchTemplateMonitoring                       `json:"Monitoring"`
	Placement                         DuploLaunchTemplatePlacement                        `json:"Placement"`
	RamDiskId                         string                                              `json:"RamDiskId,omitempty"`
	DisableApiTermination             bool                                                `json:"DisableApiTermination"`
	InstanceInitiatedShutdownBehavior string                                              `json:"InstanceInitiatedShutdownBehavior"`
	UserData                          string                                              `json:"UserData"`
	TagSpecifications                 []DuploLaunchTemplateTagSpecification               `json:"TagSpecifications"`
	ElasticGpuSpecifications          []DuploLaunchTemplateElasticGpuSpecification        `json:"ElasticGpuSpecifications"`
	ElasticInferenceAccelerators      []DuploLaunchTemplateElasticInferenceAccelerator    `json:"ElasticInferenceAccelerators"`
	SecurityGroupIds                  []string                                            `json:"SecurityGroupIds"`
	SecurityGroups                    []string                                            `json:"SecurityGroups"`
	InstanceMarketOptions             DuploLaunchTemplateInstanceMarketOptions            `json:"InstanceMarketOptions"`
	CreditSpecification               DuploLaunchTemplateCreditSpecification              `json:"CreditSpecification"`
	CpuOptions                        DuploLaunchTemplateCpuOptions                       `json:"CpuOptions"`
	CapacityReservationSpecification  DuploLaunchTemplateCapacityReservationSpecification `json:"CapacityReservationSpecification"`
	LicenseSpecifications             []DuploLaunchTemplateLicenseConfiguration           `json:"LicenseSpecifications"`
	HibernationOptions                DuploLaunchTemplateHibernationOptions               `json:"HibernationOptions"`
	MetadataOptions                   DuploLaunchTemplateInstanceMetadataOptions          `json:"MetadataOptions"`
	EnclaveOptions                    DuploLaunchTemplateEnclaveOptions                   `json:"EnclaveOptions"`
	InstanceRequirements              DuploLaunchTemplateInstanceRequirements             `json:"InstanceRequirements"`
	PrivateDnsNameOptions             DuploLaunchTemplatePrivateDnsNameOptions            `json:"PrivateDnsNameOptions"`
	MaintenanceOptions                DuploLaunchTemplateMaintenanceOptions               `json:"MaintenanceOptions"`
	DisableApiStop                    bool                                                `json:"DisableApiStop"`
	Operator                          DuploLaunchTemplateOperatorResponse                 `json:"Operator,omitempty"`
}

type DuploLaunchTemplateIamInstanceProfileSpecification struct {
	Arn  string `json:"Arn,omitempty"`
	Name string `json:"Name,omitempty"`
}

type DuploLaunchTemplateBlockDeviceMapping struct {
	DeviceName  string                            `json:"DeviceName"`
	VirtualName string                            `json:"VirtualName,omitempty"`
	Ebs         DuploLaunchTemplateEbsBlockDevice `json:"Ebs,omitempty"`
	NoDevice    string                            `json:"NoDevice,omitempty"`
}

type DuploLaunchTemplateEbsBlockDevice struct {
	Encrypted           bool             `json:"Encrypted"`
	DeleteOnTermination bool             `json:"DeleteOnTermination"`
	Iops                int              `json:"Iops,omitempty"`
	KmsKeyId            string           `json:"KmsKeyId,omitempty"`
	SnapshotId          string           `json:"SnapshotId,omitempty"`
	VolumeSize          int              `json:"VolumeSize,omitempty"`
	VolumeType          DuploStringValue `json:"VolumeType"`
	Throughput          int              `json:"Throughput,omitempty"`
}

type DuploLaunchTemplateNetworkInterfaceSpecification struct {
	AssociateCarrierIpAddress bool                                               `json:"AssociateCarrierIpAddress,omitempty"`
	AssociatePublicIpAddress  bool                                               `json:"AssociatePublicIpAddress,omitempty"`
	DeleteOnTermination       bool                                               `json:"DeleteOnTermination,omitempty"`
	Description               string                                             `json:"Description,omitempty"`
	DeviceIndex               int                                                `json:"DeviceIndex,omitempty"`
	Groups                    []string                                           `json:"Groups,omitempty"`
	InterfaceType             string                                             `json:"InterfaceType,omitempty"`
	Ipv6Addresses             []DuploLaunchTemplateInstanceIpv6Address           `json:"Ipv6Addresses,omitempty"`
	PrivateIpAddresses        []DuploLaunchTemplatePrivateIpAddressSpecification `json:"PrivateIpAddresses,omitempty"`
	SubnetId                  string                                             `json:"SubnetId,omitempty"`
}

type DuploLaunchTemplateInstanceIpv6Address struct {
	Ipv6Address   string `json:"Ipv6Address"`
	IsPrimaryIpv6 bool   `json:"IsPrimaryIpv6,omitempty"`
}

type DuploLaunchTemplatePrivateIpAddressSpecification struct {
	Primary          bool   `json:"Primary"`
	PrivateIpAddress string `json:"PrivateIpAddress"`
}

type DuploLaunchTemplateMonitoring struct {
	Enabled bool `json:"Enabled"`
}

type DuploLaunchTemplatePlacement struct {
	AvailabilityZone string `json:"AvailabilityZone,omitempty"`
	HostId           string `json:"HostId,omitempty"`
	Tenancy          string `json:"Tenancy,omitempty"`
}

type DuploLaunchTemplateTagSpecification struct {
	ResourceType string                   `json:"ResourceType,omitempty"`
	Tags         []DuploLaunchTemplateTag `json:"Tags"`
}

type DuploLaunchTemplateTag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type DuploLaunchTemplateElasticGpuSpecification struct {
	Type string `json:"Type"`
}

type DuploLaunchTemplateElasticInferenceAccelerator struct {
	Type  string `json:"Type"`
	Count int    `json:"Count"`
}

type DuploLaunchTemplateInstanceMarketOptions struct {
	MarketType  string                               `json:"MarketType,omitempty"`
	SpotOptions DuploLaunchTemplateSpotMarketOptions `json:"SpotOptions,omitempty"`
}

type DuploLaunchTemplateSpotMarketOptions struct {
	MaxPrice         string `json:"MaxPrice,omitempty"`
	SpotInstanceType string `json:"SpotInstanceType,omitempty"`
}

type DuploLaunchTemplateCreditSpecification struct {
	CpuCredits string `json:"CpuCredits"`
}

type DuploLaunchTemplateCpuOptions struct {
	CoreCount int `json:"CoreCount"`
}

type DuploLaunchTemplateCapacityReservationSpecification struct {
	CapacityReservationPreference string `json:"CapacityReservationPreference"`
}

type DuploLaunchTemplateLicenseConfiguration struct {
	LicenseConfigurationArn string `json:"LicenseConfigurationArn"`
}

type DuploLaunchTemplateHibernationOptions struct {
	Configured bool `json:"Configured"`
}

type DuploLaunchTemplateInstanceMetadataOptions struct {
	HttpEndpoint DuploStringValue `json:"HttpEndpoint"`
	HttpTokens   DuploStringValue `json:"HttpTokens"`
}

type DuploLaunchTemplateEnclaveOptions struct {
	Enabled bool `json:"Enabled"`
}

type DuploLaunchTemplateInstanceRequirements struct {
	VCpuCount DuploLaunchTemplateVCpuCountRange `json:"VCpuCount"`
}

type DuploLaunchTemplateVCpuCountRange struct {
	Min int `json:"Min"`
	Max int `json:"Max"`
}

type DuploLaunchTemplatePrivateDnsNameOptions struct {
	HostnameType string `json:"HostnameType,omitempty"`
}

type DuploLaunchTemplateMaintenanceOptions struct {
	AutoRecovery string `json:"AutoRecovery"`
}

type DuploLaunchTemplateOperatorResponse struct {
	Managed   bool   `json:"Managed,omitempty"`
	Principal string `json:"Principal,omitempty"`
}
