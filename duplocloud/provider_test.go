package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
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
				Responder: func(verb string, in interface{}) (id string, out interface{}) {
					out = deepcopy.MustAnything(in.(*duplosdk.DuploTenant))
					if verb == "POST" {
						id = uuid.New().String()
						out.(*duplosdk.DuploTenant).TenantID = id
					} else {
						id = out.(*duplosdk.DuploTenant).TenantID
					}
					return
				},
			},
			"tenant/:tenantId/aws_host": {
				Factory: func() interface{} { return &duplosdk.DuploNativeHost{} },
				Responder: func(verb string, in interface{}) (id string, out interface{}) {
					out = deepcopy.MustAnything(in.(*duplosdk.DuploNativeHost))
					host := out.(*duplosdk.DuploNativeHost)

					if verb == "POST" {
						id = "i-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:17]
						host.InstanceID = id
					} else {
						id = host.InstanceID
					}

					if host.UserAccount == "" {
						tenant := &duplosdk.DuploTenant{}
						location := "tenant/" + host.TenantID
						if err := duplosdktest.GetResource(location, tenant); err != nil {
							log.Panicf("json.Unmarshall: %s: %s", location, err)
						} else if tenant.AccountName == "" {
							log.Panicf("%s.Responder: could not get tenant", location)
						}
						host.UserAccount = tenant.AccountName
					}
					host.IdentityRole = "duploservices-" + host.UserAccount
					if !strings.HasPrefix(host.FriendlyName, host.IdentityRole) {
						host.FriendlyName = host.IdentityRole + "-" + host.FriendlyName
					}
					return
				},
			},
			"tenant/:tenantId/metadata": {
				Factory: func() interface{} { return &duplosdk.DuploKeyStringValue{} },
				Responder: func(verb string, in interface{}) (id string, out interface{}) {
					out = deepcopy.MustAnything(in.(*duplosdk.DuploKeyStringValue))
					id = out.(*duplosdk.DuploKeyStringValue).Key
					return
				},
			},
			"/adminproxy/GetInfrastructureConfig/:infraName": {
				Factory: func() interface{} { return &duplosdk.DuploInfrastructureConfig{} },
				Responder: func(verb string, in interface{}) (id string, out interface{}) {
					out = deepcopy.MustAnything(in.(*duplosdk.DuploInfrastructureConfig))
					infraConfig := out.(*duplosdk.DuploInfrastructureConfig)
					id = infraConfig.Name
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
