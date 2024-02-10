package duplocloud

import "testing"

/*
import (
	"fmt"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)
*/

func TestAccResource_duplocloud_aws_host_basic(t *testing.T) {
	/*
		rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
			acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		tenantName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
			acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

		// Tenant that is allowed to be deleted.
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
							"	 account_name = \"" + tenantName + "\"\n" +
							"	 plan_id = \"testacc1\"\n" +
							"	 wait_until_created = false\n" +
							"	 allow_deletion = true\n" +
							"}",
					),
					Check: func(state *terraform.State) error {
						tenant := duplosdktest.EmuCreated()[0].(*duplosdk.DuploTenant)
						return resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "tenant_id", tenant.TenantID),
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "plan_id", "testacc1"),
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "account_name", tenantName),
						)(state)
					},
				},
				{
					Config: testAccProvider_GenConfig(
						"resource \"duplocloud_aws_host\" \"" + rName + "\" {\n" +
							"	 account_name = \"" + tenantName + "\"\n" +
							"	 plan_id = \"testacc1\"\n" +
							"	 wait_until_created = false\n" +
							"	 allow_deletion = true\n" +
							"}",
					),
				},
			},
		})

		// Tenant that is not allowed to be deleted.
		resource.Test(t, resource.TestCase{
			IsUnitTest: true,
			Providers:  testAccProviders,
			PreCheck:   duplosdktest.ResetEmulator,
			CheckDestroy: func(state *terraform.State) error {
				if len(duplosdktest.EmuCreated()) == 0 {
					return fmt.Errorf("Should not have been deleted: %s", "duplocloud_aws_host."+rName)
				}
				return nil
			},
			Steps: []resource.TestStep{
				{
					Config: testAccProvider_GenConfig(
						"resource \"duplocloud_aws_host\" \"" + rName + "\" {\n" +
							"	 account_name = \"" + tenantName + "\"\n" +
							"	 plan_id = \"testacc1\"\n" +
							"	 wait_until_created = false\n" +
							"	 allow_deletion = false\n" +
							"}",
					),
					Check: func(state *terraform.State) error {
						tenant := duplosdktest.EmuCreated()[0].(*duplosdk.DuploTenant)
						return resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "tenant_id", tenant.TenantID),
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "plan_id", "testacc1"),
							resource.TestCheckResourceAttr("duplocloud_aws_host."+rName, "account_name", tenantName),
						)(state)
					},
				},
				{
					Config: testAccProvider_GenConfig(
						"resource \"duplocloud_aws_host\" \"" + rName + "\" {\n" +
							"	 account_name = \"" + tenantName + "\"\n" +
							"	 plan_id = \"testacc1\"\n" +
							"	 wait_until_created = false\n" +
							"	 allow_deletion = false\n" +
							"}",
					),
				},
			},
		})
	*/
}
