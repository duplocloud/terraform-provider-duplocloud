package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ucarion/jcs"
)

// ecsTaskDefinitionSchema returns a Terraform resource schema for an ECS Task Definition
func ecsTaskDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the task definition will be created in.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"family": {
			Description: "The name of the task definition to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"revision": {
			Description: "The current revision of the task definition.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the task definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The status of the task definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"container_definitions": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			StateFunc: func(v interface{}) string {
				// Sort the lists of environment variables as they are serialized to state, so we won't get
				// spurious reorderings in plans (diff is suppressed if the environment variables haven't changed,
				// but they still show in the plan if some other property changes).
				log.Printf("[TRACE] container_definitions.StateFunc: <= %s", v.(string))
				defns, _ := expandEcsContainerDefinitions(v.(string))
				for i := range defns {
					reorderEcsEnvironmentVariables(defns[i].(map[string]interface{}))
				}
				json, err := jcs.Format(defns)
				log.Printf("[TRACE] container_definitions.StateFunc: => %s (error: %s)", json, err)
				return json
			},
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				networkMode, ok := d.GetOk("network_mode")
				isAWSVPC := ok && networkMode.(string) == "awsvpc"
				equal, _ := ecsContainerDefinitionsAreEquivalent(old, new, isAWSVPC)
				return equal
			},
			ValidateFunc: func(v interface{}, k string) ([]string, []error) {
				return validateJsonObjectArray("Duplo ECS Task Definition container_definitions", v.(string))
			},
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
			Computed: true,
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
			Elem:     KeyValueSchema(),
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

