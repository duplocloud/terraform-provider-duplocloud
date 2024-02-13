package duplocloud

import (
	"fmt"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplocloudtest"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func duplocloud_aws_host_basic(rName, hostName string, attrs map[string]string) string {
	return duplocloudtest.WriteResource("duplocloud_aws_host", rName,
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
					host := duplosdktest.EmuLastCreated().(*duplosdk.DuploNativeHost)
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
					host := duplosdktest.EmuLastCreated().(*duplosdk.DuploNativeHost)
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
					host := duplosdktest.EmuLastCreated().(*duplosdk.DuploNativeHost)
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
