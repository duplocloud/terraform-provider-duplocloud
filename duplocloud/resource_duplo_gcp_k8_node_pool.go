package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpK8NodePoolFunctionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the node pool will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the node pool.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"initial_node_count": {
			Description: "The initial node count for the pool",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"machine_type": {
			Description: `The name of a Google Compute Engine machine type.
				If unspecified, the default machine type is e2-medium.`,
			Type:     schema.TypeString,
			Optional: true,
		},
		"disc_size_gb": {
			Description: `Size of the disk attached to each node, specified in GB. The smallest allowed disk size is 10GB.
				If unspecified, the default disk size is 100GB.`,
			Type:     schema.TypeInt,
			Optional: true,
		},
		"image_type": {
			Description: "The image type to use for this node. Note that for a given image type, the latest version of it will be used",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"tags": {
			Description: `The list of instance tags applied to all nodes.
				Tags are used to identify valid sources or targets for network firewalls and are specified by the client during cluster or node pool creation.
				Each tag within the list must comply with RFC1035.`,
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"disc_type": {
			Description: `Type of the disk attached to each node
				If unspecified, the default disk type is 'pd-standard'`,
			Type:     schema.TypeString,
			Optional: true,
		},
		"spot": {
			Description: "Spot flag for enabling Spot VM",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"service_account": {
			Description: ``,
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"linux_node_config": {
			Description: "Parameters that can be configured on Linux nodes",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cgroup_mode": {
						Description: "cgroupMode specifies the cgroup mode to be used on the node.",
						Type:        schema.TypeString,
						Optional:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"CGROUP_MODE_UNSPECIFIED",
							"CGROUP_MODE_V1",
							"CGROUP_MODE_V2",
						}, false),
					},
					"sysctls": {
						Description: "The Linux kernel parameters to be applied to the nodes and all pods running on the nodes.",
						Type:        schema.TypeMap,
						Optional:    true,
						Elem: &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},

		"labels": {
			Description: "The map of Kubernetes labels (key/value pairs) to be applied to each node.",
			Type:        schema.TypeMap,
			Optional:    true,
			//Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"is_autoscaling_enabled": {
			Description: "Is autoscaling enabled for this node pool.",
			Type:        schema.TypeBool,
			Optional:    true,
		},

		"location_policy": {
			Description: "Update strategy of the node pool.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "BALANCED",
			ValidateFunc: validation.StringInSlice([]string{
				"LOCATION_POLICY_UNSPECIFIED",
				"BALANCED",
				"ANY",
			}, false),
		},

		"max_node_count": {
			Description: "Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"min_node_count": {
			Description: "Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.",
			Type:        schema.TypeInt,
			Optional:    true,
		},

		"total_max_node_count": {
			Description: "Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"total_min_node_count": {
			Description: "Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"auto_upgrade": {
			Description: "Whether the nodes will be automatically upgraded.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"auto_repair": {
			Description: "Whether the nodes will be automatically repaired.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"metadata": {
			Description: "The metadata key/value pairs assigned to instances in the cluster.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"node_pool_logging_config": {
			Description: "Logging configuration.",
			Type:        schema.TypeList,
			Optional:    true,

			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"variant_config": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem: &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"CGROUP_MODE_UNSPECIFIED",
								"CGROUP_MODE_V1",
								"CGROUP_MODE_V2",
							}, false),
						},
					},
				},
			},
		},
		"oauth_scopes": {
			Description: "The set of Google API scopes to be made available on all of the node VMs under the default service account.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"zones": {
			Description: "The list of Google Compute Engine zones in which the NodePool's nodes should be located.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString}, // Define the type for each element in the list
		},
		"upgrade_settings": {
			Description: "Upgrade settings control disruption and speed of the upgrade.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"strategy": {
						Description: "Update strategy of the node pool.",
						Type:        schema.TypeString,
						Optional:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"BLUE_GREEN",
							"SURGE",
						}, false),
					},
					"max_surge": {
						Description: "",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"max_unavailable": {
						Description: "",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"blue_green_settings": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"standard_rollout_policy": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"batch_percentage": {
												Type:     schema.TypeFloat,
												Optional: true,
											},
											"batch_node_count": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"batch_soak_duration": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
								"node_pool_soak_duration": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},

		"taints": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Description: "",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"value": {
						Description: "",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"effect": {
						Description: "Update strategy of the node pool.",
						Type:        schema.TypeString,
						Optional:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"EFFECT_UNSPECIFIED",
							"NO_SCHEDULE",
							"PREFER_NO_SCHEDULE",
							"NO_EXECUTE",
						}, false),
					},
				},
			},
		},
		"accelerator": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"accelerator_count": {
						Description: "The number of the accelerator cards exposed to an instance.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"accelerator_type": {
						Description: "The accelerator type resource name.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"gpu_partition_size": {
						Description: "Size of partitions to create on the GPU",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"max_time_shared_clients_per_gpu": {
						Description: "The number of time-shared GPU resources to expose for each physical GPU.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"gpu_sharing_config": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"max_shared_clients_per_gpu": {
									Description: "The max number of containers that can share a physical GPU.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"gpu_sharing_strategy": {
									Description: "The configuration for GPU sharing options.",
									Type:        schema.TypeString,
									Optional:    true,
									ValidateFunc: validation.StringInSlice([]string{
										"GPU_SHARING_STRATEGY_UNSPECIFIED",
										"TIME_SHARING",
									}, false),
								},
							},
						},
					},
					"gpu_driver_installation_config": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"gpu_driver_version": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceGcpK8NodePools() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_k8_node_pools` manages a GCP Node Pool in Duplo.",

		ReadContext:   resourceGCPNodePoolRead,
		UpdateContext: resourceGCPK8NodePoolUpdate,
		DeleteContext: resourceGcpNodePoolDelete,
		CreateContext: resourceGCPK8NodePoolCreate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: gcpK8NodePoolFunctionSchema(),
	}
}

// CREATE resource
func resourceGCPK8NodePoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPK8NodePoolCreate ******** start")
	tenantID := d.Get("tenant_id").(string)

	c := m.(*duplosdk.Client)

	// Create the request object.
	rq, err := expandGCPNodePool(d)
	if err != nil {
		return diag.Errorf("Error fetching request for %s : %s", tenantID, err.Error())

	}
	fullName, clientErr := c.GetDuploServicesName(tenantID, rq.Name)
	if clientErr != nil {
		return diag.Errorf("Error fetching tenant prefix for %s : %s", tenantID, clientErr)
	}
	// Post the object to Duplo
	resp, err := c.GCPK8NodePoolCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s gcp node pool '%s': %s", tenantID, resp.Name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "gcp node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPK8NodePoolGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	resourceGCPNodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceGCPK8NodePoolCreate ******** end")
	return diags
}

func expandGCPNodePool(d *schema.ResourceData) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		Name:                 d.Get("name").(string),
		InitialNodeCount:     d.Get("initial_node_count").(int),
		IsAutoScalingEnabled: d.Get("is_autoscaling_enabled").(bool),
	}
	for _, zone := range d.Get("zones").([]interface{}) {
		rq.Zones = append(rq.Zones, zone.(string))
	}
	expandGCPNodePoolManagement(d, rq)
	expandGCPNodePoolConfig(d, rq)
	expandGCPNodePoolAutoScaling(d, rq)
	expandGCPNodePoolUpgradeSettings(d, rq)
	return rq, nil
}

func expandGCPNodePoolUpgradeSettings(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {
	if val, ok := d.Get("upgrade_settings").([]interface{}); ok {
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				upgradeSetting := &duplosdk.GCPNodeUpgradeSetting{
					MaxSurge:       m["max_surge"].(int),
					MaxUnavailable: m["max_unavailable"].(int),
					Strategy:       m["strategy"].(string),
				}

				if bgMap, ok := m["blue_green_settings"].(map[string]interface{}); ok {
					blueGreenSettings := &duplosdk.BlueGreenSettings{
						NodePoolSoakDuration: bgMap["node_pool_soak_duration"].(string),
					}

					if rollout, ok := bgMap["standard_rollout_policy"].(map[string]interface{}); ok {
						rolloutPolicy := &duplosdk.StandardRolloutPolicy{
							BatchPercentage:   rollout["batch_percentage"].(float32),
							BatchNodeCount:    rollout["batch_node_count"].(int),
							BatchSoakDuration: rollout["batch_soak_duration"].(string),
						}
						blueGreenSettings.StandardRolloutPolicy = rolloutPolicy
					}
					upgradeSetting.BlueGreenSettings = blueGreenSettings
				}
				req.UpgradeSettings = upgradeSetting
			}
		}
	}
}

func expandGCPNodePoolConfig(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {
	req.MachineType = d.Get("machine_type").(string)
	req.DiscSizeGb = d.Get("disc_size_gb").(int)
	req.ImageType = d.Get("image_type").(string)
	if val, ok := d.Get("tags").([]string); ok {
		req.Tags = val
	}
	if val, ok := d.Get("disc_type").(string); ok {
		req.DiscType = val
	}
	req.Spot = d.Get("spot").(bool)
	if val, ok := d.Get("linux_node_config").([]map[string]interface{}); ok {
		req.LinuxNodeConfig.CGroupMode = val[0]["cgroup_mode"].(string)
		req.LinuxNodeConfig.SysCtls = val[0]["sysctls"]
	}
	if val, ok := d.Get("labels").(map[string]string); ok {
		req.Labels = val
	}
	if val, ok := d.Get("logging_config").(map[string]interface{}); ok {
		loggingConfig := duplosdk.GCPLoggingConfig{}

		if vConfig, ok := val["variant_config"].(map[string]interface{}); ok {
			variantConfig := duplosdk.VariantConfig{}

			if variant, ok := vConfig["variant"].(string); ok {
				variantConfig.Variant = variant
			}

			loggingConfig.VariantConfig = &variantConfig
		}

		// Assign the loggingConfig to the request object
		req.LoggingConfig = &loggingConfig
	}

	if val, ok := d.Get("taints").([]duplosdk.GCPNodeTaints); ok {
		for _, dt := range val {
			taints := duplosdk.GCPNodeTaints{
				Key:    dt.Key,
				Value:  dt.Value,
				Effect: dt.Effect,
			}
			req.Taints = append(req.Taints, taints)

		}
	}

	if val, ok := d.Get("metadata").(map[string]string); ok {
		req.Metadata = val
	}
	expandGCPNodePoolAccelerator(d, req)
}

func expandGCPNodePoolAccelerator(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {
	if val, ok := d.Get("accelerator").([]interface{}); ok {
		item := val[0]
		if m, ok := item.(map[string]interface{}); ok {
			count, _ := strconv.Atoi(m["accelerator_count"].(string))

			accelerator := duplosdk.Accelerator{
				AcceleratorCount: count,
				AcceleratorType:  m["accelerator_type"].(string),
				GPUPartitionSize: m["gpu_partition_size"].(string),
			}
			if sharingConfig, ok := m["gpu_sharing_config"].([]interface{}); ok && len(sharingConfig) > 0 {
				sharingConfigMap := sharingConfig[0].(map[string]interface{})
				count, _ = strconv.Atoi(sharingConfigMap["max_shared_clients_per_gpu"].(string))
				accelerator.GPUSharingConfig = duplosdk.GPUSharingConfig{
					MaxSharedClientPerGPU: count,
					GPUSharingStrategy:    sharingConfigMap["gpu_sharing_strategy"].(string),
				}
			}
			if driverConfig, ok := m["gpu_driver_installation_config"].([]interface{}); ok && len(driverConfig) > 0 {
				driverConfigMap := driverConfig[0].(map[string]interface{})
				accelerator.GPUDriverInstallationConfig = duplosdk.GPUDriverInstallationConfig{
					GPUDriverVersion: driverConfigMap["gpu_driver_version"].(string),
				}
			}

			req.Accelerator = accelerator

		}
		log.Printf("accelrator object \n%+v", val)
	}
	log.Printf("Accelerators \n%+v", req.Accelerator)
}

func expandGCPNodePoolManagement(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {

	if val, ok := d.Get("auto_upgrade").(bool); ok {
		req.AutoUpgrade = val
	}

	if val, ok := d.Get("auto_repair").(bool); ok {
		req.AutoRepair = val
	}

}

func expandGCPNodePoolAutoScaling(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {
	if val, ok := d.Get("enabled").(bool); ok {
		req.IsAutoScalingEnabled = val
	}
	if val, ok := d.Get("location_policy").(string); ok {
		req.LocationPolicy = val
	}

	if val, ok := d.Get("min_node_count").(int); ok {
		req.MinNodeCount = &val
	}
	if val, ok := d.Get("max_node_count").(int); ok {
		req.MaxNodeCount = &val
	}
	if val, ok := d.Get("location_policy").(string); ok {
		req.LocationPolicy = val
	}

}

func resourceGCPNodePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPNodePoolRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing or deleted object
	c := m.(*duplosdk.Client)
	duplo, err := c.GCPK8NodePoolGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s GCP Node Pool Domain '%s': %s", tenantID, name, err)
	}

	if duplo == nil {
		d.SetId("") // object missing or deleted
		return nil
	}
	setGCPNodePoolStateField(d, duplo, tenantID, name)

	log.Printf("[TRACE] resourceDuploAwsElasticSearchRead ******** end")
	return nil
}
func setGCPNodePoolStateField(d *schema.ResourceData, duplo *duplosdk.DuploGCPK8NodePool, tenantID, fullName string) {
	// Set simple fields first.

	d.SetId(fmt.Sprintf("%s/%s", tenantID, fullName))
	d.Set("tenant_id", tenantID)
	d.Set("name", duplo.Name)
	d.Set("is_autoscaling_enabled", duplo.IsAutoScalingEnabled)
	d.Set("auto_upgrade", duplo.AutoUpgrade)
	d.Set("zones", duplo.Zones)
	d.Set("image_type", duplo.ImageType)
	d.Set("location_policy", duplo.LocationPolicy)
	d.Set("max_node_count", duplo.MaxNodeCount)
	d.Set("min_node_count", duplo.MinNodeCount)
	d.Set("total_max_node_count", duplo.TotalMaxNodeCount)
	d.Set("total_min_node_count", duplo.TotalMinNodeCount)
	d.Set("initial_node_count", duplo.InitialNodeCount)
	d.Set("auto_repair", duplo.AutoRepair)
	d.Set("auto_upgrade", duplo.AutoUpgrade)
	d.Set("zones", duplo.Zones)
	d.Set("disc_size_gb", duplo.DiscSizeGb)
	d.Set("disc_type", duplo.DiscType)
	d.Set("machine_type", duplo.MachineType)
	d.Set("metadata", duplo.Metadata)
	d.Set("labels", duplo.Labels)
	d.Set("spot", duplo.Spot)
	d.Set("tags", duplo.Tags)
	d.Set("taints", gcpNodePoolTaintstoState(duplo.Taints))
	d.Set("node_pool_logging_config", gcpNodePoolLoggingConfigToState(duplo.LoggingConfig))
	d.Set("linux_node_config", gcpNodePoolLinuxConfigToState(duplo.LinuxNodeConfig))
	d.Set("upgrade_settings", gcpNodePoolUpgradeSettingToState(duplo.UpgradeSettings))
	d.Set("accelerator", gcpNodePoolAcceleratortoState(duplo.Accelerator))
	// Set more complex fields next.

}

func gcpNodePoolUpgradeSettingToState(upgradeSetting *duplosdk.GCPNodeUpgradeSetting) []map[string]interface{} {
	state := make(map[string]interface{})
	state["strategy"] = upgradeSetting.Strategy
	state["max_surge"] = upgradeSetting.MaxSurge
	state["max_unavailable"] = upgradeSetting.MaxUnavailable

	if upgradeSetting.BlueGreenSettings != nil {
		blueGreenSettings := make(map[string]interface{})
		if upgradeSetting.BlueGreenSettings.StandardRolloutPolicy != nil {
			rolloutPolicy := make(map[string]interface{})
			rolloutPolicy["batch_percentage"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchPercentage
			rolloutPolicy["batch_node_count"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchNodeCount
			rolloutPolicy["batch_soak_duration"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchSoakDuration
			blueGreenSettings["standard_rollout_policy"] = []map[string]interface{}{rolloutPolicy}
		}
		blueGreenSettings["node_pool_soak_duration"] = upgradeSetting.BlueGreenSettings.NodePoolSoakDuration
		state["blue_green_settings"] = []map[string]interface{}{blueGreenSettings}
	}

	return []map[string]interface{}{state}
}

func gcpNodePoolLoggingConfigToState(loggingConfig *duplosdk.GCPLoggingConfig) []map[string]interface{} {
	state := make(map[string]interface{})
	if loggingConfig != nil && loggingConfig.VariantConfig != nil {
		variant := make(map[string]interface{})
		variant["variant"] = loggingConfig.VariantConfig.Variant
		state["variant_config"] = []map[string]interface{}{variant}
	}
	return []map[string]interface{}{state}
}

func gcpNodePoolLinuxConfigToState(linuxConfig *duplosdk.GCPLinuxNodeConfig) []map[string]interface{} {
	state := make(map[string]interface{})
	if linuxConfig != nil {
		state["cgroup_mode"] = linuxConfig.CGroupMode
		state["systctls"] = linuxConfig.SysCtls
	}
	return []map[string]interface{}{state}
}

func gcpNodePoolAcceleratortoState(accelerator duplosdk.Accelerator) []map[string]interface{} {
	state := make(map[string]interface{})

	if accelerator.AcceleratorCount != 0 {
		state["accelerator_count"] = strconv.Itoa(accelerator.AcceleratorCount)
	}
	state["accelerator_type"] = accelerator.AcceleratorType
	state["gpu_partition_size"] = accelerator.GPUPartitionSize

	gpuSharingConfigMap := make(map[string]interface{})
	if accelerator.GPUSharingConfig.GPUSharingStrategy != "" {
		gpuSharingConfigMap["gpu_sharing_strategy"] = accelerator.GPUSharingConfig.GPUSharingStrategy
	}
	if accelerator.GPUSharingConfig.MaxSharedClientPerGPU != 0 {
		gpuSharingConfigMap["max_shared_clients_per_gpu"] = strconv.Itoa(accelerator.GPUSharingConfig.MaxSharedClientPerGPU)
	}
	state["gpu_sharing_config"] = gpuSharingConfigMap

	driverConfig := make(map[string]interface{})
	if accelerator.GPUDriverInstallationConfig.GPUDriverVersion != "" {
		driverConfig["gpu_driver_version"] = accelerator.GPUDriverInstallationConfig.GPUDriverVersion
	}
	state["gpu_driver_installation_config"] = driverConfig

	return []map[string]interface{}{state}
}

func gcpNodePoolTaintstoState(taints []duplosdk.GCPNodeTaints) []interface{} {
	state := make([]interface{}, len(taints))
	for i, t := range taints {
		data := map[string]interface{}{
			"key":    t.Key,
			"value":  t.Value,
			"effect": t.Effect,
		}
		state[i] = data
	}
	return state
}

func resourceGcpNodePoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpNodePoolDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.GCPK8NodePoolDelete(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting node pool '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the node pool.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "gcp node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPK8NodePoolGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpNodePoolDelete ******** end")
	return nil
}

func resourceGCPK8NodePoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGCPK8NodePoolUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]
	rq, err := expandGCPNodePool(d)
	if err != nil {
		return diag.Errorf("Error fetching request for %s : %s", tenantID, err.Error())

	}
	c := m.(*duplosdk.Client)
	resp, err := c.GCPK8NodePoolUpdate(tenantID, name, rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s Node Pool '%s': %s", tenantID, rq.Name, err)
	}
	setGCPNodePoolStateField(d, resp, tenantID, name)
	log.Printf("[TRACE] resourceGCPK8NodePoolUpdate ******** end")

	return nil
}
