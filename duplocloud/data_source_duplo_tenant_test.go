package duplocloud

import (
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
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig("data \"duplocloud_tenant\" \"" + rName + "\" {\n" +
					"	 id = \"" + Tenant_testacc1a + "\"\n" +
					"}"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "id", Tenant_testacc1a),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "plan_id", "testacc1"),
					resource.TestCheckResourceAttr("data.duplocloud_tenant."+rName, "name", "testacc1"),
				),
			},
		},
	})
}
