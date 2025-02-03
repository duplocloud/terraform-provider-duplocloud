package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceGCPNodePools() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_node_pools` retrieves list of node pools in Duplo.",

		ReadContext: dataSourceGCPNodePoolList,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the node pool will be associated with.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"node_pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataGcpK8NodePoolsFunctionSchema(),
				},
			},
		},
	}
}

func dataGcpK8NodePoolsFunctionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The short name of the node pool.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"fullname": {
			Description: "The full name of the node pool.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"machine_type": {
			Description: `The name of a Google Compute Engine machine type.
				If unspecified, the default machine type is e2-medium.`,
			Type:     schema.TypeString,
			Computed: true,
		},
		"disc_size_gb": {
			Description: `Size of the disk attached to each node, specified in GB. The smallest allowed disk size is 10GB.
				If unspecified, the default disk size is 100GB.`,
			Type:     schema.TypeInt,
			Computed: true,
		},
		"image_type": {
			Description: "The image type to use for this node. Note that for a given image type, the latest version of it will be used",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Description: `The list of instance tags applied to all nodes.
				Tags are used to identify valid sources or targets for network firewalls and are specified by the client during cluster or node pool creation.
				Each tag within the list must comply with RFC1035.`,
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"disc_type": {
			Description: `Type of the disk attached to each node
				If unspecified, the default disk type is 'pd-standard'`,
			Type:     schema.TypeString,
			Computed: true,
		},
		"spot": {
			Description: "Spot flag for enabling Spot VM",
			Type:        schema.TypeBool,
			Computed:    true,
		},

		"linux_node_config": {
			Description: "Parameters that can be configured on Linux nodes",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cgroup_mode": {
						Description: "cgroupMode specifies the cgroup mode to be used on the node.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"sysctls": {
						Description: "The Linux kernel parameters to be applied to the nodes and all pods running on the nodes.",
						Type:        schema.TypeMap,
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
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"is_autoscaling_enabled": {
			Description: "Is autoscaling enabled for this node pool.",
			Type:        schema.TypeBool,
			Computed:    true,
		},

		"location_policy": {
			Description: "Update strategy of the node pool.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"initial_node_count": {
			Description: "The initial node count for the pool",
			Type:        schema.TypeInt,
			Computed:    true,
		},

		"max_node_count": {
			Description: "Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.",
			Type:        schema.TypeInt,
			Computed:    true,
		},

		"min_node_count": {
			Description: "Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.",
			Type:        schema.TypeInt,
			Computed:    true,
		},

		"total_max_node_count": {
			Description: "Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"total_min_node_count": {
			Description: "Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"auto_upgrade": {
			Description: "Whether the nodes will be automatically upgraded.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"auto_repair": {
			Description: "Whether the nodes will be automatically repaired.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"metadata": {
			Description: "The metadata key/value pairs assigned to instances in the cluster.",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"resource_labels": {
			Description: "Resource labels associated to node pool.",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"node_pool_logging_config": {
			Description: "Logging configuration.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"variant_config": {
						Type:     schema.TypeMap,
						Computed: true,
						Elem: &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
		"oauth_scopes": {
			Description: "The set of Google API scopes to be made available on all of the node VMs under the default service account.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"zones": {
			Description: "The list of Google Compute Engine zones in which the NodePool's nodes should be located.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString}, // Define the type for each element in the list
		},
		"upgrade_settings": {
			Description: "Upgrade settings control disruption and speed of the upgrade.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"strategy": {
						Description: "Update strategy of the node pool.",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"max_surge": {
						Description: "",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"max_unavailable": {
						Description: "",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"blue_green_settings": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"standard_rollout_policy": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"batch_percentage": {
												Type:     schema.TypeFloat,
												Computed: true,
											},
											"batch_node_count": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"batch_soak_duration": {
												Type:     schema.TypeString,
												Computed: true,
											},
										},
									},
								},
								"node_pool_soak_duration": {
									Type:     schema.TypeString,
									Computed: true,
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
						Computed: true,
					},
					"value": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"effect": {
						Description: "Update strategy of the node pool.",
						Type:        schema.TypeString,
						Computed:    true,
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
						Computed:    true,
					},
					"gpu_partition_size": {
						Description: "Size of partitions to create on the GPU",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"max_time_shared_clients_per_gpu": {
						Description: "The number of time-shared GPU resources to expose for each physical GPU.",
						Type:        schema.TypeString,
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
									Computed:    true,
								},
								"gpu_sharing_strategy": {
									Description: "The configuration for GPU sharing options.",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
					},
					"gpu_driver_installation_config": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"gpu_driver_version": {
									Type:     schema.TypeString,
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

func dataSourceGCPNodePoolList(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGCPNodePoolList ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	duploList, err := c.GCPK8NodePoolList(tenantID)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s GCP Node Pool list: %s", tenantID, err)
	}

	if duploList == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.SetId(tenantID)

	list := make([]map[string]interface{}, 0, len(*duploList))

	for _, duplo := range *duploList {
		nodepool := setGCPNodePoolStateFieldList(&duplo)
		list = append(list, nodepool)
	}
	d.Set("node_pools", list)

	log.Printf("[TRACE] dataSourceGCPNodePoolList ******** end")
	return nil
}

func setGCPNodePoolStateFieldList(duplo *duplosdk.DuploGCPK8NodePool) map[string]interface{} {
	// Set simple fields first.
	return map[string]interface{}{
		"name":                     getGCPNodePoolShortName(duplo.Name, duplo.ResourceLabels["duplo-tenant"]),
		"fullname":                 duplo.Name,
		"is_autoscaling_enabled":   duplo.IsAutoScalingEnabled,
		"auto_upgrade":             duplo.AutoUpgrade,
		"zones":                    duplo.Zones,
		"image_type":               strings.ToLower(duplo.ImageType),
		"location_policy":          duplo.LocationPolicy,
		"max_node_count":           duplo.MaxNodeCount,
		"min_node_count":           duplo.MinNodeCount,
		"total_max_node_count":     duplo.TotalMaxNodeCount,
		"total_min_node_count":     duplo.TotalMinNodeCount,
		"initial_node_count":       duplo.InitialNodeCount,
		"auto_repair":              duplo.AutoRepair,
		"disc_size_gb":             duplo.DiscSizeGb,
		"disc_type":                duplo.DiscType,
		"machine_type":             duplo.MachineType,
		"metadata":                 duplo.Metadata,
		"labels":                   filterOutDefaultLabels(duplo.Labels),
		"spot":                     duplo.Spot,
		"tags":                     filterOutDefaultTags(duplo.Tags),
		"taints":                   gcpNodePoolTaintstoState(duplo.Taints),
		"node_pool_logging_config": gcpNodePoolLoggingConfigToState(duplo.LoggingConfig),
		"linux_node_config":        gcpNodePoolLinuxConfigToState(duplo.LinuxNodeConfig),
		"upgrade_settings":         gcpNodePoolUpgradeSettingToState(duplo.UpgradeSettings),
		"accelerator":              gcpNodePoolAcceleratortoState(duplo.Accelerator),
		"oauth_scopes":             filterOutDefaultOAuth(duplo.OauthScopes),
		"resource_labels":          filterOutDefaultResourceLabels(duplo.ResourceLabels),
	}
	// Set more complex fields next.
}
