package duplosdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"
	"strings"
)

// Placement constraint holder
type DuploEcsTaskDefPlacementConstraint struct {
	Type       string `json:"Type"`
	Expression string `json:"Expression"`
}

// Proxy configuration holder
type DuploEcsTaskDefProxyConfig struct {
	ContainerName string            `json:"ContainerName"`
	Properties    *[]DuploNameValue `json:"Properties"`
	Type          string            `json:"Type"`
}

// Inference accelerator holder
type DuploEcsTaskDefInferenceAccelerator struct {
	DeviceName string `json:"DeviceName"`
	DeviceType string `json:"DeviceType"`
}

type DuploEcsTaskDef struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId string `json:"-",omitempty`

	Family                  string                                 `json:"Family"`
	Revision                int                                    `json:"Revision,omitempty"`
	Arn                     string                                 `json:"TaskDefinitionArn,omitempty"`
	ContainerDefinitions    []map[string]interface{}               `json:"ContainerDefinitions,omitempty"`
	Cpu                     string                                 `json:"Cpu,omitempty"`
	TaskRoleArn             string                                 `json:"TaskRoleArn,omitempty"`
	ExecutionRoleArn        string                                 `json:"ExecutionRoleArn,omitempty"`
	Memory                  string                                 `json:"Memory,omitempty"`
	IpcMode                 string                                 `json:"IpcMode,omitempty"`
	PidMode                 string                                 `json:"PidMode,omitempty"`
	NetworkMode             *DuploValue                            `json:"NetworkMode,omitempty"`
	PlacementConstraints    *[]DuploEcsTaskDefPlacementConstraint  `json:"PlacementConstraints,omitempty"`
	ProxyConfiguration      *DuploEcsTaskDefProxyConfig            `json:"ProxyConfiguration,omitempty"`
	RequiresAttributes      *[]DuploName                           `json:"RequiresAttributes,omitempty"`
	RequiresCompatibilities []string                               `json:"RequiresCompatibilities,omitempty"`
	Tags                    *[]DuploKeyValue                       `json:"Tags,omitempty"`
	InferenceAccelerators   *[]DuploEcsTaskDefInferenceAccelerator `json:"InferenceAccelerators,omitempty"`
	Status                  *DuploValue                            `json:"Status,omitempty"`
	Volumes                 []map[string]interface{}               `json:"Volumes,omitempty"`
}

/////------ schema ------////
func DuploEcsTaskDefinitionSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"family": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"revision": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"container_definitions": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"volumes": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  "[]",
		},
		"cpu": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"task_role_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"execution_role_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"network_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"bridge", "host", "awsvpc", "none"}, false),
			Default:      "awsvpc",
		},
		"placement_constraints": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			MaxItems: 10,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"memberOf"}, false),
					},
					"expression": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
					},
				},
			},
		},
		"requires_attributes": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						ForceNew: true,
						Required: true,
					},
				},
			},
		},
		"requires_compatibilities": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"ipc_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"host", "none", "task"}, false),
		},
		"pid_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"host", "task"}, false),
		},
		"proxy_configuration": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"container_name": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"properties": {
						Type:     schema.TypeMap,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Optional: true,
						ForceNew: true,
					},
					"type": {
						Type:         schema.TypeString,
						Default:      "APPMESH",
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringInSlice([]string{"APPMESH"}, false),
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
		"inference_accelerator": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"device_name": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"device_type": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
				},
			},
		},
	}
}

/*************************************************
 * API CALLS to duplo
 */
func (c *Client) EcsTaskDefinitionCreate(tenantId string, duploObject *DuploEcsTaskDef) (string, error) {

	// Build the request
	rqBody, err := json.Marshal(&duploObject)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 1 JSON gen : %s", err.Error())
		return "", err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/UpdateEcsTaskDefinition", c.HostURL, tenantId)
	log.Printf("[TRACE] EcsTaskDefinitionCreate 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 3 HTTP builder : %s", err.Error())
		return "", err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 4 HTTP POST : %s", err.Error())
		return "", err
	}
	bodyString := string(body)
	log.Printf("[TRACE] EcsTaskDefinitionCreate 4 HTTP RESPONSE : %s", bodyString)

	// Handle the response
	arn := ""
	if bodyString == "" {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 5 NO RESULT : %s", bodyString)
		return "", err
	}
	err = json.Unmarshal(body, &arn)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 6 JSON parse : %s", err.Error())
		return "", err
	}
	if arn == "" {
		return "", errors.New("API call returned null")
	}
	return arn, nil
}