// SCHEMA for resource crud
func resourceDuploEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecs_task_definition` manages a Amazon ECS task definition in Duplo.",

		ReadContext:   resourceDuploEcsTaskDefinitionRead,
		CreateContext: resourceDuploEcsTaskDefinitionCreate,
		DeleteContext: resourceDuploEcsTaskDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: ecsTaskDefinitionSchema(),
	}
}

/// READ resource
func resourceDuploEcsTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, arn, err := parseEcsTaskDefIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead(%s, %s): start", tenantID, arn)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.EcsTaskDefinitionGet(tenantID, arn)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.Arn == "" {
		d.SetId("")
		return nil
	}

	// Convert the object into Terraform resource data
	flattenEcsTaskDefinition(rp, d)

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead(%s, %s): end", tenantID, arn)
	return nil
}

/// CREATE resource
func resourceDuploEcsTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandEcsTaskDefinition(d)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate(%s, %s): start", tenantID, rq.Family)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	arn, err := c.EcsTaskDefinitionCreate(tenantID, rq)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", tenantID, arn))

	diags := resourceDuploEcsTaskDefinitionRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate(%s, %s): end", tenantID, rq.Family)
	return diags
}

/// DELETE resource
func resourceDuploEcsTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** start")
	// FIXME: NO-OP
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** end")
	return nil
}

func expandEcsTaskDefinition(d *schema.ResourceData) (*duplosdk.DuploEcsTaskDef, error) {
	// First, convert things into simple scalars
	duplo := duplosdk.DuploEcsTaskDef{
		Family:      d.Get("family").(string),
		CPU:         d.Get("cpu").(string),
		Memory:      d.Get("memory").(string),
		IpcMode:     d.Get("ipc_mode").(string),
		PidMode:     d.Get("pid_mode").(string),
		NetworkMode: &duplosdk.DuploStringValue{Value: d.Get("network_mode").(string)},
	}

	// Next, convert sets into lists
	rcs := d.Get("requires_compatibilities").(*schema.Set)
	dorcs := make([]string, 0, rcs.Len())
	for _, rc := range rcs.List() {
		dorcs = append(dorcs, rc.(string))
	}
	duplo.RequiresCompatibilities = dorcs

	// Next, convert things from embedded JSON
	condefs := d.Get("container_definitions").(string)
	err := json.Unmarshal([]byte(condefs), &duplo.ContainerDefinitions)
	if err != nil {
		log.Printf("[TRACE] expandEcsTaskDefinition: failed to parse container_definitions: %s", condefs)
		return nil, err
	}
	vols := d.Get("volumes").(string)
	err2 := json.Unmarshal([]byte(vols), &duplo.Volumes)
	if err2 != nil {
		log.Printf("[TRACE] expandEcsTaskDefinition: failed to parse volumes: %s", condefs)
		return nil, err
	}

	// Next, convert things into structured data.
	duplo.Tags = keyValueFromState("tags", d)
	duplo.PlacementConstraints = ecsPlacementConstraintsFromState(d)
	duplo.ProxyConfiguration = ecsProxyConfigFromState(d)
	duplo.InferenceAccelerators = ecsInferenceAcceleratorsFromState(d)
	duplo.RequiresAttributes = ecsRequiresAttributesFromState(d)

	return &duplo, nil
}

func flattenEcsTaskDefinition(duplo *duplosdk.DuploEcsTaskDef, d *schema.ResourceData) {

	// First, convert things into simple scalars
	d.Set("tenant_id", duplo.TenantID)
	d.Set("family", duplo.Family)
	d.Set("revision", duplo.Revision)
	d.Set("arn", duplo.Arn)
	d.Set("cpu", duplo.CPU)
	d.Set("task_role_arn", duplo.TaskRoleArn)
	d.Set("execution_role_arn", duplo.ExecutionRoleArn)
	d.Set("memory", duplo.Memory)
	d.Set("ipc_mode", duplo.IpcMode)
	d.Set("pid_mode", duplo.PidMode)
	d.Set("requires_compatibilities", duplo.RequiresCompatibilities)
	if duplo.NetworkMode != nil {
		d.Set("network_mode", duplo.NetworkMode.Value)
	}
	if duplo.Status != nil {
		d.Set("status", duplo.Status.Value)
	}

	// Next, convert things into embedded JSON
	toJsonStringState("container_definitions", duplo.ContainerDefinitions, d)
	toJsonStringState("volumes", duplo.Volumes, d)

	// Next, convert things into structured data.
	d.Set("placement_constraints", ecsPlacementConstraintsToState(duplo.PlacementConstraints))
	d.Set("proxy_configuration", ecsProxyConfigToState(duplo.ProxyConfiguration))
	d.Set("inference_accelerator", ecsInferenceAcceleratorsToState(duplo.InferenceAccelerators))
	d.Set("requires_attributes", ecsRequiresAttributesToState(duplo.RequiresAttributes))
	d.Set("tags", keyValueToState("tags", duplo.Tags))
}

// An internal function that compares two ECS container definitions to see if they are equivalent.
func ecsContainerDefinitionsAreEquivalent(old, new string, isAWSVPC bool) (bool, error) {

	oldCanonical, err := canonicalizeEcsContainerDefinitionsJson(old, isAWSVPC)
	if err != nil {
		return false, err
	}

	newCanonical, err := canonicalizeEcsContainerDefinitionsJson(new, isAWSVPC)
	if err != nil {
		return false, err
	}

	equal := oldCanonical == newCanonical
	if !equal {
		log.Printf("[DEBUG] Canonical definitions are not equal.\nFirst: %s\nSecond: %s\n", oldCanonical, newCanonical)
	}
	return equal, nil
}

// Internal function to expand container definitions JSON into an array of maps.
func expandEcsContainerDefinitions(encoded string) (defn []interface{}, err error) {
	err = json.Unmarshal([]byte(encoded), &defn)
	log.Printf("[DEBUG] Expanded container definition: %v", defn)
	return
}

// Internal function to unmarshal, reduce, then canonicalize container definitions JSON.
func canonicalizeEcsContainerDefinitionsJson(encoded string, isAWSVPC bool) (string, error) {
	var defns []interface{}

	// Unmarshall, reduce, then canonicalize.
	err := json.Unmarshal([]byte(encoded), &defns)
	if err != nil {
		return encoded, err
	}
	for i := range defns {
		err = reduceContainerDefinition(defns[i].(map[string]interface{}), isAWSVPC)
		if err != nil {
			return encoded, err
		}
	}
	canonical, err := jcs.Format(defns)
	if err != nil {
		return encoded, err
	}

	return canonical, nil
}

// Internal function used to re-order environment variables for an ECS task definition.
func reorderEcsEnvironmentVariables(defn map[string]interface{}) {

	// Re-order environment variables to a canonical order.
	if v, ok := defn["Environment"]; ok && v != nil {
		if env, ok := v.([]interface{}); ok && env != nil {
			sort.Slice(env, func(i, j int) bool {

				// Get both maps, ensure we are using upper camel-case.
				mi := env[i].(map[string]interface{})
				mj := env[j].(map[string]interface{})
				makeMapUpperCamelCase(mi)
				makeMapUpperCamelCase(mj)

				// Get both name keys, fall back on an empty string.
				si := ""
				sj := ""
				if v, ok = mi["Name"]; ok && !isInterfaceNil(v) {
					si = v.(string)
				}
				if v, ok = mj["Name"]; ok && !isInterfaceNil(v) {
					sj = v.(string)
				}

				// Compare the two.
				return si < sj
			})
		}
	}
}

// Internal function used to reduce a container definition down to a partially "canoncalized" form.
//
// See: https://github.com/hashicorp/terraform-provider-aws/blob/7141d1c315dc0c221979f0a4f8855a13b545dbaf/aws/ecs_task_definition_equivalency.go#L58
func reduceContainerDefinition(defn map[string]interface{}, isAWSVPC bool) error {

	// Ensure we are using upper-camel case.
	makeMapUpperCamelCase(defn)

	// Reorder the environment variables.
	reorderEcsEnvironmentVariables(defn)

	// Handle fields that have defaults.
	if v, ok := defn["Cpu"]; ok {
		if v2, ok := v.(int); ok && v2 == 0 {
			defn["Cpu"] = nil
		}
	}
	if v, ok := defn["Essential"]; !ok || isInterfaceNil(v) {
		defn["Essential"] = true
	}

	// Handle port mappings array
	if v, ok := defn["PortMappings"]; ok && v != nil {
		if pmi, ok := v.([]interface{}); ok {

			// Handle each port mapping
			for i := range pmi {
				if pms, ok := pmi[i].(map[string]interface{}); ok {

					// Ensure we are using upper-camel case.
					makeMapUpperCamelCase(pms)

					// Handle protocol == "tcp"
					if protocol, ok := pms["Protocol"]; ok {
						if v2, ok := protocol.(string); ok && v2 == "tcp" {
							pms["Protocol"] = nil
						}
					}

					// Handle HostPort
					if hostPort, ok := pms["HostPort"]; ok {

						// Handle HostPort == 0 or blank
						if v2, ok := hostPort.(int); ok && v2 == 0 {
							pms["HostPort"] = nil
						} else if v2, ok := hostPort.(string); ok && (v2 == "0" || v2 == "") {
							pms["HostPort"] = nil
						}
					}

					// Handle HostPort == null when using AWSVPC networking
					if isAWSVPC && pms["HostPort"] == nil {
						pms["HostPort"] = pms["ContainerPort"]
					}
				}
			}
		}
	}

	// Set all empty slices to nil.
	for k, v := range defn {
		if isInterfaceEmptySlice(v) {
			defn[k] = nil
		}
	}

	return nil
}

func ecsPlacementConstraintsToState(pcs *[]duplosdk.DuploEcsTaskDefPlacementConstraint) []map[string]interface{} {
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

func ecsPlacementConstraintsFromState(d *schema.ResourceData) *[]duplosdk.DuploEcsTaskDefPlacementConstraint {
	spcs := d.Get("placement_constraints").(*schema.Set)
	if spcs == nil || spcs.Len() == 0 {
		return nil
	}
	pcs := spcs.List()

	log.Printf("[TRACE] ecsPlacementConstraintsFromState ********: have data")

	duplo := make([]duplosdk.DuploEcsTaskDefPlacementConstraint, 0, len(pcs))
	for _, pc := range pcs {
		duplo = append(duplo, duplosdk.DuploEcsTaskDefPlacementConstraint{
			Type:       pc.(map[string]string)["type"],
			Expression: pc.(map[string]string)["expression"],
		})
	}

	return &duplo
}

func ecsProxyConfigToState(pc *duplosdk.DuploEcsTaskDefProxyConfig) []map[string]interface{} {
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

func ecsProxyConfigFromState(d *schema.ResourceData) *duplosdk.DuploEcsTaskDefProxyConfig {
	lpc := d.Get("proxy_configuration").([]interface{})
	if len(lpc) == 0 {
		return nil
	}
	pc := lpc[0].(map[string]interface{})

	log.Printf("[TRACE] ecsProxyConfigFromState ********: have data")

	props := pc["properties"].(map[string]interface{})
	nvs := make([]duplosdk.DuploNameStringValue, 0, len(props))
	for prop := range props {
		nvs = append(nvs, duplosdk.DuploNameStringValue{Name: prop, Value: props[prop].(string)})
	}

	return &duplosdk.DuploEcsTaskDefProxyConfig{
		ContainerName: pc["container_name"].(string),
		Properties:    &nvs,
		Type:          pc["type"].(string),
	}
}

func ecsInferenceAcceleratorsToState(ias *[]duplosdk.DuploEcsTaskDefInferenceAccelerator) []map[string]interface{} {
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

func ecsInferenceAcceleratorsFromState(d *schema.ResourceData) *[]duplosdk.DuploEcsTaskDefInferenceAccelerator {
	var ary []duplosdk.DuploEcsTaskDefInferenceAccelerator

	sias := d.Get("inference_accelerator").(*schema.Set)
	if sias == nil || sias.Len() == 0 {
		return nil
	}
	ias := sias.List()
	if len(ias) > 0 {
		log.Printf("[TRACE] ecsInferenceAcceleratorsFromState ********: have data")
		for _, ia := range ias {
			ary = append(ary, duplosdk.DuploEcsTaskDefInferenceAccelerator{
				DeviceName: ia.(map[string]string)["device_name"],
				DeviceType: ia.(map[string]string)["device_type"],
			})
		}
	}

	return &ary
}

func ecsRequiresAttributesToState(nms *[]duplosdk.DuploName) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(*nms))
	for _, nm := range *nms {
		results = append(results, map[string]interface{}{"name": nm.Name})
	}
	return results
}

func ecsRequiresAttributesFromState(d *schema.ResourceData) *[]duplosdk.DuploName {
	var ary []duplosdk.DuploName

	sras := d.Get("requires_attributes").(*schema.Set)
	if sras == nil || sras.Len() == 0 {
		return nil
	}
	ras := sras.List()
	if len(ras) > 0 {
		log.Printf("[TRACE] ecsRequiresAttributesFromState ********: have data")
		for _, ra := range ras {
			if ram, ok := ra.(map[string]interface{}); ok {
				if v, ok := ram["name"]; ok && v != nil {
					ary = append(ary, duplosdk.DuploName{Name: v.(string)})
				}
			}
		}
	}

	return &ary
}

func parseEcsTaskDefIdParts(id string) (tenantID, arn string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, arn = idParts[1], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
