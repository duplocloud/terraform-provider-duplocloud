package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploEksCredentials represents just-in-time EKS credentials in Duplo
type DuploEksCredentials struct {
	// NOTE: The PlanID field does not come from the backend - we synthesize it
	PlanID string `json:"-,omitempty"`

	Name        string `json:"Name"`
	APIServer   string `json:"ApiServer"`
	Token       string `json:"Token"`
	AwsRegion   string `json:"AwsRegion"`
	K8sProvider int    `json:"K8Provider,omitempty"`
}

// DuploInfrastructure represents a Duplo infrastructure
type DuploInfrastructure struct {
	Name               string `json:"Name"`
	AccountId          string `json:"AccountId"`
	Cloud              int    `json:"Cloud"`
	Region             string `json:"Region"`
	AzCount            int    `json:"AzCount"`
	EnableK8Cluster    bool   `json:"EnableK8Cluster"`
	AddressPrefix      string `json:"AddressPrefix"`
	SubnetCidr         int    `json:"SubnetCidr"`
	ProvisioningStatus string `json:"ProvisioningStatus"`
}

/////------ schema ------////
func InfrastructureSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"infra_name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"account_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
			ForceNew: true,
		},
		"cloud": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			ForceNew: true,
			Default:  0,
		},
		"region": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			ForceNew: true,
			Required: true,
		},
		"azcount": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: false,
			ForceNew: true,
			Required: true,
		},
		"enable_k8_cluster": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: false,
			Required: true,
		},
		"address_prefix": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			ForceNew: true,
			Required: true,
		},
		"subnet_cidr": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: false,
			ForceNew: true,
			Required: true,
		},
		"status": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
	}
}

////// convert from cloud to state and vice versa :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) InfrastructureToState(duploObject *DuploInfrastructure, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-InfrastructureToState ******** 1: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["infra_name"] = duploObject.Name
		cObj["account_id"] = duploObject.AccountId
		cObj["cloud"] = duploObject.Cloud
		cObj["region"] = duploObject.Region
		cObj["azcount"] = duploObject.AzCount
		cObj["enable_k8_cluster"] = duploObject.EnableK8Cluster
		cObj["address_prefix"] = duploObject.AddressPrefix
		cObj["subnet_cidr"] = duploObject.SubnetCidr
		cObj["status"] = duploObject.ProvisioningStatus
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-InfrastructureToState ******** 2: to-DICT %s ", jsonData2)

		return cObj
	}
	return nil
}

func (c *Client) DuploInfrastructureFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploInfrastructure, error) {
	url := c.InfrastructureByNameUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploInfrastructureFromState-Create %s ", url)
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploInfrastructureFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)
	//
	duploObject := new(DuploInfrastructure)
	///--- set
	//SKIP InstanceId? TenantId? Status? IdentityRole PrivateIpAddress
	duploObject.Name = d.Get("infra_name").(string)
	duploObject.AccountId = d.Get("account_id").(string)
	duploObject.Cloud = d.Get("cloud").(int)
	duploObject.Region = d.Get("region").(string)
	duploObject.AzCount = d.Get("azcount").(int)
	duploObject.EnableK8Cluster = d.Get("enable_k8_cluster").(bool)
	duploObject.AddressPrefix = d.Get("address_prefix").(string)
	duploObject.SubnetCidr = d.Get("subnet_cidr").(int)
	duploObject.ProvisioningStatus = d.Get("status").(string)

	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: to-CLOUD %s", api_str, jsonData)
	return duploObject, nil
}

//this is the import-id for terraform inspired from azure imports
func (c *Client) DuploInfrastructureSetIdFromCloud(duploObject *DuploInfrastructure, d *schema.ResourceData) string {
	d.Set("infra_name", duploObject.Name)
	//	d.Set("account_id", duploObject.AccountId )
	c.InfrastructureSetId(d)
	log.Printf("[TRACE] DuploInfrastructureSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) InfrastructureSetId(d *schema.ResourceData) string {
	infra_name := d.Get("infra_name").(string)
	///--- set
	id := fmt.Sprintf("v2/admin/InfrastructureV2/%s", infra_name)
	d.SetId(id)
	return id
}

//api for crud -- get + delete
func (c *Client) InfrastructureByNameUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-InfrastructureUrl %s 1 ********: %s", api, host)
	return host
}

// app for -- get list + create
func (c *Client) InfrastructureCreateOrListUrl(d *schema.ResourceData) string {
	api := "v2/admin/InfrastructureV2"
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-InfrastructureCreateOrListUrl %s 1 ********: %s", api, host)
	return host
}

//Utils convert for get-list   -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) InfrastructuresFlatten(duploObjects *[]DuploInfrastructure, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.InfrastructureToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-InfrastructureToState ******** 1 to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-InfrastructureTagsToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}

//convert  get-list item into state  -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) InfrastructureFillGet(duploObject *DuploInfrastructure, d *schema.ResourceData) error {
	if duploObject != nil {
		//create map
		ois := c.InfrastructureToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-InfrastructureFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-InfrastructureFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	err_msg := fmt.Errorf("Infrastructure not found")
	return err_msg
}

/////////  API list //////////
func (c *Client) InfrastructureGetList(d *schema.ResourceData, m interface{}) (*[]DuploInfrastructure, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] InfrastructureGetList filters 1 ********* : %s  %s", filters, filtersOk)
	//
	api := c.InfrastructureCreateOrListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-InfrastructureGetList 2 %s  ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-InfrastructureGetList 3 %s   ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-InfrastructureGetList 4 %s   ********: %s", api, bodyString)

	duploObjects := []DuploInfrastructure{}
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-InfrastructureGetList 5 %s  ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}

