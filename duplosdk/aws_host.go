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

// DuploAwsHost is a Duplo SDK object that represents an AWS host
type DuploAwsHost struct {
	InstanceID        string `json:"InstanceId"`
	UserAccount       string `json:"UserAccount,omitempty"`
	TenantID          string `json:"TenantId,omitempty"`
	FriendlyName      string `json:"FriendlyName,omitempty"`
	Capacity          string `json:"Capacity,omitempty"`
	Zone              int    `json:"Zone"`
	IsMinion          bool   `json:"IsMinion"`
	ImageID           string `json:"ImageId,omitempty"`
	Base64UserData    string `json:"Base64UserData,omitempty"`
	AgentPlatform     int    `json:"AgentPlatform"`
	IsEbsOptimized    bool   `json:"IsEbsOptimized"`
	AllocatedPublicIP bool   `json:"AllocatedPublicIp,omitempty"`
	Cloud             int    `json:"Cloud"`
	EncryptDisk       bool   `json:"EncryptDisk,omitempty"`
	Status            string `json:"Status,omitempty"`
	IdentityRole      string `json:"IdentityRole,omitempty"`
	PrivateIPAddress  string `json:"PrivateIpAddress,omitempty"`
	//json objects
	Volumes    *[]DuploAwsHostVolume `json:"Volumes,omitempty"`
	MetaData   *[]DuploAwsHostKv     `json:"MetaData,omitempty"`
	Tags       *[]DuploAwsHostKv     `json:"Tags,omitempty"`
	MinionTags *[]DuploAwsHostKv     `json:"MinionTags,omitempty"`
}

// DuploAwsHostVolume is a Duplo SDK object that represents a volume of an AWS host
type DuploAwsHostVolume struct {
	Iops       int    `json:"Iops,omitempty"`
	Name       string `json:"Name,omitempty"`
	Size       int    `Size:"Size,omitempty"`
	VolumeID   string `json:"VolumeId,omitempty"`
	VolumeType string `json:"VolumeType,omitempty"`
}

// DuploAwsHostKv is a Duplo SDK object that represents a key value pair in an AWS host
type DuploAwsHostKv struct {
	Key   string `json:"Key,omitempty"`
	Value string `json:"Value,omitempty"`
}

