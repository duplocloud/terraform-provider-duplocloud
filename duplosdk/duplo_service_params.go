package duplosdk

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
)

type DuploServiceParams struct {
	ReplicationControllerName string `json:"ReplicationControllerName"`
	TenantId                  string `json:"TenantId,omitempty"`
	WebACLId                  string `json:"WebACLId,omitempty"`
	DnsPrfx                   string `json:"DnsPrfx,omitempty"`
}

/////------ schema ------////
func DuploServiceParamsSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"webaclid": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"replication_controller_name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch service
		},
		"dns_prfx": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
	}
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceParamsToState(duploObject *DuploServiceParams, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		tenant_id := c.DuploServiceParamsGetTenantId(d)

		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-DuploServiceParamsToState ********: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["replication_controller_name"] = duploObject.ReplicationControllerName
		cObj["webaclid"] = duploObject.WebACLId
		cObj["tenant_id"] = tenant_id
		cObj["dns_prfx"] = duploObject.DnsPrfx

		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-DuploServiceParamsToState ********: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

////// convert from state to cloud :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceParamsFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceParams, error) {
	url := c.AwsHostListUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploServiceParamsFromState-Create %s ", c.AwsHostListUrl(d))
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploServiceParamsFromState-Create %s ", c.AwsHostListUrl(d))
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	//
	duploObject := new(DuploServiceParams)
	///--- set
	duploObject.ReplicationControllerName = d.Get("replication_controller_name").(string)
	duploObject.WebACLId = d.Get("webaclid").(string)
	duploObject.DnsPrfx = d.Get("dns_prfx").(string)

	return duploObject, nil
}

///////// ///////// ///////// /////////  Utils convert ////////////////////

/////////// common place to get url + Id : follow Azure  style Ids for import//////////
func (c *Client) DuploServiceParamsSetIdFromCloud(duploObject *DuploServiceParams, d *schema.ResourceData) string {
	d.Set("replication_controller_name", duploObject.ReplicationControllerName)
	d.Set("tenant_id", duploObject.TenantId)
	c.DuploServiceParamsSetId(d)
	log.Printf("[TRACE] DuploServiceParamsSetIdFromCloud 1 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) DuploServiceParamsSetId(d *schema.ResourceData) string {
	tenant_id := c.DuploServiceParamsGetTenantId(d)
	name := d.Get("replication_controller_name").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenant_id, name)
	d.SetId(id)
	return id
}
func (c *Client) DuploServiceParamsUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceParamsUrl %s 1 ********: %s", api, host)
	return host
}
func (c *Client) DuploServiceParamsListUrl(d *schema.ResourceData) string {
	tenant_id := c.DuploServiceParamsGetTenantId(d)
	api := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2", tenant_id)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceParamsListUrl %s 1 ********: %s", api, host)
	return host
}

func (c *Client) DuploServiceParamsGetTenantId(d *schema.ResourceData) string {
	tenant_id := d.Get("tenant_id").(string)
	//tenant_id is local only field --- should be returned from server
	if tenant_id == "" {
		id := d.Id()
		id_array := strings.Split(id, "/")
		for i, s := range id_array {
			if s == "subscriptions" {
				next_i := i + 1
				if id_array[next_i] != "" {
					d.Set("tenant_id", id_array[next_i])
				}
				return id_array[next_i]
			}
			fmt.Println(i, s)
		}
	}
	return tenant_id
}

/////////// common place to get url + Id : follow Azure  style Ids for import//////////

/////////  Utils convert //////////
func (c *Client) DuploServiceParamsFlatten(duploObject *[]DuploServiceParams, d *schema.ResourceData) []interface{} {
	if duploObject != nil {
		ois := make([]interface{}, len(*duploObject), len(*duploObject))
		for i, duploObject := range *duploObject {
			ois[i] = c.DuploServiceParamsToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServiceParamsToState ******** jsonData: \n%s", jsonData)
		return ois
	}
	return make([]interface{}, 0)
}

func (c *Client) DuploServiceParamsFillGet(duploObject *DuploServiceParams, d *schema.ResourceData) error {
	if duploObject != nil {

		//create map
		ois := c.DuploServiceParamsToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServiceParamsFillGet ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-DuploServiceParamsFillGet ******** Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	err_msg := fmt.Errorf("DuploServiceParams not found 2")
	return err_msg
}

/////////  API list //////////
func (c *Client) DuploServiceParamsGetList(d *schema.ResourceData, m interface{}) (*[]DuploServiceParams, error) {
	//
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] DuploServiceParamsGetList filters ********* : %s  %s", filters, filtersOk)
	//
	api := c.DuploServiceParamsListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-DuploServiceParamsGetList %s 1 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceParamsGetList %s 2 ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceParamsGetList %s 3 ********: %s", api, bodyString)

	duploObjects := []DuploServiceParams{}
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		log.Printf("[TRACE] ERROR duplo-DuploServiceParamsGetList %s 4 ********: %s", api, err.Error())
		return nil, err
	}
	log.Printf("[TRACE] duplo-DuploServiceParamsGetList %s 5 ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}

func (c *Client) DuploServiceParamsGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.DuploServiceParamsUrl(d)
	log.Printf("[TRACE] duplo-DuploServiceParamsUpdate %s 1 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceParamsGet %s 2 ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceParamsGet %s 3 ********: bodyString %s", api, bodyString)

	duploObject := DuploServiceParams{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] ERROR: duplo-DuploServiceParamsGet %s 4 ********: error: %s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-DuploServiceParamsGet %s 5 ******** ", api)
	if duploObject.TenantId != "" {
		c.DuploServiceParamsFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-DuploServiceParamsGet 6 FOUND ***** : %s", api)
		return nil
	}
	err_msg := fmt.Errorf("DuploServiceParams not found 7 : %s  bodyString:%s", api, bodyString)
	return err_msg
}

/////////  API Create //////////
func (c *Client) DuploServiceParamsCreate(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {
	return c.DuploServiceParamsCreateOrUpdate(d, m, false)
}
func (c *Client) DuploServiceParamsUpdate(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {
	return c.DuploServiceParamsCreateOrUpdate(d, m, true)
}
func (c *Client) DuploServiceParamsCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceParams, error) {

	url := c.DuploServiceParamsListUrl(d)
	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-DuploServiceParamsCreate %s ", api)
	if isUpdate {
		action = "PUT"
		api_str = fmt.Sprintf("duplo-DuploServiceParamsUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	//
	duploObject, _ := c.DuploServiceParamsFromState(d, m, isUpdate)
	//
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: %s", api_str, jsonData)

	//
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s", api_str, api, err.Error())
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

		duploObject := DuploServiceParams{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] ERROR: %s 8 ********:  %s  ", api_str, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ********: ", api_str)
		c.DuploServiceParamsSetIdFromCloud(&duploObject, d)
		log.Printf("[TRACE] DONE:  %s 10 ********:", api_str)
		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create/update %s, body: %s", api_str, body)
	return nil, err_msg
}

func (c *Client) DuploServiceParamsDelete(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {

	var api = d.Id()
	url := c.DuploServiceParamsUrl(d)
	log.Printf("[TRACE] duplo-DuploServiceParamsDelete %s 1 ********: %s", api, url)

	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceParamsDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] ERROR: duplo-DuploServiceParamsDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE: duplo-DuploServiceParamsDelete %s 4 ********: ", api)
	return nil, nil
}
