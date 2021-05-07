package duplosdk

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
)

func isInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

// GetDuploServicesName builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesName(tenantID, name string) (string, error) {
	return c.GetResourceName("duploservices", tenantID, name)
}

// GetResourceNamebuilds a duplo resource name, given a tenant ID.
func (c *Client) GetResourceName(prefix, tenantID, name string) (string, error) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{prefix, tenant.AccountName, name}, "-"), nil
}

// GetDuploServicesPrefix builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesPrefix(tenantID string) (string, error) {
	return c.GetResourcePrefix("duploservices", tenantID)
}

// GetResourcePrefix builds a duplo resource prefix, given a tenant ID.
func (c *Client) GetResourcePrefix(prefix, tenantID string) (string, error) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{prefix, tenant.AccountName}, "-"), nil
}

// UnprefixName removes a duplo resource prefix from a name.
func UnprefixName(prefix, name string) (string, bool) {
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix)+1:], true
	}

	return name, false
}

func hashForData(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	var apiStr = fmt.Sprintf("%d==", h.Sum32())
	return apiStr
}
