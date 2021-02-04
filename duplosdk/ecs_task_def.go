package duplosdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
 	"encoding/json"
 	"fmt"
	"log"
 	"net/http"
 	"strings"
)

// Placement constraint holder
type DuploEcsTaskDefPlacementConstraint struct {
	Type           string `json:"Type"`
	Expression     string `json:"Expression"`
}

// Proxy configuration holder
type DuploEcsTaskDefProxyConfig struct {
	ContainerName  string `json:"ContainerName"`
	Properties     *[]DuploNameValue `json:"Properties"`
	Type           string `json:"Type"`
}

// Inference accelerator holder
type DuploEcsTaskDefInferenceAccelerator struct {
	DeviceName     string `json:"DeviceName"`
	DeviceType     string `json:"DeviceType"`
}

type DuploEcsTaskDef struct {
    TenantId                string                   `json:"TenantId",omitempty`
	Family                  string                   `json:"Family"`
	Revision                int                      `json:"Revision"`
	Arn                     string                   `json:"TaskDefinitionArn"`
	ContainerDefinitions    []map[string]interface{} `json:"ContainerDefinitions,omitempty"`
	Cpu                     string                   `json:"Cpu,omitempty"`
	TaskRoleArn             string                   `json:"TaskRoleArn"`
	ExecutionRoleArn        string                   `json:"ExecutionRoleArn"`
	Memory                  string                   `json:"Memory,omitempty"`
	IpcMode                 string                   `json:"IpcMode,omitempty"`
	PidMode                 string                   `json:"PidMode,omitempty"`
	NetworkMode             *DuploValue              `json:"NetworkMode,omitempty"`
	PlacementConstraints    *[]DuploEcsTaskDefPlacementConstraint                 `json:"PlacementConstraints,omitempty"`
    ProxyConfiguration      *DuploEcsTaskDefProxyConfig  `json:"ProxyConfiguration,omitempty"`
	RequiresAttributes      *[]DuploName             `json:"RequiresAttributes,omitempty"`
	RequiresCompatibilities []string                 `json:"RequiresCompatibilities,omitempty"`
	Tags                    *[]DuploKeyValue         `json:"Tags,omitempty"`
	InferenceAcclerator     *[]DuploEcsTaskDefInferenceAccelerator         `json:"InferenceAcclerator,omitempty"`
	Status                  *DuploValue              `json:"Status,omitempty"`
	Volumes                 []map[string]interface{} `json:"Volumes,omitempty"`
}

/////------ schema ------////
func DuploEcsTaskDefinitionSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"tenant_id": &schema.Schema{
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
            Optional: true,
            ForceNew: true,
        },
        "cpu": {
            Type:     schema.TypeString,
            Optional: true,
            ForceNew: true,
        },
        "task_role_arn": {
            Type:         schema.TypeString,
            Optional:     true,
            ForceNew:     true,
        },
        "execution_role_arn": {
            Type:         schema.TypeString,
            Optional:     true,
            ForceNew:     true,
        },
        "memory": {
            Type:     schema.TypeString,
            Optional: true,
            ForceNew: true,
        },
        "network_mode": {
            Type:     schema.TypeString,
            Optional: true,
            Computed: true,
            ForceNew: true,
            ValidateFunc: validation.StringInSlice([]string{"bridge","host","awsvpc","none"}, false),
        },
        "placement_constraints": {
            Type:     schema.TypeSet,
            Optional: true,
            ForceNew: true,
            MaxItems: 10,
            Elem: &schema.Resource{
                Schema: map[string]*schema.Schema{
                    "type": {
                        Type:     schema.TypeString,
                        ForceNew: true,
                        Required: true,
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
            Type:     schema.TypeString,
            Optional: true,
            ForceNew: true,
            ValidateFunc: validation.StringInSlice([]string{"host","none","task"}, false),
        },
        "pid_mode": {
            Type:     schema.TypeString,
            Optional: true,
            ForceNew: true,
            ValidateFunc: validation.StringInSlice([]string{"host","task"}, false),
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
                        Type:     schema.TypeString,
                        Default:  "APPMESH",
                        Optional: true,
                        ForceNew: true,
                        ValidateFunc: validation.StringInSlice([]string{"APPMESH"}, false),
                    },
                },
            },
        },
		"tags": {
			Type:             schema.TypeList,
			Computed:         true,
			Required:         false,
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
func (c *Client) EcsTaskDefinitionGet(id string) (*DuploEcsTaskDef, error) {
    idParts := strings.SplitN(id, "/", 4)
    tenantId := idParts[1]
    arn := idParts[3]

    // Build the request
	url := fmt.Sprintf("%s/subscriptions/%s/FindEcsTaskDefinition", c.HostURL, tenantId)
	rqBody := fmt.Sprintf("{\"Arn\":\"%s\"}", arn)
	log.Printf("[TRACE] duploEcsTaskDefinitionGet 1 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
        log.Printf("[TRACE] duploEcsTaskDefinitionGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

    // Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
        log.Printf("[TRACE] duploEcsTaskDefinitionGet 3 HTTP POST : %s", err.Error())
		return nil, err
	}
    log.Printf("[TRACE] duploEcsTaskDefinitionGet 4 HTTP RESPONSE : %s", string(body))

    // Parse the response into a duplo object
	duploObject := DuploEcsTaskDef{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
        log.Printf("[TRACE] duploEcsTaskDefinitionGet 5 JSON PARSE : %s", string(body))
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
    jo["container_definitions"] = condefs
    volumes,_ := json.Marshal(duploObject.Volumes)
    jo["volumes"] = volumes

    // Next, convert things into structured data.
    jo["placement_constraints"] = ecsPlacementConstraintsToState(duploObject.PlacementConstraints)
    jo["proxy_configuration"] = ecsProxyConfigToState(duploObject.ProxyConfiguration)
    jo["inference_accelerator"] = ecsInferenceAcceleratorsToState(duploObject.InferenceAcclerator)
    jo["tags"] = duploKeyValueToState("tags", duploObject.Tags)
    jo["requires_attributes"] = ecsRequiresAttributesToState(duploObject.RequiresAttributes)

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

	return []map[string]interface{}{ config }
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
