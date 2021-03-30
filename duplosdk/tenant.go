package duplosdk

import (
	"fmt"
)

const (
	// SGSourceTypeTenant represents a Duplo Tenant as an SG rule source
	SGSourceTypeTenant int = 0
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

// DuploTenantConfig represents a Duplo tenant's configuration
type DuploTenantConfig struct {
	TenantID string                 `json:"TenantId,omitempty"`
	Metadata *[]DuploKeyStringValue `json:"MetaData,omitempty"`
}

// DuploTenantConfigUpdateRequest represents a request to update a Duplo tenant's configuration
type DuploTenantConfigUpdateRequest struct {
	TenantID string `json:"ComponentId,omitempty"`
	Key      string `json:"Key,omitempty"`
	State    string `json:"State,omitempty"`
	Value    string `json:"Value,omitempty"`
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

// DuploTenantEksCredentials represents just-in-time EKS credentials in Duplo
type DuploTenantK8sCredentials struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name        string `json:"Name"`
	APIServer   string `json:"ApiServer"`
	Token       string `json:"Token"`
	AwsRegion   string `json:"AwsRegion"`
	K8sProvider int    `json:"K8Provider,omitempty"`
}

// DuploTenantEksSecret represents just-in-time EKS credentials in Duplo
type DuploTenantEksSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name        string            `json:"SecretName"`
	Type        string            `json:"SecretType"`
	Data        map[string]string `json:"SecretData"`
	Annotations map[string]string `json:"SecretAnnotations"`
}

// DuploTenantExtConnSecurityGroupSource represents an external connection SG source for a Duplo tenant.
type DuploTenantExtConnSecurityGroupSource struct {
	Description string `json:"Description"`
	Type        int    `json:"Type"`
	Value       string `json:"Value"`
}

// DuploTenantExtConnSecurityGroupRule represents just-in-time EKS credentials in Duplo
type DuploTenantExtConnSecurityGroupRule struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	State    string                                   `json:"State,omitempty"`
	Protocol string                                   `json:"ServiceProtocol,omitempty"`
	Type     int                                      `json:"ServiceType,omitempty"`
	FromPort int                                      `json:"FromPort,omitempty"`
	Sources  *[]DuploTenantExtConnSecurityGroupSource `json:"SourceInfos,omitempty"`
	ToPort   int                                      `json:"ToPort,omitempty"`
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

// TenantGetConfig retrieves tenant configuration metadata via the Duplo API.
func (c *Client) TenantGetConfig(tenantID string) (*DuploTenantConfig, error) {
	list := []DuploKeyStringValue{}
	err := c.getAPI(fmt.Sprintf("TenantGetConfig(%s)", tenantID), fmt.Sprintf("adminproxy/GetTenantMetadata/%s", tenantID), &list)
	if err != nil {
		return nil, err
	}
	return &DuploTenantConfig{TenantID: tenantID, Metadata: &list}, nil
}

// TenantReplaceConfig replaces tenant configuration metadata via the Duplo API.
func (c *Client) TenantReplaceConfig(config DuploTenantConfig) error {
	existing, err := c.TenantGetConfig(config.TenantID)
	if err != nil {
		return err
	}
	return c.TenantChangeConfig(config.TenantID, existing.Metadata, config.Metadata)
}

// TenantReplaceConfig changes tenant configuration metadata via the Duplo API, using the supplied
// oldConfig and newConfig, for the given tenantID.
func (c *Client) TenantChangeConfig(tenantID string, oldConfig, newConfig *[]DuploKeyStringValue) error {

	// Next, update all keys that are present, keeping a record of each one that is present
	present := map[string]struct{}{}
	if newConfig != nil {
		for _, kv := range *newConfig {
			if err := c.TenantSetConfigKey(tenantID, kv.Key, kv.Value); err != nil {
				return err
			}
			present[kv.Key] = struct{}{}
		}
	}

	// Finally, delete any keys that are no longer present.
	if oldConfig != nil {
		for _, kv := range *oldConfig {
			if _, ok := present[kv.Key]; !ok {
				if err := c.TenantDeleteConfigKey(tenantID, kv.Key); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// TenantDeleteConfigKey deletes a specific configuration key for a tenant via the Duplo API.
func (c *Client) TenantDeleteConfigKey(tenantID, key string) error {
	rq := DuploTenantConfigUpdateRequest{TenantID: tenantID, State: "delete", Key: key}
	return c.postAPI(fmt.Sprintf("TenantDeleteConfigKey(%s, %s)", tenantID, key), "adminproxy/TenantMetadataUpdate", &rq, nil)
}

// TenantSetConfigKey set a specific configuration key for a tenant via the Duplo API.
func (c *Client) TenantSetConfigKey(tenantID, key, value string) error {
	rq := DuploTenantConfigUpdateRequest{TenantID: tenantID, Key: key, Value: value}
	return c.postAPI(fmt.Sprintf("TenantSetConfigKey(%s, %s)", tenantID, key), "adminproxy/TenantMetadataUpdate", &rq, nil)
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

// GetTenantK8sCredentials retrieves just-in-time K8S cluster credentials via the Duplo API..
func (c *Client) GetTenantK8sCredentials(tenantID string) (*DuploTenantK8sCredentials, error) {
	creds := DuploTenantK8sCredentials{}
	err := c.getAPI(fmt.Sprintf("GetTenantEksCredentials(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetK8ClusterConfigByTenant", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// GetEksCredentials retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetTenantEksSecret(tenantID string) (*DuploTenantEksSecret, error) {
	creds := DuploTenantEksSecret{}
	err := c.getAPI(fmt.Sprintf("GetTenantEksSecret(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetEksSecret", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// TenantGetExtConnSecurityGroups retrieves a list of the external connection security group rules for a Duplo tenant.
func (c *Client) TenantGetExtConnSecurityGroups(tenantID string) (*[]DuploTenantExtConnSecurityGroupRule, error) {
	list := []DuploTenantExtConnSecurityGroupRule{}
	err := c.getAPI(fmt.Sprintf("TenantGetExtConnSecurityGroups(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetAllTenantExtConnSgRules", tenantID), &list)
	return &list, err
}

// TenantUpdateExtConnSecurityGroupRule creates or updates an external connection security group rule for a Duplo tenant.
func (c *Client) TenantUpdateExtConnSecurityGroupRule(rq DuploTenantExtConnSecurityGroupRule) error {
	rq.State = ""
	return c.postAPI(fmt.Sprintf("TenantUpdateExtConnSecurityGroupRule(%s, %v)", rq.TenantID, rq.Sources),
		fmt.Sprintf("subscriptions/%s/TenantExtConnSgRuleUpdate", rq.TenantID),
		&rq,
		nil)
}

// TenantDeleteExtConnSecurityGroupRule deletes an external connection security group rule for a Duplo tenant.
func (c *Client) TenantDeleteExtConnSecurityGroupRule(rq DuploTenantExtConnSecurityGroupRule) error {
	rq.State = "delete"
	return c.postAPI(fmt.Sprintf("TenantDeleteExtConnSecurityGroupRule(%s, %v)", rq.TenantID, rq.Sources),
		fmt.Sprintf("subscriptions/%s/TenantExtConnSgRuleUpdate", rq.TenantID),
		&rq,
		nil)
}
