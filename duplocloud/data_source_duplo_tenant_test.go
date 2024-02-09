package duplocloud

import (
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	Tenant_testacc1a = "302c6a63-ffc4-4276-b52e-48a84108b658"
)

func TestAccDatasource_duplocloud_tenant_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetFixtures,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig("data \"duplocloud_tenant\" \"" + rName + "\" {\n" +
					"	 id = \"" + Tenant_testacc1a + "\"\n" +
					"}"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "id", Tenant_testacc1a),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "plan_id", "testacc1"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "name", "testacc1a"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.#", "2"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.0.key", "foo"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.0.value", "bar"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.1.key", "my"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.1.value", "tag"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "policy.0.allow_volume_mapping", "false"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "policy.0.block_external_ep", "false"),
				),
			},

			// Ensure that we don't crash when tenant policy is nil
			{
				Config: testAccProvider_GenConfig("data \"duplocloud_tenant\" \"" + rName + "\" {\n" +
					"	 id = \"" + Tenant_testacc1a + "\"\n" +
					"}"),
				PreConfig: func() {
					tenant := duplosdk.DuploTenant{}
					duplosdktest.PatchFixture("tenant/"+Tenant_testacc1a, &tenant, func() {
						tenant.TenantPolicy = nil
					})
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "id", Tenant_testacc1a),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "plan_id", "testacc1"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "name", "testacc1a"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.#", "2"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.0.key", "foo"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.0.value", "bar"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.1.key", "my"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "tags.1.value", "tag"),
					resource.TestCheckNoResourceAttr("data.duplocloud_tenant."+rName, "policy.0.allow_volume_mapping"),
					resource.TestCheckNoResourceAttr("data.duplocloud_tenant."+rName, "policy.0.block_external_ep"),
				),
			},
		},
	})
}
