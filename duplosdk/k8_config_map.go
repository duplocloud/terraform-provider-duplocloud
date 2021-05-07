package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// K8ConfigMapToState converts a Duplo SDK object respresenting a k8s configmap to terraform resource data.
func (c *Client) K8ConfigMapToState(pduploObject *map[string]interface{}, d *schema.ResourceData) map[string]interface{} {
	duploObject := *pduploObject
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1: from-CLOUD %s ", jsonData)

		///--- set
		cObj := make(map[string]interface{})
		var metadata = make(map[string]interface{})
		var data = make(map[string]interface{})
		if duploObject["metadata"] != nil {
			metadata = duploObject["metadata"].(map[string]interface{})
		}
		if duploObject["data"] != nil {
			data = duploObject["data"].(map[string]interface{})
		}
		dataStr, _ := json.Marshal(data)
		metadataStr, _ := json.Marshal(metadata)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1: from-CLOUD-data dataStr %s metadata_str %s", dataStr, metadataStr)
		///--- set
		cObj["tenant_id"] = c.K8ConfigMapGetTenantID(d)
		cObj["data"] = string(dataStr)
		cObj["metadata"] = string(metadataStr)
		cObj["name"] = metadata["name"].(string)

		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 2: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

// DuploK8ConfigMapFromState converts resource data respresenting a k8s configmap to a Duplo SDK object.
func (c *Client) DuploK8ConfigMapFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (string, error) {
	url := c.K8ConfigMapListURL(d)
	var apiStr = fmt.Sprintf("duplo-DuploK8ConfigMapFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploK8ConfigMapFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	duploObject := make(map[string]interface{})
	///--- set
	data := make(map[string]interface{})
	metadata := make(map[string]interface{})
	metadata["name"] = d.Get("name").(string)
	//data
	dataStr := d.Get("data").(string)
	log.Printf("[TRACE] %s 2 ********: dataStr %s", apiStr, dataStr)
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: err %s", apiStr, err.Error())
	}
	//
	duploObject["data"] = &data
	duploObject["metadata"] = &metadata

	//log
	dataStr2, _ := json.Marshal(&data)
	log.Printf("[TRACE] %s 4 ********: to-DICT-data %s", apiStr, dataStr2)
	metadataStr2, _ := json.Marshal(&metadata)
	log.Printf("[TRACE] %s 4 ********: to-DICT-metadata %s", apiStr, metadataStr2)
	//log
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 5 ********: to-CLOUD-all %s", apiStr, jsonData)
	return string(jsonData), nil
}

// DuploK8ConfigMapSetIDFromCloud populates the resource ID based on name and tenant_id
func (c *Client) DuploK8ConfigMapSetIDFromCloud(duploObject *map[string]interface{}, d *schema.ResourceData) string {
	d.Set("tenant_id", c.K8ConfigMapGetTenantID(d))
	c.K8ConfigMapSetID(d)
	log.Printf("[TRACE] DuploK8ConfigMapSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}

// K8ConfigMapSetID populates the resource ID based on name and tenant_id
func (c *Client) K8ConfigMapSetID(d *schema.ResourceData) string {
	tenantID := c.K8ConfigMapGetTenantID(d)
	name := d.Get("name").(string)
	///--- set
	id := fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2/%s", tenantID, name)
	d.SetId(id)
	return id
}

// K8ConfigMapURL returns the base API URL for crud -- get + delete
func (c *Client) K8ConfigMapURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8ConfigMapUrl %s 1 ********: %s", api, host)
	return host
}

// K8ConfigMapListURL returns the base API URL for crud -- get list + create + update
func (c *Client) K8ConfigMapListURL(d *schema.ResourceData) string {
	tenantID := c.K8ConfigMapGetTenantID(d)
	api := fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8ConfigMapListUrl %s 1 ********: %s", api, host)
	return host
}

// K8ConfigMapGetTenantID tries to retrieve (or synthesize) a tenant_id based on resource data
// - tenant_id or any parents in import url should be handled if not part of get json
func (c *Client) K8ConfigMapGetTenantID(d *schema.ResourceData) string {
	tenantID := d.Get("tenant_id").(string)
	//tenant_id is local only field --- should be returned from server
	if tenantID == "" {
		id := d.Id()
		idArray := strings.Split(id, "/")
		for i, s := range idArray {
			if s == "subscriptions" {
				j := i + 1
				if idArray[j] != "" {
					d.Set("tenant_id", idArray[j])
				}
				return idArray[j]
			}
			fmt.Println(i, s)
		}
	}
	return tenantID
}

// K8ConfigMapsFlatten converts a list of Duplo SDK objects into Terraform resource data
func (c *Client) K8ConfigMapsFlatten(duploObjects *[]map[string]interface{}, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.K8ConfigMapToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-K8ConfigMapToState ******** 1 to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-K8ConfigMapTagsToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}

// K8ConfigMapFillGet converts a Duplo SDK object into Terraform resource data
func (c *Client) K8ConfigMapFillGet(duploObject *map[string]interface{}, d *schema.ResourceData) error {
	if duploObject != nil {
		//create map
		ois := c.K8ConfigMapToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-K8ConfigMapFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-K8ConfigMapFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	errMsg := fmt.Errorf("K8ConfigMap not found")
	return errMsg
}

/////////  API list //////////

// K8ConfigMapGetList retrieves a list of AWS hosts via the Duplo API.
func (c *Client) K8ConfigMapGetList(d *schema.ResourceData, m interface{}) (*[]map[string]interface{}, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] K8ConfigMapGetList filters 1 ********* : %s  %v", filters, filtersOk)
	//
	api := c.K8ConfigMapListURL(d)
	url := api
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 2 %s  ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGetList 3 %s   ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 4 %s   ********: %s", api, bodyString)

	duploObjects := make([]map[string]interface{}, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-K8ConfigMapGetList 5 %s  ********: %d", api, len(duploObjects))

	return &duploObjects, nil
}

/////////   list DONE //////////

/////////  API Item //////////

// K8ConfigMapGet retrieves a service's load balancer via the Duplo API.
func (c *Client) K8ConfigMapGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.K8ConfigMapURL(d)
	log.Printf("[TRACE] duplo-K8ConfigMapGet 1  %s ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGet 2 %s ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-K8ConfigMapGet 3 %s ********: bodyString %s", api, bodyString)

	duploObject := make(map[string]interface{})
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapGet 4 %s ********:  error:%s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-K8ConfigMapGet 5 %s ******** ", api)
	if duploObject["metadata"] != nil {
		c.K8ConfigMapFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-K8ConfigMapGet 6 %s FOUND *****", api)
		return nil
	}
	return fmt.Errorf("K8ConfigMap not found %s body: %s", api, bodyString)
}

/////////  API Item //////////

/////////  API  Create/update //////////

// K8ConfigMapCreate creates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapCreate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8ConfigMapCreateOrUpdate(d, m, false)
}

// K8ConfigMapUpdate updates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapUpdate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8ConfigMapCreateOrUpdate(d, m, true)
}

