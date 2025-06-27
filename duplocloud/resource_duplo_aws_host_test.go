package duplocloud

import (
	"fmt"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/internal/duplocloudtest"
	"github.com/duplocloud/terraform-provider-duplocloud/internal/duplosdktest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type duploNativeHost struct {
	InstanceID         string                             `json:"InstanceId"`
	UserAccount        string                             `json:"UserAccount,omitempty"`
	TenantID           string                             `json:"TenantId,omitempty"`
	FriendlyName       string                             `json:"FriendlyName,omitempty"`
	Capacity           string                             `json:"Capacity,omitempty"`
	Zone               int                                `json:"Zone"`
	IsMinion           bool                               `json:"IsMinion"`
	ImageID            string                             `json:"ImageId,omitempty"`
	Base64UserData     string                             `json:"Base64UserData,omitempty"`
	PrependUserData    bool                               `json:"IsUserDataCombined,omitempty"`
	AgentPlatform      int                                `json:"AgentPlatform"`
	IsEbsOptimized     bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP  bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud              int                                `json:"Cloud"`
	KeyPairType        int                                `json:"KeyPairType"`
	EncryptDisk        bool                               `json:"EncryptDisk,omitempty"`
	Status             string                             `json:"Status,omitempty"`
	IdentityRole       string                             `json:"IdentityRole,omitempty"`
	PrivateIPAddress   string                             `json:"PrivateIpAddress,omitempty"`
	PublicIPAddress    string                             `json:"PublicIpAddress,omitempty"`
	NetworkInterfaceId string                             `json:"NetworkInterfaceId,omitempty"`
	NetworkInterfaces  *[]duploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes            *[]duploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData           *[]duploKeyStringValue             `json:"MetaData,omitempty"`
	Tags               *[]duploKeyStringValue             `json:"Tags,omitempty"`
	TagsEx             *[]duploKeyStringValue             `json:"TagsEx,omitempty"`
	MinionTags         *[]duploKeyStringValue             `json:"MinionTags,omitempty"`
	SecurityType       string                             `json:"SecurityType"`
	IsEncryptAtHost    bool                               `json:"IsEncryptAtHost"`
	IsSecureBoot       bool                               `json:"IsSecureBoot"`
	IsvTPM             bool                               `json:"IsvTPM"`
	DiskControlType    string                             `json:"DiskControllerType,omitempty"`
	ExtraNodeLabels    *[]duploKeyStringValue             `json:"ExtraNodeLabels,omitempty"`
	Taints             *[]duploTaints                     `json:"Taints,omitempty"`
	AvailabilitySetId  string                             `json:"AvailabilitySetId"`
}

type duploTaints struct {
	Key    string `json:"Key"`
	Value  string `json:"Value"`
	Effect string `json:"Effect"`
}
type duploKeyStringValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value,omitempty"`
}

// DuploNativeHostNetworkInterface is a Duplo SDK object that represents a network interface of a native host
type duploNativeHostNetworkInterface struct {
	NetworkInterfaceID string                 `json:"NetworkInterfaceId,omitempty"`
	SubnetID           string                 `json:"SubnetId,omitempty"`
	AssociatePublicIP  bool                   `json:"AssociatePublicIpAddress,omitempty"`
	Groups             *[]string              `json:"Groups,omitempty"`
	DeviceIndex        int                    `json:"DeviceIndex,omitempty"`
	MetaData           *[]duploKeyStringValue `json:"MetaData,omitempty"`
}

// DuploNativeHostVolume is a Duplo SDK object that represents a volume of a native host
type duploNativeHostVolume struct {
	Iops       int    `json:"Iops,omitempty"`
	Name       string `json:"Name,omitempty"`
	Size       int    `Size:"Size,omitempty"`
	VolumeID   string `json:"VolumeId,omitempty"`
	VolumeType string `json:"VolumeType,omitempty"`
}

