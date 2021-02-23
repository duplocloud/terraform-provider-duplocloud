package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploServiceParams represents an service's parameters in the Duplo SDK
type DuploServiceParams struct {
	ReplicationControllerName string `json:"ReplicationControllerName"`
	TenantID                  string `json:"TenantId,omitempty"`
	WebACLId                  string `json:"WebACLId,omitempty"`
	DNSPrfx                   string `json:"DnsPrfx,omitempty"`
}

// DuploServiceParamsSchema returns a Terraform resource schema for a service's parameters
func DuploServiceParamsSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"webaclid": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"replication_controller_name": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch service
		},
		"dns_prfx": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
	}
}

// DuploServiceParamsToState converts a Duplo SDK object respresenting a service's parameters to terraform resource data.
func (c *Client) DuploServiceParamsToState(duploObject *DuploServiceParams, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		tenantID := c.DuploServiceParamsGetTenantID(d)

		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-DuploServiceParamsToState ********: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["replication_controller_name"] = duploObject.ReplicationControllerName
		cObj["webaclid"] = duploObject.WebACLId
		cObj["tenant_id"] = tenantID
		cObj["dns_prfx"] = duploObject.DNSPrfx

		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-DuploServiceParamsToState ********: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

// DuploServiceParamsFromState converts resource data respresenting a service's parameters to a Duplo SDK object.
func (c *Client) DuploServiceParamsFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceParams, error) {
	url := c.AwsHostListURL(d)
	var apiStr = fmt.Sprintf("duplo-DuploServiceParamsFromState-Create %s ", c.AwsHostListURL(d))
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploServiceParamsFromState-Create %s ", c.AwsHostListURL(d))
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)

	//
	duploObject := new(DuploServiceParams)
	///--- set
	duploObject.ReplicationControllerName = d.Get("replication_controller_name").(string)
	duploObject.WebACLId = d.Get("webaclid").(string)
	duploObject.DNSPrfx = d.Get("dns_prfx").(string)

	return duploObject, nil
}

///////// ///////// ///////// /////////  Utils convert ////////////////////

// DuploServiceParamsSetIDFromCloud populates the resource ID based on name and tenant_id
func (c *Client) DuploServiceParamsSetIDFromCloud(duploObject *DuploServiceParams, d *schema.ResourceData) string {
	d.Set("replication_controller_name", duploObject.ReplicationControllerName)
	d.Set("tenant_id", duploObject.TenantID)
	c.DuploServiceParamsSetID(d)
	log.Printf("[TRACE] DuploServiceParamsSetIdFromCloud 1 ********: %s", d.Id())
	return d.Id()
}

// DuploServiceParamsSetID populates the resource ID based on name and tenant_id
func (c *Client) DuploServiceParamsSetID(d *schema.ResourceData) string {
	tenantID := c.DuploServiceParamsGetTenantID(d)
	name := d.Get("replication_controller_name").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2/%s", tenantID, name)
	d.SetId(id)
	return id
}

// DuploServiceParamsURL returns the base API URL for crud -- get + delete
func (c *Client) DuploServiceParamsURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceParamsUrl %s 1 ********: %s", api, host)
	return host
}

// DuploServiceParamsListURL returns the base API URL for crud -- get list + create + update
func (c *Client) DuploServiceParamsListURL(d *schema.ResourceData) string {
	tenantID := c.DuploServiceParamsGetTenantID(d)
	api := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerParamsV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceParamsListUrl %s 1 ********: %s", api, host)
	return host
}

// DuploServiceParamsGetTenantID tries to retrieve (or synthesize) a tenant_id based on resource data
// - tenant_id or any parents in import url should be handled if not part of get json
func (c *Client) DuploServiceParamsGetTenantID(d *schema.ResourceData) string {
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

/////////// common place to get url + Id : follow Azure  style Ids for import//////////

// DuploServiceParamsFlatten converts a list of Duplo SDK objects into Terraform resource data
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

// DuploServiceParamsFillGet converts a Duplo SDK object into Terraform resource data
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
	errMsg := fmt.Errorf("DuploServiceParams not found 2")
	return errMsg
}

// DuploServiceParamsGetList retrieves a list of AWS hosts via the Duplo API.
func (c *Client) DuploServiceParamsGetList(d *schema.ResourceData, m interface{}) (*[]DuploServiceParams, error) {
	//
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] DuploServiceParamsGetList filters ********* : %s  %v", filters, filtersOk)
	//
	api := c.DuploServiceParamsListURL(d)
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
	log.Printf("[TRACE] duplo-DuploServiceParamsGetList %s 5 ********: %d", api, len(duploObjects))

	return &duploObjects, nil
}

// DuploServiceParamsGet retrieves a service's load balancer via the Duplo API.
func (c *Client) DuploServiceParamsGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.DuploServiceParamsURL(d)
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
	if duploObject.TenantID != "" {
		c.DuploServiceParamsFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-DuploServiceParamsGet 6 FOUND ***** : %s", api)
		return nil
	}
	errMsg := fmt.Errorf("DuploServiceParams not found 7 : %s  bodyString:%s", api, bodyString)
	return errMsg
}

/////////  API Create //////////

// DuploServiceParamsCreate creates a service's load balancer via the Duplo API.
func (c *Client) DuploServiceParamsCreate(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {
	return c.DuploServiceParamsCreateOrUpdate(d, m, false)
}

// DuploServiceParamsUpdate updates an service's load balancer via the Duplo API.
func (c *Client) DuploServiceParamsUpdate(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {
	return c.DuploServiceParamsCreateOrUpdate(d, m, true)
}

// DuploServiceParamsCreateOrUpdate updates an service's load balancer  via the Duplo API.
func (c *Client) DuploServiceParamsCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceParams, error) {

	url := c.DuploServiceParamsListURL(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-DuploServiceParamsCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-DuploServiceParamsUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)

	//
	duploObject, _ := c.DuploServiceParamsFromState(d, m, isUpdate)
	//
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: %s", apiStr, jsonData)

	//
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s %s", apiStr, api, err.Error())
		return nil, err
	}

	jsonStr := string(rb)
	log.Printf("[TRACE] %s 4 ********: %s", apiStr, jsonStr)

	req, err := http.NewRequest(action, url, strings.NewReader(string(rb)))
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

		duploObject := DuploServiceParams{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] ERROR: %s 8 ********:  %s  ", apiStr, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ********: ", apiStr)
		c.DuploServiceParamsSetIDFromCloud(&duploObject, d)
		log.Printf("[TRACE] DONE:  %s 10 ********:", apiStr)
		return nil, nil
	}
	errMsg := fmt.Errorf("ERROR: in create/update %s, body: %s", apiStr, body)
	return nil, errMsg
}

// DuploServiceParamsDelete deletes a service's load balancer via the Duplo API.
func (c *Client) DuploServiceParamsDelete(d *schema.ResourceData, m interface{}) (*DuploServiceParams, error) {

	var api = d.Id()
	url := c.DuploServiceParamsURL(d)
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
