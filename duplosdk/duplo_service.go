package duplosdk

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"strings"
)

type DuploService struct {
	Name                    string                   `json:"Name"`
	TenantId                string                   `json:"TenantId,omitempty"`
	OtherDockerHostConfig   string                   `json:"OtherDockerHostConfig,omitempty"`
	OtherDockerConfig       string                   `json:"OtherDockerConfig,omitempty"`
	AllocationTags          string                   `json:"AllocationTags,omitempty"`
	ExtraConfig             string                   `json:"ExtraConfig,omitempty"`
	Commands                string                   `json:"Commands,omitempty"`
	Volumes                 string                   `json:"Volumes,omitempty"`
	DockerImage             string                   `json:"DockerImage"`
	ReplicasMatchingAsgName string                   `json:"ReplicasMatchingAsgName,omitempty"`
	Replicas                int                      `json:"Replicas"`
	AgentPlatform           int                      `json:"AgentPlatform"`
	Cloud                   int                      `json:"Cloud"`
	Tags                    []map[string]interface{} `json:"Tags,omitempty"`
}

/////------ schema ------////
func DuploServiceSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"tenant_id": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"other_docker_host_config": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"other_docker_config": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"extra_config": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"allocation_tags": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"volumes": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"commands": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"cloud": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"agent_platform": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"replicas": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"replicas_matching_asg_name": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"docker_image": &schema.Schema{
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		//
		"tags": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
	}
}

////// convert from cloud to state :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceToState(duploObject *DuploService, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		//log
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-DuploServiceToState 1 ********: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["name"] = duploObject.Name
		cObj["other_docker_host_config"] = duploObject.OtherDockerHostConfig
		cObj["tenant_id"] = duploObject.TenantId
		cObj["other_docker_config"] = duploObject.OtherDockerConfig
		cObj["allocation_tags"] = duploObject.AllocationTags
		cObj["extra_config"] = duploObject.ExtraConfig
		cObj["commands"] = duploObject.Commands
		cObj["volumes"] = duploObject.Volumes
		cObj["docker_image"] = duploObject.DockerImage
		cObj["agent_platform"] = duploObject.AgentPlatform
		cObj["replicas_matching_asg_name"] = duploObject.ReplicasMatchingAsgName
		cObj["replicas"] = duploObject.Replicas
		cObj["cloud"] = duploObject.Cloud
		cObj["tags"] = duploObject.Tags
		//log
		jsonData2, _ := json.Marshal(cObj)
		log.Printf("[TRACE] duplo-DuploServiceToState 2 ********: to-DICT %s ", jsonData2)
		return cObj
	}
	return nil
}

////// convert from state to cloud :  cloud names (CamelCase) to tf names (SnakeCase)
func (c *Client) DuploServiceFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploService, error) {
	url := c.DuploServiceUrl(d)
	var api_str = fmt.Sprintf("duplo-DuploServiceFromState-Create %s ", url)
	if isUpdate {
		api_str = fmt.Sprintf("duplo-DuploServiceFromState-Create %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: ", api_str)

	//object
	duploObject := new(DuploService)
	///--- set
	duploObject.Name = d.Get("name").(string)
	duploObject.OtherDockerHostConfig = d.Get("other_docker_host_config").(string)
	duploObject.OtherDockerConfig = d.Get("other_docker_config").(string)
	duploObject.AllocationTags = d.Get("allocation_tags").(string)
	duploObject.ExtraConfig = d.Get("extra_config").(string)
	duploObject.Commands = d.Get("commands").(string)
	duploObject.Volumes = d.Get("volumes").(string)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.DockerImage = d.Get("docker_image").(string)
	duploObject.AgentPlatform = d.Get("agent_platform").(int)
	duploObject.ReplicasMatchingAsgName = d.Get("replicas_matching_asg_name").(string)
	duploObject.Cloud = d.Get("cloud").(int)
	duploObject.Replicas = d.Get("replicas").(int)
	//log
	jsonData2, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] %s 2 ********: %s to-CLOUD", api_str, jsonData2)

	return duploObject, nil
}

///////// ///////// ///////// /////////  Utils convert ////////////////////

/////////// common place to get url + Id : follow Azure  style Ids for import//////////
func (c *Client) DuploServiceSetIdFromCloud(duploObject *DuploService, d *schema.ResourceData) string {
	d.Set("name", duploObject.Name)
	d.Set("tenant_id", duploObject.TenantId)
	c.DuploServiceSetId(d)
	log.Printf("[TRACE] DuploServiceSetIdFromCloud 1 ********: %s", d.Id())
	return d.Id()
}
func (c *Client) DuploServiceSetId(d *schema.ResourceData) string {
	tenant_id := c.DuploServiceGetTenantId(d)
	name := d.Get("name").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenant_id, name)
	d.SetId(id)
	return id
}
func (c *Client) DuploServiceUrl(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceUrl %s 1 ********: %s", api, host)
	return host
}
func (c *Client) DuploServiceListUrl(d *schema.ResourceData) string {
	tenant_id := c.DuploServiceParamsGetTenantId(d)
	api := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2", tenant_id)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceListUrl %s 1 ********: %s", api, host)
	return host
}

