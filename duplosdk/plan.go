package duplosdk

import (
	"fmt"
)

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
	DnsConfig         *DuploPlanDnsConfig         `json:"DnsConfig,omitempty"`
}

type DuploPlanDnsConfig struct {
	DomainId          string `json:"DomainId,omitempty"`
	InternalDnsSuffix string `json:"InternalDnsSuffix,omitempty"`
	ExternalDnsSuffix string `json:"ExternalDnsSuffix,omitempty"`
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
	K8sVersion                     string `json:"K8sVersion,omitempty"`
	CertificateAuthorityDataBase64 string `json:"CertificateAuthorityDataBase64,omitempty"`
}

type DuploPlanWafInfo struct {
	WebAclName string `json:"WebAclName,omitempty"`
	WebAclId   string `json:"WebAclId,omitempty"`
}

type DuploPlanCloudPlatform struct {
	Platform     int                    `json:"Cloud"`
	Images       *[]DuploPlanImage      `json:"Images,omitempty"`
	AzureConfig  map[string]interface{} `json:"AzureConfig,omitempty"`
	GoogleConfig map[string]interface{} `json:"GoogleConfig,omitempty"`
}

type DuploPlanNgwAddress struct {
	AllocationId       string `json:"AllocationId"`
	NetworkInterfaceId string `json:"NetworkInterfaceId"`
	PrivateIP          string `json:"PrivateIp"`
	PublicIP           string `json:"PublicIp"`
}

type DuploPlanNgw struct {
	NatGatewayId        string                 `json:"NatGatewayId,omitempty"`
	State               *DuploStringValue      `json:"State,omitempty"`
	SubnetId            string                 `json:"SubnetId"`
	VpcId               string                 `json:"VpcId"`
	Tags                *[]DuploKeyStringValue `json:"Tags,omitempty"`
	NatGatewayAddresses *[]DuploPlanNgwAddress `json:"NatGatewayAddresses,omitempty"`
}

func (c *Client) PlanUpdate(rq *DuploPlan) ClientError {
	return c.postAPI(
		fmt.Sprintf("PlanUpdate(%s)", rq.Name),
		"adminproxy/UpdatePlan",
		&rq,
		nil)
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

// PlanGetCertificate retrieves a certificate of plan via the Duplo API.
func (c *Client) PlanCertificateGet(planID string, name string) (*DuploPlanCertificate, ClientError) {
	cert := DuploPlanCertificate{}
	err := c.getAPI("PlanCertificateGet()", fmt.Sprintf("v3/admin/plans/%s/certificates/%s", planID, name), &cert)
	if err != nil {
		return nil, err
	}
	return &cert, nil
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

// PlanImageGetList retrieves a list of plan images via the Duplo API.
func (c *Client) PlanImageGetList(planID string) (*[]DuploPlanImage, ClientError) {
	list := []DuploPlanImage{}
	err := c.getAPI(fmt.Sprintf("PlanImageGetList(%s)", planID), fmt.Sprintf("v3/admin/plans/%s/images", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// PlanImageGet retrieves a plan images via the Duplo API.
func (c *Client) PlanImageGet(planID, name string) (*DuploPlanImage, ClientError) {
	rp := DuploPlanImage{}
	err := c.getAPI(fmt.Sprintf("PlanImageGet(%s, %s)", planID, name), fmt.Sprintf("v3/admin/plans/%s/images/%s", planID, name), &rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// PlanReplaceImages replaces plan certificates via the Duplo API.
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

// PlanGetConfigList retrieves a list of plan configs via the Duplo API.
func (c *Client) PlanConfigGetList(planID string) (*[]DuploCustomDataEx, ClientError) {
	list := []DuploCustomDataEx{}
	err := c.getAPI("PlanConfigGetList()", fmt.Sprintf("v3/admin/plans/%s/configs", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// TenantReplaceConfig replaces plan configs via the Duplo API.
func (c *Client) PlanReplaceConfigs(planID string, newConfigs *[]DuploCustomDataEx) ClientError {
	existing, err := c.PlanConfigGetList(planID)
	if err != nil {
		return err
	}
	return c.PlanChangeConfigs(planID, existing, newConfigs)
}

// PlanChangeConfigs changes plan configs via the Duplo API, using the supplied
// oldConfigs and newConfigs, for the given planID.
func (c *Client) PlanChangeConfigs(planID string, oldConfigs, newConfigs *[]DuploCustomDataEx) ClientError {

	// Next, update all certs that are present, keeping a record of each one that is present
	present := map[string]struct{}{}
	if newConfigs != nil {
		for _, pc := range *newConfigs {
			if err := c.PlanSetConfig(planID, pc); err != nil {
				return err
			}
			present[pc.Key] = struct{}{}
		}
	}

	// Finally, delete any certs that are no longer present.
	if oldConfigs != nil {
		for _, pc := range *oldConfigs {
			if _, ok := present[pc.Key]; !ok {
				if err := c.PlanDeleteConfig(planID, pc.Key); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// PlanDeleteConfig deletes a specific configuration key for a plan via the Duplo API.
func (c *Client) PlanDeleteConfig(planID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("PlanDeleteConfig(%s, %s)", planID, name),
		fmt.Sprintf("v3/admin/plans/%s/configs/%s", planID, name),
		nil)
}

// PlanSetConfig set a specific configuration key for a tenant via the Duplo API.
func (c *Client) PlanSetConfig(planID string, item DuploCustomDataEx) ClientError {
	var rp DuploCustomDataEx
	return c.postAPI(
		fmt.Sprintf("PlanSetConfig(%s, %s)", planID, item.Key),
		fmt.Sprintf("v3/admin/plans/%s/configs", planID),
		&item,
		&rp)
}

func (c *Client) PlanNgwGetList(planID string) (*[]DuploPlanNgw, ClientError) {
	list := []DuploPlanNgw{}
	err := c.getAPI("PlanNgwGetList()", fmt.Sprintf("v3/admin/plans/%s/nat-gateways", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
