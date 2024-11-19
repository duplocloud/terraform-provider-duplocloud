package duplocloud

import (
	"context"
	"encoding/json"
	"errors"
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
		"full_family_name": {
			Description: "The name of the task definition to create.",
			Type:        schema.TypeString,
			Computed:    true,
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
		"prevent_tf_destroy": {
			Description: "Prevent this resource to be deleted from terraform destroy. Default value is `true`.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
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
				isLogConfigProvided := isLogConfigProvided(d)
				log.Printf("[TRACE] Log configuration provided: => %v", isLogConfigProvided)
				equal, _ := ecsContainerDefinitionsAreEquivalent(old, new, isAWSVPC, isLogConfigProvided)
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
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"FARGATE"}, false),
			},
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
		"runtime_platform": {
			Description: "Configuration block for runtime_platform that containers in your task may use. Required on ecs tasks that are hosted on Fargate.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cpu_architecture": {
						Description:  "Valid values are 'X86_64','ARM64'",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice([]string{"X86_64", "ARM64"}, false),
					},
					"operating_system_family": {
						Description: "Valid values are <br>For FARGATE: 'LINUX','WINDOWS_SERVER_2019_FULL','WINDOWS_SERVER_2019_CORE','WINDOWS_SERVER_2022_FULL','WINDOWS_SERVER_2022_CORE'", // <br> For EC2 : 'LINUX','WINDOWS_SERVER_2022_CORE','WINDOWS_SERVER_2022_FULL','WINDOWS_SERVER_2019_FULL','WINDOWS_SERVER_2019_CORE','WINDOWS_SERVER_2016_FULL','WINDOWS_SERVER_2004_CORE','WINDOWS_SERVER_20H2_CORE'",
						Type:        schema.TypeString,
						Optional:    true,
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
		UpdateContext: resourceDuploEcsTaskDefinitionUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        ecsTaskDefinitionSchema(),
		CustomizeDiff: validateInput,
	}
}

// READ resource
func resourceDuploEcsTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, arn, err := parseEcsTaskDefIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead(%s, %s): start", tenantID, arn)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.EcsTaskDefinitionGetV2(tenantID, arn)
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

// CREATE resource
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

// Update resource
func resourceDuploEcsTaskDefinitionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

// DELETE resource
func resourceDuploEcsTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** start")
	var diags diag.Diagnostics
	id := d.Id()
	tenantID, arn, err := parseEcsTaskDefIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	preventDestroy := d.Get("prevent_tf_destroy").(bool)
	log.Printf("[TRACE] Prevent destroy is %t", preventDestroy)
	if !preventDestroy {
		c := m.(*duplosdk.Client)
		err = c.EcsTaskDefinitionDelete(tenantID, arn)
		if err != nil {
			return diag.FromErr(err)
		}
		// Wait for the task definition to be missing
		diags = waitForResourceToBeMissingAfterDelete(ctx, d, "ECS Task Defnition", id, func() (interface{}, duplosdk.ClientError) {
			if rp, err := c.EcsTaskDefinitionExists(tenantID, arn); rp || err != nil {
				return rp, err
			}
			return nil, nil
		})
	} else {
		log.Printf("[WARN] resourceDuploEcsTaskDefinitionDelete(%s): will NOT delete the task definition - because 'prevent_tf_destroy' is 'true'", arn)
	}

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** end")
	return diags
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
	platform := d.Get("runtime_platform").([]interface{})
	if len(platform) > 0 {
		duplo.RuntimePlatform = &duplosdk.DuploEcsTaskDefRuntimePlatform{}
		obj := platform[0].(map[string]interface{})
		if v, ok := obj["cpu_architecture"]; ok && v.(string) != "" {
			duplo.RuntimePlatform.CPUArchitecture.Value = v.(string)

		}
		if v, ok := obj["operating_system_family"]; ok && v.(string) != "" {
			duplo.RuntimePlatform.OSFamily.Value = v.(string)
		}
	}
	return &duplo, nil
}

