package duplosdk

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
	IsOtpNeeded                    bool   `json:"IsOtpNeeded"`
	IsAwsAdminJITEnabled           bool   `json:"IsAwsAdminJITEnabled"`
	IsDuploOpsEnabled              bool   `json:"IsDuploOpsEnabled"`
	IsTagsBasedResourceMgmtEnabled bool   `json:"IsTagsBasedResourceMgmtEnabled"`
	DevopsManagerHostname          string `json:"DevopsManagerHostname"`
	TenantNameMaxLength            int    `json:"TenantNameMaxLength"`
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
