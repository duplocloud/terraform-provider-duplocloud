package duplosdk

import (
	"fmt"
)

// DuploNativeHost is a Duplo SDK object that represents an nativehost
type DuploNativeHost struct {
	InstanceID        string                             `json:"InstanceId"`
	UserAccount       string                             `json:"UserAccount,omitempty"`
	TenantID          string                             `json:"TenantId,omitempty"`
	FriendlyName      string                             `json:"FriendlyName,omitempty"`
	Capacity          string                             `json:"Capacity,omitempty"`
	Zone              int                                `json:"Zone"`
	IsMinion          bool                               `json:"IsMinion"`
	ImageID           string                             `json:"ImageId,omitempty"`
	Base64UserData    string                             `json:"Base64UserData,omitempty"`
	AgentPlatform     int                                `json:"AgentPlatform"`
	IsEbsOptimized    bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud             int                                `json:"Cloud"`
	EncryptDisk       bool                               `json:"EncryptDisk,omitempty"`
	Status            string                             `json:"Status,omitempty"`
	IdentityRole      string                             `json:"IdentityRole,omitempty"`
	PrivateIPAddress  string                             `json:"PrivateIpAddress,omitempty"`
	NetworkInterfaces *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes           *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData          *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	Tags              *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	MinionTags        *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
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

// NativeHostGetList retrieves a list of native hosts via the Duplo API.
func (c *Client) NativeHostGetList(tenantID string) (*[]DuploNativeHost, error) {
	rp := []DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", tenantID),
		&rp)
	return &rp, err
}

// NativeHostGet retrieves an native host via the Duplo API.
func (c *Client) NativeHostGet(tenantID, instanceID string) (*DuploNativeHost, error) {
	rp := DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGet(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenantID, instanceID),
		&rp)
	return &rp, err
}

// NativeHostCreate creates an native host via the Duplo API.
func (c *Client) NativeHostCreate(rq *DuploNativeHost) (*DuploNativeHost, error) {
	return c.NativeHostCreateOrUpdate(rq, false)
}

// NativeHostUpdate updates an native host via the Duplo API.
func (c *Client) NativeHostUpdate(rq *DuploNativeHost) (*DuploNativeHost, error) {
	return c.NativeHostCreateOrUpdate(rq, true)
}

// NativeHostCreateOrUpdate creates or updates a native host via the Duplo API.
func (c *Client) NativeHostCreateOrUpdate(rq *DuploNativeHost, updating bool) (*DuploNativeHost, error) {

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
func (c *Client) NativeHostDelete(tenantID, instanceID string) error {
	return c.deleteAPI(fmt.Sprintf("NativeHostDelete(%s, %s)", tenantID, instanceID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenantID, instanceID),
		nil)
}
