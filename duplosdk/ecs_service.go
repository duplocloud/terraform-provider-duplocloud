package duplosdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
 	"encoding/json"
 	"fmt"
	"log"
 	"net/http"
 	"strings"
)

type DuploEcsServiceLbConfig struct {
	ReplicationControllerName string                 `json:"ReplicationControllerName"`
	Port                      string                 `json:"Port,omitempty"`
	Protocol                  string                 `json:"Protocol,omitempty"`
	ExternalPort              int                    `json:"ExternalPort,omitempty"`
	IsInternal                bool                   `json:"IsInternal,omitempty"`
	HealthCheckUrl            string                 `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string                 `json:"CertificateArn,omitempty"`
	LbType                    int                    `json:"LbType,omitempty"`
}

type DuploEcsService struct {
    // NOTE: The TenantId field does not come from the backend - we synthesize it
    TenantId                string                   `json:"-",omitempty`

	Name                    string                   `json:"Name"`
	TaskDefinition          string                   `json:"TaskDefinition,omitempty"`
	Replicas                int                      `json:"Replicas,omitempty"`
	HealthCheckGracePeriodSeconds int                `json:"HealthCheckGracePeriodSeconds,omitempty"`
	OldTaskDefinitionBufferSize int                  `json:"OldTaskDefinitionBufferSize,omitempty"`
	DnsPrfx                 string                   `json:"DnsPrfx,omitempty"`
	LBConfigurations        *[]DuploEcsServiceLbConfig `json:"LBConfigurations,omitempty"`
}

/////------ schema ------////
func DuploEcsServiceSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"task_definition": &schema.Schema{
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
			Computed: true,
        },
        "dns_prfx": &schema.Schema{
            Type:     schema.TypeString,
            Optional: true,
            Required: false,
        },
        "load_balancer": {
            Type:     schema.TypeSet,
            Optional: true,
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
func (c *Client) EcsServiceCreate(tenantId string, duploObject *DuploEcsService) (*DuploEcsService, error) {
    return c.EcsServiceCreateOrUpdate(tenantId, duploObject, false)
}

func (c *Client) EcsServiceUpdate(tenantId string, duploObject *DuploEcsService) (*DuploEcsService, error) {
    return c.EcsServiceCreateOrUpdate(tenantId, duploObject, true)
}

func (c *Client) EcsServiceCreateOrUpdate(tenantId string, duploObject *DuploEcsService, updating bool) (*DuploEcsService, error) {

    // Build the request
    verb := "POST"
    if (updating) {
        verb = "PUT"
    }
    rqBody, err := json.Marshal(&duploObject)
    if err != nil {
        log.Printf("[TRACE] EcsServiceCreateOrUpdate 1 JSON gen : %s", err.Error())
		return nil, err
    }
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2", c.HostURL, tenantId)
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

func (c *Client) EcsServiceDelete(id string) (*DuploEcsService, error) {
    idParts := strings.SplitN(id, "/", 5)
    tenantId := idParts[2]
    name := idParts[4]

    // Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2/%s", c.HostURL, tenantId, name)
	log.Printf("[TRACE] EcsServiceGet 1 : %s", url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
        log.Printf("[TRACE] EcsServiceGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

    // Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
        log.Printf("[TRACE] EcsServiceGet 3 HTTP DELETE : %s", err.Error())
		return nil, err
	}
    log.Printf("[TRACE] EcsServiceGet 4 HTTP RESPONSE : %s", string(body))

    // Parse the response into a duplo object
	duploObject := DuploEcsService{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
        log.Printf("[TRACE] EcsServiceGet 5 JSON PARSE : %s", string(body))
		return nil, err
	}

    // Fill in the tenant ID and return the object
    duploObject.TenantId = tenantId
    return &duploObject, nil
}

func (c *Client) EcsServiceGet(id string) (*DuploEcsService, error) {
    idParts := strings.SplitN(id, "/", 5)
    tenantId := idParts[2]
    name := idParts[4]

    // Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/EcsServiceApiV2/%s", c.HostURL, tenantId, name)
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
    log.Printf("[TRACE] EcsServiceGet 4 HTTP RESPONSE : %s", string(body))

    // Parse the response into a duplo object
	duploObject := DuploEcsService{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
        log.Printf("[TRACE] EcsServiceGet 5 JSON PARSE : %s", string(body))
		return nil, err
	}

    // Fill in the tenant ID and return the object
    duploObject.TenantId = tenantId
    return &duploObject, nil
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */
func EcsServiceFromState(d *schema.ResourceData) (*DuploEcsService, error) {
    duploObject := new(DuploEcsService)

    // First, convert things into simple scalars
    duploObject.Name = d.Get("name").(string)
    duploObject.TaskDefinition = d.Get("task_definition").(string)
    duploObject.Replicas = d.Get("replicas").(int)
    duploObject.HealthCheckGracePeriodSeconds = d.Get("health_check_grace_period_seconds").(int)
    duploObject.OldTaskDefinitionBufferSize = d.Get("old_task_definition_buffer_size").(int)
    duploObject.DnsPrfx = d.Get("dns_prfx").(string)

    // Next, convert things into structured data.
    duploObject.LBConfigurations = ecsLoadBalancersFromState(d)

    return duploObject, nil
}

func EcsServiceToState(duploObject *DuploEcsService, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
	    return nil
    }
    jsonData, _ := json.Marshal(duploObject)
    log.Printf("[TRACE] duplo-EcsServiceToState ******** 1: INPUT <= %s ", jsonData)

    jo := make(map[string]interface{})

    // First, convert things into simple scalars
    jo["tenant_id"] = duploObject.TenantId
    jo["name"] = duploObject.Name
    jo["task_definition"] = duploObject.TaskDefinition
    jo["replicas"] = duploObject.Replicas
    jo["health_check_grace_period_seconds"] = duploObject.HealthCheckGracePeriodSeconds
    jo["old_task_definition_buffer_size"] = duploObject.OldTaskDefinitionBufferSize
    jo["dns_prfx"] = duploObject.DnsPrfx

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
        jo["health_check_url"] = lbc.HealthCheckUrl
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
            HealthCheckUrl:            lb["health_check_url"].(string),
            CertificateArn:            lb["certificate_arn"].(string),
        })
	}

	return &ary
}
