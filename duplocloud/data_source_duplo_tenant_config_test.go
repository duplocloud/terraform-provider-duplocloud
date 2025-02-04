package duplocloud

import (
	"github.com/duplocloud/terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSource_duplocloud_tenant_config_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetEmulator,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider_GenConfig(
					"data \"duplocloud_tenant_config\" \"" + rName + "\" {\n" +
						"	 tenant_id = \"" + Tenant_testacc1a + "\"\n" +
						"}",
				),
				Check: func(state *terraform.State) error {
					r := "data.duplocloud_tenant_config." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "metadata.#", "2"),
						resource.TestCheckResourceAttr(r, "metadata.0.key", "foo"),
						resource.TestCheckResourceAttr(r, "metadata.0.value", "bar"),
						resource.TestCheckResourceAttr(r, "metadata.1.key", "xyz"),
						resource.TestCheckResourceAttr(r, "metadata.1.value", "abc"),
					)(state)
				},
			},
		},
	})
}