func (c *Client) DuploServiceGetTenantId(d *schema.ResourceData) string {
	tenant_id := d.Get("tenant_id").(string)
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
func (c *Client) DuploServicesFlatten(duploObjects *[]DuploService, d *schema.ResourceData) []interface{} {
	if duploObjects != nil {
		ois := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			ois[i] = c.DuploServiceToState(&duploObject, d)
		}
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServicesFlatten ******** jsonData: %s", jsonData)
		return ois
	}
	return make([]interface{}, 0)
}

func (c *Client) DuploServiceFillGet(duploObject *DuploService, d *schema.ResourceData) error {
	if duploObject != nil {

		//create map
		ois := c.DuploServiceToState(duploObject, d)
		//log
		jsonData, _ := json.Marshal(ois)
		log.Printf("[TRACE] duplo-DuploServiceFillGet 1 ********: to-DICT %s ", jsonData)
		// transfer from map to state
		for key, element := range ois {
			fmt.Println("[TRACE] duplo-DuploServiceFillGet 2 ******** Key:", key, "=>", "Element:", element)
			d.Set(key, ois[key])
		}
		return nil
	}
	err_msg := fmt.Errorf("DuploService not found 2")
	return err_msg
}

/////////  API list //////////
func (c *Client) DuploServiceGetList(d *schema.ResourceData, m interface{}) (*[]DuploService, error) {
	//
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] DuploServiceGetList filters ********* 1 : %s  %s", filters, filtersOk)
	//
	api := c.DuploServiceListUrl(d)
	url := api
	log.Printf("[TRACE] duplo-DuploServiceGetList %s 2 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceGetList %s 1 ********: %s", api, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceGetList %s 2 ********: bodyString %s", api, bodyString)

	duploObjects := []DuploService{}
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-DuploServiceGetList %s 3 ********: %s", api, len(duploObjects))

	return &duploObjects, nil
}

func (c *Client) DuploServiceGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.DuploServiceUrl(d)
	log.Printf("[TRACE] duplo-DuploServiceUpdate %s 1 ********: %s", api, url)
	//
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceGet %s 2 ********: %s", api, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-DuploServiceGet %s 3 ********: bodyString %s", api, bodyString)

	duploObject := DuploService{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceGet %s 4 ********: error %s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-DuploServiceGet %s 5 ******** ", api)
	if duploObject.TenantId != "" {
		c.DuploServiceFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-DuploServiceGet 6 FOUND ***** : %s ", api)
		return nil
	}
	err_msg := fmt.Errorf("DuploService not found 7 : %s bodyString %s ", api, bodyString)
	return err_msg
}

/////////  API Create //////////
func (c *Client) DuploServiceCreate(d *schema.ResourceData, m interface{}) (*DuploService, error) {
	return c.DuploServiceCreateOrUpdate(d, m, false)
}
func (c *Client) DuploServiceUpdate(d *schema.ResourceData, m interface{}) (*DuploService, error) {
	return c.DuploServiceCreateOrUpdate(d, m, true)
}
func (c *Client) DuploServiceCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploService, error) {
	url := c.DuploServiceListUrl(d)
	api := url
	var action = "POST"
	var api_str = fmt.Sprintf("duplo-DuploServiceCreate %s ", api)
	if isUpdate {
		action = "PUT"
		api_str = fmt.Sprintf("duplo-DuploServiceUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", api_str, url)

	//
	duploObject, _ := c.DuploServiceFromState(d, m, isUpdate)
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

		duploObject := DuploService{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			return nil, err
		}
		log.Printf("[TRACE] %s 8 ********: ", api_str)
		c.DuploServiceSetIdFromCloud(&duploObject, d)
		return nil, nil
	}
	err_msg := fmt.Errorf("ERROR: in create %s body: %s", api_str, body)
	return nil, err_msg
}

func (c *Client) DuploServiceDelete(d *schema.ResourceData, m interface{}) (*DuploService, error) {

	var api = d.Id()
	url := c.DuploServiceUrl(d)
	log.Printf("[TRACE] duplo-DuploServiceDelete %s 1 ********: %s", api, url)

	//
	req, err := http.NewRequest("DELETE", url, strings.NewReader("{\"a\":\"b\"}"))
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceDelete %s 4 ********: %s", api, err.Error())
		return nil, err
	}

	body, err := c.doRequestWithStatus(req, 204)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceDelete %s 5 ********: %s", api, err.Error())
		return nil, err
	}

	if body != nil {
		//nothing ?
	}

	log.Printf("[TRACE] DONE duplo-DuploServiceDelete %s 5 ********:  ", api)
	return nil, nil
}
