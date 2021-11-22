package duplosdk

import (
	"fmt"
)

// DuploNativeHost is a Duplo SDK object that represents an nativehost
type DuploNativeHost struct {
	InstanceID         string                             `json:"InstanceId"`
	UserAccount        string                             `json:"UserAccount,omitempty"`
	TenantID           string                             `json:"TenantId,omitempty"`
	FriendlyName       string                             `json:"FriendlyName,omitempty"`
	Capacity           string                             `json:"Capacity,omitempty"`
	Zone               int                                `json:"Zone"`
	IsMinion           bool                               `json:"IsMinion"`
	ImageID            string                             `json:"ImageId,omitempty"`
	Base64UserData     string                             `json:"Base64UserData,omitempty"`
	AgentPlatform      int                                `json:"AgentPlatform"`
	IsEbsOptimized     bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP  bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud              int                                `json:"Cloud"`
	EncryptDisk        bool                               `json:"EncryptDisk,omitempty"`
	Status             string                             `json:"Status,omitempty"`
	IdentityRole       string                             `json:"IdentityRole,omitempty"`
	PrivateIPAddress   string                             `json:"PrivateIpAddress,omitempty"`
	NetworkInterfaceId string                             `json:"NetworkInterfaceId,omitempty"`
	NetworkInterfaces  *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes            *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData           *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	Tags               *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	MinionTags         *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
}

// DuploNativeHostNetworkInterface is a Duplo SDK object that represents a network interface of a native host
type DuploNativeHostNetworkInterface struct {
	NetworkInterfaceID string                 `json:"NetworkInterfaceId,omitempty"`
	SubnetID           string                 `json:"SubnetId,omitempty"`
	AssociatePublicIP  bool                   `json:"AssociatePublicIpAddress,omitempty"`
	Groups             *[]string              `json:"Groups,omitempty"`
	DeviceIndex        int                    `json:"DeviceIndex,omitempty"`
	MetaData           *[]DuploKeyStringValue `json:"MetaData,omitempty"`
}

// DuploNativeHostVolume is a Duplo SDK object that represents a volume of a native host
type DuploNativeHostVolume struct {
	Iops       int    `json:"Iops,omitempty"`
	Name       string `json:"Name,omitempty"`
	Size       int    `Size:"Size,omitempty"`
	VolumeID   string `json:"VolumeId,omitempty"`
	VolumeType string `json:"VolumeType,omitempty"`
}

type DuploAzureVitualMachine struct {
	PropertiesHardwareProfile struct {
		VMSize string `json:"vmSize"`
	} `json:"properties.hardwareProfile"`
	PropertiesStorageProfile struct {
		ImageReference struct {
			Publisher    string `json:"publisher"`
			Offer        string `json:"offer"`
			Sku          string `json:"sku"`
			Version      string `json:"version"`
			ExactVersion string `json:"exactVersion"`
		} `json:"imageReference"`
		OsDisk struct {
			OsType       string `json:"osType"`
			Name         string `json:"name"`
			Caching      string `json:"caching"`
			CreateOption string `json:"createOption"`
			DiskSizeGB   int    `json:"diskSizeGB"`
			ManagedDisk  struct {
				StorageAccountType string `json:"storageAccountType"`
				ID                 string `json:"id"`
			} `json:"managedDisk"`
		} `json:"osDisk"`
		DataDisks []interface{} `json:"dataDisks"`
	} `json:"properties.storageProfile"`
	PropertiesAdditionalCapabilities struct {
		UltraSSDEnabled bool `json:"ultraSSDEnabled"`
	} `json:"properties.additionalCapabilities"`
	PropertiesOsProfile struct {
		ComputerName       string `json:"computerName"`
		AdminUsername      string `json:"adminUsername"`
		LinuxConfiguration struct {
			DisablePasswordAuthentication bool `json:"disablePasswordAuthentication"`
			ProvisionVMAgent              bool `json:"provisionVMAgent"`
		} `json:"linuxConfiguration"`
		Secrets                     []interface{} `json:"secrets"`
		AllowExtensionOperations    bool          `json:"allowExtensionOperations"`
		RequireGuestProvisionSignal bool          `json:"requireGuestProvisionSignal"`
	} `json:"properties.osProfile"`
	PropertiesNetworkProfile struct {
		NetworkInterfaces []struct {
			ID string `json:"id"`
		} `json:"networkInterfaces"`
	} `json:"properties.networkProfile"`
	PropertiesProvisioningState string `json:"properties.provisioningState"`
	PropertiesVMID              string `json:"properties.vmId"`
	Identity                    struct {
		Type                   string `json:"type"`
		UserAssignedIdentities struct {
			Subscriptions0C84B91E95F5409E9Cff6C2E60Affbb3ResourcegroupsDuploservicesBase01ProvidersMicrosoftManagedIdentityUserAssignedIdentitiesDuploservicesBase01 struct {
				PrincipalID string `json:"principalId"`
				ClientID    string `json:"clientId"`
			} `json:"/subscriptions/0c84b91e-95f5-409e-9cff-6c2e60affbb3/resourcegroups/duploservices-base01/providers/Microsoft.ManagedIdentity/userAssignedIdentities/duploservices-base01"`
		} `json:"userAssignedIdentities"`
	} `json:"identity"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location"`
	Tags     struct {
		TENANTNAME        string `json:"TENANT_NAME"`
		TENANTID          string `json:"TENANT_ID"`
		DuploProject      string `json:"duplo-project"`
		Owner             string `json:"owner"`
		DuploCreationTime string `json:"duplo_creation_time"`
		DuploSyncVM       string `json:"duplo_sync_vm"`
	} `json:"tags"`
}

