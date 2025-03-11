package duplocloud

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/duplocloud/terraform-provider-duplocloud/internal/duplosdktest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResource_duplocloud_tenant_basic(t *testing.T) {
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
			for _, s := range deleted {
				if !strings.Contains(s, "tenant/") {
					return fmt.Errorf("Should  have been deleted: %s", "duplocloud_tenant."+rName)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_tenant\" \"" + rName + "\" {\n" +
						"	 account_name = \"" + tenantName + "\"\n" +
						"	 plan_id = \"testacc1\"\n" +
						"	 wait_until_created = false\n" +
						"	 allow_deletion = true\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					tenant := duplosdktest.EmuCreated()[0].(*duplosdk.DuploTenant)
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "tenant_id", tenant.TenantID),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "plan_id", "testacc1"),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "account_name", tenantName),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "allow_deletion", "true"),
					)(state)
				},
			},
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_tenant\" \"" + rName + "\" {\n" +
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
			if len(duplosdktest.EmuDeleted()) == 0 {
				return nil
			} else {
				return fmt.Errorf("Should not have been deleted: %s", "duplocloud_tenant."+rName)
			}

		},
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_tenant\" \"" + rName + "\" {\n" +
						"	 account_name = \"" + tenantName + "\"\n" +
						"	 plan_id = \"testacc1\"\n" +
						"	 wait_until_created = false\n" +
						"	 allow_deletion = false\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					tenant := duplosdktest.EmuCreated()[0].(*duplosdk.DuploTenant)
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "tenant_id", tenant.TenantID),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "plan_id", "testacc1"),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "account_name", tenantName),
						resource.TestCheckResourceAttr("duplocloud_tenant."+rName, "allow_deletion", "false"),
					)(state)
				},
				ExpectError: regexp.MustCompile("Will NOT delete the tenant - because allow_deletion is false"),
			},
			{
				Config: testAccProvider_GenConfig(
					"resource \"duplocloud_tenant\" \"" + rName + "\" {\n" +
						"	 account_name = \"" + tenantName + "\"\n" +
						"	 plan_id = \"testacc1\"\n" +
						"	 wait_until_created = false\n" +
						"	 allow_deletion = false\n" +
						"}",
				),
			},
		},
	})
}
