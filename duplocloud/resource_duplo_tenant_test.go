package duplocloud

import (
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResource_duplocloud_tenant_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tenantName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetEmulator,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_tenant\" \"" + rName + "\" {\n" +
						"	 account_name = \"" + tenantName + "\"\n" +
						"	 plan_id = \"testacc1\"\n" +
						"	 wait_until_created = false\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					tenant := duplosdktest.EmuCreated()[0].(*duplosdk.DuploTenant)
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "tenant_id", tenant.TenantID),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "plan_id", "testacc1"),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "account_name", tenantName),
					)(state)
				},
			},
		},
	})
}
