package duplosdk

import (
	"fmt"
)

type DuploSubResource struct {
	Id string `json:"id,omitempty"`
}
type DuploAzureVirtualMachineScaleSetSku struct {
	Capacity int    `json:"capacity,omitempty"`
	Tier     string `json:"tier,omitempty"`
	Name     string `json:"name,omitempty"`
}

type DuploAzureVirtualMachineScaleSetUpgradePolicy struct {
	Mode                     string                                                    `json:"mode,omitempty"`
	RollingUpgradePolicy     *DuploAzureVirtualMachineScaleSetRollingUpgradePolicy     `json:"rollingUpgradePolicy,omitempty"`
	AutomaticOSUpgradePolicy *DuploAzureVirtualMachineScaleSetAutomaticOSUpgradePolicy `json:"automaticOSUpgradePolicy,omitempty"`
}

type DuploAzureVirtualMachineScaleSetRollingUpgradePolicy struct {
	MaxBatchInstancePercent             int    `json:"maxBatchInstancePercent,omitempty"`
	MaxUnhealthyInstancePercent         int    `json:"maxUnhealthyInstancePercent,omitempty"`
	MaxUnhealthyUpgradedInstancePercent int    `json:"maxUnhealthyUpgradedInstancePercent,omitempty"`
	PauseTimeBetweenBatches             string `json:"pauseTimeBetweenBatches,omitempty"`
}

type DuploAzureVirtualMachineScaleSetAutomaticOSUpgradePolicy struct {
	EnableAutomaticOSUpgrade bool `json:"enableAutomaticOSUpgrade,omitempty"`
	DisableAutomaticRollback bool `json:"disableAutomaticRollback,omitempty"`
}

type DuploAzureVirtualMachineScaleSetIdentityValue struct {
	PrincipalId string `json:"principalId,omitempty"`
	ClientId    string `json:"clientId,omitempty"`
}

type DuploAzureVirtualMachineScaleSetIdentity struct {
	PrincipalId            string                                         `json:"principalId,omitempty"`
	TenantId               string                                         `json:"tenantId,omitempty"`
	Type                   string                                         `json:"type,omitempty"`
	UserAssignedIdentities *DuploAzureVirtualMachineScaleSetIdentityValue `json:"userAssignedIdentities,omitempty"`
}

type DuploOSProfileWinRMListener struct {
	Protocol       string `json:"protocol,omitempty"`
	CertificateUrl string `json:"certificateUrl,omitempty"`
}

type DuploOSProfileWinRMConfiguration struct {
	Listeners *[]DuploOSProfileWinRMListener `json:"listeners,omitempty"`
}

type DuploWinConfigAdditionalUnattendContent struct {
	PassName      string `json:"passName,omitempty"`
	ComponentName string `json:"componentName,omitempty"`
	SettingName   string `json:"settingName,omitempty"`
	Content       string `json:"content,omitempty"`
}

type DuploOSProfileWindowsConfiguration struct {
	ProvisionVMAgent          bool                                       `json:"provisionVMAgent,omitempty"`
	EnableAutomaticUpdates    bool                                       `json:"enableAutomaticUpdates,omitempty"`
	TimeZone                  string                                     `json:"timeZone,omitempty"`
	WinRM                     *DuploOSProfileWinRMConfiguration          `json:"winRM,omitempty"`
	AdditionalUnattendContent *[]DuploWinConfigAdditionalUnattendContent `json:"additionalUnattendContent,omitempty"`
}

type DuploSshPublicKey struct {
	Path    string `json:"path,omitempty"`
	KeyData string `json:"keyData,omitempty"`
}

type DuploOSProfileSshConfiguration struct {
	PublicKeys *[]DuploSshPublicKey
}

type DuploOSProfileLinuxConfiguration struct {
	DisablePasswordAuthentication bool                            `json:"disablePasswordAuthentication,omitempty"`
	ProvisionVMAgent              bool                            `json:"provisionVMAgent,omitempty"`
	Ssh                           *DuploOSProfileSshConfiguration `json:"ssh,omitempty"`
}