func flattenEcsTaskDefinition(duplo *duplosdk.DuploEcsTaskDef, d *schema.ResourceData) {

	// First, convert things into simple scalars
	d.Set("tenant_id", duplo.TenantID)
	d.Set("full_family_name", duplo.Family)
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
	d.Set("runtime_platform", ecsPlatformRuntimeToState)
}

// An internal function that compares two ECS container definitions to see if they are equivalent.
func ecsContainerDefinitionsAreEquivalent(old, new string, isAWSVPC, isLogConfigProvided bool) (bool, error) {

	oldCanonical, err := canonicalizeEcsContainerDefinitionsJson(old, isAWSVPC, isLogConfigProvided)
	if err != nil {
		return false, err
	}

	newCanonical, err := canonicalizeEcsContainerDefinitionsJson(new, isAWSVPC, isLogConfigProvided)
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
func canonicalizeEcsContainerDefinitionsJson(encoded string, isAWSVPC, isLogConfigProvided bool) (string, error) {
	var defns []interface{}

	// Unmarshall, reduce, then canonicalize.
	err := json.Unmarshal([]byte(encoded), &defns)
	if err != nil {
		return encoded, err
	}
	for i := range defns {
		err = reduceContainerDefinition(defns[i].(map[string]interface{}), isAWSVPC, isLogConfigProvided)
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
			sort.SliceStable(env, func(i, j int) bool {

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
func reduceContainerDefinition(defn map[string]interface{}, isAWSVPC, isLogConfigProvided bool) error {

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

	// Handle computed fields
	if !isLogConfigProvided {
		delete(defn, "LogConfiguration")
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

func ecsPlatformRuntimeToState(p *duplosdk.DuploEcsTaskDefRuntimePlatform) []interface{} {
	if p != nil {
		return nil
	}

	results := make([]interface{}, 0)
	c := make(map[string]interface{})
	c["cpu_architecture"] = p.CPUArchitecture.Value
	c["operating_system_family"] = p.OSFamily.Value
	results = append(results, c)

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

func isLogConfigProvided(d *schema.ResourceData) bool {
	condefs := d.Get("container_definitions").(string)
	return strings.Contains(condefs, "LogConfiguration")
}

func validateInput(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	obj := diff.Get("requires_compatibilities").(*schema.Set)
	pf := diff.Get("runtime_platform").([]interface{})

	fmp := map[string]bool{
		"LINUX":                    true,
		"WINDOWS_SERVER_2019_FULL": true,
		"WINDOWS_SERVER_2019_CORE": true,
		"WINDOWS_SERVER_2022_FULL": true,
		"WINDOWS_SERVER_2022_CORE": true,
	}
	emp := map[string]bool{
		"LINUX":                    true,
		"WINDOWS_SERVER_2022_CORE": true,
		"WINDOWS_SERVER_2022_FULL": true,
		"WINDOWS_SERVER_2019_FULL": true,
		"WINDOWS_SERVER_2019_CORE": true,
		"WINDOWS_SERVER_2016_FULL": true,
		"WINDOWS_SERVER_2004_CORE": true,
		"WINDOWS_SERVER_20H2_CORE": true,
	}
	for _, o := range obj.List() {
		if o.(string) == "FARGATE" {
			if len(pf) == 0 {
				v := []interface{}{}
				m := map[string]interface{}{
					"operating_system_family": "LINUX",
					"cpu_architecture":        "X86_64",
				}
				v = append(v, m)
				e := diff.SetNew("runtime_platform", v)
				if e != nil {
					return e
				}
			} else {
				os := pf[0].(map[string]interface{})
				f := os["operating_system_family"].(string)
				if _, ok := fmp[f]; !ok {
					return errors.New("Invalid operating_system_family")
				}
			}
		} else if o.(string) == "EC2" {
			if len(pf) > 0 {
				os := pf[0].(map[string]interface{})
				f := os["operating_system_family"].(string)
				if f != "" {
					if _, ok := emp[f]; !ok {
						return errors.New("Invalid operating_system_family")

					}
				}
			}
		}
	}
	return nil
}