// NativeHostGetList retrieves a list of native hosts via the Duplo API.
func (c *Client) NativeHostGetList(tenantID string) (*[]DuploNativeHost, ClientError) {
	rp := []DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", tenantID),
		&rp)
	return &rp, err
}

// NativeHostExists checks if a native host exists via the Duplo API.
func (c *Client) NativeHostExists(tenantID, instanceID string) (bool, ClientError) {

	// Get the list of hosts
	// TODO: change the backend error to a 404
	list, err := c.NativeHostGetList(tenantID)
	if err != nil {
		return false, err
	}

	// Check if the host exists
	if list != nil {
		for _, host := range *list {
			if host.InstanceID == instanceID {
				return true, nil
			}
		}
	}
	return false, nil
}

// NativeHostGet retrieves an native host via the Duplo API.
func (c *Client) NativeHostGet(tenantID, instanceID string) (*DuploNativeHost, ClientError) {
	rp := DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGet(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenantID, instanceID),
		&rp)
	return &rp, err
}

// NativeHostCreate creates an native host via the Duplo API.
func (c *Client) NativeHostCreate(rq *DuploNativeHost) (*DuploNativeHost, ClientError) {
	return c.NativeHostCreateOrUpdate(rq, false)
}

// NativeHostUpdate updates an native host via the Duplo API.
func (c *Client) NativeHostUpdate(rq *DuploNativeHost) (*DuploNativeHost, ClientError) {
	return c.NativeHostCreateOrUpdate(rq, true)
}

// NativeHostCreateOrUpdate creates or updates a native host via the Duplo API.
func (c *Client) NativeHostCreateOrUpdate(rq *DuploNativeHost, updating bool) (*DuploNativeHost, ClientError) {

	// Build the request
	var verb, msg, api string

	if updating {
		verb = "PUT"
		msg = fmt.Sprintf("NativeHostUpdate(%s, %s)", rq.TenantID, rq.InstanceID)
		api = fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", rq.TenantID, rq.InstanceID)
	} else {
		verb = "POST"
		msg = fmt.Sprintf("NativeHostCreate(%s, %s)", rq.TenantID, rq.FriendlyName)
		api = fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", rq.TenantID)
	}

	// Call the API.
	rp := DuploNativeHost{}
	err := c.doAPIWithRequestBody(verb, msg, api, &rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// NativeHostDelete deletes a native host via the Duplo API.
func (c *Client) NativeHostDelete(tenantID, instanceID string) ClientError {
	return c.deleteAPI(fmt.Sprintf("NativeHostDelete(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenantID, instanceID),
		nil)
}

func (c *Client) AzureVitualMachineList(tenantID string) (*[]DuploAzureVitualMachine, ClientError) {
	rp := []DuploAzureVitualMachine{}
	err := c.getAPI(fmt.Sprintf("AzureVitualMachineList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAzureVirtualMachinesEx", tenantID),
		&rp)
	return &rp, err
}

func (c *Client) AzureVitualMachineGet(tenantID, name string) (*DuploAzureVitualMachine, ClientError) {
	list, err := c.AzureVitualMachineList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, vm := range *list {
			if vm.Name == name {
				return &vm, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureNativeHostGet(tenantID, name string) (*DuploNativeHost, ClientError) {
	list, err := c.AzureNativeHostList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, vm := range *list {
			if vm.FriendlyName == name {
				return &vm, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) AzureNativeHostList(tenantID string) (*[]DuploNativeHost, ClientError) {
	rp := []DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGet(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetSyncNativeHosts", tenantID),
		&rp)
	return &rp, err
}

func (c *Client) AzureNativeHostCreate(rq *DuploNativeHost) ClientError {
	rp := ""
	return c.postAPI(fmt.Sprintf("AzureNativeHostCreate(%s, %s)", rq.TenantID, rq.FriendlyName),
		fmt.Sprintf("subscriptions/%s/CreateAzureVmSync", rq.TenantID),
		&rq,
		&rp)
}

func (c *Client) AzureNativeHostDelete(tenantID, name string) ClientError {
	return c.postAPI(fmt.Sprintf("AzureNativeHostDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteAzureVmSync/%s", tenantID, name),
		nil,
		nil)
}

func (c *Client) AzureNativeHostExists(tenantID, name string) (bool, ClientError) {

	list, err := c.AzureNativeHostList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, host := range *list {
			if host.InstanceID == name {
				return true, nil
			}
		}
	}
	return false, nil
}