func (c *Client) EcsTaskDefinitionGet(id string) (*DuploEcsTaskDef, error) {
	idParts := strings.SplitN(id, "/", 4)
	tenantId := idParts[1]
	arn := idParts[3]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/FindEcsTaskDefinition", c.HostURL, tenantId)
	rqBody := fmt.Sprintf("{\"Arn\":\"%s\"}", arn)
	log.Printf("[TRACE] EcsTaskDefinitionGet 1 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionGet 3 HTTP POST : %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] EcsTaskDefinitionGet 4 HTTP RESPONSE : %s", bodyString)

	// Parse the response into a duplo object, detecting a missing object
	if bodyString == "null" {
		return nil, nil
	}
	duploObject := DuploEcsTaskDef{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionGet 5 JSON PARSE : %s", bodyString)
		return nil, err
	}
	if duploObject.Arn == "" {
		return nil, fmt.Errorf("ECS task definition %s not found in tenant %s", arn, tenantId)
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantId = tenantId
	return &duploObject, nil
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */
func EcsTaskDefFromState(d *schema.ResourceData) (*DuploEcsTaskDef, error) {
	duploObject := new(DuploEcsTaskDef)

	// First, convert things into simple scalars
	duploObject.Family = d.Get("family").(string)
	duploObject.Cpu = d.Get("cpu").(string)
	duploObject.Memory = d.Get("memory").(string)
	duploObject.IpcMode = d.Get("ipc_mode").(string)
	duploObject.PidMode = d.Get("pid_mode").(string)
	duploObject.NetworkMode = &DuploValue{Value: d.Get("network_mode").(string)}

	// Next, convert sets into lists
	rcs := d.Get("requires_compatibilities").(*schema.Set)
	dorcs := make([]string, 0, rcs.Len())
	for _, rc := range rcs.List() {
		dorcs = append(dorcs, rc.(string))
	}
	duploObject.RequiresCompatibilities = dorcs

	// Next, convert things from embedded JSON
	condefs := d.Get("container_definitions").(string)
	err := json.Unmarshal([]byte(condefs), &duploObject.ContainerDefinitions)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefFromState 1 JSON PARSE container_definitions: %s", condefs)
		return nil, err
	}
	vols := d.Get("volumes").(string)
	err2 := json.Unmarshal([]byte(vols), &duploObject.Volumes)
	if err2 != nil {
		log.Printf("[TRACE] EcsTaskDefFromState 2 JSON PARSE container_definitions: %s", condefs)
		return nil, err
	}

	// Next, convert things into structured data.
	duploObject.Tags = duploKeyValueFromState("tags", d)
	duploObject.PlacementConstraints = ecsPlacementConstraintsFromState(d)
	duploObject.ProxyConfiguration = ecsProxyConfigFromState(d)
	duploObject.InferenceAccelerators = ecsInferenceAcceleratorsFromState(d)
	duploObject.RequiresAttributes = ecsRequiresAttributesFromState(d)

	return duploObject, nil
}

func EcsTaskDefToState(duploObject *DuploEcsTaskDef, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-EcsTaskDefToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantId
	jo["family"] = duploObject.Family
	jo["revision"] = duploObject.Revision
	jo["arn"] = duploObject.Arn
	jo["cpu"] = duploObject.Cpu
	jo["task_role_arn"] = duploObject.TaskRoleArn
	jo["execution_role_arn"] = duploObject.ExecutionRoleArn
	jo["memory"] = duploObject.Memory
	jo["ipc_mode"] = duploObject.IpcMode
	jo["pid_mode"] = duploObject.PidMode
	jo["requires_compatibilities"] = duploObject.RequiresCompatibilities
	if duploObject.NetworkMode != nil {
		jo["network_mode"] = duploObject.NetworkMode.Value
	}
	if duploObject.Status != nil {
		jo["status"] = duploObject.Status.Value
	}

	// Next, convert things into embedded JSON
	condefs, _ := json.Marshal(duploObject.ContainerDefinitions)
	jo["container_definitions"] = string(condefs)
	volumes, _ := json.Marshal(duploObject.Volumes)
	jo["volumes"] = string(volumes)

	// Next, convert things into structured data.
	jo["placement_constraints"] = ecsPlacementConstraintsToState(duploObject.PlacementConstraints)
	jo["proxy_configuration"] = ecsProxyConfigToState(duploObject.ProxyConfiguration)
	jo["inference_accelerator"] = ecsInferenceAcceleratorsToState(duploObject.InferenceAccelerators)
	jo["requires_attributes"] = ecsRequiresAttributesToState(duploObject.RequiresAttributes)
	jo["tags"] = duploKeyValueToState("tags", duploObject.Tags)

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-EcsTaskDefToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}

func ecsPlacementConstraintsToState(pcs *[]DuploEcsTaskDefPlacementConstraint) []map[string]interface{} {
	if len(*pcs) == 0 {
		return nil
	}

	results := make([]map[string]interface{}, 0)
	for _, pc := range *pcs {
		c := make(map[string]interface{})
		c["type"] = pc.Type
		c["expression"] = pc.Expression
		results = append(results, c)
	}
	return results
}

func ecsPlacementConstraintsFromState(d *schema.ResourceData) *[]DuploEcsTaskDefPlacementConstraint {
	spcs := d.Get("placement_constraints").(*schema.Set)
	if spcs == nil || spcs.Len() == 0 {
		return nil
	}
	pcs := spcs.List()

	log.Printf("[TRACE] ecsPlacementConstraintsFromState ********: have data")

	duplo := make([]DuploEcsTaskDefPlacementConstraint, 0, len(pcs))
	for _, pc := range pcs {
		duplo = append(duplo, DuploEcsTaskDefPlacementConstraint{
			Type:       pc.(map[string]string)["type"],
			Expression: pc.(map[string]string)["expression"],
		})
	}

	return &duplo
}

func ecsProxyConfigToState(pc *DuploEcsTaskDefProxyConfig) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	props := make(map[string]string)
	if pc.Properties != nil {
		for _, prop := range *pc.Properties {
			props[prop.Name] = prop.Value
		}
	}

	config := make(map[string]interface{})
	config["container_name"] = pc.ContainerName
	config["type"] = pc.Type
	config["properties"] = props

	return []map[string]interface{}{config}
}

