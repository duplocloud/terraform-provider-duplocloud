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

type DuploServiceLBConfigs struct {
	ReplicationControllerName string                  `json:"ReplicationControllerName"`
	TenantID                  string                  `json:"TenantId,omitempty"`
	LBConfigs                 *[]DuploLBConfiguration `json:"LBConfigs,omitempty"`
	Arn                       string                  `json:"Arn,omitempty"`
	Status                    string                  `json:"Status,omitempty"`
}

type DuploLBConfiguration struct {
	LBType                    int    `LBType:"LBType,omitempty"`
	Protocol                  string `json:"Protocol,omitempty"`
	Port                      string `Port:"Port,omitempty"`
	ExternalPort              int    `ExternalPort:"ExternalPort,omitempty"`
	HealthCheckURL            string `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string `json:"CertificateArn,omitempty"`
	ReplicationControllerName string `json:"ReplicationControllerName"`
	IsNative                  bool   `json:"IsNative"`
	IsInternal                bool   `json:"IsInternal"`
}

/////------ schema ------////
func DuploServiceLBConfigsSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"replication_controller_name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"lbconfigs": {
			Type:     schema.TypeList,
			Optional: false,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"lb_type": {
						Type:     schema.TypeInt,
						Required: true,
						ForceNew: true,
					},
					"protocol": {
						Type:     schema.TypeString,
						Required: true,
					},
					"port": {
						Type:     schema.TypeString,
						Required: true,
					},
					"external_port": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"health_check_url": {
						Type:     schema.TypeString,
						Required: false,
						Optional: true,
					},
					"certificate_arn": {
						Type:     schema.TypeString,
						Required: false,
						Optional: true,
					},
					"replication_controller_name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"is_native": {
						Type:     schema.TypeBool,
						Required: false,
						Optional: true,
					},
					"is_internal": {
						Type:     schema.TypeBool,
						Required: false,
						Optional: true,
					},
				},
			},
		},
	}
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceLBConfigsToState(duploObject *DuploServiceLBConfigs, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		//log
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsToState ******** 1 : from-CLOUD %s ", jsonData)
		//create map
		cObj := make(map[string]interface{})
		///--- set
		cObj["tenant_id"] = duploObject.TenantID
		cObj["replication_controller_name"] = duploObject.ReplicationControllerName
		cObj["arn"] = duploObject.Arn
		cObj["status"] = duploObject.Status

		cObj["lbconfigs"] = c.DuploLBConfigurationToState(duploObject.LBConfigs, d)
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsToState ******** 2 : to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploLBConfigurationToState(duploObjects *[]DuploLBConfiguration, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			cObj := make(map[string]interface{})
			cObj["lb_type"] = duploObject.LBType
			cObj["protocol"] = duploObject.Protocol
			cObj["port"] = duploObject.Port
			cObj["external_port"] = duploObject.ExternalPort
			cObj["health_check_url"] = duploObject.HealthCheckURL
			cObj["certificate_arn"] = duploObject.CertificateArn
			cObj["replication_controller_name"] = duploObject.ReplicationControllerName
			cObj["is_native"] = duploObject.IsNative
			cObj["is_internal"] = duploObject.IsInternal
			ois[i] = cObj
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploLBConfigurationToState ******** to-DICT 1:  %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-DuploLBConfigurationToState ??? empty ?? ******** from-CLOUD 2: %s", jsonData)
	return make([]interface{}, 0)
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceLBConfigsFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceLBConfigs, error) {
	url := c.DuploServiceLBConfigsListUrl(d)
	var apiStr = fmt.Sprintf("duplo-DuploServiceLBConfigsFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploServiceLBConfigsFromState-Update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	duploObject := new(DuploServiceLBConfigs)
	///--- set
	duploObject.ReplicationControllerName = d.Get("replication_controller_name").(string)
	duploObject.TenantID = d.Get("tenant_id").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Status = d.Get("status").(string)
	lbconfigs := d.Get("lbconfigs").([]interface{})
	if len(lbconfigs) > 0 {
		var lbc []DuploLBConfiguration
		for _, raw := range lbconfigs {
			p := raw.(map[string]interface{})
			lbc = append(lbc, DuploLBConfiguration{
				LBType:                    p["lb_type"].(int),
				Protocol:                  p["protocol"].(string),
				Port:                      p["port"].(string),
				ExternalPort:              p["external_port"].(int),
				HealthCheckURL:            p["health_check_url"].(string),
				CertificateArn:            p["certificate_arn"].(string),
				ReplicationControllerName: p["replication_controller_name"].(string),
				IsNative:                  p["is_native"].(bool),
				IsInternal:                p["is_internal"].(bool),
			})
		}
		duploObject.LBConfigs = &lbc
	}
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: to-CLOUD %s", apiStr, jsonData)
	return duploObject, nil
}

/////////// common place to get url + Id : follow Azure  style Ids for import//////////
func (c *Client) DuploServiceLBConfigsSetIdFromCloud(duploObject *DuploServiceLBConfigs, d *schema.ResourceData) string {
	d.Set("name", duploObject.ReplicationControllerName)
	d.Set("tenant_id", duploObject.TenantID)
	c.DuploServiceLBConfigsSetId(d)
	log.Printf("[TRACE] DuploServiceLBConfigsSetIdFromCloud 1 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) DuploServiceLBConfigsSetId(d *schema.ResourceData) string {
	tenantID := c.DuploServiceLBConfigsGetTenantId(d)
	replicationControllerName := d.Get("replication_controller_name").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, replicationControllerName)
	d.SetId(id)
	return id
}
func (c *Client) DuploServiceLBConfigsUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsUrl 1 %s ********: %s", api, host)
	return host
}
func (c *Client) DuploServiceLBConfigsListUrl(d *schema.ResourceData) string {
	tenantID := c.DuploServiceLBConfigsGetTenantId(d)
	api := fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigs 1 %s ********: %s", api, host)
	return host
}
func (c *Client) DuploServiceLBConfigsGetTenantId(d *schema.ResourceData) string {
	tenantID := d.Get("tenant_id").(string)
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

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////
func DuploServiceLBConfigsRefreshFunc(c *Client, url string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		api := url
		req2, _ := http.NewRequest("GET", url, nil)
		body, err := c.doRequest(req2)
		if err != nil {
			log.Printf("[TRACE] duplo-DuploServiceLBConfigsRefreshFunc 2 %s ********: %s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		bodyString := string(body)
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsRefreshFunc 3 %s ********: bodyString %s", api, bodyString)

		duploObject := DuploServiceLBConfigs{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] duplo-DuploServiceLBConfigsRefreshFunc 4 %s ********:  error:%s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsRefreshFunc 5 %s ******** ", api)
		var status string
		status = "pending"
		if duploObject.Status == "Ready" {
			status = "ready"
		}
		return duploObject, status, nil
	}
}
func DuploServiceLBConfigsWaitForCreation(c *Client, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: DuploServiceLBConfigsRefreshFunc(c, url),
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
	}
	log.Printf("[DEBUG] LBConfigsRefreshFuncWaitForCreation (%s)", url)
	_, err := stateConf.WaitForState()
	return err
}

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////

/////////  Utils convert //////////
func (c *Client) DuploServiceLBConfigsListFlatten(duploObjects *[]DuploServiceLBConfigs, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.DuploServiceLBConfigsToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsListFlatten 1 ******** to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsListFlatten ??? empty ?? 2 ******** from-CLOUD-LIST: %s", jsonData)
	return make([]interface{}, 0)
}
func (c *Client) DuploServiceLBConfigsFillGet(duploObject *DuploServiceLBConfigs, d *schema.ResourceData) error {
	if duploObject != nil {
		//create map
		ois := c.DuploServiceLBConfigsToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-DuploServiceLBConfigsFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	errMsg := fmt.Errorf("DuploServiceLBConfigs not found 2")
	return errMsg
}

func (c *Client) DuploServiceLBConfigsGetList(d *schema.ResourceData, m interface{}) (*[]DuploServiceLBConfigs, error) {
	//
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] DuploServiceLBConfigsGetList filters 1 ********* : %s  %v", filters, filtersOk)
	//
	api := c.DuploServiceLBConfigsListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsGetList %s 2 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsGetList %s  3 ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsGetList %s 4 ********: %s", api, bodyString)

	duploObject := []DuploServiceLBConfigs{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsGetList %s 5 ********: %d", api, len(duploObject))

	return &duploObject, nil
}

/////////   list DONE //////////

/////////  API Item //////////
func (c *Client) DuploServiceLBConfigsGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.DuploServiceLBConfigsUrl(d)
	log.Printf("[TRACE] duplo-AwsHostUpdate %s 1 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsGet %s 2 ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigsGet %s 3 ********: bodyString %s", api, bodyString)

	duploObject := DuploServiceLBConfigs{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsGet %s 4 ********: error:%s", api, err.Error())
		return err
	}

	if duploObject.TenantID != "" {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigsGet %s 4 ******** ", api)
		c.DuploServiceLBConfigsFillGet(&duploObject, d)
	}
	log.Printf("[TRACE] duplo-AwsHostGet %s 5 FOUND ***** : body: %s", api, bodyString)
	return nil
}

/////////  API Create / Update//////////
func (c *Client) DuploServiceLBConfigsCreate(d *schema.ResourceData, m interface{}) (*DuploServiceLBConfigs, error) {
	return c.DuploServiceLBConfigsCreateOrUpdate(d, m, false)
}
func (c *Client) DuploServiceLBConfigsUpdate(d *schema.ResourceData, m interface{}) (*DuploServiceLBConfigs, error) {
	return c.DuploServiceLBConfigsCreateOrUpdate(d, m, true)
}
func (c *Client) DuploServiceLBConfigsCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploServiceLBConfigs, error) {

	url := c.DuploServiceLBConfigsListUrl(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-DuploServiceLBConfigsCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-DuploServiceLBConfigsUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	duploObject, _ := c.DuploServiceLBConfigsFromState(d, m, isUpdate)

	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s", apiStr, err.Error())
		return nil, err
	}

	json_str := string(rb)
	log.Printf("[TRACE] %s 4 ********: %s", apiStr, json_str)

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

		duploObject := DuploServiceLBConfigs{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********: error:%s", apiStr, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ********  ", apiStr)
		c.DuploServiceLBConfigsSetIdFromCloud(&duploObject, d)

		////////DuploServiceLBConfigsWaitForCreation////////
		DuploServiceLBConfigsWaitForCreation(c, c.DuploServiceLBConfigsUrl(d))
		////////DuploServiceLBConfigsWaitForCreation////////

		return nil, nil
	}
	errMsg := fmt.Errorf("ERROR: in create %s body:%s error:%s", apiStr, body, err.Error())
	return nil, errMsg
}

func (c *Client) DuploServiceLBConfigsDelete(d *schema.ResourceData, m interface{}) (*DuploServiceLBConfigs, error) {

	var api = d.Id()
	url := c.DuploServiceLBConfigsUrl(d)
	log.Printf("[TRACE] duplo-DuploServiceLBConfigs %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigs %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceLBConfigs %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-DuploServiceLBConfigs %s 4 ********: ", api)
	return nil, nil
}
