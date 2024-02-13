package duplocloud

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplocloudtest"
	"terraform-provider-duplocloud/internal/duplosdktest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func duplocloud_tenant_config_basic(rName string, deleteUnspecified bool, settings []duplosdk.DuploKeyStringValue) string {
	return duplocloudtest.WriteCustomResource("duplocloud_tenant_config", rName, func(sb *strings.Builder) {
		duplocloudtest.WriteAttr(sb, "tenant_id", "\""+Tenant_testacc1a+"\"")
		duplocloudtest.WriteAttr(sb, "delete_unspecified_settings", strconv.FormatBool(deleteUnspecified))

		for i := range settings {
			sb.WriteString("  setting {\n")
			duplocloudtest.WriteAttr(sb, "  key", "\""+settings[i].Key+"\"")
			duplocloudtest.WriteAttr(sb, "  value", "\""+settings[i].Value+"\"")
			sb.WriteString("  }\n")
		}
	})
}

func TestAccResource_duplocloud_tenant_config_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(2, acctest.CharSetAlpha) +
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	// Tenant that is allowed to be deleted.
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   duplosdktest.ResetEmulator,
		CheckDestroy: func(state *terraform.State) error {
			deleted := duplosdktest.EmuDeleted()
			if len(deleted) == 0 {
				return fmt.Errorf("Not deleted: %s", "duplocloud_tenant_conig."+rName)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// No deletion of unspecified settings.
			// No modification of unspecified settings.
			{
				Config: testAccProvider_GenConfig(
					duplocloud_tenant_config_basic(rName, false, []duplosdk.DuploKeyStringValue{
						{Key: "abc", Value: "xyz"},
						{Key: "mermaid", Value: "cousin"},
					}),
				),
				Check: func(state *terraform.State) error {
					r := "duplocloud_tenant_config." + rName
					return resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(r, "tenant_id", Tenant_testacc1a),
						resource.TestCheckResourceAttr(r, "delete_unspecified_settings", "false"),
						resource.TestCheckResourceAttr(r, "setting.#", "2"),
						resource.TestCheckResourceAttr(r, "setting.0.key", "abc"),
						resource.TestCheckResourceAttr(r, "setting.0.value", "xyz"),
						resource.TestCheckResourceAttr(r, "setting.1.key", "mermaid"),
						resource.TestCheckResourceAttr(r, "setting.1.value", "cousin"),
						resource.TestCheckResourceAttr(r, "specified_settings.#", "2"),
						resource.TestCheckResourceAttr(r, "specified_settings.0", "abc"),
						resource.TestCheckResourceAttr(r, "specified_settings.1", "mermaid"),
						resource.TestCheckResourceAttr(r, "metadata.#", "4"),
						resource.TestCheckResourceAttr(r, "metadata.0.key", "abc"),
						resource.TestCheckResourceAttr(r, "metadata.0.value", "xyz"),
						resource.TestCheckResourceAttr(r, "metadata.1.key", "foo"),
						resource.TestCheckResourceAttr(r, "metadata.1.value", "bar"),
						resource.TestCheckResourceAttr(r, "metadata.2.key", "mermaid"),
						resource.TestCheckResourceAttr(r, "metadata.2.value", "cousin"),
						resource.TestCheckResourceAttr(r, "metadata.3.key", "xyz"),
						resource.TestCheckResourceAttr(r, "metadata.3.value", "abc"),
					)(state)
				},
			},
		},
	})
}
