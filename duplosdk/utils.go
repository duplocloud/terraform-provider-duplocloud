package duplosdk

import (
	"net/url"
	"reflect"
	"strings"
)

// handle path parameter encoding when it might contain slashes
func EncodePathParam(param string) string {
	return url.PathEscape(url.PathEscape(param))
}

func QueryEscape(param string) string {
	return url.QueryEscape(url.QueryEscape(param))
}

//nolint:deadcode,unused // utility function
func isInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

// GetDuploServicesNameWithAws builds a duplo resource name, given a tenant ID. The name includes the AWS account ID suffix.
func (c *Client) GetDuploServicesNameWithAws(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, true)
}

// GetDuploServicesName builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesName(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, false)
}

// GetResourceName builds a duplo resource name, given a tenant ID.  It can optionally include the AWS account ID suffix.
func (c *Client) GetResourceName(prefix, tenantID, name string, withAccountSuffix bool) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	if withAccountSuffix {
		accountID, err := c.TenantGetAwsAccountID(tenantID)
		if err != nil {
			return "", err
		}
		return strings.Join([]string{prefix, tenant.AccountName, name, accountID}, "-"), nil
	}
	return strings.Join([]string{prefix, tenant.AccountName, name}, "-"), nil
}

// GetDuploServicesNameWithGcp builds a duplo resource name, given a tenant ID. The name includes the Gcp project ID suffix.
func (c *Client) GetDuploServicesNameWithGcp(tenantID, name string) (string, ClientError) {
	return c.GetResourceNameWithGcp("duploservices", tenantID, name)
}

// GetDuploServicesNameWithGcp builds a duplo resource name, given a tenant ID. The name includes the Gcp project ID suffix.
func (c *Client) GetResourceNameWithGcp(prefix, tenantID, name string) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	projectID, err := c.TenantGetGcpProjectID(tenantID)
	if err != nil {
		return "", err
	}

	return strings.Join([]string{prefix, tenant.AccountName, name, projectID}, "-"), nil
}

// GetDuploServicesPrefix builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesPrefix(tenantID string) (string, ClientError) {
	return c.GetResourcePrefix("duploservices", tenantID)
}

// GetResourcePrefix builds a duplo resource prefix, given a tenant ID.
func (c *Client) GetResourcePrefix(prefix, tenantID string) (string, ClientError) {
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

// UnwrapName removes a duplo resource prefix and AWS account ID suffix from a name.
func UnwrapName(prefix, accountID, name string, optionalAccountID bool) (string, bool) {
	suffix := "-" + accountID

	var part string
	if !strings.HasSuffix(name, suffix) {
		if !optionalAccountID {
			return name, false
		} else {
			part = name
		}
	} else {
		part = name[0 : len(name)-len(suffix)]
	}

	if !strings.HasPrefix(part, prefix) {
		return name, false
	}

	return part[len(prefix)+1:], true
}
