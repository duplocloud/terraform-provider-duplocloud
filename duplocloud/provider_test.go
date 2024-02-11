package duplocloud

import (
	"context"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"terraform-provider-duplocloud/internal/duplosdktest"

	"github.com/barkimedes/go-deepcopy"
	"github.com/google/uuid"
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
	srv := duplosdktest.NewEmulator(duplosdktest.EmuConfig{
		Types: map[string]duplosdktest.EmuType{
			"tenant": {
				Factory: func() interface{} { return &duplosdk.DuploTenant{} },
				Responder: func(in interface{}) (id string, out interface{}) {
					id = uuid.New().String()
					out = deepcopy.MustAnything(in.(*duplosdk.DuploTenant))
					out.(*duplosdk.DuploTenant).TenantID = id
					return
				},
			},
			"tenant/:tenantId/aws_host": {
				Factory: func() interface{} { return &duplosdk.DuploNativeHost{} },
				Responder: func(in interface{}) (id string, out interface{}) {
					id = "i-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:17]
					out = deepcopy.MustAnything(in.(*duplosdk.DuploNativeHost))
					out.(*duplosdk.DuploNativeHost).InstanceID = id
					return
				},
			},
		},
	})

	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		c, diags := orig(ctx, d)

		if client, ok := c.(*duplosdk.Client); ok {
			client.HostURL = srv.URL
		}

		return c, diags
	}
}

func testAccProvider_GenConfig(body string) string {
	return TestAccProvider_PREAMBLE + body
}
