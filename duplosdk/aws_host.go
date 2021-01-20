package duplosdk

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
	"time"
)

type DuploAwsHost struct {
	InstanceId        string `json:"InstanceId"`
	UserAccount       string `json:"UserAccount,omitempty"`
	TenantId          string `json:"TenantId,omitempty"`
	FriendlyName      string `json:"FriendlyName,omitempty"`
	Capacity          string `json:"Capacity,omitempty"`
	Zone              int    `json:"Zone"`
	IsMinion          bool   `json:"IsMinion"`
	ImageId           string `json:"ImageId,omitempty"`
	Base64UserData    string `json:"Base64UserData,omitempty"`
	AgentPlatform     int    `json:"AgentPlatform"`
	IsEbsOptimized    bool   `json:"IsEbsOptimized"`
	AllocatedPublicIp bool   `json:"AllocatedPublicIp,omitempty"`
	Cloud             int    `json:"Cloud"`
	EncryptDisk       bool   `json:"EncryptDisk,omitempty"`
	Status            string `json:"Status,omitempty"`
	IdentityRole      string `json:"IdentityRole,omitempty"`
	PrivateIpAddress  string `json:"PrivateIpAddress,omitempty"`
	//json objects
	Volumes    *[]DuploAwsHostVolume `json:"Volumes,omitempty"`
	Tags       *[]DuploAwsHostTag    `json:"Tags,omitempty"`
	MinionTags *[]DuploAwsHostTag    `json:"MinionTags,omitempty"`
}

type DuploAwsHostVolume struct {
	Iops       int    `json:"Iops,omitempty"`
	Name       string `json:"Name,omitempty"`
	Size       int    `Size:"Size,omitempty"`
	VolumeId   string `json:"VolumeId,omitempty"`
	VolumeType string `json:"VolumeType,omitempty"`
}

type DuploAwsHostTag struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

