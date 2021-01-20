package duplosdk

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
)

//Tenant
type DuploTenant struct {
	TenantId    string `json:"TenantId",omitempty`
	AccountName string `json:"AccountName"`
	PlanID      string `json:"PlanID"`
}

/////------ schema ------////
func TenantSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"account_name": &schema.Schema{
			Type:     schema.TypeString,
			Computed: false,
			Required: true,
		},
		"plan_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: false,
			Required: true,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
	}
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) TenantFlatten(duploTenant *DuploTenant, d *schema.ResourceData) map[string]interface{} {
	if duploTenant != nil {
		c := make(map[string]interface{})
		///--- set
		c["account_name"] = duploTenant.AccountName
		c["tenant_id"] = duploTenant.TenantId
		c["plan_id"] = duploTenant.PlanID

		jsonData, _ := json.Marshal(duploTenant)
		log.Printf("[TRACE] duplo-TenantFlatten ********: jsonData %s ", jsonData)
		return c
	}
	return nil
}

////// convert from state to cloud :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) TenantFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploTenant, error) {
	url := c.TenantUrl(d)
	var api_str = fmt.Sprintf("duplo-TenantFromState-Create %s ", url)
	if isUpdate {
		api_str = fmt.Sprintf("duplo-TenantFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	duploObject := new(DuploTenant)
	///--- set
	duploObject.AccountName = d.Get("account_name").(string)
	duploObject.PlanID = d.Get("plan_id").(string)

	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] %s ********: jsonData %s ", api_str, jsonData)
	return duploObject, nil
}

/////////// common place to get url + Id : follow Azure  style Ids for import//////////
func (c *Client) TenantSetId(d *schema.ResourceData) string {
	tenant_id := d.Get("tenant_id").(string)
	///--- set
	id := fmt.Sprintf("v2/admin/TenantV2/%s", tenant_id)
	log.Printf("[TRACE] duplo-TenantSetId %s  ********: %s", id, tenant_id)
	d.SetId(id)
	return id
}

func (c *Client) TenantUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-TenantUrl %s 1 ********: %s", api, host)
	return host
}

func (c *Client) TenantUrlList(d *schema.ResourceData) string {
	api := "v2/admin/TenantV2"
	url := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-TenantUrlList %s 1 ********: %s", api, url)
	return url
}

/////////// common place to get url + Id : follow Azure  style Ids for import//////////

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
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

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
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
	err_msg := fmt.Errorf("Tenant not found 2")
	return err_msg
}

/////////  API list //////////
func (c *Client) TenantGetList(d *schema.ResourceData, m interface{}) (*[]DuploTenant, error) {

	url := c.TenantUrlList(d)

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
	log.Printf("[TRACE] duplo-GetListTenant %s 3 ********: %s", url, len(duploTenants))

	return &duploTenants, nil
}

/////////  API Item //////////
func (c *Client) TenantGet(d *schema.ResourceData, m interface{}) error {

	api := d.Id()
	url := c.TenantUrl(d)
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
	if duploTenant.TenantId != "" {
		c.TenantFillGet(&duploTenant, d)
		log.Printf("[TRACE] duplo-TenantGet 5 FOUND ***** : '%s' '%s'", duploTenant.AccountName, duploTenant.TenantId)
		return nil
	}

	account_name := d.Get("account_name").(string)
	err_msg := fmt.Errorf("Tenant not found %s : %s. Please ", account_name, duploTenant.TenantId)
	return err_msg
}

/////////  API Create / update //////////
func (c *Client) TenantCreate(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	return c.TenantCreateOrUpdate(d, m, false)
}

func (c *Client) TenantUpdate(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	//do nothing
	return nil, nil
}

func (c *Client) TenantCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploTenant, error) {
	duploObject, _ := c.TenantFromState(d, m, isUpdate)
	log.Printf("[TRACE] duplo-TenantCreate duploObject.AccountName %s PlanID %s ********:", duploObject.AccountName, duploObject.PlanID)

	url := c.TenantUrlList(d)
	log.Printf("[TRACE] duplo-TenantCreate %s 1 ********:", url)

	//
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostCreate %s 2 ********: %s", url, err.Error())
		return nil, err
	}

	json_str := string(rb) //strings.NewReader(string(rb))
	log.Printf("[TRACE] duplo-AwsHostCreate %s 3 ********: %s", url, json_str)

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
		log.Printf("[TRACE] duplo-AwsHostCreate %s 5 ********: %s %s %s", url, duploTenant.AccountName, duploTenant.TenantId, duploTenant.PlanID)
		//todo: move this part up
		d.Set("tenant_id", duploTenant.TenantId)
		c.TenantSetId(d) //??
		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %d, %s  body: %s", url, duploObject.AccountName, body)
	return nil, err_msg
}

func (c *Client) TenantDelete(d *schema.ResourceData, m interface{}) (*DuploTenant, error) {
	//do nothing
	return nil, nil
}
