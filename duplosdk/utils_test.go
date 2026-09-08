package duplosdk

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duplocloud/terraform-provider-duplocloud/internal/duplosdktest"
)

// setupTenantListServer serves the user-scoped tenant list and rejects every
// other path the way the portal rejects a non-admin user (DUPLO-44452).
func setupTenantListServer(t *testing.T, body string) (*httptest.Server, *Client, *[]string) {
	paths := []string{}
	srv := duplosdktest.SetupHttptest(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		paths = append(paths, req.URL.Path)
		if req.URL.Path != "/admin/GetTenantsForUser" {
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte(`{"Message":"Access is forbidden for this user"}`)) // nolint
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(body)) // nolint
	}))
	c, err := NewClient(srv.URL, "FAKE")
	assert.Nil(t, err, err)
	return srv, c, &paths
}

// GetResourcePrefix must resolve the tenant name through the user-scoped tenant
// list so that non-admin tenant users can read prefixed resources.
func TestGetResourcePrefix_UsesUserScopedTenantLookup(t *testing.T) {
	srv, c, paths := setupTenantListServer(t,
		`[{"TenantId":"other","AccountName":"other"},{"TenantId":"t1","AccountName":"myapp"}]`)
	defer duplosdktest.TeardownHttptest(srv)

	prefix, err := c.GetResourcePrefix("duploservices", "t1")

	assert.Nil(t, err)
	assert.Equal(t, "duploservices-myapp", prefix)
	assert.Equal(t, []string{"/admin/GetTenantsForUser"}, *paths)
}

// A tenant the user cannot see must surface as a 404 so callers can clear state.
func TestGetResourcePrefix_TenantNotVisible(t *testing.T) {
	srv, c, _ := setupTenantListServer(t, `[{"TenantId":"other","AccountName":"other"}]`)
	defer duplosdktest.TeardownHttptest(srv)

	prefix, err := c.GetResourcePrefix("duploservices", "t1")

	assert.Equal(t, "", prefix)
	if assert.NotNil(t, err) {
		assert.Equal(t, http.StatusNotFound, err.Status())
	}
}
