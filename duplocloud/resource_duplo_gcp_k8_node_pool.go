package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
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
			ValidateFunc: validation.IsUUID,
			ForceNew:     true,
		},
		"name": {
			Description: "The short name of the node pool.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The short name of the node pool.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"machine_type": {
			Description: `The name of a Google Compute Engine machine type.
				If unspecified, the default machine type is e2-medium.`,
			Type:     schema.TypeString,
			Required: true,
		},
		"disc_size_gb": {
			Description: `Size of the disk attached to each node, specified in GB. The smallest allowed disk size is 10GB.
				If unspecified, the default disk size is 100GB.`,
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"image_type": {
			Description: "The image type to use for this node. Note that for a given image type, the latest version of it will be used",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"ubuntu_containerd",
				"cos_containerd",
			}, false),
		},
		"tags": {
			Description: `The list of instance tags applied to all nodes.
				Tags are used to identify valid sources or targets for network firewalls and are specified by the client during cluster or node pool creation.
				Each tag within the list must comply with RFC1035.`,
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"disc_type": {
			Description: `Type of the disk attached to each node
				If unspecified, the default disk type is 'pd-standard'`,
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"allocation_tags": {
			Description: `Allocation tag to give to the nodes 
			if specified it would be added as a label and that can be used while creating services`,
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"spot": {
			Description: "Spot flag for enabling Spot VM",
			Type:        schema.TypeBool,
			Optional:    true,
		},

		"linux_node_config": {
			Description: "Parameters that can be configured on Linux nodes",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cgroup_mode": {
						Description: "cgroupMode specifies the cgroup mode to be used on the node.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
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
						Computed:    true,
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
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"is_autoscaling_enabled": {
			Description: "Is autoscaling enabled for this node pool.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
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
		"initial_node_count": {
			Description: "The initial node count for the pool",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     1,
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
			Computed:    true,
		},
		"auto_repair": {
			Description: "Whether the nodes will be automatically repaired.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"metadata": {
			Description: "The metadata key/value pairs assigned to instances in the cluster.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"node_pool_logging_config": {
			Description: "Logging configuration.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"variant_config": {
						Type:     schema.TypeMap,
						Optional: true,
						Computed: true,
						Elem: &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"VARIANT_UNSPECIFIED",
								"DEFAULT",
								"MAX_THROUGHPUT",
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
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"resource_labels": {
			Description: "Resource labels associated to node pool",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"zones": {
			Description: "The list of Google Compute Engine zones in which the NodePool's nodes should be located.",
			Type:        schema.TypeList,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString}, // Define the type for each element in the list
		},
		"upgrade_settings": {
			Description: "Upgrade settings control disruption and speed of the upgrade.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"strategy": {
						Description: "Update strategy of the node pool.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"BLUE_GREEN",
							"SURGE",
						}, false),
					},
					"max_surge": {
						Description: "The maximum number of nodes that can be created beyond the current size of the node pool during the upgrade process.",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},
					"max_unavailable": {
						Description: "The maximum number of nodes that can be simultaneously unavailable during the upgrade process. A node is considered available if its status is Ready",
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
					},
					"blue_green_settings": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"standard_rollout_policy": {
									Description: "Note: The standard_rollout_policy should not be used along with node_pool_soak_duration",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"batch_percentage": {
												Description: "Note: The batch_percentage should not be used along with batch_node_count",

												Type:         schema.TypeFloat,
												Optional:     true,
												Computed:     true,
												ValidateFunc: validation.FloatBetween(0.1, 1.0),
											},
											"batch_node_count": {
												Description: "Note: The batch_node_count should not be used along with batch_percentage",

												Type:     schema.TypeInt,
												Optional: true,
												Computed: true,
											},
											"batch_soak_duration": {
												Type:         schema.TypeString,
												Optional:     true,
												Computed:     true,
												ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d+(\.\d{1,9})?s$`), "Invalid seconds format, valid format : seconds with up to nine fractional digits, ending with 's'. Example: `3.5s`."),
											},
										},
									},
								},
								"node_pool_soak_duration": {
									Description:  "Note: The node_pool_soak_duration should not be used along with standard_rollout_policy",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d+(\.\d{1,9})?s$`), "Invalid seconds format, valid format : seconds with up to nine fractional digits, ending with 's'. Example: `3.5s`."),
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
						Type:     schema.TypeString,
						Optional: true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"effect": {
						Description: "Update strategy of the node pool. Supported effect's are : \n\t- EFFECT_UNSPECIFIED \n\t- NO_SCHEDULE \n\t- PREFER_NO_SCHEDULE\n\t- NO_EXECUTE",
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
						Computed:    true,
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
						Computed:    true,
					},
					"max_time_shared_clients_per_gpu": {
						Description: "The number of time-shared GPU resources to expose for each physical GPU.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"gpu_sharing_config": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"max_shared_clients_per_gpu": {
									Description: "The max number of containers that can share a physical GPU.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"gpu_sharing_strategy": {
									Description: "The configuration for GPU sharing options.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
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
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"gpu_driver_version": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceGcpK8NodePool() *schema.Resource {
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
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
	// Post the object to Duplo
	shortName := rq.Name
	resp, err := c.GCPK8NodePoolCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s gcp node pool '%s': %s", tenantID, resp.Name, err)
	}
	fullName, clientErr := c.GetDuploServicesName(tenantID, shortName)
	if clientErr != nil {
		return diag.Errorf("Error fetching tenant prefix for %s : %s", tenantID, clientErr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "gcp node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPK8NodePoolGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	d.Set("name", shortName)
	resourceGCPNodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceGCPK8NodePoolCreate ******** end")
	return diags
}

func expandGCPNodePool(d *schema.ResourceData) (*duplosdk.DuploGCPK8NodePool, error) {
	fullName := d.Get("name").(string)

	id := d.Id()
	if id != "" {
		idParts := strings.SplitN(id, "/", 2)
		fullName = idParts[1]
	}
	rq := &duplosdk.DuploGCPK8NodePool{
		Name:                 fullName,
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
	_, err := autoScalingHelper(rq.IsAutoScalingEnabled, rq)
	if err != nil {
		return nil, err
	}
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

				if bgl, ok := m["blue_green_settings"].([]interface{}); ok {
					if len(bgl) > 0 {
						bgMap := bgl[0].(map[string]interface{})
						blueGreenSettings := &duplosdk.BlueGreenSettings{
							NodePoolSoakDuration: bgMap["node_pool_soak_duration"].(string),
						}

						if rolloutList, ok := bgMap["standard_rollout_policy"].([]interface{}); ok {
							if len(rolloutList) > 0 {
								rollout := rolloutList[0].(map[string]interface{})
								rolloutPolicy := &duplosdk.StandardRolloutPolicy{
									BatchPercentage:   float32(rollout["batch_percentage"].(float64)),
									BatchNodeCount:    rollout["batch_node_count"].(int),
									BatchSoakDuration: rollout["batch_soak_duration"].(string),
								}
								blueGreenSettings.StandardRolloutPolicy = rolloutPolicy
							}
						}
						upgradeSetting.BlueGreenSettings = blueGreenSettings
					}
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

	if val, ok := d.Get("allocation_tags").(string); ok {
		req.AllocationTags = val
	}
	for _, tag := range d.Get("tags").([]interface{}) {
		req.Tags = append(req.Tags, tag.(string))
	}

	for _, oauth := range d.Get("oauth_scopes").([]interface{}) {
		req.OauthScopes = append(req.OauthScopes, oauth.(string))
	}

	if val, ok := d.Get("disc_type").(string); ok {
		req.DiscType = val
	}
	req.Spot = d.Get("spot").(bool)
	if val, ok := d.Get("linux_node_config").([]interface{}); ok {
		if len(val) > 0 {
			if val[0] != nil {
				m := val[0].(map[string]interface{})
				lnc := duplosdk.GCPLinuxNodeConfig{
					CGroupMode: m["cgroup_mode"].(string),
					SysCtls:    m["sysctls"],
				}
				req.LinuxNodeConfig = &lnc
			}
		}
	}
	if val, ok := d.Get("labels").(map[string]interface{}); ok {
		req.Labels = make(map[string]string)
		for k, v := range val {
			req.Labels[k] = v.(string)
		}
	}
	if val, ok := d.Get("node_pool_logging_config").([]interface{}); ok {
		loggingConfig := duplosdk.GCPLoggingConfig{}
		//for _, v := range val {
		if len(val) > 0 {
			if val[0] != nil {
				m := val[0].(map[string]interface{})
				m1 := m["variant_config"].(map[string]interface{})
				variantConfig := duplosdk.VariantConfig{}

				if variant, ok := m1["variant"].(string); ok {
					variantConfig.Variant = variant
				}

				loggingConfig.VariantConfig = &variantConfig
			}
		}
		// Assign the loggingConfig to the request object
		req.LoggingConfig = &loggingConfig
	}

	if val, ok := d.Get("taints").([]interface{}); ok {
		for _, dt := range val {
			m := dt.(map[string]interface{})
			taints := duplosdk.GCPNodeTaints{
				Key:    m["key"].(string),
				Value:  m["value"].(string),
				Effect: m["effect"].(string),
			}
			req.Taints = append(req.Taints, taints)

		}
	}
	if val, ok := d.Get("metadata").(map[string]interface{}); ok {
		metadata := make(map[string]string)
		for key, value := range val {
			if strVal, ok := value.(string); ok {
				metadata[key] = strVal
			}
		}
		req.Metadata = metadata
	}

	if val, ok := d.Get("resource_labels").(map[string]interface{}); ok {
		resourceLabels := make(map[string]string)
		for key, value := range val {
			if strVal, ok := value.(string); ok {
				resourceLabels[key] = strVal
			}
		}
		req.ResourceLabels = resourceLabels
	}
	expandGCPNodePoolAccelerator(d, req)
}

func expandGCPNodePoolAccelerator(d *schema.ResourceData, req *duplosdk.DuploGCPK8NodePool) {
	if val, ok := d.Get("accelerator").([]interface{}); ok && len(val) > 0 {
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

			req.Accelerator = &accelerator

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
	tenantID, fullName := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing or deleted object
	c := m.(*duplosdk.Client)
	duplo, err := c.GCPK8NodePoolGet(tenantID, fullName)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s GCP Node Pool Domain '%s': %s", tenantID, fullName, err)
	}

	if duplo == nil {
		d.SetId("") // object missing or deleted
		return nil
	}
	setGCPNodePoolStateField(d, duplo, tenantID)

	log.Printf("[TRACE] resourceDuploAwsElasticSearchRead ******** end")
	return nil
}
func getGCPNodePoolShortName(fullName, tenantName string) string {
	shortName := strings.Split(fullName, tenantName+"-")
	return shortName[len(shortName)-1]
}
func setGCPNodePoolStateField(d *schema.ResourceData, duplo *duplosdk.DuploGCPK8NodePool, tenantID string) {
	// Set simple fields first.

	d.SetId(fmt.Sprintf("%s/%s", tenantID, duplo.Name))
	d.Set("tenant_id", tenantID)
	d.Set("name", getGCPNodePoolShortName(duplo.Name, duplo.ResourceLabels["duplo-tenant"]))
	d.Set("fullname", duplo.Name)
	d.Set("is_autoscaling_enabled", duplo.IsAutoScalingEnabled)
	d.Set("auto_upgrade", duplo.AutoUpgrade)
	d.Set("zones", duplo.Zones)
	d.Set("image_type", strings.ToLower(duplo.ImageType))
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
	d.Set("labels", filterOutDefaultLabels(duplo.Labels))
	d.Set("spot", duplo.Spot)
	d.Set("tags", filterOutDefaultTags(duplo.Tags))
	d.Set("taints", gcpNodePoolTaintstoState(duplo.Taints))
	d.Set("node_pool_logging_config", gcpNodePoolLoggingConfigToState(duplo.LoggingConfig))
	d.Set("linux_node_config", gcpNodePoolLinuxConfigToState(duplo.LinuxNodeConfig))
	d.Set("upgrade_settings", gcpNodePoolUpgradeSettingToState(duplo.UpgradeSettings))
	d.Set("accelerator", gcpNodePoolAcceleratortoState(duplo.Accelerator))
	d.Set("oauth_scopes", filterOutDefaultOAuth(duplo.OauthScopes))
	d.Set("resource_labels", filterOutDefaultResourceLabels(duplo.ResourceLabels))
	// Set more complex fields next.

}

func gcpNodePoolUpgradeSettingToState(upgradeSetting *duplosdk.GCPNodeUpgradeSetting) []interface{} {
	us := make([]interface{}, 0, 1)
	if upgradeSetting == nil {
		return nil
	}
	state := make(map[string]interface{})
	state["strategy"] = upgradeSetting.Strategy
	if upgradeSetting.MaxSurge > 0 {
		state["max_surge"] = upgradeSetting.MaxSurge
	}
	if upgradeSetting.MaxUnavailable > 0 {
		state["max_unavailable"] = upgradeSetting.MaxUnavailable
	}
	if upgradeSetting.BlueGreenSettings != nil {
		bGSettings := make([]interface{}, 0, 1)

		blueGreenSettings := make(map[string]interface{})
		if upgradeSetting.BlueGreenSettings.StandardRolloutPolicy != nil {
			sRP := make([]interface{}, 0, 1)
			rolloutPolicy := make(map[string]interface{})
			if upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchPercentage > 0 {
				rolloutPolicy["batch_percentage"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchPercentage
			}
			if upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchNodeCount > 0 {
				rolloutPolicy["batch_node_count"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchNodeCount
			}
			rolloutPolicy["batch_soak_duration"] = upgradeSetting.BlueGreenSettings.StandardRolloutPolicy.BatchSoakDuration
			sRP = append(sRP, rolloutPolicy)
			blueGreenSettings["standard_rollout_policy"] = sRP
		}
		blueGreenSettings["node_pool_soak_duration"] = upgradeSetting.BlueGreenSettings.NodePoolSoakDuration
		bGSettings = append(bGSettings, blueGreenSettings)
		state["blue_green_settings"] = bGSettings
	}
	us = append(us, state)
	return us
}

func gcpNodePoolLoggingConfigToState(loggingConfig *duplosdk.GCPLoggingConfig) []map[string]interface{} {
	state := make(map[string]interface{})
	if loggingConfig != nil && loggingConfig.VariantConfig != nil {
		variant := make(map[string]interface{})
		variant["variant"] = loggingConfig.VariantConfig.Variant
		state["variant_config"] = variant
	}
	return []map[string]interface{}{state}
}

func gcpNodePoolLinuxConfigToState(linuxConfig *duplosdk.GCPLinuxNodeConfig) []map[string]interface{} {
	state := make(map[string]interface{})
	if linuxConfig != nil {
		state["cgroup_mode"] = []string{linuxConfig.CGroupMode}
		state["sysctls"] = linuxConfig.SysCtls
	}
	return []map[string]interface{}{state}
}

func gcpNodePoolAcceleratortoState(accelerator *duplosdk.Accelerator) []map[string]interface{} {
	if accelerator == nil {
		return nil
	}
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
	state := make([]interface{}, 0, len(taints))
	for _, t := range taints {
		data := map[string]interface{}{
			"key":    t.Key,
			"value":  t.Value,
			"effect": t.Effect,
		}
		state = append(state, data)
	}
	return state
}

func resourceGcpNodePoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpNodePoolDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	tenantID := idParts[0]
	fullName := idParts[1]
	resp, err := c.GCPK8NodePoolGet(tenantID, fullName)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s gcp node pool %s : %s", tenantID, resp.Name, err)
	}

	err = c.GCPK8NodePoolDelete(tenantID, fullName)
	if err != nil {
		return diag.Errorf("Error deleting node pool '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the node pool.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "gcp node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.GCPK8NodePoolGet(tenantID, fullName)
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
	tenantID, fullName := idParts[0], idParts[1]
	rq, err := expandGCPNodePool(d)
	if err != nil {
		return diag.Errorf("Error fetching request for %s : %s", tenantID, err.Error())

	}
	c := m.(*duplosdk.Client)
	if d.HasChange("zones") {
		_, err = gcpNodePoolZoneUpdate(c, tenantID, fullName, rq.Zones)
		if err != nil {
			return diag.Errorf("Error updating request for %s : %s", tenantID, err.Error())

		}
	}
	if d.HasChange("image_type") {
		_, err = gcpNodePoolImageTypeUpdate(c, tenantID, fullName, rq.ImageType)
		if err != nil {
			return diag.Errorf("Error updating request for %s : %s", tenantID, err.Error())

		}
	}

	if d.HasChange("taints") || d.HasChange("labels") || d.HasChange("tags") || d.HasChange("resource_labels") {
		_, err = gcpNodePoolUpdateTaintAndTags(c, tenantID, fullName, rq)
		if err != nil {
			return diag.Errorf("Error updating request for %s : %s", tenantID, err.Error())

		}
	}
	if d.HasChange("upgrade_settings") {
		_, err = gcpNodePoolUpdateUpgradeSetting(c, tenantID, fullName, rq.UpgradeSettings)
		if err != nil {
			return diag.Errorf("Error updating request for %s : %s", tenantID, err.Error())

		}
	}
	if d.HasChange("initial_node_count") {
		_, err = gcpNodePoolUpdateInitialNodeCount(c, tenantID, fullName, rq.InitialNodeCount)
		if err != nil {
			return diag.Errorf("Error updating request for %s : %s", tenantID, err.Error())

		}
	}

	err = gcpNodePoolAutoScalingUpdate(c, tenantID, fullName, d, *rq)
	if err != nil {
		return diag.Errorf("error: %s", err.Error())
	}
	duplo, err := c.GCPK8NodePoolGet(tenantID, fullName)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s GCP Node Pool Domain '%s': %s", tenantID, fullName, err)
	}

	if duplo == nil {
		d.SetId("") // object missing or deleted
		return nil
	}

	setGCPNodePoolStateField(d, duplo, tenantID)
	log.Printf("[TRACE] resourceGCPK8NodePoolUpdate ******** end")

	return nil
}

func gcpNodePoolAutoScalingUpdate(c *duplosdk.Client, tenantID, fullName string, d *schema.ResourceData, r duplosdk.DuploGCPK8NodePool) error {

	if d.HasChange("is_autoscaling_enabled") {
		isAutoScalingEnabled := r.IsAutoScalingEnabled
		rq, err := autoScalingHelper(isAutoScalingEnabled, &r)
		if err != nil {
			return err
		}
		_, err = c.GCPK8NodePoolUpdate(tenantID, fullName, "/autoscaling", rq)
		if err != nil {
			return err
		}
	}

	return nil
}

func autoScalingHelper(isAutoScalingEnabled bool, r *duplosdk.DuploGCPK8NodePool) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		IsAutoScalingEnabled: isAutoScalingEnabled,
		MaxNodeCount:         nil,
		MinNodeCount:         nil,
		TotalMaxNodeCount:    nil,
		TotalMinNodeCount:    nil,
	}
	if isAutoScalingEnabled {
		if r.MaxNodeCount != nil && r.MinNodeCount != nil {
			rq.MaxNodeCount = r.MaxNodeCount
			rq.MinNodeCount = r.MinNodeCount
		} else if r.TotalMaxNodeCount != nil && r.TotalMinNodeCount != nil {
			rq.TotalMaxNodeCount = r.TotalMaxNodeCount
			rq.TotalMinNodeCount = r.TotalMinNodeCount
		} else {
			return nil, fmt.Errorf("on autoscaling enabled set (max_node_count, min_node_count) or (total_max_node_count, total_min_node_count)")
		}
	}
	return rq, nil
}

func gcpNodePoolUpdateInitialNodeCount(c *duplosdk.Client, tenantID, fullName string, nodeCount int) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		InitialNodeCount: nodeCount,
	}
	resp, err := c.GCPK8NodePoolUpdate(tenantID, fullName, "/nodeCount", rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func gcpNodePoolUpdateUpgradeSetting(c *duplosdk.Client, tenantID, fullName string, upgradeSetting *duplosdk.GCPNodeUpgradeSetting) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		UpgradeSettings: upgradeSetting,
	}
	resp, err := c.GCPK8NodePoolUpdate(tenantID, fullName, "/upgradeSettings", rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func gcpNodePoolUpdateTaintAndTags(c *duplosdk.Client, tenantID, fullName string, rq *duplosdk.DuploGCPK8NodePool) (*duplosdk.DuploGCPK8NodePool, error) {

	resp, err := c.GCPK8NodePoolUpdate(tenantID, fullName, "", rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func gcpNodePoolImageTypeUpdate(c *duplosdk.Client, tenantID, fullName string, imageType string) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		ImageType: imageType,
	}
	resp, err := c.GCPK8NodePoolUpdate(tenantID, fullName, "/ImageType", rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func gcpNodePoolZoneUpdate(c *duplosdk.Client, tenantID, fullName string, zones []string) (*duplosdk.DuploGCPK8NodePool, error) {
	rq := &duplosdk.DuploGCPK8NodePool{
		Zones: zones,
	}
	resp, err := c.GCPK8NodePoolUpdate(tenantID, fullName, "/zones", rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func filterOutDefaultTags(tags []string) []string {
	return trimStringsByPosition(tags, 3)
}

func filterOutDefaultLabels(labels map[string]string) map[string]string {
	delete(labels, "tenantname")
	delete(labels, "duplo-tenant")
	return labels
}

func filterOutDefaultOAuth(oAuths []string) []string {
	oauthMap := map[string]struct{}{
		//"https://www.googleapis.com/auth/compute":              {},
		//"https://www.googleapis.com/auth/devstorage.read_only": {},
		//"https://www.googleapis.com/auth/logging.write":        {},
		//"https://www.googleapis.com/auth/monitoring":           {},
	}
	filters := []string{}
	for _, oAuth := range oAuths {
		if _, ok := oauthMap[oAuth]; !ok {
			filters = append(filters, oAuth)
		}
	}
	return filters
}

func filterOutDefaultResourceLabels(labels map[string]string) map[string]string {
	delete(labels, "duplo-tenant")
	return labels
}