/////------ schema ------////
func AwsHostSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"instance_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"user_account": &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Required:         false,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"friendly_name": &schema.Schema{
			Type:             schema.TypeString,
			Optional:         false,
			Required:         true,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"capacity": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, // relaunch instnace
		},
		"zone": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			ForceNew: true, // relaunch instance
			Default:  0,
		},
		"is_minion": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  true,
		},
		"image_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, // relaunch instance
		},
		"base64_user_data": &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Required:         false,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"agent_platform": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"is_ebs_optimized": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"allocated_public_ip": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"cloud": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
			ForceNew: true, // relaunch instance
		},
		"encrypt_disk": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"status": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"identity_role": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"private_ip_address": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},

		//
		"tags": {
			Type:             schema.TypeList,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},

		"minion_tags": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},

		"volumes": {
			Type:             schema.TypeSet,
			Optional:         true,
			ForceNew:         true, // relaunch instance
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"iops": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"size": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"volume_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"volume_type": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

////// convert from cloud to state and vice versa :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) AwsHostToState(duploObject *DuploAwsHost, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-AwsHostToState ******** 1: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["instance_id"] = duploObject.InstanceId
		cObj["user_account"] = duploObject.UserAccount
		cObj["tenant_id"] = duploObject.TenantId
		cObj["friendly_name"] = duploObject.FriendlyName
		cObj["capacity"] = duploObject.Capacity
		cObj["zone"] = duploObject.Zone
		cObj["is_minion"] = duploObject.IsMinion
		cObj["image_id"] = duploObject.ImageId
		cObj["base64_user_data"] = duploObject.Base64UserData
		cObj["agent_platform"] = duploObject.AgentPlatform
		cObj["is_ebs_optimized"] = duploObject.IsEbsOptimized
		cObj["allocated_public_ip"] = duploObject.AllocatedPublicIp
		cObj["cloud"] = duploObject.Cloud
		cObj["encrypt_disk"] = duploObject.EncryptDisk
		cObj["status"] = duploObject.Status
		cObj["identity_role"] = duploObject.IdentityRole
		cObj["private_ip_address"] = duploObject.PrivateIpAddress
		//
		cObj["tags"] = c.AwsHostTagsToState(duploObject.Tags, d)
		cObj["minion_tags"] = c.AwsHostTagsToState(duploObject.MinionTags, d)
		cObj["volumes"] = c.AwsHostVolumesToState(duploObject.Volumes, d)
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-AwsHostToState ******** 2: to-DICT %s ", jsonData2)

		return cObj
	}
	return nil
}
func (c *Client) AwsHostTagsToState(duploObjects *[]DuploAwsHostTag, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			cObj := make(map[string]interface{})
			///--- set
			cObj["key"] = duploObject.Key
			cObj["value"] = duploObject.Value
			ois[i] = cObj
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-AwsHostTagsToState ******** to-DICT: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-AwsHostTagsToState ??? empty ?? ******** from-CLOUD: %s", jsonData)
	return make([]interface{}, 0)
}
func (c *Client) AwsHostVolumesToState(duploObjects *[]DuploAwsHostVolume, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			cObj := make(map[string]interface{})
			///--- set
			cObj["iops"] = duploObject.Iops
			cObj["name"] = duploObject.Name
			cObj["size"] = duploObject.Size
			cObj["volume_id"] = duploObject.VolumeId
			cObj["volume_type"] = duploObject.VolumeType
			ois[i] = cObj
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-AwsHostVolumesToState ******** 1 to-DICT: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-AwsHostVolumesToState ??? empty ?? ******** 2 from-CLOUD: \n%s", jsonData)
	return make([]interface{}, 0)
}
func (c *Client) DuploAwsHostFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploAwsHost, error) {
	url := c.AwsHostListUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploAwsHostFromState-Create %s ", url)
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploAwsHostFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)
	//
	duploObject := new(DuploAwsHost)
	///--- set
	//SKIP InstanceId? TenantId? Status? IdentityRole PrivateIpAddress
	duploObject.TenantId = d.Get("tenant_id").(string)
	duploObject.InstanceId = d.Get("instance_id").(string)
	duploObject.UserAccount = d.Get("user_account").(string)
	duploObject.FriendlyName = d.Get("friendly_name").(string)
	duploObject.Capacity = d.Get("capacity").(string)
	duploObject.Zone = d.Get("zone").(int)
	duploObject.IsMinion = d.Get("is_minion").(bool)
	duploObject.ImageId = d.Get("image_id").(string)
	duploObject.Base64UserData = d.Get("base64_user_data").(string)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.IsEbsOptimized = d.Get("is_ebs_optimized").(bool)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.AllocatedPublicIp = d.Get("allocated_public_ip").(bool)
	duploObject.Cloud = d.Get("cloud").(int)
	duploObject.EncryptDisk = d.Get("encrypt_disk").(bool)

	//todo: tags
	miniontags := d.Get("minion_tags").([]interface{})
	if len(miniontags) > 0 {
		var pc []DuploAwsHostTag
		for _, raw := range miniontags {
			p := raw.(map[string]interface{})
			pc = append(pc, DuploAwsHostTag{
				Key:   p["key"].(string),
				Value: p["value"].(string),
			})
		}
		duploObject.MinionTags = &pc
	}
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: to-CLOUD %s", api_str, jsonData)
	return duploObject, nil
}

//this is the import-id for terraform inspired from azure imports
func (c *Client) DuploAwsHostSetIdFromCloud(duploObject *DuploAwsHost, d *schema.ResourceData) string {
	d.Set("instance_id", duploObject.InstanceId)
	d.Set("tenant_id", duploObject.TenantId)
	c.AwsHostSetId(d)
	log.Printf("[TRACE] DuploAwsHostSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) AwsHostSetId(d *schema.ResourceData) string {
	tenant_id := c.AwsHostGetTenantId(d)
	instance_id := d.Get("instance_id").(string)
	///--- set
	id := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenant_id, instance_id)
	d.SetId(id)
	return id
}

//api for crud -- get + delete
func (c *Client) AwsHostUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-AwsHostUrl %s 1 ********: %s", api, host)
	return host
}

// app for -- get list + create + update
func (c *Client) AwsHostListUrl(d *schema.ResourceData) string {
	tenant_id := d.Get("tenant_id").(string)
	api := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", tenant_id)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-AwsHostListUrl %s 1 ********: %s", api, host)
	return host
}