// AwsHostSchema returns a Terraform resource schema for an AWS host
func AwsHostSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"instance_id": {
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"user_account": {
			Type:             schema.TypeString,
			Optional:         true,
			Required:         false,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"friendly_name": {
			Type:             schema.TypeString,
			Optional:         false,
			Required:         true,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"capacity": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, // relaunch instnace
		},
		"zone": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			ForceNew: true, // relaunch instance
			Default:  0,
		},
		"is_minion": {
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  true,
		},
		"image_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, // relaunch instance
		},
		"base64_user_data": {
			Type:             schema.TypeString,
			Optional:         true,
			Required:         false,
			DiffSuppressFunc: diffIgnoreIfAlreadySet,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"is_ebs_optimized": {
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"allocated_public_ip": {
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
			ForceNew: true, // relaunch instance
		},
		"encrypt_disk": {
			Type:     schema.TypeBool,
			Optional: true,
			Required: false,
			Default:  false,
			ForceNew: true, // relaunch instance
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"identity_role": {
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},
		"private_ip_address": {
			Type:     schema.TypeString,
			Computed: true,
			Required: false,
		},

		"metadata": {
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

		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Required: false,
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
			Required: false,
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

// AwsHostToState converts a Duplo SDK object representing an AWS Host to Terraform resource data
func (c *Client) AwsHostToState(duploObject *DuploAwsHost, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-AwsHostToState ******** 1: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["instance_id"] = duploObject.InstanceID
		cObj["user_account"] = duploObject.UserAccount
		cObj["tenant_id"] = duploObject.TenantID
		cObj["friendly_name"] = duploObject.FriendlyName
		cObj["capacity"] = duploObject.Capacity
		cObj["zone"] = duploObject.Zone
		cObj["is_minion"] = duploObject.IsMinion
		cObj["image_id"] = duploObject.ImageID
		cObj["base64_user_data"] = duploObject.Base64UserData
		cObj["agent_platform"] = duploObject.AgentPlatform
		cObj["is_ebs_optimized"] = duploObject.IsEbsOptimized
		cObj["allocated_public_ip"] = duploObject.AllocatedPublicIP
		cObj["cloud"] = duploObject.Cloud
		cObj["encrypt_disk"] = duploObject.EncryptDisk
		cObj["status"] = duploObject.Status
		cObj["identity_role"] = duploObject.IdentityRole
		cObj["private_ip_address"] = duploObject.PrivateIPAddress
		//
		cObj["metadata"] = c.AwsHostKvToState("metadata", duploObject.MetaData, d)
		cObj["tags"] = c.AwsHostKvToState("tags", duploObject.Tags, d)
		cObj["minion_tags"] = c.AwsHostKvToState("minion_tags", duploObject.MinionTags, d)
		cObj["volumes"] = c.AwsHostVolumesToState(duploObject.Volumes, d)
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-AwsHostToState ******** 2: to-DICT %s ", jsonData2)

		return cObj
	}
	return nil
}

// AwsHostKvFromState converts partial Terraform resource data representing a key value pair to a Duplo SDK object
func (c *Client) AwsHostKvFromState(fieldName string, d *schema.ResourceData) *[]DuploAwsHostKv {
	var ary []DuploAwsHostKv

	kvs := d.Get(fieldName).([]interface{})
	if len(kvs) > 0 {
		log.Printf("[TRACE] AwsHostKvFromState ********: found %s", fieldName)
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, DuploAwsHostKv{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

// AwsHostKvToState converts a Duplo SDK object representing a key value pair to partial Terraform resource data
func (c *Client) AwsHostKvToState(fieldName string, duploObjects *[]DuploAwsHostKv, d *schema.ResourceData) []interface{} {
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
		log.Printf("[TRACE] duplo-AwsHostKvToState[%s] ******** to-DICT: %s", fieldName, jsonData)
		return ois
	}
	jsonData, _ := json.Marshal(&duploObjects)
	log.Printf("[TRACE] duplo-AwsHostKvToState[%s] ??? empty ?? ******** from-CLOUD: %s", fieldName, jsonData)
	return make([]interface{}, 0)
}

// AwsHostVolumesToState converts a Duplo SDK object representing a volume to partial Terraform resource data
func (c *Client) AwsHostVolumesToState(duploObjects *[]DuploAwsHostVolume, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			cObj := make(map[string]interface{})
			///--- set
			cObj["iops"] = duploObject.Iops
			cObj["name"] = duploObject.Name
			cObj["size"] = duploObject.Size
			cObj["volume_id"] = duploObject.VolumeID
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

// DuploAwsHostFromState converts Terraform resource data representing an AWS host into partial Terraform resource data
func (c *Client) DuploAwsHostFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploAwsHost, error) {
	url := c.AwsHostListURL(d)
	var apiStr = fmt.Sprintf("duplo-DuploAwsHostFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploAwsHostFromState-update %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)
	//
	duploObject := new(DuploAwsHost)
	///--- set
	//SKIP InstanceId? TenantId? Status? IdentityRole PrivateIpAddress
	duploObject.TenantID = d.Get("tenant_id").(string)
	duploObject.InstanceID = d.Get("instance_id").(string)
	duploObject.UserAccount = d.Get("user_account").(string)
	duploObject.FriendlyName = d.Get("friendly_name").(string)
	duploObject.Capacity = d.Get("capacity").(string)
	duploObject.Zone = d.Get("zone").(int)
	duploObject.IsMinion = d.Get("is_minion").(bool)
	duploObject.ImageID = d.Get("image_id").(string)
	duploObject.Base64UserData = d.Get("base64_user_data").(string)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.IsEbsOptimized = d.Get("is_ebs_optimized").(bool)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.AllocatedPublicIP = d.Get("allocated_public_ip").(bool)
	duploObject.Cloud = d.Get("cloud").(int)
	duploObject.EncryptDisk = d.Get("encrypt_disk").(bool)

	//todo: tags
	//todo: volumes

	duploObject.MinionTags = c.AwsHostKvFromState("minion_tags", d)
	duploObject.MetaData = c.AwsHostKvFromState("metadata", d)

	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: to-CLOUD %s", apiStr, jsonData)
	return duploObject, nil
}

// DuploAwsHostSetIDFromCloud populates the resource ID based on instance_id and tenant_id
func (c *Client) DuploAwsHostSetIDFromCloud(duploObject *DuploAwsHost, d *schema.ResourceData) string {
	d.Set("instance_id", duploObject.InstanceID)
	d.Set("tenant_id", duploObject.TenantID)
	c.AwsHostSetID(d)
	log.Printf("[TRACE] DuploAwsHostSetIdFromCloud 2 ********: %s", d.Id())
	return d.Id()
}

// AwsHostSetID populates the resource ID based on instance_id and tenant_id
func (c *Client) AwsHostSetID(d *schema.ResourceData) string {
	tenantID := c.AwsHostGetTenantID(d)
	instanceID := d.Get("instance_id").(string)

	///--- set
	id := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2/%s", tenantID, instanceID)
	d.SetId(id)
	return id
}

// AwsHostURL returns the base API URL for crud -- get + delete
func (c *Client) AwsHostURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-AwsHostUrl %s 1 ********: %s", api, host)
	return host
}

// AwsHostListURL returns the base API URL for crud -- get list + create + update
func (c *Client) AwsHostListURL(d *schema.ResourceData) string {
	tenantID := d.Get("tenant_id").(string)
	api := fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-AwsHostListUrl %s 1 ********: %s", api, host)
	return host
}

// AwsHostGetTenantID tries to retrieve (or synthesize) a tenant_id based on resource data
// - tenant_id or any parents in import url should be handled if not part of get json
func (c *Client) AwsHostGetTenantID(d *schema.ResourceData) string {
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

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////

// AwsHostRefreshFunc refreshes AWS host information from the Duplo API.
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

// AwsHostWaitForCreation waits for creation of an AWS Host by the Duplo API
func AwsHostWaitForCreation(c *Client, url string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: AwsHostRefreshFunc(c, url),
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      30 * time.Minute,
	}
	log.Printf("[DEBUG] AwsHostRefreshFuncWaitForCreation (%s)", url)
	_, err := stateConf.WaitForState()
	return err
}

//////////////////////////////////////////////////////////////////////////
///////////////////////////////////// refresh state //////////////////////
/////////////////////////////////////////////////////////////////////////

// AwsHostsFlatten converts a list of Duplo SDK objects into Terraform resource data
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
	log.Printf("[TRACE] duplo-AwsHostKvToState ??? empty ?? 2 ******** from-CLOUD-LIST: \n%s", jsonData)
	return make([]interface{}, 0)
}

// AwsHostFillGet converts a Duplo SDK object into Terraform resource data
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
	errMsg := fmt.Errorf("AwsHost not found")
	return errMsg
}

// AwsHostGetList retrieves a list of AWS hosts via the Duplo API.
func (c *Client) AwsHostGetList(d *schema.ResourceData, m interface{}) (*[]DuploAwsHost, error) {
	//todo: filter other than tenant
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] AwsHostGetList filters 1 ********* : %s  %v", filters, filtersOk)
	//
	api := c.AwsHostListURL(d)
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
	log.Printf("[TRACE] duplo-AwsHostGetList 5 %s  ********: %d", api, len(duploObjects))

	return &duploObjects, nil
}

/////////   list DONE //////////

// AwsHostGet retrieves an AWS host via the Duplo API.
func (c *Client) AwsHostGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.AwsHostURL(d)
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
	if duploObject.TenantID != "" {
		c.AwsHostFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-AwsHostGet 6 %s FOUND *****", api)
		return nil
	}
	errMsg := fmt.Errorf("AwsHost not found  : %s body:%s", api, bodyString)
	return errMsg
}

/////////  API Item //////////

/////////  API  Create/update //////////

// AwsHostCreate creates an AWS host via the Duplo API.
func (c *Client) AwsHostCreate(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	return c.AwsHostCreateOrUpdate(d, m, false)
}

// AwsHostUpdate updates an AWS host via the Duplo API.
func (c *Client) AwsHostUpdate(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	return c.AwsHostCreateOrUpdate(d, m, true)
}

// AwsHostCreateOrUpdate creates or updates an AWS host via the Duplo API.
func (c *Client) AwsHostCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploAwsHost, error) {
	url := c.AwsHostListURL(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-AwsHostCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-AwsHostUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)

	//
	duploObject, _ := c.DuploAwsHostFromState(d, m, isUpdate)
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s", apiStr, err.Error())
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

		duploObject := DuploAwsHost{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] %s 8 ********:  error: %s", apiStr, err.Error())
			return nil, err
		}
		log.Printf("[TRACE] %s 9 ******** ", apiStr)
		c.DuploAwsHostSetIDFromCloud(&duploObject, d)

		////////AwsHostWaitForCreation////////
		//todo: test AwsHostWaitForCreation(c, c.AwsHostUrl(d))
		////////AwsHostWaitForCreation////////

		return nil, nil
	}
	errMsg := fmt.Errorf("ERROR: in create %s,   body: %s", api, body)
	return nil, errMsg
}

// AwsHostDelete deletes an AWS host via the Duplo API.
func (c *Client) AwsHostDelete(d *schema.ResourceData, m interface{}) (*DuploAwsHost, error) {
	var api = d.Id()
	url := c.AwsHostURL(d)
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

	log.Printf("[TRACE] DONE duplo-AwsHostDelete %s 4 ********", api)
	return nil, nil
}

/////////  API Delete //////////
