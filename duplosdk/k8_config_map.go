package duplosdk

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
)

/////------ schema ------////
func  K8ConfigMapSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema {
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true,
		},
		"data": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"metadata": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
	}
}
////// convert from cloud to state and vice versa :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8ConfigMapToState(pduploObject *map[string]interface{}, d *schema.ResourceData) map[string]interface{} {
	duploObject :=*pduploObject
	if duploObject != nil {
		jsonData , _ := json.Marshal(duploObject )
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1: from-CLOUD %s ", jsonData)

		///--- set
		cObj := make(map[string]interface{})
		var metadata = make(map[string]interface{})
		var data = make(map[string]interface{})
		if duploObject["metadata"] != nil{
			metadata = duploObject["metadata"].(map[string]interface{})
		}
		if duploObject["data"] != nil{
			data = duploObject["data"].(map[string]interface{})
		}
		data_str , _ := json.Marshal(data)
		metadata_str , _ := json.Marshal(metadata)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1: from-CLOUD-data data_str %s metadata_str %s", data_str, metadata_str)
		///--- set
		cObj["tenant_id"] = c.K8ConfigMapGetTenantId(d)
		cObj["data"] = string(data_str)
		cObj["metadata"] = string(metadata_str)
		cObj["name"] = metadata["name"].(string)

		//log
		jsonData2 , _ := json.Marshal(cObj )
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 2: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

func (c *Client) DuploK8ConfigMapFromState(d *schema.ResourceData, m interface{}, isUpdate bool)( string, error) {
	url := c.K8ConfigMapListUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploK8ConfigMapFromState-Create %s ", url )
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploK8ConfigMapFromState-update %s ", url )
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str,  url)
	//
	duploObject := make(map[string]interface{})
	///--- set
	data := make(map[string]interface{})
	metadata := make(map[string]interface{})
	metadata["name"] = d.Get("name").(string)
	//data
	data_str := d.Get("data").(string)
	log.Printf("[TRACE] %s 2 ********: data_str %s", api_str,  data_str)
	err := json.Unmarshal([]byte(data_str), &data )
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: err %s", api_str,  err.Error())
	}
	//
	duploObject["data"] = &data
	duploObject["metadata"] =&metadata

	//log
	data_str2, _ := json.Marshal(&data)
	log.Printf("[TRACE] %s 4 ********: to-DICT-data %s", api_str,  data_str2)
	metadata_str2, _ := json.Marshal(&metadata)
	log.Printf("[TRACE] %s 4 ********: to-DICT-metadata %s", api_str,  metadata_str2)
	//log
	jsonData, _ := json.Marshal( &duploObject)
	log.Printf("[TRACE] %s 5 ********: to-CLOUD-all %s", api_str,  jsonData)
	return string(jsonData), nil
}

//this is the import-id for terraform inspired from azure imports
func (c * Client) DuploK8ConfigMapSetIdFromCloud(duploObject *map[string]interface{}, d *schema.ResourceData) string{
	d.Set("tenant_id", c.K8ConfigMapGetTenantId(d) )
	c.K8ConfigMapSetId(d)
	log.Printf("[TRACE] DuploK8ConfigMapSetIdFromCloud 2 ********: %s",  d.Id())
	return d.Id()
}
func (c *Client) K8ConfigMapSetId(d *schema.ResourceData) string{
	tenant_id := c.K8ConfigMapGetTenantId(d)
	name := d.Get("name").(string)
	///--- set
	id  := fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2/%s",tenant_id , name)
	d.SetId(id)
	return id
}
//api for crud -- get + delete
func (c *Client) K8ConfigMapUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api )
	log.Printf("[TRACE] duplo-K8ConfigMapUrl %s 1 ********: %s",api, host)
	return host
}
// app for -- get list + create + update
func (c *Client) K8ConfigMapListUrl(d *schema.ResourceData) string {
	tenant_id := c.K8ConfigMapGetTenantId(d)
	api:=fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2",tenant_id )
	host := fmt.Sprintf("%s/%s", c.HostURL,api )
	log.Printf("[TRACE] duplo-K8ConfigMapListUrl %s 1 ********: %s",api, host)
	return host
}
// tenant_id or any  parents in import url should be handled if not part of get json
func (c *Client) K8ConfigMapGetTenantId(d *schema.ResourceData) string{
	tenant_id := d.Get("tenant_id").(string)
	//tenant_id is local only field --- should be returned from server
	if tenant_id == ""{
		id := d.Id()
		id_array := strings.Split(id, "/")
		for i, s := range id_array {
			if s == "subscriptions"{
				next_i := i +1
				if id_array[next_i] != ""{
					d.Set("tenant_id", id_array[next_i] )
				}
				return id_array[next_i]
			}
			fmt.Println(i, s)
		}
	}
	return tenant_id
}

