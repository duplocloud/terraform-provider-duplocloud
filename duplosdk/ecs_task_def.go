package duplosdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploEcsTaskDefPlacementConstraint represents an ECS placement constraint in the Duplo SDK
type DuploEcsTaskDefPlacementConstraint struct {
	Type       string `json:"Type"`
	Expression string `json:"Expression"`
}

// DuploEcsTaskDefProxyConfig represents an ECS proxy configuration in the Duplo SDK
type DuploEcsTaskDefProxyConfig struct {
	ContainerName string                  `json:"ContainerName"`
	Properties    *[]DuploNameStringValue `json:"Properties"`
	Type          string                  `json:"Type"`
}

// DuploEcsTaskDefInferenceAccelerator represents an inference accelerator in the Duplo SDK
type DuploEcsTaskDefInferenceAccelerator struct {
	DeviceName string `json:"DeviceName"`
	DeviceType string `json:"DeviceType"`
}

// DuploEcsTaskDef represents an ECS task definition in the Duplo SDK
type DuploEcsTaskDef struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Family                  string                                 `json:"Family"`
	Revision                int                                    `json:"Revision,omitempty"`
	Arn                     string                                 `json:"TaskDefinitionArn,omitempty"`
	ContainerDefinitions    []map[string]interface{}               `json:"ContainerDefinitions,omitempty"`
	CPU                     string                                 `json:"Cpu,omitempty"`
	TaskRoleArn             string                                 `json:"TaskRoleArn,omitempty"`
	ExecutionRoleArn        string                                 `json:"ExecutionRoleArn,omitempty"`
	Memory                  string                                 `json:"Memory,omitempty"`
	IpcMode                 string                                 `json:"IpcMode,omitempty"`
	PidMode                 string                                 `json:"PidMode,omitempty"`
	NetworkMode             *DuploStringValue                      `json:"NetworkMode,omitempty"`
	PlacementConstraints    *[]DuploEcsTaskDefPlacementConstraint  `json:"PlacementConstraints,omitempty"`
	ProxyConfiguration      *DuploEcsTaskDefProxyConfig            `json:"ProxyConfiguration,omitempty"`
	RequiresAttributes      *[]DuploName                           `json:"RequiresAttributes,omitempty"`
	RequiresCompatibilities []string                               `json:"RequiresCompatibilities,omitempty"`
	Tags                    *[]DuploKeyStringValue                 `json:"Tags,omitempty"`
	InferenceAccelerators   *[]DuploEcsTaskDefInferenceAccelerator `json:"InferenceAccelerators,omitempty"`
	Status                  *DuploStringValue                      `json:"Status,omitempty"`
	Volumes                 []map[string]interface{}               `json:"Volumes,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcsTaskDefinitionCreate creates an ECS task definition via the Duplo API.
func (c *Client) EcsTaskDefinitionCreate(tenantID string, duploObject *DuploEcsTaskDef) (string, error) {

	// Build the request
	rqBody, err := json.Marshal(&duploObject)
	if err != nil {
		log.Printf("[TRACE] EcsTaskDefinitionCreate 1 JSON gen : %s", err.Error())
		return "", err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/UpdateEcsTaskDefinition", c.HostURL, tenantID)
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

// EcsTaskDefinitionGet retrieves an ECS task definition via the Duplo API.
func (c *Client) EcsTaskDefinitionGet(id string) (*DuploEcsTaskDef, error) {
	idParts := strings.SplitN(id, "/", 4)
	tenantID := idParts[1]
	arn := idParts[3]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/FindEcsTaskDefinition", c.HostURL, tenantID)
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
		return nil, fmt.Errorf("ECS task definition %s not found in tenant %s", arn, tenantID)
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// EcsTaskDefFromState converts resource data respresenting an ECS Task Definition to a Duplo SDK object.
func EcsTaskDefFromState(d *schema.ResourceData) (*DuploEcsTaskDef, error) {
	duploObject := new(DuploEcsTaskDef)

	// First, convert things into simple scalars
	duploObject.Family = d.Get("family").(string)
	duploObject.CPU = d.Get("cpu").(string)
	duploObject.Memory = d.Get("memory").(string)
	duploObject.IpcMode = d.Get("ipc_mode").(string)
	duploObject.PidMode = d.Get("pid_mode").(string)
	duploObject.NetworkMode = &DuploStringValue{Value: d.Get("network_mode").(string)}

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
	duploObject.Tags = KeyValueFromState("tags", d)
	duploObject.PlacementConstraints = ecsPlacementConstraintsFromState(d)
	duploObject.ProxyConfiguration = ecsProxyConfigFromState(d)
	duploObject.InferenceAccelerators = ecsInferenceAcceleratorsFromState(d)
	duploObject.RequiresAttributes = ecsRequiresAttributesFromState(d)

	return duploObject, nil
}

// EcsTaskDefToState converts a Duplo SDK object respresenting an ECS Task Definition to terraform resource data.
func EcsTaskDefToState(duploObject *DuploEcsTaskDef, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-EcsTaskDefToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["family"] = duploObject.Family
	jo["revision"] = duploObject.Revision
	jo["arn"] = duploObject.Arn
	jo["cpu"] = duploObject.CPU
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
	jo["tags"] = KeyValueToState("tags", duploObject.Tags)

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
	nvs := make([]DuploNameStringValue, 0, len(props))
	for prop := range props {
		nvs = append(nvs, DuploNameStringValue{Name: prop, Value: props[prop].(string)})
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
