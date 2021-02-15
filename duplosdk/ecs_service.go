package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploEcsServiceLbConfig is a Duplo SDK object that represents load balancer configuration for an ECS service
type DuploEcsServiceLbConfig struct {
	ReplicationControllerName string `json:"ReplicationControllerName"`
	Port                      string `json:"Port,omitempty"`
	Protocol                  string `json:"Protocol,omitempty"`
	ExternalPort              int    `json:"ExternalPort,omitempty"`
	IsInternal                bool   `json:"IsInternal,omitempty"`
	HealthCheckURL            string `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string `json:"CertificateArn,omitempty"`
	LbType                    int    `json:"LbType,omitempty"`
}

// DuploEcsService is a Duplo SDK object that represents an ECS service
type DuploEcsService struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name                          string                     `json:"Name"`
	TaskDefinition                string                     `json:"TaskDefinition,omitempty"`
	Replicas                      int                        `json:"Replicas,omitempty"`
	HealthCheckGracePeriodSeconds int                        `json:"HealthCheckGracePeriodSeconds,omitempty"`
	OldTaskDefinitionBufferSize   int                        `json:"OldTaskDefinitionBufferSize,omitempty"`
	IsTargetGroupOnly             bool                       `json:"IsTargetGroupOnly,omitempty"`
	DNSPrfx                       string                     `json:"DnsPrfx,omitempty"`
	LBConfigurations              *[]DuploEcsServiceLbConfig `json:"LBConfigurations,omitempty"`
}

// DuploEcsServiceSchema returns a Terraform resource schema for an ECS Service
func DuploEcsServiceSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"task_definition": {
			Type:     schema.TypeString,
			Required: true,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"health_check_grace_period_seconds": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"old_task_definition_buffer_size": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  10,
		},
		"is_target_group_only": {
			Type:     schema.TypeBool,
			ForceNew: true,
			Optional: true,
			Required: false,
			Default:  false,
		},
		"dns_prfx": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"load_balancer": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"replication_controller_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"lb_type": {
						Type:     schema.TypeInt,
						Optional: false,
						Required: true,
					},
					"port": {
						Type:     schema.TypeString,
						Optional: false,
						Required: true,
					},
					"protocol": {
						Type:     schema.TypeString,
						Optional: false,
						Required: true,
					},
					"external_port": {
						Type:     schema.TypeInt,
						Optional: false,
						Required: true,
					},
					"is_internal": {
						Type:     schema.TypeBool,
						Optional: true,
						Required: false,
						Default:  false,
					},
					"health_check_url": {
						Type:     schema.TypeString,
						Optional: true,
						Required: false,
					},
					"certificate_arn": {
						Type:     schema.TypeString,
						Optional: true,
						Required: false,
					},
				},
			},
		},
	}
}

/*************************************************
 * API CALLS to duplo
 */

// EcsServiceCreate creates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, error) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, false)
}

// EcsServiceUpdate updates an ECS service via the Duplo API.
func (c *Client) EcsServiceUpdate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, error) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, true)
}

// EcsServiceCreateOrUpdate creates or updates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreateOrUpdate(tenantID string, duploObject *DuploEcsService, updating bool) (*DuploEcsService, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}
	rqBody, err := json.Marshal(&duploObject)
	if err != nil {
		log.Printf("[TRACE] EcsServiceCreateOrUpdate 1 JSON gen : %s", err.Error())
		return nil, err
	}
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2", c.HostURL, tenantID)
	log.Printf("[TRACE] EcsServiceCreate 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest(verb, url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] EcsServiceCreateOrUpdate 3 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] EcsServiceCreateOrUpdate 4 HTTP %s : %s", verb, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] EcsServiceCreateOrUpdate 4 HTTP RESPONSE : %s", bodyString)

	// Handle the response
	rpObject := DuploEcsService{}
	if bodyString == "" {
		log.Printf("[TRACE] EcsServiceCreateOrUpdate 5 NO RESULT : %s", bodyString)
		return nil, err
	}
	err = json.Unmarshal(body, &rpObject)
	if err != nil {
		log.Printf("[TRACE] EcsServiceCreateOrUpdate 6 JSON parse : %s", err.Error())
		return nil, err
	}
	return &rpObject, nil
}

// EcsServiceDelete deletes an ECS service via the Duplo API.
func (c *Client) EcsServiceDelete(id string) (*DuploEcsService, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2/%s", c.HostURL, tenantID, name)
	log.Printf("[TRACE] EcsServiceGet 1 : %s", url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("[TRACE] EcsServiceGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	bodyString := string(body)
	if err != nil {
		log.Printf("[TRACE] EcsServiceGet 3 HTTP DELETE : %s", err.Error())
		return nil, err
	}
	log.Printf("[TRACE] EcsServiceGet 4 HTTP RESPONSE : %s", bodyString)

	// Parse the response into a duplo object
	duploObject := DuploEcsService{}
	if bodyString == "" {
		// tolerate an empty response from DELETE
		duploObject.Name = name
	} else {
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] EcsServiceGet 5 JSON PARSE : %s", bodyString)
			return nil, err
		}
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

// EcsServiceGet retrieves an ECS service via the Duplo API.
func (c *Client) EcsServiceGet(id string) (*DuploEcsService, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2/%s", c.HostURL, tenantID, name)
	log.Printf("[TRACE] EcsServiceGet 1 : %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[TRACE] EcsServiceGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] EcsServiceGet 3 HTTP GET : %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] EcsServiceGet 4 HTTP RESPONSE : %s", bodyString)

	// Parse the response into a duplo object, detecting a missing object
	if bodyString == "null" {
		return nil, nil
	}
	duploObject := DuploEcsService{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] EcsServiceGet 5 JSON PARSE : %s", bodyString)
		return nil, err
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// EcsServiceFromState converts resource data respresenting an ECS Service to a Duplo SDK object.
func EcsServiceFromState(d *schema.ResourceData) (*DuploEcsService, error) {
	duploObject := new(DuploEcsService)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.TaskDefinition = d.Get("task_definition").(string)
	duploObject.Replicas = d.Get("replicas").(int)
	duploObject.HealthCheckGracePeriodSeconds = d.Get("health_check_grace_period_seconds").(int)
	duploObject.OldTaskDefinitionBufferSize = d.Get("old_task_definition_buffer_size").(int)
	duploObject.IsTargetGroupOnly = d.Get("is_target_group_only").(bool)
	duploObject.DNSPrfx = d.Get("dns_prfx").(string)

	// Next, convert things into structured data.
	duploObject.LBConfigurations = ecsLoadBalancersFromState(d)

	return duploObject, nil
}

// EcsServiceToState converts a Duplo SDK object respresenting an ECS Service to terraform resource data.
func EcsServiceToState(duploObject *DuploEcsService, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-EcsServiceToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["name"] = duploObject.Name
	jo["task_definition"] = duploObject.TaskDefinition
	jo["replicas"] = duploObject.Replicas
	jo["health_check_grace_period_seconds"] = duploObject.HealthCheckGracePeriodSeconds
	jo["old_task_definition_buffer_size"] = duploObject.OldTaskDefinitionBufferSize
	jo["is_target_group_only"] = duploObject.IsTargetGroupOnly
	jo["dns_prfx"] = duploObject.DNSPrfx

	// Next, convert things into structured data.
	jo["load_balancer"] = ecsLoadBalancersToState(duploObject.Name, duploObject.LBConfigurations)

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-EcsServiceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}

func ecsLoadBalancersToState(name string, lbcs *[]DuploEcsServiceLbConfig) []map[string]interface{} {
	if lbcs == nil {
		return nil
	}

	var ary []map[string]interface{}

	for _, lbc := range *lbcs {
		jo := make(map[string]interface{})
		jo["replication_controller_name"] = name
		jo["lb_type"] = lbc.LbType
		jo["port"] = lbc.Port
		jo["protocol"] = lbc.Protocol
		jo["external_port"] = lbc.ExternalPort
		jo["is_internal"] = lbc.IsInternal
		jo["health_check_url"] = lbc.HealthCheckURL
		jo["certificate_arn"] = lbc.CertificateArn
		ary = append(ary, jo)
	}

	return ary
}

func ecsLoadBalancersFromState(d *schema.ResourceData) *[]DuploEcsServiceLbConfig {
	var ary []DuploEcsServiceLbConfig

	slb := d.Get("load_balancer").(*schema.Set)
	if slb == nil {
		return nil
	}

	log.Printf("[TRACE] ecsLoadBalancersFromState ********: have data")

	for _, _lb := range slb.List() {
		lb := _lb.(map[string]interface{})
		ary = append(ary, DuploEcsServiceLbConfig{
			ReplicationControllerName: lb["replication_controller_name"].(string),
			LbType:                    lb["lb_type"].(int),
			Port:                      lb["port"].(string),
			Protocol:                  lb["protocol"].(string),
			ExternalPort:              lb["external_port"].(int),
			IsInternal:                lb["is_internal"].(bool),
			HealthCheckURL:            lb["health_check_url"].(string),
			CertificateArn:            lb["certificate_arn"].(string),
		})
	}

	return &ary
}