// K8ConfigMapCreateOrUpdate creates or updates a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*map[string]interface{}, error) {
	url := c.K8ConfigMapListURL(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-K8ConfigMapCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-K8ConfigMapUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	jsonStr, _ := c.DuploK8ConfigMapFromState(d, m, isUpdate)
	log.Printf("[TRACE] %s 4 ********: %s", apiStr, jsonStr)
	req, err := http.NewRequest(action, url, strings.NewReader(jsonStr))
	if err != nil {
		log.Printf("[TRACE] %s 5 ********: %s", apiStr, err.Error())
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] %s 6 ********: %s", apiStr, err.Error())
		return nil, err
	}
	if body != nil {
		bodyString := string(body)
		log.Printf("[TRACE] %s 7 ********: %s", apiStr, bodyString)

		duploObject := map[string]interface{}{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s", apiStr, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ", apiStr)
		c.DuploK8ConfigMapSetIDFromCloud(&duploObject, d)
		return nil, nil
	}
	return nil, fmt.Errorf("ERROR: in create %s, body: %s", api, body)
}

/////////  API Create/update //////////

/////////  API Delete //////////

// K8ConfigMapDelete deletes a k8s configmap via the Duplo API.
func (c *Client) K8ConfigMapDelete(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	var api = d.Id()
	url := c.K8ConfigMapURL(d)
	log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-K8ConfigMapDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-K8ConfigMapDelete %s 4 ********", api)
	return nil, nil
}

/////////  API Delete //////////