type DuploVaultCertificate struct {
	CertificateUrl   string `json:"certificateUrl,omitempty"`
	CertificateStore string `json:"certificateStore,omitempty"`
}

type DuploVaultSecretGroup struct {
	SourceVault       *DuploSubResource        `json:"sourceVault,omitempty"`
	VaultCertificates *[]DuploVaultCertificate `json:"vaultCertificates,omitempty"`
}

type DuploVirtualMachineScaleSetOSProfile struct {
	ComputerNamePrefix   string                              `json:"computerNamePrefix,omitempty"`
	AdminUsername        string                              `json:"adminUsername,omitempty"`
	AdminPassword        string                              `json:"adminPassword,omitempty"`
	CustomData           string                              `json:"customData,omitempty"`
	WindowsConfiguration *DuploOSProfileWindowsConfiguration `json:"windowsConfiguration,omitempty"`
	LinuxConfiguration   *DuploOSProfileLinuxConfiguration   `json:"linuxConfiguration,omitempty"`
	Secrets              *[]DuploVaultSecretGroup            `json:"secrets,omitempty"`
}

type DuploStorageProfileImageReference struct {
	Id        string `json:"id,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Offer     string `json:"offer,omitempty"`
	Sku       string `json:"sku,omitempty"`
	Version   string `json:"version,omitempty"`
}

type DuploDiffDiskSettings struct {
	Option    string `json:"option,omitempty"`
	Placement string `json:"placement,omitempty"`
}

type DuploVirtualHardDisk struct {
	Uri string `json:"uri,omitempty"`
}

type DuploVirtualMachineScaleSetManagedDiskParameters struct {
	StorageAccountType string            `json:"storageAccountType,omitempty"`
	DiskEncryptionSet  *DuploSubResource `json:"diskEncryptionSet,omitempty"`
}

type DuploVirtualMachineScaleSetOSDisk struct {
	Name                    string                                            `json:"name,omitempty"`
	Caching                 string                                            `json:"caching,omitempty"`
	WriteAcceleratorEnabled bool                                              `json:"writeAcceleratorEnabled,omitempty"`
	CreateOption            string                                            `json:"createOption,omitempty"`
	DiffDiskSettings        *DuploDiffDiskSettings                            `json:"diffDiskSettings,omitempty"`
	DiskSizeGB              int                                               `json:"diskSizeGB,omitempty"`
	OsType                  string                                            `json:"osType,omitempty"`
	Image                   *DuploVirtualHardDisk                             `json:"image,omitempty"`
	VhdContainers           []string                                          `json:"vhdContainers,omitempty"`
	ManagedDisk             *DuploVirtualMachineScaleSetManagedDiskParameters `json:"managedDisk,omitempty"`
}

type DuploVirtualMachineScaleSetDataDisk struct {
	Name                    string                                            `json:"name,omitempty"`
	Lun                     int                                               `json:"lun,omitempty"`
	Caching                 string                                            `json:"caching,omitempty"`
	WriteAcceleratorEnabled bool                                              `json:"writeAcceleratorEnabled,omitempty"`
	CreateOption            string                                            `json:"createOption,omitempty"`
	DiskSizeGB              int                                               `json:"diskSizeGB,omitempty"`
	ManagedDisk             *DuploVirtualMachineScaleSetManagedDiskParameters `json:"managedDisk,omitempty"`
	DiskIOPSReadWrite       int                                               `json:"diskIOPSReadWrite,omitempty"`
	DiskMBpsReadWrite       int                                               `json:"diskMBpsReadWrite,omitempty"`
}

type DuploVirtualMachineScaleSetStorageProfile struct {
	ImageReference *DuploStorageProfileImageReference     `json:"imageReference,omitempty"`
	OsDisk         *DuploVirtualMachineScaleSetOSDisk     `json:"osDisk,omitempty"`
	DataDisks      *[]DuploVirtualMachineScaleSetDataDisk `json:"dataDisks,omitempty"`
}

type DuploApiEntityReference struct {
	Id string `json:"id,omitempty"`
}

type DuploVirtualMachineScaleSetNetworkConfigurationDnsSettings struct {
	DnsServers []string `json:"dnsServers,omitempty"`
}

type DuploVirtualMachineScaleSetPublicIPAddressConfigurationDnsSettings struct {
	DomainNameLabel string `json:"domainNameLabel,omitempty"`
}

type DuploVirtualMachineScaleSetIpTag struct {
	IpTagType string `json:"ipTagType,omitempty"`
	Tag       string `json:"tag,omitempty"`
}

type DuploVirtualMachineScaleSetPublicIPAddressConfiguration struct {
	Name                   string                                                              `json:"name,omitempty"`
	IdleTimeoutInMinutes   int                                                                 `json:"idleTimeoutInMinutes,omitempty"`
	DnsSettings            *DuploVirtualMachineScaleSetPublicIPAddressConfigurationDnsSettings `json:"dnsSettings,omitempty"`
	IpTags                 *[]DuploVirtualMachineScaleSetIpTag                                 `json:"ipTags,omitempty"`
	PublicIPPrefix         *DuploSubResource                                                   `json:"publicIPPrefix,omitempty"`
	PublicIPAddressVersion string                                                              `json:"publicIPAddressVersion,omitempty"`
}

type DuploVirtualMachineScaleSetIPConfiguration struct {
	Name                                  string
	Subnet                                *DuploApiEntityReference
	Primary                               bool
	PublicIPAddressConfiguration          *DuploVirtualMachineScaleSetPublicIPAddressConfiguration
	PrivateIPAddressVersion               string
	ApplicationGatewayBackendAddressPools *[]DuploSubResource
	ApplicationSecurityGroups             *[]DuploSubResource
	LoadBalancerBackendAddressPools       *[]DuploSubResource
	LoadBalancerInboundNatPools           *[]DuploSubResource
}

type DuploVirtualMachineScaleSetNetworkConfiguration struct {
	Name                        string                                                      `json:"name,omitempty"`
	Primary                     bool                                                        `json:"primary,omitempty"`
	EnableAcceleratedNetworking bool                                                        `json:"enableAcceleratedNetworking,omitempty"`
	NetworkSecurityGroup        *DuploSubResource                                           `json:"networkSecurityGroup,omitempty"`
	DnsSettings                 *DuploVirtualMachineScaleSetNetworkConfigurationDnsSettings `json:"dnsSettings,omitempty"`
	IpConfigurations            *[]DuploVirtualMachineScaleSetIPConfiguration               `json:"ipConfigurations,omitempty"`
	EnableIPForwarding          bool                                                        `json:"enableIPForwarding,omitempty"`
}

type DuploVirtualMachineScaleSetNetworkProfile struct {
	HealthProbe                    *DuploApiEntityReference                           `json:"healthProbe,omitempty"`
	NetworkInterfaceConfigurations *[]DuploVirtualMachineScaleSetNetworkConfiguration `json:"networkInterfaceConfigurations,omitempty"`
}

type DuploSecurityProfile struct {
	EncryptionAtHost bool `json:"encryptionAtHost,omitempty"`
}

type DuploBootDiagnostics struct {
	Enabled    bool   `json:"enabled,omitempty"`
	StorageUri string `json:"storageUri,omitempty"`
}

type DuploDiagnosticsProfile struct {
	BootDiagnostics *DuploBootDiagnostics `json:"bootDiagnostics,omitempty"`
}

type DuploBillingProfile struct {
	MaxPrice float64 `json:"maxPrice,omitempty"`
}

type DuploTerminateNotificationProfile struct {
	NotBeforeTimeout string `json:"notBeforeTimeout,omitempty"`
	Enable           bool   `json:"enable,omitempty"`
}

type DuploScheduledEventsProfile struct {
	TerminateNotificationProfile *DuploTerminateNotificationProfile `json:"terminateNotificationProfile,omitempty"`
}

type DuploAzureScaleSetVirtualMachineProfile struct {
	OsProfile              *DuploVirtualMachineScaleSetOSProfile      `json:"osProfile,omitempty"`
	StorageProfile         *DuploVirtualMachineScaleSetStorageProfile `json:"storageProfile,omitempty"`
	NetworkProfile         *DuploVirtualMachineScaleSetNetworkProfile `json:"networkProfile,omitempty"`
	SecurityProfile        *DuploSecurityProfile                      `json:"securityProfile,omitempty"`
	DiagnosticsProfile     *DuploDiagnosticsProfile                   `json:"diagnosticsProfile,omitempty"`
	LicenseType            string                                     `json:"licenseType,omitempty"`
	Priority               string                                     `json:"priority,omitempty"`
	EvictionPolicy         string                                     `json:"evictionPolicy,omitempty"`
	BillingProfile         *DuploBillingProfile                       `json:"billingProfile,omitempty"`
	ScheduledEventsProfile *DuploScheduledEventsProfile               `json:"scheduledEventsProfile,omitempty"`
}

type DuploAzureVirtualMachineScaleSet struct {
	ID                                     string                                                    `json:"id,omitempty"`
	Type                                   string                                                    `json:"type,omitempty"`
	Location                               string                                                    `json:"location,omitempty"`
	Name                                   string                                                    `json:"name,omitempty"`
	Sku                                    *DuploAzureVirtualMachineScaleSetSku                      `json:"sku,omitempty"`
	UpgradePolicy                          *DuploAzureVirtualMachineScaleSetAutomaticOSUpgradePolicy `json:"properties.upgradePolicy,omitempty"`
	DoNotRunExtensionsOnOverprovisionedVMs bool                                                      `json:"properties.doNotRunExtensionsOnOverprovisionedVMs,omitempty"`
	Overprovision                          bool                                                      `json:"properties.overprovision,omitempty"`
	ProvisioningState                      string                                                    `json:"properties.provisioningState,omitempty"`
	SinglePlacementGroup                   bool                                                      `json:"properties.singlePlacementGroup,omitempty"`
	UniqueId                               string                                                    `json:"properties.uniqueId,omitempty"`
	ZoneBalance                            bool                                                      `json:"properties.zoneBalance,omitempty"`
	PlatformFaultDomainCount               int                                                       `json:"properties.platformFaultDomainCount,omitempty"`
	Zones                                  []string                                                  `json:"zones,omitempty"`
	Identity                               *DuploAzureVirtualMachineScaleSetIdentity                 `json:"identity,omitempty"`
	VirtualMachineProfile                  *DuploAzureScaleSetVirtualMachineProfile                  `json:"virtualMachineProfile,omitempty"`
	NameEx                                 string                                                    `json:"NameEx,omitempty"`
	IsMinion                               bool                                                      `json:"IsMinion"`
	AgentPlatform                          int                                                       `json:"AgentPlatform"`
	AllocationTags                         string                                                    `json:"AllocationTags,omitempty"`
}

func (c *Client) AzureVirtualMachineScaleSetCreate(tenantID string, rq *DuploAzureVirtualMachineScaleSet) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureVirtualMachineScaleSetCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/CreateAzureVmScaleSetSync", tenantID),
		&rq,
		nil,
	)
}

func (c *Client) AzureVirtualMachineScaleSetGet(tenantID string, name string) (*DuploAzureVirtualMachineScaleSet, ClientError) {
	list, err := c.AzureVirtualMachineScaleSetList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Name == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureVirtualMachineScaleSetList(tenantID string) (*[]DuploAzureVirtualMachineScaleSet, ClientError) {
	rp := []DuploAzureVirtualMachineScaleSet{}
	err := c.getAPI(
		fmt.Sprintf("AzureVirtualMachineScaleSetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetVirtualMachineScaleSetsSync", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) AzureVirtualMachineScaleSetExists(tenantID, name string) (bool, ClientError) {
	list, err := c.AzureVirtualMachineScaleSetList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, element := range *list {
			if element.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) AzureVirtualMachineScaleSetDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("AzureVirtualMachineScaleSetDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteAzureVmScaleSetSync/%s", tenantID, name),
		nil,
		nil,
	)
}