/////////   list DONE //////////

/////////  API Item //////////
func (c *Client) InfrastructureGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.InfrastructureByNameUrl(d)
	log.Printf("[TRACE] duplo-InfrastructureGet 1  %s ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-InfrastructureGet 2 %s ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-InfrastructureGet 3 %s ********: bodyString %s", api, bodyString)

	duploObject := DuploInfrastructure{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-InfrastructureGet 4 %s ********:  error:%s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-InfrastructureGet 5 %s ******** ", api)
	if duploObject.Name != "" {
		c.InfrastructureFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-InfrastructureGet 6 FOUND *****", api)
		return nil
	}
	err_msg := fmt.Errorf("Infrastructure not found  : %s body:%s", api, bodyString)
	return err_msg
}

/////////  API Item //////////

/////////  API  Create/update //////////

func (c *Client) InfrastructureCreate(d *schema.ResourceData, m interface{}) (*DuploInfrastructure, error) {
	return c.InfrastructureCreateOrUpdate(d, m, false)
}

func (c *Client) InfrastructureUpdate(d *schema.ResourceData, m interface{}) (*DuploInfrastructure, error) {
	return c.InfrastructureCreateOrUpdate(d, m, true)
}

func (c *Client) InfrastructureCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploInfrastructure, error) {
	duploObject, _ := c.DuploInfrastructureFromState(d, m, isUpdate)
	//log.Printf("[TRACE] duplo-InfrastructureCreate duploObject.InfraName %s AccountID %s ********:", duploObject. ,duploObject )

	url := c.InfrastructureCreateOrListUrl(d)
	log.Printf("[TRACE] duplo-InfrastructureCreate %s 1 ********:", url)

	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-InfrastructureCreate %s ", api)
	if isUpdate {

		action = "PUT"
		api_str = fmt.Sprintf("duplo-InfrastructureUpdate %s ", api)
	}

	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	//
	//duploObject , _ := c.DuploInfrastructureFromState(d,m,isUpdate)
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s", api_str, err.Error())
		return nil, err
	}

	json_str := string(rb)
	log.Printf("[TRACE] %s 4 ********: %s", api_str, json_str)

	req, err := http.NewRequest(action, url, strings.NewReader(string(rb)))
	if err != nil {
		log.Printf("[TRACE] %s 5 ********: %s", api_str, err.Error())
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] %s 6 ********: %s", api_str, err.Error())
		return nil, err
	}
	if body != nil {
		bodyString := string(body)
		log.Printf("[TRACE] %s 7 ********: %s", api_str, bodyString)

		duploObject := DuploInfrastructure{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s", api_str, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ", api_str)
		c.DuploInfrastructureSetIdFromCloud(&duploObject, d)

		////////DuploInfrastructureWaitForCreation////////
		DuploInfrastructureWaitForCreation(c, c.InfrastructureByNameUrl(d))
		////////DuploInfrastructureWaitForCreation////////

		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %d,   body: %s", api, body)
	return nil, err_msg
}

/////////  API Create/update //////////

/////////  API Delete //////////
func (c *Client) InfrastructureDelete(d *schema.ResourceData, m interface{}) (*DuploInfrastructure, error) {
	var api = d.Id()
	url := c.InfrastructureByNameUrl(d)
	log.Printf("[TRACE] duplo-InfrastructureDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-InfrastructureDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-InfrastructureDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-InfrastructureDelete %s 4 ********: %s", api)
	return nil, nil
}

/////////  API Delete //////////

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////
func DuploInfrastructureRefreshFunc(c *Client, url string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		api := url
		req2, _ := http.NewRequest("GET", url, nil)
		body, err := c.doRequest(req2)
		if err != nil {
			log.Printf("[TRACE] duplo-DuploInfrastructureRefreshFunc 2 %s ********: %s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		bodyString := string(body)
		log.Printf("[TRACE] duplo-DuploInfrastructureRefreshFunc 3 %s ********: bodyString %s", api, bodyString)

		duploObject := DuploInfrastructure{}

		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] duplo-DuploInfrastructureRefreshFunc 4 %s ********:  error:%s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		log.Printf("[TRACE] duplo-DuploInfrastructureRefreshFunc 5 %s ******** ", api)
		var status string
		status = "pending"
		if duploObject.ProvisioningStatus == "Complete" {
			status = "ready"
		}
		return duploObject, status, nil
	}
}

func DuploInfrastructureWaitForCreation(c *Client, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: DuploInfrastructureRefreshFunc(c, url),
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
	}
	log.Printf("[DEBUG] InfrastructureRefreshFuncWaitForCreation (%s)", url)
	_, err := stateConf.WaitForState()
	return err
}

// GetEksCredentials retrieves just-in-time EKS credentials via the Duplo API.
func (c *Client) GetEksCredentials(planID string) (*DuploEksCredentials, error) {

	// Format the URL
	url := fmt.Sprintf("%s/adminproxy/%s/GetEksClusterByInfra", c.HostURL, planID)
	log.Printf("[TRACE] duplo-GetEksCredentials 1 ********: %s ", url)

	// Get the AWS region from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-GetEksCredentials 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-GetEksCredentials 3 ********: %s", bodyString)

	// Return it as an object.
	creds := DuploEksCredentials{}
	err = json.Unmarshal(body, &creds)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-GetEksCredentials 4 ********: %s", creds.Name)
	creds.PlanID = planID
	return &creds, nil
}
