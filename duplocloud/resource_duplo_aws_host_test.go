package duplocloud

import (
	"fmt"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_aws_host\" \"" + rName + "\" {\n" +
						"	 tenant_id = \"" + Tenant_testacc1a + "\"\n" +
						"	 user_account = \"testacc1a\"\n" +
						"	 friendly_name = \"" + hostName + "\"\n" +
						"	 zone = 0\n" +
						"	 image_id = \"ami-1234abc\"\n" +
						"	 capacity = \"t4g.small\"\n" +
						"	 allocated_public_ip = true\n" +
						"	 wait_until_connected = false\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					host := duplosdktest.EmuCreated()[0].(*duplosdk.DuploNativeHost)
					r := "duplocloud_aws_host." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "instance_id", host.InstanceID),
						resource.TestCheckResourceAttr(r, "friendly_name", roleName+"-"+hostName),
						resource.TestCheckResourceAttr(r, "image_id", "ami-1234abc"),
						resource.TestCheckResourceAttr(r, "capacity", "t4g.small"),
						resource.TestCheckResourceAttr(r, "allocated_public_ip", "true"),
						resource.TestCheckResourceAttr(r, "network_interface.0.subnet_id", "subnet-ext1"),
					)(state)
				},
			},
		},
	})

}
