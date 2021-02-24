package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploTenant represents a Duplo tenant
type DuploTenant struct {
	TenantID    string `json:"TenantId,omitempty"`
	AccountName string `json:"AccountName"`
	PlanID      string `json:"PlanID"`
}

// DuploAwsKmsKey represents an AWS KMS key for a Duplo tenant
type DuploAwsKmsKey struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Arn         string `json:"Arn,omitempty"`
	KeyName     string `json:"KeyName,omitempty"`
	KeyID       string `json:"KeyId,omitempty"`
	KeyArn      string `json:"KeyArn,omitempty"`
	Description string `json:"Description,omitempty"`
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

// TenantSchema returns a Terraform resource schema for a Duplo Tenant
func TenantSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"account_name": {
			Type:     schema.TypeString,
			Computed: false,
			Required: true,
		},
		"plan_id": {
			Type:     schema.TypeString,
			Computed: false,
			Required: true,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
	}
}

// TenantFlatten converts a list of Duplo SDK objects into Terraform resource data
func (c *Client) TenantFlatten(duploTenant *DuploTenant, d *schema.ResourceData) map[string]interface{} {
	if duploTenant != nil {
		c := make(map[string]interface{})
		///--- set
		c["account_name"] = duploTenant.AccountName
		c["tenant_id"] = duploTenant.TenantID
		c["plan_id"] = duploTenant.PlanID

		jsonData, _ := json.Marshal(duploTenant)
		log.Printf("[TRACE] duplo-TenantFlatten ********: jsonData %s ", jsonData)
		return c
	}
	return nil
}

// TenantFromState converts Terraform resource data representing an AWS host into partial Terraform resource data
func (c *Client) TenantFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploTenant, error) {
	url := c.TenantURL(d)
	var apiStr = fmt.Sprintf("duplo-TenantFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-TenantFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)

	duploObject := new(DuploTenant)
	///--- set
	duploObject.AccountName = d.Get("account_name").(string)
	duploObject.PlanID = d.Get("plan_id").(string)

	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] %s ********: jsonData %s ", apiStr, jsonData)
	return duploObject, nil
}

// TenantSetID populates the resource ID based on tenant_id
func (c *Client) TenantSetID(d *schema.ResourceData) string {
	tenantID := d.Get("tenant_id").(string)
	///--- set
	id := fmt.Sprintf("v2/admin/TenantV2/%s", tenantID)
	log.Printf("[TRACE] duplo-TenantSetId %s  ********: %s", id, tenantID)
	d.SetId(id)
	return id
}

// TenantURL returns the base API URL for crud -- get + delete
func (c *Client) TenantURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-TenantUrl %s 1 ********: %s", api, host)
	return host
}

// TenantURLList returns the base API URL for crud -- get list + create + update
func (c *Client) TenantURLList(d *schema.ResourceData) string {
	api := "v2/admin/TenantV2"
	url := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-TenantUrlList %s 1 ********: %s", api, url)
	return url
}

/////////// common place to get url + Id : follow Azure  style Ids for import//////////

// TenantsFlatten converts a list of Duplo SDK objects into Terraform resource data
func (c *Client) TenantsFlatten(duploTenants *[]DuploTenant, d *schema.ResourceData) []interface{} {
	if duploTenants != nil {
		ois := make([]interface{}, len(*duploTenants), len(*duploTenants))
		for i, duploTenant := range *duploTenants {
			ois[i] = c.TenantFlatten(&duploTenant, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-TenantFlatten ******** jsonData: \n%s", jsonData)
		return ois
	}
	return make([]interface{}, 0)
}

// TenantFillGet converts a Duplo SDK object into Terraform resource data
func (c *Client) TenantFillGet(duploTenant *DuploTenant, d *schema.ResourceData) error {
	if duploTenant != nil {
		ois := c.TenantFlatten(duploTenant, d)
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-TenantFillGet ********: ois %s ", jsonData)
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-TenantFillGet Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	return fmt.Errorf("Tenant not found 2")
}

/////////  API list //////////

// TenantGetList retrieves a list of tenants via the Duplo API.
func (c *Client) TenantGetList(d *schema.ResourceData, m interface{}) (*[]DuploTenant, error) {

	url := c.TenantURLList(d)

	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-GetListTenant %s 1 ********: %s", url, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-GetListTenant %s 2 ********: \n%s", url, bodyString)

	duploTenants := []DuploTenant{}
	err = json.Unmarshal(body, &duploTenants)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-GetListTenant %s 3 ********: %d", url, len(duploTenants))

	return &duploTenants, nil
}

// TenantGet retrieves a tenant via the Duplo API.
func (c *Client) TenantGet(d *schema.ResourceData, m interface{}) error {

	api := d.Id()
	url := c.TenantURL(d)
	log.Printf("[TRACE] duplo-TenantGet %s 1 ********: %s ", api, url)

	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGet %s 2 ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGet %s 3 ********: %s", api, bodyString)

	duploTenant := DuploTenant{}
	err = json.Unmarshal(body, &duploTenant)
	if err != nil {
		return err
	}
	log.Printf("[TRACE] duplo-TenantGet %s 4 ********: %s", api, duploTenant.AccountName)
	if duploTenant.TenantID != "" {
		c.TenantFillGet(&duploTenant, d)
		log.Printf("[TRACE] duplo-TenantGet 5 FOUND ***** : '%s' '%s'", duploTenant.AccountName, duploTenant.TenantID)
		return nil
	}

	accountName := d.Get("account_name").(string)
	return fmt.Errorf("Tenant not found %s : %s. Please ", accountName, duploTenant.TenantID)
}

/////////  API Create / update //////////

// TenantCreate creates a tenant via the Duplo API.
func (c *Client) TenantCreate(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	return c.TenantCreateOrUpdate(d, m, false)
}

// TenantUpdate updates a tenant via the Duplo API.
func (c *Client) TenantUpdate(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	// No-op
	return nil, nil
}

// TenantCreateOrUpdate creates or updates a tenant via the Duplo API.
func (c *Client) TenantCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploTenant, error) {
	duploObject, _ := c.TenantFromState(d, m, isUpdate)
	log.Printf("[TRACE] duplo-TenantCreate duploObject.AccountName %s PlanID %s ********:", duploObject.AccountName, duploObject.PlanID)

	url := c.TenantURLList(d)
	log.Printf("[TRACE] duplo-TenantCreate %s 1 ********:", url)

	//
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostCreate %s 2 ********: %s", url, err.Error())
		return nil, err
	}

	jsonStr := string(rb) //strings.NewReader(string(rb))
	log.Printf("[TRACE] duplo-AwsHostCreate %s 3 ********: %s", url, jsonStr)

	req, err := http.NewRequest("POST", url, strings.NewReader(string(rb)))
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostCreate %s 4 ********: %s", url, err.Error())
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostCreate %s 5 ********: %s", url, err.Error())
		return nil, err
	}
	if body != nil {
		duploTenant := DuploTenant{}
		err = json.Unmarshal(body, &duploTenant)
		if err != nil {
			return nil, err
		}
		log.Printf("[TRACE] duplo-AwsHostCreate %s 5 ********: %s %s %s", url, duploTenant.AccountName, duploTenant.TenantID, duploTenant.PlanID)
		//todo: move this part up
		d.Set("tenant_id", duploTenant.TenantID)
		c.TenantSetID(d) //??
		return nil, nil
	}
	errMsg := fmt.Errorf("ERROR: in create %s, %s  body: %s", url, duploObject.AccountName, body)
	return nil, errMsg
}