// tenant_id or any  parents in import url should be handled if not part of get json
func (c *Client) AwsHostGetTenantId(d *schema.ResourceData) string {
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

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////
func AwsHostRefreshFunc(c *Client, url string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		api := url
		req2, _ := http.NewRequest("GET", url, nil)
		body, err := c.doRequest(req2)
		if err != nil {
			log.Printf("[TRACE] duplo-AwsHostRefreshFunc 2 %s ********: %s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		bodyString := string(body)
		log.Printf("[TRACE] duplo-AwsHostRefreshFunc 3 %s ********: bodyString %s", api, bodyString)

		duploObject := DuploAwsHost{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] duplo-AwsHostRefreshFunc 4 %s ********:  error:%s", api, err.Error())
			return nil, "", fmt.Errorf("error reading 1 (%s): %s", url, err)
		}
		log.Printf("[TRACE] duplo-AwsHostRefreshFunc 5 %s ******** ", api)
		var status string
		status = "pending"
		if duploObject.Status == "running" {
			status = "ready"
		}
		return duploObject, status, nil
	}
}
func AwsHostWaitForCreation(c *Client, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: AwsHostRefreshFunc(c, url),
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      30 * time.Minute,
	}
	log.Printf("[DEBUG] AwsHostRefreshFuncWiatForCreation (%s)", url)
	_, err := stateConf.WaitForState()
	return err
}

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////

//Utils convert for get-list   -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) AwsHostsFlatten(duploObjects *[]DuploAwsHost, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.AwsHostToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-AwsHostToState ******** 1 to-DICT-LIST: %s", jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-AwsHostTagsToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}

//convert  get-list item into state  -- cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) AwsHostFillGet(duploObject *DuploAwsHost, d *schema.ResourceData) error {
	if duploObject != nil {
		//create map
		ois := c.AwsHostToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-AwsHostFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-AwsHostFillGet 2 Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	err_msg := fmt.Errorf("AwsHost not found")
	return err_msg
}

/////////  API list //////////
func (c *Client) AwsHostGetList(d *schema.ResourceData, m interface{}) (*[]DuploAwsHost, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] AwsHostGetList filters 1 ********* : %s  %s", filters, filtersOk)
	//
	api := c.AwsHostListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-AwsHostGetList 2 %s  ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostGetList 3 %s   ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-AwsHostGetList 4 %s   ********: %s", api, bodyString)

	duploObjects := []DuploAwsHost{}
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-AwsHostGetList 5 %s  ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}

/////////   list DONE //////////

/////////  API Item //////////
func (c *Client) AwsHostGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.AwsHostUrl(d)
	log.Printf("[TRACE] duplo-AwsHostGet 1  %s ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostGet 2 %s ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-AwsHostGet 3 %s ********: bodyString %s", api, bodyString)

	duploObject := DuploAwsHost{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostGet 4 %s ********:  error:%s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-AwsHostGet 5 %s ******** ", api)
	if duploObject.TenantId != "" {
		c.AwsHostFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-AwsHostGet 6 FOUND *****", api)
		return nil
	}
	err_msg := fmt.Errorf("AwsHost not found  : %s body:%s", api, bodyString)
	return err_msg
}

/////////  API Item //////////

/////////  API  Create/update //////////

func (c *Client) AwsHostCreate(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	return c.AwsHostCreateOrUpdate(d, m, false)
}

func (c *Client) AwsHostUpdate(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	return c.AwsHostCreateOrUpdate(d, m, true)
}

func (c *Client) AwsHostCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploAwsHost, error) {
	url := c.AwsHostListUrl(d)
	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-AwsHostCreate %s ", api)
	if isUpdate {
		action = "PUT"
		api_str = fmt.Sprintf("duplo-AwsHostUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	//
	duploObject, _ := c.DuploAwsHostFromState(d, m, isUpdate)
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

		duploObject := DuploAwsHost{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s", api_str, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ", api_str)
		c.DuploAwsHostSetIdFromCloud(&duploObject, d)

		////////AwsHostWaitForCreation////////
		//todo: test AwsHostWaitForCreation(c, c.AwsHostUrl(d))
		////////AwsHostWaitForCreation////////

		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %d,   body: %s", api, body)
	return nil, err_msg
}

/////////  API Create/update //////////

/////////  API Delete //////////
func (c *Client) AwsHostDelete(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	var api = d.Id()
	url := c.AwsHostUrl(d)
	log.Printf("[TRACE] duplo-AwsHostDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostDelete %s 2 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-AwsHostDelete %s 3 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-AwsHostDelete %s 4 ********: %s", api)
	return nil, nil
}

/////////  API Delete //////////
