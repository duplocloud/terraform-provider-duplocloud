package duplosdk

type DuploPlan struct {
	Name              string                      `json:"Name"`
	NwProvider        int                         `json:"NwProvider,omitempty"` // FIXME: put a proper enum here.
	BlockBYOHosts     bool                        `json:"BlockBYOHosts"`
	Images            *[]DuploPlanImage           `json:"Images,omitempty"`
	AwsConfig         map[string]interface{}      `json:"AwsConfig,omitempty"`
	UnrestrictedExtLB bool                        `json:"UnrestrictedExtLB,omitempty"`
	Capabilities      map[string]interface{}      `json:"Capabilities,omitempty"`
	Certificates      *[]DuploPlanCertificate     `json:"Certificates,omitempty"`
	KmsKeyInfos       *[]DuploPlanKmsKeyInfo      `json:"KmsKeyInfos,omitempty"`
	MetaData          *[]DuploKeyStringValue      `json:"MetaData,omitempty"`
	PlanConfigData    *[]DuploCustomDataEx        `json:"PlanConfigData,omitempty"`
	WafInfos          *[]DuploPlanWafInfo         `json:"WafInfos,omitempty"`
	K8ClusterConfigs  *[]DuploPlanK8ClusterConfig `json:"K8ClusterConfigs,omitempty"`
	CloudPlatforms    *[]DuploPlanCloudPlatform   `json:"CloudPlatforms,omitempty"`
}

type DuploPlanImage struct {
	Name     string                 `json:"Name"`
	ImageId  string                 `json:"ImageId,omitempty"`
	OS       string                 `json:"OS,omitempty"`
	Tags     *[]DuploKeyStringValue `json:"Tags,omitempty"`
	Username string                 `json:"Username,omitempty"`
}

type DuploPlanCertificate struct {
	CertificateName string `json:"CertificateName"`
	CertificateArn  string `json:"CertificateArn"`
}

type DuploPlanKmsKeyInfo struct {
	KeyName string `json:"KeyName,omitempty"`
	KeyArn  string `json:"KeyArn,omitempty"`
	KeyId   string `json:"KeyId,omitempty"`
}

type DuploPlanK8ClusterConfig struct {
	Name                           string `json:"Name,omitempty"`
	ApiServer                      string `json:"ApiServer,omitempty"`
	Token                          string `json:"Token,omitempty"`
	K8Provider                     int    `json:"K8Provider,omitempty"`
	AwsRegion                      string `json:"AwsRegion,omitempty"`
	CertificateAuthorityDataBase64 string `json:"CertificateAuthorityDataBase64,omitempty"`
}

type DuploPlanWafInfo struct {
	WebAclName string `json:"WebAclName,omitempty"`
	WebAclId   string `json:"WebAclId,omitempty"`
}

type DuploPlanCloudPlatform struct {
	Platform     int                    `json:"Cloud"`
	Images       *[]DuploPlanImage      `json:"Images,omitempty"`
	AzureConfig  map[string]interface{} `json:"PlanAzureConfig,omitempty"`
	GoogleConfig map[string]interface{} `json:"PlanGoogleConfig,omitempty"`
}

// PlanGetList retrieves a list of plans via the Duplo API.
func (c *Client) PlanGetList() (*[]DuploPlan, ClientError) {
	list := []DuploPlan{}
	err := c.getAPI("PlanGetList()", "adminproxy/GetPlans", &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// PlanGet retrieves a plan by name via the Duplo API.
func (c *Client) PlanGet(name string) (*DuploPlan, ClientError) {
	list, err := c.PlanGetList()
	if err != nil {
		return nil, err
	}

	for _, plan := range *list {
		if plan.Name == name {
			return &plan, nil
		}
	}

	return nil, nil
}
