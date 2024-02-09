package duplocloud

import (
	"context"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

const (
	TestAccProvider_PREAMBLE = `
`
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"duplocloud": testAccProvider,
	}

	testAccProvider.ConfigureContextFunc = testAccProvider_ConfigureContextFunc(testAccProvider)
}

func testAccProvider_ConfigureContextFunc(d *schema.Provider) schema.ConfigureContextFunc {
	orig := d.ConfigureContextFunc
	srv := duplosdktest.SetupHttptestWithFixtures()

	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		c, diags := orig(ctx, d)

		client := c.(*duplosdk.Client)
		client.HostURL = srv.URL

		return client, diags
	}
}

func testAccProvider_GenConfig(body string) string {
	return TestAccProvider_PREAMBLE + body
}
