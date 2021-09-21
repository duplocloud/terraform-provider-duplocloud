package duplosdk

import (
	"fmt"
)

const (
	// SGSourceTypeTenant represents a Duplo Tenant as an SG rule source
	SGSourceTypeTenant int = 0

	// SGSourceTypeIPAddress represents an IP Address as an SG rule source
	SGSourceTypeIPAddress int = 1
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
	TenantID string `json:"-"`

	ConsoleURL      string `json:"ConsoleUrl,omitempty"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Region          string `json:"Region"`
	SessionToken    string `json:"SessionToken,omitempty"`
}

// DuploTenantEksCredentials represents just-in-time EKS credentials in Duplo
type DuploTenantK8sCredentials struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                           string `json:"Name"`
	APIServer                      string `json:"ApiServer"`
	Token                          string `json:"Token"`
	AwsRegion                      string `json:"AwsRegion"`
	K8sProvider                    int    `json:"K8Provider,omitempty"`
	CertificateAuthorityDataBase64 string `json:"CertificateAuthorityDataBase64,omitempty"`
	DefaultNamespace               string `json:"DefaultNamespace,omitempty"`
}

// DuploTenantEksSecret represents just-in-time EKS credentials in Duplo
type DuploTenantEksSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

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
	TenantID string `json:"-"`

	State    string                                   `json:"State,omitempty"`
	Protocol string                                   `json:"ServiceProtocol,omitempty"`
	Type     int                                      `json:"ServiceType,omitempty"`
	FromPort int                                      `json:"FromPort,omitempty"`
	Sources  *[]DuploTenantExtConnSecurityGroupSource `json:"SourceInfos,omitempty"`
	ToPort   int                                      `json:"ToPort,omitempty"`
}

// TenantGet retrieves a tenant via the Duplo API.
func (c *Client) TenantGet(tenantID string) (*DuploTenant, ClientError) {
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
func (c *Client) TenantCreate(rq DuploTenant) (string, ClientError) {
	rp := ""
	err := c.postAPI(fmt.Sprintf("TenantCreate(%s, %s)", rq.AccountName, rq.PlanID), "admin/AddTenant", &rq, &rp)
	if err != nil {
		return "", err
	}
	return rp, err
}

// TenantDelete deletes an AWS host via the Duplo API.
func (c *Client) TenantDelete(tenantID string) ClientError {
	return c.postAPI(fmt.Sprintf("TenantDelete(%s)", tenantID), fmt.Sprintf("admin/DeleteTenant/%s", tenantID), "", nil)
}

// ListTenantsForUser retrieves a list of tenants for the current user via the Duplo API.
func (c *Client) ListTenantsForUser() (*[]DuploTenant, ClientError) {
	list := []DuploTenant{}
	err := c.getAPI("ListTenantsForUser()", "admin/GetTenantsForUser", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// ListTenantsForUserByPlan retrieves a list of tenants with the given plan for the current user via the Duplo API.
// If the planID is an empty string, returns all
func (c *Client) ListTenantsForUserByPlan(planID string) (*[]DuploTenant, ClientError) {
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
func (c *Client) GetTenantByNameForUser(name string) (*DuploTenant, ClientError) {
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
func (c *Client) GetTenantForUser(tenantID string) (*DuploTenant, ClientError) {
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
func (c *Client) TenantGetConfig(tenantID string) (*DuploTenantConfig, ClientError) {
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
func (c *Client) TenantChangeConfig(tenantID string, oldConfig, newConfig *[]DuploKeyStringValue) ClientError {

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
func (c *Client) TenantDeleteConfigKey(tenantID, key string) ClientError {
	rq := DuploTenantConfigUpdateRequest{TenantID: tenantID, State: "delete", Key: key}
	return c.postAPI(fmt.Sprintf("TenantDeleteConfigKey(%s, %s)", tenantID, key), "adminproxy/TenantMetadataUpdate", &rq, nil)
}

// TenantSetConfigKey set a specific configuration key for a tenant via the Duplo API.
func (c *Client) TenantSetConfigKey(tenantID, key, value string) ClientError {
	rq := DuploTenantConfigUpdateRequest{TenantID: tenantID, Key: key, Value: value}
	return c.postAPI(fmt.Sprintf("TenantSetConfigKey(%s, %s)", tenantID, key), "adminproxy/TenantMetadataUpdate", &rq, nil)
}

// TenantGetAwsRegion retrieves a tenant's AWS region via the Duplo API.
func (c *Client) TenantGetAwsRegion(tenantID string) (string, ClientError) {
	awsRegion := ""
	err := c.getAPI(fmt.Sprintf("TenantGetAwsRegion(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetAwsRegionId", tenantID), &awsRegion)
	return awsRegion, err
}

// TenantGetAwsCredentials retrieves just-in-time AWS credentials for a tenant via the Duplo API.
func (c *Client) TenantGetAwsCredentials(tenantID string) (*DuploTenantAwsCredentials, ClientError) {
	creds := DuploTenantAwsCredentials{}
	err := c.getAPI(fmt.Sprintf("TenantGetAwsCredentials(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetAwsConsoleTokenUrl", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// TenantGetInternalSubnets retrieves a list of the internal subnets for a tenant via the Duplo API.
func (c *Client) TenantGetInternalSubnets(tenantID string) ([]string, ClientError) {
	list := []string{}
	err := c.getAPI(fmt.Sprintf("TenantGetInternalSubnets(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetInternalSubnets", tenantID), &list)
	return list, err
}

// TenantGetExternalSubnets retrieves a list of the internal subnets for a tenant via the Duplo API.
func (c *Client) TenantGetExternalSubnets(tenantID string) ([]string, ClientError) {
	list := []string{}
	err := c.getAPI(fmt.Sprintf("TenantGetExternalSubnets(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetExternalSubnets", tenantID), &list)
	return list, err
}

// TenantGetAwsAccountID retrieves the AWS account ID via the Duplo API.
func (c *Client) TenantGetAwsAccountID(tenantID string) (string, ClientError) {
	awsAccountID := ""
	err := c.getAPI(fmt.Sprintf("TenantGetAwsAccountID(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetTenantAwsAccountId", tenantID), &awsAccountID)
	return awsAccountID, err
}

// TenantGetGcpProjectID retrieves the GCP project ID via the Duplo API.
func (c *Client) TenantGetGcpProjectID(tenantID string) (string, ClientError) {
	plan, err := c.GetTenantPlan(tenantID)
	if err != nil {
		return "", err
	}

	var gcpConfig map[string]interface{}
	if plan.CloudPlatforms != nil && len(*plan.CloudPlatforms) > 0 {
		for _, cp := range *plan.CloudPlatforms {
			if len(cp.GoogleConfig) > 0 {
				gcpConfig = cp.GoogleConfig
				break
			}
		}
	}

	if projectID, ok := gcpConfig["GcpProjectId"]; ok && projectID != nil && projectID.(string) != "" {
		return projectID.(string), nil
	}

	return "", clientError{message: "No such GCP project"}
}

// GetTenantK8sCredentials retrieves just-in-time K8S cluster credentials via the Duplo API..
func (c *Client) GetTenantK8sCredentials(tenantID string) (*DuploTenantK8sCredentials, ClientError) {
	creds := DuploTenantK8sCredentials{}
	err := c.getAPI(fmt.Sprintf("GetTenantEksCredentials(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetK8ClusterConfigByTenant", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// GetTenantK8sServiceAccountToken retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetTenantK8sJitAccess(tenantID string) (*DuploTenantK8sCredentials, ClientError) {
	creds := DuploTenantK8sCredentials{}
	err := c.getAPI(fmt.Sprintf("GetTenantK8sJitAccess(%s)", tenantID), fmt.Sprintf("v3/subscriptions/%s/k8s/jitAccess", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// GetTenantEksSecret retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetTenantEksSecret(tenantID string) (*DuploTenantEksSecret, ClientError) {
	creds := DuploTenantEksSecret{}
	err := c.getAPI(fmt.Sprintf("GetTenantEksSecret(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetEksSecret", tenantID), &creds)
	if err != nil {
		return nil, err
	}
	creds.TenantID = tenantID
	return &creds, nil
}

// TenantGetExtConnSecurityGroupRules retrieves a list of the external connection security group rules for a Duplo tenant.
func (c *Client) TenantGetExtConnSecurityGroupRules(tenantID string) (*[]DuploTenantExtConnSecurityGroupRule, ClientError) {
	list := []DuploTenantExtConnSecurityGroupRule{}
	err := c.postAPI(fmt.Sprintf("TenantGetExtConnSecurityGroups(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAllTenantExtConnSgRules", tenantID),
		map[string]interface{}{},
		&list)
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, err
}

// TenantGetExtConnSecurityGroupRule retrieves an external connection security group rule for a Duplo tenant.
func (c *Client) TenantGetExtConnSecurityGroupRule(rq *DuploTenantExtConnSecurityGroupRule) (*DuploTenantExtConnSecurityGroupRule, ClientError) {
	list, err := c.TenantGetExtConnSecurityGroupRules(rq.TenantID)
	if err != nil {
		return nil, err
	}

	for _, rule := range *list {
		if rule.Type == rq.Type && rule.Protocol == rq.Protocol && rule.FromPort == rq.FromPort && rule.ToPort == rq.ToPort && rule.Sources != nil && len(*rule.Sources) > 0 {
			for _, source := range *rq.Sources {
				if source.Type == rq.Type && source.Value == (*rq.Sources)[0].Value {
					matched := DuploTenantExtConnSecurityGroupRule{
						TenantID: rq.TenantID,
						Protocol: rq.Protocol,
						Type:     rq.Type,
						FromPort: rq.FromPort,
						ToPort:   rq.ToPort,
						Sources: &[]DuploTenantExtConnSecurityGroupSource{{
							Description: source.Description,
							Type:        source.Type,
							Value:       source.Value,
						}},
					}
					return &matched, nil
				}
			}
		}
	}

	return nil, nil
}

// TenantUpdateExtConnSecurityGroupRule creates or updates an external connection security group rule for a Duplo tenant.
func (c *Client) TenantUpdateExtConnSecurityGroupRule(rq *DuploTenantExtConnSecurityGroupRule) ClientError {
	rq.State = ""
	return c.postAPI(fmt.Sprintf("TenantUpdateExtConnSecurityGroupRule(%s, %v)", rq.TenantID, rq.Sources),
		fmt.Sprintf("subscriptions/%s/TenantExtConnSgRuleUpdate", rq.TenantID),
		rq,
		nil)
}

// TenantDeleteExtConnSecurityGroupRule deletes an external connection security group rule for a Duplo tenant.
func (c *Client) TenantDeleteExtConnSecurityGroupRule(rq *DuploTenantExtConnSecurityGroupRule) ClientError {
	rq.State = "delete"
	return c.postAPI(fmt.Sprintf("TenantDeleteExtConnSecurityGroupRule(%s, %v)", rq.TenantID, rq.Sources),
		fmt.Sprintf("subscriptions/%s/TenantExtConnSgRuleUpdate", rq.TenantID),
		rq,
		nil)
}

// GetTenantPlan retrieves non-admin plan details via the Duplo API.
func (c *Client) GetTenantPlan(tenantID string) (*DuploPlan, ClientError) {
	plan := DuploPlan{}
	err := c.getAPI(fmt.Sprintf("GetTenantPlan(%s)", tenantID), fmt.Sprintf("subscriptions/%s/GetTenantPlan", tenantID), &plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (c *Client) TenantGetDockerCredentials(tenantId string) (map[string]interface{}, ClientError) {
	rp := map[string]interface{}{}
	err := c.getAPI(fmt.Sprintf("TenantGetDockerCredentials(%s)", tenantId), fmt.Sprintf("subscriptions/%s/GetDockerCredentialsAnonymized", tenantId), &rp)
	if err != nil {
		return nil, err
	}
	return rp, nil
}

func (c *Client) TenantUpdateDockerCredentials(tenantId string, data map[string]interface{}) ClientError {
	return c.postAPI(fmt.Sprintf("TenantUpdateDockerCredentials(%s)", tenantId),
		fmt.Sprintf("subscriptions/%s/UpdateDockerCredentials", tenantId),
		data,
		nil)
}