func ecsProxyConfigFromState(d *schema.ResourceData) *DuploEcsTaskDefProxyConfig {
	lpc := d.Get("proxy_configuration").([]interface{})
	if lpc == nil || len(lpc) == 0 {
		return nil
	}
	pc := lpc[0].(map[string]interface{})

	log.Printf("[TRACE] ecsProxyConfigFromState ********: have data")

	props := pc["properties"].(map[string]interface{})
	nvs := make([]DuploNameValue, 0, len(props))
	for prop := range props {
		nvs = append(nvs, DuploNameValue{Name: prop, Value: props[prop].(string)})
	}

	return &DuploEcsTaskDefProxyConfig{
		ContainerName: pc["container_name"].(string),
		Properties:    &nvs,
		Type:          pc["type"].(string),
	}
}

func ecsInferenceAcceleratorsToState(ias *[]DuploEcsTaskDefInferenceAccelerator) []map[string]interface{} {
	if ias == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(*ias))
	for _, iAcc := range *ias {
		l := map[string]interface{}{
			"device_name": iAcc.DeviceName,
			"device_type": iAcc.DeviceType,
		}

		result = append(result, l)
	}
	return result
}

func ecsInferenceAcceleratorsFromState(d *schema.ResourceData) *[]DuploEcsTaskDefInferenceAccelerator {
	var ary []DuploEcsTaskDefInferenceAccelerator

	sias := d.Get("inference_accelerator").(*schema.Set)
	if sias == nil || sias.Len() == 0 {
		return nil
	}
	ias := sias.List()
	if len(ias) > 0 {
		log.Printf("[TRACE] ecsInferenceAcceleratorsFromState ********: have data")
		for _, ia := range ias {
			ary = append(ary, DuploEcsTaskDefInferenceAccelerator{
				DeviceName: ia.(map[string]string)["device_name"],
				DeviceType: ia.(map[string]string)["device_type"],
			})
		}
	}

	return &ary
}

func ecsRequiresAttributesToState(nms *[]DuploName) []string {
	if len(*nms) == 0 {
		return nil
	}
	results := make([]string, 0, len(*nms))
	for _, nm := range *nms {
		results = append(results, nm.Name)
	}
	return results
}

func ecsRequiresAttributesFromState(d *schema.ResourceData) *[]DuploName {
	var ary []DuploName

	sras := d.Get("requires_attributes").(*schema.Set)
	if sras == nil || sras.Len() == 0 {
		return nil
	}
	ras := sras.List()
	if len(ras) > 0 {
		log.Printf("[TRACE] ecsRequiresAttributesFromState ********: have data")
		for _, ra := range ras {
			ary = append(ary, DuploName{Name: ra.(string)})
		}
	}

	return &ary
}