//Utils convert for get-list   -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8ConfigMapsFlatten(duploObjects *[]map[string]interface{}, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.K8ConfigMapToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal( ois)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1 to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-K8ConfigMapTagsToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}
//convert  get-list item into state  -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8ConfigMapFillGet(duploObject *map[string]interface{}, d *schema.ResourceData)   error  {
	if duploObject != nil {
		//create map
		ois := c.K8ConfigMapToState(duploObject, d)
		//log
		jsonData , _ := json.Marshal(ois )
		log.Printf("[TRACE] duplo-K8ConfigMapFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-K8ConfigMapFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key] )
		}
		return nil
	}
	err_msg:= fmt.Errorf("K8ConfigMap not found")
	return err_msg
}

/////////  API list //////////
func (c *Client) K8ConfigMapGetList(d *schema.ResourceData, m interface{}, )( *[]map[string]interface{}, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] K8ConfigMapGetList filters 1 ********* : %s  %s", filters , filtersOk)
	//
	api:= c.K8ConfigMapListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 2 %s  ********: %s",api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGetList 3 %s   ********: %s",api, err.Error())
		return nil , err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 4 %s   ********: %s", api, bodyString)

	duploObjects := make([]map[string]interface{}, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 5 %s  ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}
/////////   list DONE //////////

/////////  API Item //////////
func (c *Client) K8ConfigMapGet( d *schema.ResourceData, m interface{}, )  error  {
	var api = d.Id()
	url := c.K8ConfigMapUrl(d)
	log.Printf("[TRACE] duplo-K8ConfigMapGet 1  %s ********: %s",api,  url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGet 2 %s ********: %s",api, err.Error())
		return  err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8ConfigMapGet 3 %s ********: bodyString %s",api, bodyString)

	duploObject := make(map[string]interface{})
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGet 4 %s ********:  error:%s",api,  err.Error())
		return  err
	}
	log.Printf("[TRACE] duplo-K8ConfigMapGet 5 %s ******** ",api)
	if duploObject["metadata"] != nil {
		c.K8ConfigMapFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-K8ConfigMapGet 6 FOUND *****",api)
		return  nil
	}
	err_msg := fmt.Errorf("K8ConfigMap not found  : %s body:%s",api, bodyString)
	return  err_msg
}
/////////  API Item //////////

/////////  API  Create/update //////////

func (c *Client) K8ConfigMapCreate(d *schema.ResourceData, m interface{}, )( *map[string]interface{}, error) {
	return  c.K8ConfigMapCreateOrUpdate(d, m, false)
}

func (c *Client) K8ConfigMapUpdate(d *schema.ResourceData, m interface{}, )( *map[string]interface{}, error) {
	return  c.K8ConfigMapCreateOrUpdate(d, m, true)
}

func (c *Client) K8ConfigMapCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool)( *map[string]interface{}, error) {
	url := c.K8ConfigMapListUrl(d)
	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-K8ConfigMapCreate %s ",api )
	if isUpdate {
		action = "PUT"
		api_str = fmt.Sprintf("duplo-K8ConfigMapUpdate %s ",api )
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str,  url)
	//
	json_str , _ := c.DuploK8ConfigMapFromState(d,m,isUpdate)
	log.Printf("[TRACE] %s 4 ********: %s",api_str, json_str )
	req, err := http.NewRequest(action, url, strings.NewReader(json_str))
	if err != nil {
		log.Printf("[TRACE] %s 5 ********: %s",api_str, err.Error())
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] %s 6 ********: %s",api_str,  err.Error())
		return nil, err
	}
	if body != nil {
		bodyString := string(body)
		log.Printf("[TRACE] %s 7 ********: %s",api_str, bodyString)

		duploObject := map[string]interface{}{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s",api_str ,  err.Error())
			return  nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ",api_str )
		c.DuploK8ConfigMapSetIdFromCloud(&duploObject, d)
		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %d,   body: %s", api,  body)
	return nil,  err_msg
}
/////////  API Create/update //////////

/////////  API Delete //////////
func (c *Client) K8ConfigMapDelete(d *schema.ResourceData, m interface{}, )( *map[string]interface{}, error) {
	var api = d.Id()
	url := c.K8ConfigMapUrl(d)
	log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 1 ********: %s",api,  url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 2 ********: %s",api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 3 ********: %s",api,  err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-K8ConfigMapDelete %s 4 ********: %s",api  )
	return nil,  nil
}
/////////  API Delete //////////