// TenantDelete deletes an AWS host via the Duplo API.
func (c *Client) TenantDelete(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	// No-op
	return nil, nil
}

// TenantGetAwsRegion retrieves a tenant's AWS region via the Duplo API.
func (c *Client) TenantGetAwsRegion(tenantID string) (string, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetAwsRegionId", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantGetAwsRegion 1 ********: %s ", url)

	// Get the AWS region from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGetAwsRegion 2 ********: %s", err.Error())
		return "", err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGetAwsRegion 3 ********: %s", bodyString)

	// Return it as a string.
	awsRegion := ""
	err = json.Unmarshal(body, &awsRegion)
	if err != nil {
	}
	log.Printf("[TRACE] duplo-TenantGetAwsRegion 4 ********: %s", awsRegion)

	return awsRegion, nil
}

// TenantGetAwsCredentials retrieves just-in-time AWS credentials for a tenant via the Duplo API.
func (c *Client) TenantGetAwsCredentials(tenantID string) (*DuploTenantAwsCredentials, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetAwsConsoleTokenUrl", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantGetAwsCredentials 1 ********: %s ", url)

	// Get the AWS region from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGetAwsCredentials 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGetAwsCredentials 3 ********: %s", bodyString)

	// Return it as an object.
	creds := DuploTenantAwsCredentials{}
	err = json.Unmarshal(body, &creds)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetAwsCredentials 4 ********: %s", creds.AccessKeyID)
	creds.TenantID = tenantID
	return &creds, nil
}

// TenantGetInternalSubnets retrieves a list of the internal subnets for a tenant via the Duplo API.
func (c *Client) TenantGetInternalSubnets(tenantID string) ([]string, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetInternalSubnets", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantGetInternalSubnets 1 ********: %s ", url)

	// Get the list from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGetInternalSubnets 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGetInternalSubnets 3 ********: %s", bodyString)

	// Return it as an object.
	list := []string{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetInternalSubnets 4 ********: %s", strings.Join(list, " "))
	return list, nil
}

// TenantGetPlanKmsKeys retrieves a list of the AWS KMS keys for a tenant via the Duplo API.
func (c *Client) TenantGetPlanKmsKeys(tenantID string) (*[]DuploAwsKmsKey, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetPlanKmsKeys", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantGetPlanKmsKeys 1 ********: %s ", url)

	// Get the list from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGetPlanKmsKeys 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGetPlanKmsKeys 3 ********: %s", bodyString)

	// Return it as a list.
	list := []DuploAwsKmsKey{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetPlanKmsKeys 4 ********: %d items", len(list))
	for i := range list {
		list[i].TenantID = tenantID
		list[i].Arn = list[i].KeyArn
		list[i].Description = list[i].KeyName
	}
	return &list, nil
}

// TenantGetTenantKmsKey retrieves a tenant specific AWS KMS keys via the Duplo API.
func (c *Client) TenantGetTenantKmsKey(tenantID string) (*DuploAwsKmsKey, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetTenantKmsKey", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantGetTenantKmsKey 1 ********: %s ", url)

	// Get the key from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantGetTenantKmsKey 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantGetTenantKmsKey 3 ********: %s", bodyString)

	// Return it as an object.
	kms := DuploAwsKmsKey{}
	err = json.Unmarshal(body, &kms)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetTenantKmsKey 4 ********")
	kms.TenantID = tenantID
	kms.KeyArn = kms.Arn
	kms.KeyName = kms.Description
	return &kms, nil
}
