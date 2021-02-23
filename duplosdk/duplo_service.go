package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploService represents a service in the Duplo SDK
type DuploService struct {
	Name                    string                   `json:"Name"`
	TenantID                string                   `json:"TenantId,omitempty"`
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

// DuploServiceSchema returns a Terraform resource schema for a service's parameters
func DuploServiceSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"name": {
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
		"other_docker_host_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"other_docker_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"extra_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"allocation_tags": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"volumes": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"commands": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"replicas_matching_asg_name": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"docker_image": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		//
		"tags": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
	}
}

// DuploServiceToState converts a Duplo SDK object respresenting a service to terraform resource data.
func (c *Client) DuploServiceToState(duploObject *DuploService, d *schema.ResourceData) map[string]interface{} {
	if duploObject != nil {
		//log
		jsonData, _ := json.Marshal(duploObject)
		log.Printf("[TRACE] duplo-DuploServiceToState 1 ********: from-CLOUD %s ", jsonData)

		cObj := make(map[string]interface{})
		///--- set
		cObj["name"] = duploObject.Name
		cObj["other_docker_host_config"] = duploObject.OtherDockerHostConfig
		cObj["tenant_id"] = duploObject.TenantID
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

// DuploServiceFromState converts resource data respresenting a service to a Duplo SDK object.
func (c *Client) DuploServiceFromState(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploService, error) {
	url := c.DuploServiceURL(d)
	var apiStr = fmt.Sprintf("duplo-DuploServiceFromState-Create %s ", url)
	if isUpdate {
		apiStr = fmt.Sprintf("duplo-DuploServiceFromState-Create %s ", url)
	}
	log.Printf("[TRACE] %s 1 ********: ", apiStr)

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
	log.Printf("[TRACE] %s 2 ********: %s to-CLOUD", apiStr, jsonData2)

	return duploObject, nil
}

///////// ///////// ///////// /////////  Utils convert ////////////////////

// DuploServiceSetIDFromCloud populates the resource ID based on name and tenant_id
func (c *Client) DuploServiceSetIDFromCloud(duploObject *DuploService, d *schema.ResourceData) string {
	d.Set("name", duploObject.Name)
	d.Set("tenant_id", duploObject.TenantID)
	c.DuploServiceSetID(d)
	log.Printf("[TRACE] DuploServiceSetIdFromCloud 1 ********: %s", d.Id())
	return d.Id()
}

// DuploServiceSetID populates the resource ID based on name and tenant_id
func (c *Client) DuploServiceSetID(d *schema.ResourceData) string {
	tenantID := c.DuploServiceGetTenantID(d)
	name := d.Get("name").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenantID, name)
	d.SetId(id)
	return id
}

// DuploServiceURL returns the base API URL for crud -- get + delete
func (c *Client) DuploServiceURL(d *schema.ResourceData) string {
	api := d.Id()
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceUrl %s 1 ********: %s", api, host)
	return host
}

// DuploServiceListURL returns the base API URL for crud -- get list + create + update
func (c *Client) DuploServiceListURL(d *schema.ResourceData) string {
	tenantID := c.DuploServiceParamsGetTenantID(d)
	api := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2", tenantID)
	host := fmt.Sprintf("%s/%s", c.HostURL, api)
	log.Printf("[TRACE] duplo-DuploServiceListUrl %s 1 ********: %s", api, host)
	return host
}

// DuploServiceGetTenantID tries to retrieve (or synthesize) a tenant_id based on resource data
// - tenant_id or any parents in import url should be handled if not part of get json
func (c *Client) DuploServiceGetTenantID(d *schema.ResourceData) string {
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

/////////// common place to get url + Id : follow Azure  style Ids for import//////////

/////////  Utils convert //////////

// DuploServicesFlatten converts a list of Duplo SDK objects into Terraform resource data
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

// DuploServiceFillGet converts a Duplo SDK object into Terraform resource data
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
	return fmt.Errorf("DuploService not found 2")
}

/////////  API list //////////

// DuploServiceGetList retrieves a list of services via the Duplo API.
func (c *Client) DuploServiceGetList(d *schema.ResourceData, m interface{}) (*[]DuploService, error) {
	//
	filters, filtersOk := d.GetOk("filter")
	log.Printf("[TRACE] DuploServiceGetList filters ********* 1 : %s  %v", filters, filtersOk)
	//
	api := c.DuploServiceListURL(d)
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
	log.Printf("[TRACE] duplo-DuploServiceGetList %s 3 ********: %d", api, len(duploObjects))

	return &duploObjects, nil
}

// DuploServiceGet retrieves a service's load balancer via the Duplo API.
func (c *Client) DuploServiceGet(d *schema.ResourceData, m interface{}) error {
	var api = d.Id()
	url := c.DuploServiceURL(d)
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
	if bodyString == "" || bodyString == "null" {
		d.Set("name", "")
		return nil
	}

	duploObject := DuploService{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] duplo-DuploServiceGet %s 4 ********: error %s", api, err.Error())
		return err
	}
	log.Printf("[TRACE] duplo-DuploServiceGet %s 5 ******** ", api)
	if duploObject.TenantID != "" {
		c.DuploServiceFillGet(&duploObject, d)
		log.Printf("[TRACE] duplo-DuploServiceGet 6 FOUND ***** : %s ", api)
		return nil
	}
	return fmt.Errorf("DuploService not found 7 : %s bodyString %s ", api, bodyString)
}

/////////  API Create //////////

// DuploServiceCreate creates a service via the Duplo API.
func (c *Client) DuploServiceCreate(d *schema.ResourceData, m interface{}) (*DuploService, error) {
	return c.DuploServiceCreateOrUpdate(d, m, false)
}

// DuploServiceUpdate updates a service via the Duplo API.
func (c *Client) DuploServiceUpdate(d *schema.ResourceData, m interface{}) (*DuploService, error) {
	return c.DuploServiceCreateOrUpdate(d, m, true)
}

// DuploServiceCreateOrUpdate creates or updates a service via the Duplo API.
func (c *Client) DuploServiceCreateOrUpdate(d *schema.ResourceData, m interface{}, isUpdate bool) (*DuploService, error) {
	url := c.DuploServiceListURL(d)
	api := url
	var action = "POST"
	var apiStr = fmt.Sprintf("duplo-DuploServiceCreate %s ", api)
	if isUpdate {
		action = "PUT"
		apiStr = fmt.Sprintf("duplo-DuploServiceUpdate %s ", api)
	}
	log.Printf("[TRACE] %s 1 ********: %s", apiStr, url)

	//
	duploObject, _ := c.DuploServiceFromState(d, m, isUpdate)
	//
	jsonData, _ := json.Marshal(&duploObject)
	log.Printf("[TRACE] %s 2 ********: %s", apiStr, jsonData)

	//
	rb, err := json.Marshal(duploObject)
	if err != nil {
		log.Printf("[TRACE] %s 3 ********: %s: %s", apiStr, api, err.Error())
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

		duploObject := DuploService{}
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			return nil, err
		}
		log.Printf("[TRACE] %s 8 ********: ", apiStr)
		c.DuploServiceSetIDFromCloud(&duploObject, d)
		return nil, nil
	}
	return nil, fmt.Errorf("ERROR: in create %s body: %s", apiStr, body)
}

// DuploServiceDelete deletes a service via the Duplo API.
func (c *Client) DuploServiceDelete(d *schema.ResourceData, m interface{}) (*DuploService, error) {

	var api = d.Id()
	url := c.DuploServiceURL(d)
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
