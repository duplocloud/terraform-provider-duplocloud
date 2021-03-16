package duplosdk

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploTenant represents a Duplo tenant
type DuploTenant struct {
	TenantID     string                 `json:"TenantId,omitempty"`
	AccountName  string                 `json:"AccountName"`
	PlanID       string                 `json:"PlanID"`
	InfraOwner   string                 `json:"InfraOwner,omitempty"`
	TenantPolicy *DuploTenantPolicy     `json:"TenantPolicy,omitempty"`
	Tags         *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploTenantPolicy reprsents policies for a Duplo Tenant
type DuploTenantPolicy struct {
	AllowVolumeMapping bool `json:"AllowVolumeMapping,omitempty"`
	BlockExternalEp    bool `json:"BlockExternalEp,omitempty"`
}

// DuploTenantAwsCredentials represents AWS credentials for a Duplo tenant
type DuploTenantAwsCredentials struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	ConsoleURL      string `json:"ConsoleUrl,omitempty"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Region          string `json:"Region"`
	SessionToken    string `json:"SessionToken,omitempty"`
}

// TenantURLList returns the base API URL for crud -- get list + create + update
func (c *Client) TenantURLList(d *schema.ResourceData) string {
	api := "v2/admin/TenantV2"
	url := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-TenantUrlList %s 1 ********: %s", api, url)
	return url
}

// TenantGet retrieves a tenant via the Duplo API.
func (c *Client) TenantGet(tenantID string) (*DuploTenant, error) {
	apiName := fmt.Sprintf("TenantGet(%s)", tenantID)
	rp := DuploTenant{}

	// Get the tenant from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("v2/admin/TenantV2/%s", tenantID), &rp)
	if err != nil || rp.TenantID == "" {
		return nil, err
	}
	return &rp, nil
}

// TenantCreate creates a tenant via the Duplo API.
func (c *Client) TenantCreate(rq DuploTenant) (*DuploTenant, error) {
	rp := DuploTenant{}
	err := c.postAPI(fmt.Sprintf("TenantCreate(%s, %s)", rq.AccountName, rq.PlanID), "v2/admin/TenantV2", &rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// TenantUpdate updates a tenant via the Duplo API.
func (c *Client) TenantUpdate(rq DuploTenant) (*DuploTenant, error) {
	// No-op
	return nil, nil
}

// TenantDelete deletes an AWS host via the Duplo API.
func (c *Client) TenantDelete(tenantID string) (*DuploTenant, error) {
	// No-op
	return nil, nil
}

// ListTenantsForUser retrieves a list of tenants for the current user via the Duplo API.
func (c *Client) ListTenantsForUser() (*[]DuploTenant, error) {
	list := []DuploTenant{}
	err := c.getAPI("ListTenantsForUser()", "admin/GetTenantsForUser", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// ListTenantsForUserByPlan retrieves a list of tenants with the given plan for the current user via the Duplo API.
// If the planID is an empty string, returns all
func (c *Client) ListTenantsForUserByPlan(planID string) (*[]DuploTenant, error) {
	// Get all tenants.
	allTenants, err := c.ListTenantsForUser()
	if err != nil {
		return nil, err
	}
	if planID == "" {
		return allTenants, nil
	}

	// Build a new list of tenants with the given plan ID.
	planTenants := make([]DuploTenant, 0, len(*allTenants))
	for _, tenant := range *allTenants {
		if tenant.PlanID == planID {
			planTenants = append(planTenants, tenant)
		}
	}

	// Return the new list
	return &planTenants, nil
}

// GetTenantByNameForUser retrieves a single tenant by name for the current user via the Duplo API.
func (c *Client) GetTenantByNameForUser(name string) (*DuploTenant, error) {
	// Get all tenants.
	allTenants, err := c.ListTenantsForUser()
	if err != nil {
		return nil, err
	}

	// Find and return the tenant with the specific name.
	for _, tenant := range *allTenants {
		if tenant.AccountName == name {
			return &tenant, nil
		}
	}

	// No tenant was found.
	return nil, nil
}

// GetTenantForUser retrieves a single tenant by ID for the current user via the Duplo API.
func (c *Client) GetTenantForUser(tenantID string) (*DuploTenant, error) {
	// Get all tenants.
	allTenants, err := c.ListTenantsForUser()
	if err != nil {
		return nil, err
	}

	// Find and return the tenant with the specific name.
	for _, tenant := range *allTenants {
		if tenant.TenantID == tenantID {
			return &tenant, nil
		}
	}

	// No tenant was found.
	return nil, nil
}

// TenantGetAwsRegion retrieves a tenant's AWS region via the Duplo API.
func (c *Client) TenantGetAwsRegion(tenantID string) (string, error) {
	awsRegion := ""
	err := c.getAPI(fmt.Sprintf("TenantGetAwsRegion(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetAwsRegionId", tenantID), &awsRegion)
	return awsRegion, err
}

// TenantGetAwsCredentials retrieves just-in-time AWS credentials for a tenant via the Duplo API.
func (c *Client) TenantGetAwsCredentials(tenantID string) (*DuploTenantAwsCredentials, error) {
	creds := DuploTenantAwsCredentials{}
	err := c.getAPI(fmt.Sprintf("TenantGetAwsCredentials(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetAwsConsoleTokenUrl", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// TenantGetInternalSubnets retrieves a list of the internal subnets for a tenant via the Duplo API.
func (c *Client) TenantGetInternalSubnets(tenantID string) ([]string, error) {
	list := []string{}
	err := c.getAPI(fmt.Sprintf("TenantGetInternalSubnets(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetInternalSubnets", tenantID), &list)
	return list, err
}

// TenantGetAwsAccountID retrieves the AWS account ID via the Duplo API.
func (c *Client) TenantGetAwsAccountID(tenantID string) (string, error) {
	awsAccountID := ""
	err := c.getAPI(fmt.Sprintf("TenantGetAwsAccountID(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetTenantAwsAccountId", tenantID), &awsAccountID)
	return awsAccountID, err
}