func duplocloud_aws_host_basic(rName, hostName string, attrs map[string]string) string {
	return duplocloudtest.WriteFlatResource("duplocloud_aws_host", rName,
		map[string]string{
			"tenant_id":            "\"" + Tenant_testacc1a + "\"",
			"user_account":         "\"testacc1a\"",
			"friendly_name":        "\"duploservices-testacc1a-" + hostName + "\"",
			"zone":                 "0",
			"image_id":             "\"ami-1234abc\"",
			"capacity":             "\"t4g.small\"",
			"allocated_public_ip":  "true",
			"wait_until_connected": "false",
		},
		attrs,
	)
}

func TestAccResource_duplocloud_aws_host_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	hostName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	roleName := "duploservices-testacc1a"

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetEmulator,
		CheckDestroy: func(state *terraform.State) error {
			deleted := duplosdktest.EmuDeleted()
			if len(deleted) == 0 {
				return fmt.Errorf("Not deleted: %s", "duplocloud_aws_host."+rName)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// No diffs given when friendly_name is the long name.
			// Public subnets work.
			{
				Config: testAccProvider_GenConfig(
					duplocloud_aws_host_basic(rName, hostName, map[string]string{}),
				),
				Check: func(state *terraform.State) error {
					host := duplosdktest.EmuLastCreated().(*duploNativeHost)
					r := "duplocloud_aws_host." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "instance_id", host.InstanceID),
						resource.TestCheckResourceAttr(r, "friendly_name", roleName+"-"+hostName),
						resource.TestCheckResourceAttr(r, "image_id", "ami-1234abc"),
						resource.TestCheckResourceAttr(r, "capacity", "t4g.small"),
						resource.TestCheckResourceAttr(r, "allocated_public_ip", "true"),
						resource.TestCheckResourceAttr(r, "zone", "0"),
						resource.TestCheckResourceAttr(r, "network_interface.0.subnet_id", "subnet-ext1"),
					)(state)
				},
			},

			// No diffs given when friendly_name is the short name.
			// Private subnets work.
			{
				Config: testAccProvider_GenConfig(
					duplocloud_aws_host_basic(rName, hostName, map[string]string{
						"friendly_name":       "\"" + hostName + "\"",
						"allocated_public_ip": "false",
					}),
				),
				Check: func(state *terraform.State) error {
					host := duplosdktest.EmuLastCreated().(*duploNativeHost)
					r := "duplocloud_aws_host." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "instance_id", host.InstanceID),
						resource.TestCheckResourceAttr(r, "friendly_name", roleName+"-"+hostName),
						resource.TestCheckResourceAttr(r, "image_id", "ami-1234abc"),
						resource.TestCheckResourceAttr(r, "capacity", "t4g.small"),
						resource.TestCheckResourceAttr(r, "allocated_public_ip", "false"),
						resource.TestCheckResourceAttr(r, "zone", "0"),
						resource.TestCheckNoResourceAttr(r, "network_interface.0"),
					)(state)
				},
			},

			// Zone selection works.
			{
				Config: testAccProvider_GenConfig(
					duplocloud_aws_host_basic(rName, hostName, map[string]string{
						"zone": "1",
					}),
				),
				Check: func(state *terraform.State) error {
					host := duplosdktest.EmuLastCreated().(*duploNativeHost)
					r := "duplocloud_aws_host." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "instance_id", host.InstanceID),
						resource.TestCheckResourceAttr(r, "friendly_name", roleName+"-"+hostName),
						resource.TestCheckResourceAttr(r, "image_id", "ami-1234abc"),
						resource.TestCheckResourceAttr(r, "capacity", "t4g.small"),
						resource.TestCheckResourceAttr(r, "allocated_public_ip", "true"),
						resource.TestCheckResourceAttr(r, "network_interface.0.subnet_id", "subnet-ext2"),
					)(state)
				},
			},
		},
	})
}
