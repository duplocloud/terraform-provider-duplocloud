package duplosdk

import (
	"strings"
)

type DuploSystemFeatures struct {
	IsKatkitEnabled      bool     `json:"IsKatkitEnabled"`
	IsSignupEnabled      bool     `json:"IsSignupEnabled"`
	IsComplianceEnabled  bool     `json:"IsComplianceEnabled"`
	IsBillingEnabled     bool     `json:"IsBillingEnabled"`
	IsSiemEnabled        bool     `json:"IsSiemEnabled"`
	IsAwsCloudEnabled    bool     `json:"IsAwsCloudEnabled"`
	AwsRegions           []string `json:"AwsRegions"`
	DefaultAwsAccount    string   `json:"DefaultAwsAccount"`
	DefaultAwsRegion     string   `json:"DefaultAwsRegion"`
	IsAzureCloudEnabled  bool     `json:"IsAzureCloudEnabled"`
	AzureRegions         []string `json:"AzureRegions"`
	IsGoogleCloudEnabled bool     `json:"IsGoogleCloudEnabled"`
	EksVersions          struct {
		DefaultVersion    string   `json:"DefaultVersion"`
		SupportedVersions []string `json:"SupportedVersions"`
	} `json:"EksVersions"`
	IsOtpNeeded                    bool                 `json:"IsOtpNeeded"`
	IsAwsAdminJITEnabled           bool                 `json:"IsAwsAdminJITEnabled"`
	IsDuploOpsEnabled              bool                 `json:"IsDuploOpsEnabled"`
	IsTagsBasedResourceMgmtEnabled bool                 `json:"IsTagsBasedResourceMgmtEnabled"`
	DevopsManagerHostname          string               `json:"DevopsManagerHostname"`
	TenantNameMaxLength            int                  `json:"TenantNameMaxLength"`
	AzureResourcePrefix            *AzureResourcePrefix `json:"AzureResourcePrefix,omitempty"`
	S3BucketNamePrefix             string               `json:"S3BucketNamePrefix"`
}

type AzureResourcePrefix struct {
	IsAzureCustomPrefixesEnabled bool                      `json:"IsAzureCustomPrefixesEnabled"`
	InfraRgPrefix                string                    `json:"InfraRgPrefix"`
	TenantRgPrefix               string                    `json:"TenantRgPrefix"`
	BackupRgPrefix               string                    `json:"BackupRgPrefix"`
	ResourceTypePrefix           []AzureResourceTypePrefix `json:"ResourceTypePrefix"`
}

type AzureResourceTypePrefix struct {
	Type   string `json:"Type"`
	Prefix string `json:"Prefix"`
}

// DuploAdminAwsCredentials represents just-in-time admin AWS credentials from Duplo
type DuploAdminAwsCredentials struct {
	ConsoleURL      string `json:"ConsoleUrl,omitempty"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Region          string `json:"Region"`
	SessionToken    string `json:"SessionToken,omitempty"`
	Validity        int    `json:"Validity,omitempty"`
}

// GetAwsAccountID retrieves the AWS account ID via the Duplo API.
func (c *Client) GetAwsAccountID() (string, ClientError) {
	awsAccount := ""
	err := c.getAPI("GetAwsAccountID()", "adminproxy/GetAwsAccountId", &awsAccount)
	return awsAccount, err
}

// AdminGetAwsCredentials retrieves just-in-time admin AWS credentials via the Duplo API.
func (c *Client) AdminGetAwsCredentials() (*DuploAdminAwsCredentials, ClientError) {
	creds := DuploAdminAwsCredentials{}
	err := c.getAPI("AdminGetAwsCredentials()", "adminproxy/GetJITAwsConsoleAccessUrl", &creds)
	if err != nil {
		return nil, err
	}
	return &creds, nil
}

func (c *Client) AdminGetSystemFeatures() (*DuploSystemFeatures, ClientError) {
	features := DuploSystemFeatures{}
	err := c.getAPI("AdminGetSystemFeatures()", "v3/features/system", &features)
	if err != nil {
		return nil, err
	}
	return &features, nil
}

func (c *Client) IsAzureCustomPrefixesEnabled() bool {
	features := DuploSystemFeatures{}
	err := c.getAPI("AdminGetSystemFeatures()", "v3/features/system", &features)
	if err != nil || features.AzureResourcePrefix == nil || !features.AzureResourcePrefix.IsAzureCustomPrefixesEnabled {
		return false
	}
	return true
}

func (c *Client) getPrefixFromResourceType(resourceType string, isInfraResource bool) (string, error) {
	features := DuploSystemFeatures{}
	resourcePrefixEnabled := false
	err := c.getAPI("AdminGetSystemFeatures()", "v3/features/system", &features)
	if err != nil {
		return "", err
	}
	azureResourcePrefix := features.AzureResourcePrefix
	if azureResourcePrefix == nil {
		return "", nil
	}

	prefix := ""
	resourceTypePrefix := azureResourcePrefix.ResourceTypePrefix
	for i := range resourceTypePrefix {
		if resourceTypePrefix[i].Type == resourceType {
			prefix = resourceTypePrefix[i].Prefix
			resourcePrefixEnabled = true
			break
		}
	}
	if !resourcePrefixEnabled && isInfraResource && (resourceType == "subnet" || resourceType == "nsg") {
		prefix += azureResourcePrefix.InfraRgPrefix
	}

	return prefix, nil
}

func (c *Client) TrimPrefixSuffixFromResourceName(resourceName, resourceType string, isInfraResource bool) string {
	prefix, err := c.getPrefixFromResourceType(resourceType, isInfraResource)
	if err != nil || prefix == "" {
		return resourceName
	}
	return strings.TrimPrefix(resourceName, prefix)
}

func (c *Client) AddPrefixSuffixFromResourceName(resourceName, resourceType string, isInfraResource bool) string {
	prefix, err := c.getPrefixFromResourceType(resourceType, isInfraResource)
	if err != nil || prefix == "" {
		return resourceName
	}
	return prefix + resourceName
}
