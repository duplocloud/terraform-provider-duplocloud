package duplocloud

import (
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	AwsHost_testacc1a_host1        = "i-a4a64e4407f5ae678"
	AwsHost_testacc1a_host1_subnet = "subnet-0c63c3139342fb12b"
)

func TestAccDataSource_duplocloud_native_hosts_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetEmulator,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig(
					"data \"duplocloud_native_hosts\" \"" + rName + "\" {\n" +
						"	 tenant_id = \"" + Tenant_testacc1a + "\"\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					r := "data.duplocloud_native_hosts." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "hosts.0.instance_id", AwsHost_testacc1a_host1),
						resource.TestCheckResourceAttr(r, "hosts.0.identity_role", "duploservices-testacc1a"),
						resource.TestCheckResourceAttr(r, "hosts.0.friendly_name", "duploservices-testacc1a-host1"),
						resource.TestCheckResourceAttr(r, "hosts.0.is_minion", "true"),
						resource.TestCheckResourceAttr(r, "hosts.0.cloud", "0"),
						resource.TestCheckResourceAttr(r, "hosts.0.agent_platform", "0"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.0.key", "DUPLO_TENANT"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.0.value", "testacc1a"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.1.key", "Name"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.1.value", "duploservices-testacc1a-host1"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.2.key", "duplo-project"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.2.value", "testacc1a"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.3.key", "TENANT_NAME"),
						resource.TestCheckResourceAttr(r, "hosts.0.tags.3.value", "testacc1a"),
						resource.TestCheckResourceAttr(r, "hosts.0.network_interface.0.subnet_id", AwsHost_testacc1a_host1_subnet),
						resource.TestCheckResourceAttr(r, "hosts.0.volume.0.name", "xvda1"),
					)(state)
				},
			},
		},
	})
}
