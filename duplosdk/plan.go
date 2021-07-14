package duplosdk

import "fmt"

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

// GetK8sCredentials retrieves just-in-time kubernetes credentials via the Duplo API.
func (c *Client) GetPlanK8sJitAccess(planID string) (*DuploEksCredentials, ClientError) {
	creds := DuploEksCredentials{}
	err := c.getAPI(fmt.Sprintf("GetK8sCredentials(%s)", planID), fmt.Sprintf("v3/admin/plans/%s/k8sConfig", planID), &creds)
	if err != nil {
		return nil, err
	}
	creds.PlanID = planID
	return &creds, nil
}

// PlanGetCertificateList retrieves a list of plan certificates via the Duplo API.
func (c *Client) PlanCertificateGetList(planID string) (*[]DuploPlanCertificate, ClientError) {
	list := []DuploPlanCertificate{}
	err := c.getAPI("PlanCertificateGetList()", fmt.Sprintf("v3/admin/plans/%s/certificates", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// TenantReplaceConfig replaces plan certificates via the Duplo API.
func (c *Client) PlanReplaceCertificates(planID string, newCerts *[]DuploPlanCertificate) ClientError {
	existing, err := c.PlanCertificateGetList(planID)
	if err != nil {
		return err
	}
	return c.PlanChangeCertificates(planID, existing, newCerts)
}

// PlanChangeCertificates changes plan certificates via the Duplo API, using the supplied
// oldConfig and newConfig, for the given planID.
func (c *Client) PlanChangeCertificates(planID string, oldCerts, newCerts *[]DuploPlanCertificate) ClientError {

	// Next, update all certs that are present, keeping a record of each one that is present
	present := map[string]struct{}{}
	if newCerts != nil {
		for _, pc := range *newCerts {
			if err := c.PlanSetCertificate(planID, pc); err != nil {
				return err
			}
			present[pc.CertificateName] = struct{}{}
		}
	}

	// Finally, delete any certs that are no longer present.
	if oldCerts != nil {
		for _, pc := range *oldCerts {
			if _, ok := present[pc.CertificateName]; !ok {
				if err := c.PlanDeleteCertificate(planID, pc.CertificateName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// PlanDeleteCertificate deletes a specific certificate for a plan via the Duplo API.
func (c *Client) PlanDeleteCertificate(planID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PlanDeleteCertificate(%s, %s)", planID, name),
		fmt.Sprintf("v3/admin/plans/%s/certificates/%s", planID, name),
		nil)
}

// PlanSetCertificate set a specific configuration key for a tenant via the Duplo API.
func (c *Client) PlanSetCertificate(planID string, cert DuploPlanCertificate) ClientError {
	var rp DuploPlanCertificate
	return c.postAPI(
		fmt.Sprintf("PlanSetCertificate(%s, %s)", planID, cert.CertificateName),
		fmt.Sprintf("v3/admin/plans/%s/certificates", planID),
		&cert,
		&rp)
}

// PlanGetImageList retrieves a list of plan images via the Duplo API.
func (c *Client) PlanImageGetList(planID string) (*[]DuploPlanImage, ClientError) {
	list := []DuploPlanImage{}
	err := c.getAPI("PlanImageGetList()", fmt.Sprintf("v3/admin/plans/%s/images", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// TenantReplaceConfig replaces plan certificates via the Duplo API.
func (c *Client) PlanReplaceImages(planID string, newImages *[]DuploPlanImage) ClientError {
	existing, err := c.PlanImageGetList(planID)
	if err != nil {
		return err
	}
	return c.PlanChangeImages(planID, existing, newImages)
}

// PlanChangeImages changes plan certificates via the Duplo API, using the supplied
// oldConfig and newConfig, for the given planID.
func (c *Client) PlanChangeImages(planID string, oldImages, newImages *[]DuploPlanImage) ClientError {

	// Next, update all certs that are present, keeping a record of each one that is present
	present := map[string]struct{}{}
	if newImages != nil {
		for _, pc := range *newImages {
			if err := c.PlanSetImage(planID, pc); err != nil {
				return err
			}
			present[pc.Name] = struct{}{}
		}
	}

	// Finally, delete any certs that are no longer present.
	if oldImages != nil {
		for _, pc := range *oldImages {
			if _, ok := present[pc.Name]; !ok {
				if err := c.PlanDeleteImage(planID, pc.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// PlanDeleteImage deletes a specific certificate for a plan via the Duplo API.
func (c *Client) PlanDeleteImage(planID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PlanDeleteImage(%s, %s)", planID, name),
		fmt.Sprintf("v3/admin/plans/%s/images/%s", planID, name),
		nil)
}

// PlanSetImage set a specific configuration key for a tenant via the Duplo API.
func (c *Client) PlanSetImage(planID string, cert DuploPlanImage) ClientError {
	var rp DuploPlanImage
	return c.postAPI(
		fmt.Sprintf("PlanSetImage(%s, %s)", planID, cert.Name),
		fmt.Sprintf("v3/admin/plans/%s/images", planID),
		&cert,
		&rp)
}
