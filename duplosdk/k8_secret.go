package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploK8sSecret represents a kubernetes secret in a Duplo tenant
type DuploK8sSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

	SecretName        string                 `json:"SecretName"`
	SecretType        string                 `json:"SecretType"`
	SecretData        map[string]interface{} `json:"SecretData,omitempty"`
	SecretAnnotations map[string]string      `json:"SecretAnnotations,omitempty"`
}

// K8SecretToState converts a Duplo SDK object respresenting a k8s secret to terraform resource data.
func (c *Client) K8SecretToState(pduploObject *map[string]interface{}, d *schema.ResourceData) map[string]interface{} {
	duploObject := *pduploObject
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-K8SecretToState ******** 1: from-CLOUD %s ", jsonData)
		///--- set
		cObj := make(map[string]interface{})
		dataNew := map[string]interface{}{}
		if duploObject["SecretData"] != nil {
			data := duploObject["SecretData"].(map[string]interface{})
			for key, value := range data {
				fmt.Println("Key:", key, "=>", "value:", value)
				dataNew[key] = ""
			}
		}
		dataNewStr, _ := json.Marshal(dataNew)
		///--- set
		cObj["tenant_id"] = c.K8SecretGetTenantID(d)
		cObj["secret_data"] = string(dataNewStr)
		cObj["secret_type"] = duploObject["SecretType"].(string)
		cObj["secret_name"] = duploObject["SecretName"].(string)
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-K8SecretToState ******** 2: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

// DuploK8SecretFromState converts resource data respresenting a k8s secret to a Duplo SDK object.
func (c *Client) DuploK8SecretFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (string, error) {
	url := c.K8SecretListURL(d)
	var apiStr = fmt.Sprintf("duplo-DuploK8SecretFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploK8SecretFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	duploObject := make(map[string]interface{})
	///--- set
	prev, newSecretData := d.GetChange("secret_data")
	log.Printf("[TRACE] %s 2 ********: api_str,  prev, new_secret_data %s %s ", apiStr, prev, newSecretData)
	duploObject["SecretData"] = make(map[string]interface{})
	data := make(map[string]interface{})
	dataStr := d.Get("secret_data").(string)
	log.Printf("[TRACE] %s 2 ********: data_str %s", apiStr, dataStr)
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: err %s", apiStr, err.Error())
	}
	//
	duploObject["SecretData"] = data
	duploObject["SecretName"] = d.Get("secret_name").(string)
	duploObject["SecretVersion"] = d.Get("secret_version").(string)
	duploObject["SecretType"] = d.Get("secret_type").(string)
	//keep hash
	d.Set("client_secret_version", hashForData(string(dataStr)))

	//log
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 5 ********: to-CLOUD-all %s", apiStr, jsonData)
	return string(jsonData), nil
}

// DuploK8SecretSetIDFromCloud populates the resource ID based on secret_name and tenant_id
func (c *Client) DuploK8SecretSetIDFromCloud(duploObject *map[string]interface{}, d *schema.ResourceData) string {
	d.Set("tenant_id", c.K8SecretGetTenantID(d))
	c.K8SecretSetID(d)
	log.Printf("[TRACE] DuploK8SecretSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}

// K8SecretSetID populates the resource ID based on secret_name and tenant_id
func (c *Client) K8SecretSetID(d *schema.ResourceData) string {
	tenantID := c.K8SecretGetTenantID(d)
	name := d.Get("secret_name").(string)
	///--- set
	id := fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantID, name)
	d.SetId(id)
	return id
}

// K8SecretURL returns the base API URL for crud -- get + delete
func (c *Client) K8SecretURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8SecretUrl %s 1 ********: %s", api, host)
	return host
}

// K8SecretListURL returns the base API URL for crud -- get list + create + update
func (c *Client) K8SecretListURL(d *schema.ResourceData) string {
	tenantID := c.K8SecretGetTenantID(d)
	api := fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-K8SecretListUrl %s 1 ********: %s", api, host)
	return host
}

// K8SecretGetTenantID tries to retrieve (or synthesize) a tenant_id based on resource data
// - tenant_id or any parents in import url should be handled if not part of get json
func (c *Client) K8SecretGetTenantID(d *schema.ResourceData) string {
	tenantID := d.Get("tenant_id").(string)
	//tenant_id is local only field --- should be returned from server
	if tenantID == "" {
		id := d.Id()
		idArray := strings.Split(id, "/")
		for i, s := range idArray {
			if s == "subscriptions" {
				j := i + 1
				if idArray[j] != "" {
					_ = d.Set("tenant_id", idArray[j])
				}
				return idArray[j]
			}
			fmt.Println(i, s)
		}
	}
	return tenantID
}

// K8SecretsFlatten converts a list of Duplo SDK objects into Terraform resource data
func (c *Client) K8SecretsFlatten(duploObjects *[]map[string]interface{}, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects))
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

// K8SecretFillGet converts a Duplo SDK object into Terraform resource data
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
	errMsg := fmt.Errorf("K8Secret not found")
	return errMsg
}

// K8SecretGetList retrieves a list of k8s secrets via the Duplo API.
func (c *Client) K8SecretGetList(tenantID string) (*[]DuploK8sSecret, ClientError) {
	rp := []DuploK8sSecret{}
	err := c.getAPI(
		fmt.Sprintf("K8SecretGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAllK8Secrets", tenantID),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
		}
	}

	return &rp, err
}

// K8SecretGet retrieves a k8s secret via the Duplo API.
func (c *Client) K8SecretGet(tenantID, secretName string) (*DuploK8sSecret, ClientError) {

	// Retrieve the list of secrets
	list, err := c.K8SecretGetList(tenantID)
	if err != nil || list == nil {
		return nil, err
	}

	// Return the secret, if it exists.
	for i := range *list {
		if (*list)[i].SecretName == secretName {
			return &(*list)[i], nil
		}
	}

	return nil, nil
}

/////////  API Item //////////

/////////  API  Create/update //////////

// K8SecretCreate creates a k8s secret via the Duplo API.
func (c *Client) K8SecretCreate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8SecretCreateOrUpdate(d, m, false)
}

// K8SecretUpdate updates a k8s secret via the Duplo API.
func (c *Client) K8SecretUpdate(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	return c.K8SecretCreateOrUpdate(d, m, true)
}

// K8SecretCreateOrUpdate creates or updates a k8s secret via the Duplo API.
func (c *Client) K8SecretCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*map[string]interface{}, error) {
	url := c.K8SecretListURL(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-K8SecretCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-K8SecretUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	jsonStr, _ := c.DuploK8SecretFromState(d, m, isUpdate)
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
		c.DuploK8SecretSetIDFromCloud(&duploObject, d)
		return nil, nil
	}
	errMsg := fmt.Errorf("ERROR: in create %s,   body: %s", api, body)
	return nil, errMsg
}

/////////  API Create/update //////////

/////////  API Delete //////////

// K8SecretDelete deletes a k8s secret via the Duplo API.
func (c *Client) K8SecretDelete(d *schema.ResourceData, m interface{}) (*map[string]interface{}, error) {
	var api = d.Id()
	url := c.K8SecretURL(d)
	log.Printf("[TRACE] duplo-K8SecretDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	_, err = c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-K8SecretDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	log.Printf("[TRACE] DONE duplo-K8SecretDelete %s 4 ********", api)
	return nil, nil
}

/////////  API Delete //////////
