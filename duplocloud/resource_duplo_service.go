package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ucarion/jcs"
)

// DuploServiceSchema returns a Terraform resource schema for a service's parameters
func duploServiceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The name of the service to create.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"tenant_id": {
			Description: "The GUID of the tenant that the service will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, //switch tenant
		},
		"other_docker_host_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"other_docker_config": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			StateFunc: func(v interface{}) string {
				// Creates canonical JSON as it is serialized to state, so we won't get
				// spurious reorderings in plans (diff is suppressed if the environment variables haven't changed,
				// but they still show in the plan if some other property changes).
				log.Printf("[TRACE] duplocloud_duplo_service.other_docker_config.StateFunc: <= %v", v)
				defn, _ := expandOtherDockerConfig(v.(string))
				reorderOtherDockerConfigsEnvironmentVariables(defn)
				json, err := jcs.Format(defn)
				if json == "{}" {
					json = ""
				}
				log.Printf("[TRACE] duplocloud_duplo_service.other_docker_config.StateFunc: => %s (error: %s)", json, err)
				return json
			},
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				equal, _ := otherDockerConfigsAreEquivalent(old, new)
				return equal
			},
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
			Description: "The numeric ID of the cloud provider to launch the service in.",
			Type:        schema.TypeInt,
			Optional:    true,
			Required:    false,
			Default:     0,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent to use for deployment.",
			Type:        schema.TypeInt,
			Optional:    true,
			Required:    false,
			Default:     0,
		},
		"replicas": {
			Description: "The number of container replicas to deploy.",
			Type:        schema.TypeInt,
			Optional:    false,
			Required:    true,
		},
		"replicas_matching_asg_name": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"docker_image": {
			Description: "The docker image to use for the launched container(s).",
			Type:        schema.TypeString,
			Optional:    false,
			Required:    true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// SCHEMA for resource crud
func resourceDuploService() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_duplo_service` manages a container-based service in Duplo.\n\n" +
			"NOTE: For Amazon ECS services, see the `duplocloud_ecs_service` resource.",

		ReadContext:   resourceDuploServiceRead,
		CreateContext: resourceDuploServiceCreate,
		UpdateContext: resourceDuploServiceUpdate,
		DeleteContext: resourceDuploServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: duploServiceSchema(),
	}
}

/// READ resource
func resourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceRead ******** start")

	// Parse the identifying attributes
	tenantID, name := parseDuploServiceIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploServiceGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s service '%s': %s", tenantID, name, err)
	}

	// Record the state of the object.
	d.Set("name", duplo.Name)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("other_docker_host_config", duplo.OtherDockerHostConfig)
	d.Set("other_docker_config", duplo.OtherDockerConfig)
	d.Set("allocation_tags", duplo.AllocationTags)
	d.Set("extra_config", duplo.ExtraConfig)
	d.Set("commands", duplo.Commands)
	d.Set("volumes", duplo.Volumes)
	d.Set("docker_image", duplo.DockerImage)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("replicas_matching_asg_name", duplo.ReplicasMatchingAsgName)
	d.Set("replicas", duplo.Replicas)
	d.Set("cloud", duplo.Cloud)
	d.Set("tags", keyValueToState("tags", duplo.Tags))

	log.Printf("[TRACE] resourceDuploServiceRead ******** start")
	return nil
}

/// CREATE resource
func resourceDuploServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceCreate ******** start")
	diags := resourceDuploServiceCreateOrUpdate(ctx, d, m, false)
	log.Printf("[TRACE] resourceDuploServiceCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceUpdate ******** start")
	diags := resourceDuploServiceCreateOrUpdate(ctx, d, m, true)
	log.Printf("[TRACE] resourceDuploServiceUpdate ******** end")
	return diags
}

func resourceDuploServiceCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, updating bool) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceCreateOrUpdate ******** start")

	// Build the request.
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	rq := duplosdk.DuploService{
		Name:                    name,
		OtherDockerHostConfig:   d.Get("other_docker_host_config").(string),
		OtherDockerConfig:       d.Get("other_docker_config").(string),
		AllocationTags:          d.Get("allocation_tags").(string),
		ExtraConfig:             d.Get("extra_config").(string),
		Commands:                d.Get("commands").(string),
		Volumes:                 d.Get("volumes").(string),
		AgentPlatform:           d.Get("agent_platform").(int),
		DockerImage:             d.Get("docker_image").(string),
		ReplicasMatchingAsgName: d.Get("replicas_matching_asg_name").(string),
		Cloud:                   d.Get("cloud").(int),
		Replicas:                d.Get("replicas").(int),
	}

	// Post the object to Duplo
	id := fmt.Sprintf("v2/subscriptions/%s/ReplicationControllerApiV2/%s", tenantID, name)
	c := m.(*duplosdk.Client)
	_, err := c.DuploServiceCreateOrUpdate(tenantID, &rq, updating)
	if err != nil {
		return diag.Errorf("Error applying Duplo service '%s': %s", id, err)
	}
	if !updating {
		d.SetId(id)
	}

	// Read the latest status from Duplo
	diags := resourceDuploServiceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceCreateOrUpdate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceDelete ******** start")

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name := parseDuploServiceIdParts(id)
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	err := c.DuploServiceDelete(tenantID, name)
	if err != nil {
		return diag.Errorf("Error deleting Duplo service '%s': %s", id, err)
	}

	// Wait for it to be deleted
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "duplo service", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploServiceGet(tenantID, name)
	})

	// Wait 40 more seconds to deal with consistency issues.
	if diags == nil {
		time.Sleep(40 * time.Second)
	}

	log.Printf("[TRACE] resourceDuploServiceDelete ******** end")
	return nil
}

// Internal function to expand other_docker_config JSON into a structure.
func expandOtherDockerConfig(encoded string) (defn map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(encoded), &defn)
	log.Printf("[DEBUG] Expanded duplocloud_duplo_service.other_docker_config: %v", defn)
	return
}

// Internal function to unmarshal, reduce, then canonicalize other_docker_config JSON.
func canonicalizeOtherDockerConfigJson(encoded string) (string, error) {
	var defn interface{}

	// Unmarshall, reduce, then canonicalize.
	err := json.Unmarshal([]byte(encoded), &defn)
	if err != nil {
		return encoded, err
	}
	err = reduceOtherDockerConfig(defn.(map[string]interface{}))
	if err != nil {
		return encoded, err
	}
	canonical, err := jcs.Format(defn)
	if err != nil {
		return encoded, err
	}
	if canonical == "{}" {
		canonical = ""
	}

	return canonical, nil
}

func reduceOtherDockerConfig(defn map[string]interface{}) error {

	// Ensure we are using upper-camel case.
	makeMapUpperCamelCase(defn)

	// Reorder the environment variables.
	reorderOtherDockerConfigsEnvironmentVariables(defn)

	// Handle fields that have defaults.
	if v, ok := defn["HostNetwork"]; !ok || isInterfaceNil(v) {
		defn["HostNetwork"] = false
	}

	// Handle probe entries.
	probes := []string{"LivenessProbe", "ReadinessProbe"}
	for _, pk := range probes {
		if pv, ok := defn[pk]; ok {
			if probe, ok := pv.(map[string]interface{}); ok {
				makeMapUpperCamelCase(probe)

				// Reduce HTTP Get keys
				if hg, ok := probe["HttpGet"]; ok {
					if hgv, ok := hg.(map[string]interface{}); ok {
						reduceNilOrEmptyMapEntries(hgv)
						makeMapUpperCamelCase(hgv)
					}
				}

				reduceNilOrEmptyMapEntries(probe)
			}
		}
	}

	// Handle env entries.
	if v, ok := defn["Env"]; ok {
		if list, ok := v.([]interface{}); ok {
			for _, item := range list {
				if entry, ok := item.(map[string]interface{}); ok {
					reduceNilOrEmptyMapEntries(entry)

					// Reduce ValueFrom keys.
					if ev, ok := entry["ValueFrom"]; ok {
						if vf, ok := ev.(map[string]interface{}); ok {
							makeMapUpperCamelCase(vf)

							// Reduce SecretKeyRef keys.
							if skr, ok := vf["SecretKeyRef"]; ok {
								if skrv, ok := skr.(map[string]interface{}); ok {
									reduceNilOrEmptyMapEntries(skrv)
									makeMapUpperCamelCase(skrv)
								}
							}

							// Reduce ConfigMapKeyRef keys.
							if skr, ok := vf["ConfigMapKeyRef"]; ok {
								if skrv, ok := skr.(map[string]interface{}); ok {
									reduceNilOrEmptyMapEntries(skrv)
									makeMapUpperCamelCase(skrv)
								}
							}

							reduceNilOrEmptyMapEntries(vf)
						}
					}
				}
			}
		}
	}

	// Handle fields that have nil entries.
	reduceNilOrEmptyMapEntries(defn)

	return nil
}

// An internal function that compares two other_docker_config values to see if they are equivalent.
func otherDockerConfigsAreEquivalent(old, new string) (bool, error) {

	oldCanonical, err := canonicalizeOtherDockerConfigJson(old)
	if err != nil {
		return false, err
	}

	newCanonical, err := canonicalizeOtherDockerConfigJson(new)
	if err != nil {
		return false, err
	}

	equal := oldCanonical == newCanonical
	if !equal {
		log.Printf("[DEBUG] Canonical definitions are not equal.\nFirst: %s\nSecond: %s\n", oldCanonical, newCanonical)
	}
	return equal, nil
}

// Internal function used to re-order environment variables for an ECS task definition.
func reorderOtherDockerConfigsEnvironmentVariables(defn map[string]interface{}) {

	// Re-order environment variables to a canonical order.
	if v, ok := defn["Env"]; ok && v != nil {
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

func parseDuploServiceIdParts(id string) (tenantID, name string) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	}
	return
}
