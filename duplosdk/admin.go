package duplosdk

// DuploAdminAwsCredentials represents just-in-time admin AWS credentials from Duplo
type DuploAdminAwsCredentials struct {
	ConsoleURL      string `json:"ConsoleUrl,omitempty"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Region          string `json:"Region"`
	SessionToken    string `json:"SessionToken,omitempty"`
}

// GetAwsAccountID retrieves the AWS account ID via the Duplo API.
func (c *Client) GetAwsAccountID() (string, error) {
	awsAccount := ""
	err := c.getAPI("GetAwsAccountID()", "admin/GetAwsAccountId", &awsAccount)
	return awsAccount, err
}

// AdminGetAwsCredentials retrieves just-in-time admin AWS credentials via the Duplo API.
func (c *Client) AdminGetAwsCredentials() (*DuploAdminAwsCredentials, error) {
	creds := DuploAdminAwsCredentials{}
	err := c.getAPI("AdminGetAwsCredentials()", "adminproxy/GetJITAwsConsoleAccessUrl", &creds)
	if err != nil {
		return nil, err
	}
	return &creds, nil
}
