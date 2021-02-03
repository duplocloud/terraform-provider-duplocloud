package duplosdk

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
)

/////------ schema ------////
func K8SecretSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"secret_name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"secret_type": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"client_secret_version": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_version": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_data": &schema.Schema{
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
			//DiffSuppressFunc: diffIgnoreIfSameHash,
			DiffSuppressFunc: diffIgnoreForSecretMap,
		},
	}
}

////// convert from cloud to state and vice versa :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8SecretToState(pduploObject *map[string]interface{}, d *schema.ResourceData) map[string]interface{} {
	duploObject := *pduploObject
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-K8SecretToState ******** 1: from-CLOUD %s ", jsonData)
		///--- set
		cObj := make(map[string]interface{})
		var data = make(map[string]interface{})
		var data_new = make(map[string]interface{})
		if duploObject["SecretData"] != nil {
			data = duploObject["SecretData"].(map[string]interface{})
			data_new = make(map[string]interface{})
			for key, value := range data {
				fmt.Println("Key:", key, "=>", "value:", value)
				data_new[key] = ""
			}
		}
		data_new_str, _ := json.Marshal(data_new)
		///--- set
		cObj["tenant_id"] = c.K8SecretGetTenantId(d)
		cObj["secret_data"] = string(data_new_str)
		cObj["secret_type"] = duploObject["SecretType"].(string)
		cObj["secret_name"] = duploObject["SecretName"].(string)
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-K8SecretToState ******** 2: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

func diffIgnoreForSecretMap(k, old, new string, d *schema.ResourceData) bool {
	mapFieldName := "client_secret_version"
	hashFieldName := "secret_data"
	_, data_new := d.GetChange(hashFieldName)
	hash_old := d.Get(mapFieldName).(string)
	hash_new := hashForData(data_new.(string))
	log.Printf("[TRACE] duplo-diffIgnoreForSecretMap ******** 1: hash_old %s hash_new   ", hash_new, hash_old)
	return hash_old == hash_new
}

func (c *Client) DuploK8SecretFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (string, error) {
	url := c.K8SecretListUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploK8SecretFromState-Create %s ", url)
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploK8SecretFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)
	//
	duploObject := make(map[string]interface{})
	///--- set
	prev, new_secret_data := d.GetChange("secret_data")
	log.Printf("[TRACE] %s 2 ********: api_str,  prev, new_secret_data %s %s ", api_str, prev, new_secret_data)
	duploObject["SecretData"] = make(map[string]interface{})
	data := make(map[string]interface{})
	data_str := d.Get("secret_data").(string)
	log.Printf("[TRACE] %s 2 ********: data_str %s", api_str, data_str)
	err := json.Unmarshal([]byte(data_str), &data)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: err %s", api_str, err.Error())
	}
	//
	duploObject["SecretData"] = data
	duploObject["SecretName"] = d.Get("secret_name").(string)
	duploObject["SecretVersion"] = d.Get("secret_version").(string)
	duploObject["SecretType"] = d.Get("secret_type").(string)
	//keep hash
	d.Set("client_secret_version", hashForData(string(data_str)))

	//log
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 5 ********: to-CLOUD-all %s", api_str, jsonData)
	return string(jsonData), nil
}

//this is the import-id for terraform inspired from azure imports
func (c *Client) DuploK8SecretSetIdFromCloud(duploObject *map[string]interface{}, d *schema.ResourceData) string {
	d.Set("tenant_id", c.K8SecretGetTenantId(d))
	c.K8SecretSetId(d)
	log.Printf("[TRACE] DuploK8SecretSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) K8SecretSetId(d *schema.ResourceData) string {
	tenant_id := c.K8SecretGetTenantId(d)
	name := d.Get("secret_name").(string)
	///--- set
	id := fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenant_id, name)
	d.SetId(id)
	return id
}

//api for crud -- get + delete
func (c *Client) K8SecretUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8SecretUrl %s 1 ********: %s", api, host)
	return host
}

// app for -- get list + create + update
func (c *Client) K8SecretListUrl(d *schema.ResourceData) string {
	tenant_id := c.K8SecretGetTenantId(d)
	api := fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2", tenant_id)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8SecretListUrl %s 1 ********: %s", api, host)
	return host
}

// tenant_id or any  parents in import url should be handled if not part of get json
func (c *Client) K8SecretGetTenantId(d *schema.ResourceData) string {
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

//Utils convert for get-list   -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8SecretsFlatten(duploObjects *[]map[string]interface{}, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.K8SecretToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-K8SecretToState ******** 1 to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-K8SecretTagsToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}

//convert  get-list item into state  -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) K8SecretFillGet(duploObject *map[string]interface{}, d *schema.ResourceData) error {
	if duploObject != nil {
		//create map
		ois := c.K8SecretToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-K8SecretFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-K8SecretFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	err_msg := fmt.Errorf("K8Secret not found")
	return err_msg
}

/////////  API list //////////
func (c *Client) K8SecretGetList(d *schema.ResourceData, m interface{}) (*[]map[string]interface{}, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] K8SecretGetList filters 1 ********* : %s  %s", filters, filtersOk)
	//
	api := c.K8SecretListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-K8SecretGetList 2 %s  ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretGetList 3 %s   ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8SecretGetList 4 %s   ********: %s", api, bodyString)

	duploObjects := make([]map[string]interface{}, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-K8SecretGetList 5 %s  ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}

/////////   list DONE //////////

/////////  API Item //////////
func (c *Client) K8SecretGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.K8SecretUrl(d)
	log.Printf("[TRACE] duplo-K8SecretGet 1  %s ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretGet 2 %s ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8SecretGet 3 %s ********: bodyString %s", api, bodyString)

	duploObject := make(map[string]interface{})
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretGet 4 %s ********:  error:%s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-K8SecretGet 5 %s ******** ", api)
	if duploObject["SecretData"] != nil {
		c.K8SecretFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-K8SecretGet 6 FOUND *****", api)
		return nil
	}
	err_msg := fmt.Errorf("K8Secret not found  : %s body:%s", api, bodyString)
	return err_msg
}

/////////  API Item //////////

/////////  API  Create/update //////////

func (c *Client) K8SecretCreate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8SecretCreateOrUpdate(d, m, false)
}

func (c *Client) K8SecretUpdate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8SecretCreateOrUpdate(d, m, true)
}

func (c *Client) K8SecretCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*map[string]interface{}, error) {
	url := c.K8SecretListUrl(d)
	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-K8SecretCreate %s ", api)
	if isUpdate {
		action = "PUT"
		api_str = fmt.Sprintf("duplo-K8SecretUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)
	//
	json_str, _ := c.DuploK8SecretFromState(d, m, isUpdate)
	log.Printf("[TRACE] %s 4 ********: %s", api_str, json_str)
	req, err := http.NewRequest(action, url, strings.NewReader(json_str))
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

		duploObject := map[string]interface{}{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s", api_str, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ", api_str)
		c.DuploK8SecretSetIdFromCloud(&duploObject, d)
		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %d,   body: %s", api, body)
	return nil, err_msg
}

/////////  API Create/update //////////

/////////  API Delete //////////
func (c *Client) K8SecretDelete(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	var api = d.Id()
	url := c.K8SecretUrl(d)
	log.Printf("[TRACE] duplo-K8SecretDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-K8SecretDelete %s 4 ********: %s", api)
	return nil, nil
}

/////////  API Delete //////////
