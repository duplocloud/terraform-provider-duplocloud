package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
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
		ReadContext:   resourceDuploEcsTaskDefinitionRead,
		CreateContext: resourceDuploEcsTaskDefinitionCreate,
		DeleteContext: resourceDuploEcsTaskDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsTaskDefinitionGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.EcsTaskDefToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", duplo.TenantID, duplo.Arn))

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcsTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcsTaskDefFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	arn, err := c.EcsTaskDefinitionCreate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", tenantID, arn))

	diags := resourceDuploEcsTaskDefinitionRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploEcsTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** start")
	// FIXME: NO-OP
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** end")
	return nil
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

// Internal function to unmarshall, reduce, then canonicalize container definitions JSON.
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
		env := v.([]interface{})
		sort.Slice(env, func(i, j int) bool {
			mi := env[i].(map[string]interface{})
			mj := env[j].(map[string]interface{})
			si := ""
			sj := ""
			if v, ok = mi["Name"]; ok && !isInterfaceNil(v) {
				si = v.(string)
			}
			if v, ok = mj["Name"]; ok && !isInterfaceNil(v) {
				sj = v.(string)
			}
			return si < sj
		})
	}
}

// Internal function used to reduce a container definition down to a partially "canoncalized" form.
//
// See: https://github.com/hashicorp/terraform-provider-aws/blob/7141d1c315dc0c221979f0a4f8855a13b545dbaf/aws/ecs_task_definition_equivalency.go#L58
func reduceContainerDefinition(defn map[string]interface{}, isAWSVPC bool) error {
	reorderEcsEnvironmentVariables(defn)

	// Handle fields that have defaults.
	if v, ok := defn["Cpu"]; ok && v != nil && v.(int) == 0 {
		defn["Cpu"] = nil
	}
	if v, ok := defn["Essential"]; !(ok && v != nil) {
		defn["Essential"] = true
	}
	if v, ok := defn["PortMappings"]; ok && v != nil {
		pmi := v.([]interface{})
		for i := range pmi {
			pms := pmi[i].(map[string]interface{})

			if v2, ok2 := pms["Protocol"]; ok2 && !isInterfaceNil(v2) && v2.(string) == "tcp" {
				pms["Protocol"] = nil
			}
			if v2, ok2 := pms["HostPort"]; ok2 {
				if !isInterfaceNil(v2) && v2.(int) == 0 {
					pms["HostPort"] = nil
				}
				if isAWSVPC && pms["HostPort"] == nil {
					pms["HostPort"] = pms["ContainerPort"]
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
