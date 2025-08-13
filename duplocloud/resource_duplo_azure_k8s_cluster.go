package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureK8sClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description: "The name of the infrastructure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the aks. If not specified default name would be infra name",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"resource_group_name": {
			Description: "The name of the aks resource group.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},

		"vm_size": {
			Description: "The size of the Virtual Machine.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"kubernetes_version": {
			Description: "Version of Kubernetes specified when creating the AKS managed cluster.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"network_plugin": {
			Description: "Network plugin to use for networking. Valid values are: `azure` and `kubenet`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"azure",
				"kubenet",
			}, false),
		},
		"outbound_type": {
			Description: "The outbound (egress) routing method which should be used for this Kubernetes Cluster. Valid values are: `loadBalancer` and `userDefinedRouting`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"loadBalancer",
				"userDefinedRouting",
			}, false),
		},
		"private_cluster_enabled": {
			Description: "Should this Kubernetes Cluster have its API server only exposed on internal IP addresses? This provides a Private IP Address for the Kubernetes API on the Virtual Network where the Kubernetes Cluster is located.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_workload_identity": {
			Description: "Enable Workload Identity for the AKS cluster. This allows Kubernetes workloads to access Azure resources using Azure AD identities.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_blob_csi_driver": {
			Description: "Enable the Azure Blob CSI driver for the AKS cluster. This allows Kubernetes workloads to use Azure Blob Storage as persistent storage.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"disable_run_command": {
			Description: "Disable the Run Command feature for the AKS cluster. This prevents the use of the Azure CLI to run commands directly on the nodes.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"add_critical_taint_to_system_agent_pool": {
			Description: "Add a critical taint to the system agent pool. This prevents the scheduler from scheduling non-critical pods on the system agent pool.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_image_cleaner": {
			Description: "Enable the image cleaner for the AKS cluster. This helps to clean up unused container images in the cluster.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"image_cleaner_interval_in_days": {
			Description: "Interval in days for the image cleaner to run. This determines how often the image cleaner will check for unused images.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     false,
		},

		"pricing_tier": {
			Description: "Pricing tier for the AKS cluster. Valid values are: `Free`, `Standard`, and `Premium`. This determines the level of support and features available for the AKS cluster.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"linux_admin_username": {
			Description: "The username for the Linux administrator of the AKS cluster. This user will have administrative access to the nodes in the cluster.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"linux_ssh_public_key": {
			Description: "The SSH public key for the Linux administrator of the AKS cluster. This key will be used to access the nodes in the cluster via SSH.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		/*Future enhancement
		"system_agent_pool_taints": {
				Description:      "Taints to be applied to the system agent pool.",
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressWhenNotCreating,
				Elem:             &schema.Schema{Type: schema.TypeString},
			},*/
		"active_directory_config": {
			Description: "Azure Active Directory configuration for the AKS cluster.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ad_tenant_id": {
						Description: "The Azure Active Directory tenant ID.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"enable_ad": {
						Description: "Enable Azure Active Directory integration.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"enable_rbac": {
						Description: "Enable Azure RBAC for Kubernetes authorization.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"admin_group_object_ids": {
						Description:      "List of Azure AD group object IDs that have admin access to the AKS cluster.",
						Type:             schema.TypeList,
						Optional:         true,
						DiffSuppressFunc: diffSuppressWhenNotCreating,
						Computed:         true,
						Elem:             &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
	}
}

func resourceAzureK8sCluster() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_k8s_cluster` manages an azure kubernetes cluster in Duplo.",

		ReadContext:   resourceAzureK8sClusterRead,
		CreateContext: resourceAzureK8sClusterCreate,
		UpdateContext: resourceAzureK8sClusterUpdate,
		DeleteContext: resourceAzureK8sClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureK8sClusterSchema(),
	}
}

func resourceAzureK8sClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	name := idParts[3]

	log.Printf("[TRACE] resourceAzureK8sClusterRead(%s): start", name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	config, err := c.InfrastructureGetConfig(name)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure '%s': %s", name, err)
	}
	if config == nil {
		d.SetId("") // object missing
		return nil
	}

	flattenAzureK8sCluster(d, config.AksConfig)
	d.Set("infra_name", name)
	log.Printf("[TRACE] resourceAzureK8sClusterRead(%s): end", name)
	return nil
}

func resourceAzureK8sClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	log.Printf("[TRACE] resourceAzureK8sClusterCreate(%s): start", infraName)
	c := m.(*duplosdk.Client)
	rq := expandAzureK8sCluster(d)
	err = c.AzureK8sClusterCreate(infraName, rq)
	if err != nil {
		return diag.Errorf("Error creating k8s azure cluster for infra %s': %s", infraName, err)
	}

	id := fmt.Sprintf("v2/admin/InfrastructureV2/%s", infraName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure k8s cluster", id, func() (interface{}, duplosdk.ClientError) {
		plan, err := c.PlanGet(infraName)
		if err != nil {
			return nil, err
		}
		if plan.K8ClusterConfigs != nil && len(*plan.K8ClusterConfigs) > 0 && len((*plan.K8ClusterConfigs)[0].ApiServer) > 0 {
			return plan, nil
		}
		return nil, nil
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureK8sClusterRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureK8sClusterCreate(%s): end", infraName)
	return diags
}

func resourceAzureK8sClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureK8sClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandAzureK8sCluster(d *schema.ResourceData) *duplosdk.DuploAksConfig {
	body := &duplosdk.DuploAksConfig{
		Name:                              d.Get("name").(string),
		PrivateCluster:                    d.Get("private_cluster_enabled").(bool),
		K8sVersion:                        d.Get("kubernetes_version").(string),
		VmSize:                            d.Get("vm_size").(string),
		NetworkPlugin:                     d.Get("network_plugin").(string),
		OutboundType:                      d.Get("outbound_type").(string),
		NodeResourceGroup:                 d.Get("resource_group_name").(string),
		CreateAndManage:                   true,
		EnableWorkloadIdentity:            d.Get("enable_workload_identity").(bool),
		EnableBlobCsiDriver:               d.Get("enable_blob_csi_driver").(bool),
		DisableRunCommand:                 d.Get("disable_run_command").(bool),
		AddCriticalTaintToSystemAgentPool: d.Get("add_critical_taint_to_system_agent_pool").(bool),
		EnableImageCleaner:                d.Get("enable_image_cleaner").(bool),
		ImageCleanerIntervalInDays:        d.Get("image_cleaner_interval_in_days").(int),
		PricingTier:                       d.Get("pricing_tier").(string),
		LinuxAdminUsername:                d.Get("linux_admin_username").(string),
		LinuxSshPublicKey:                 d.Get("linux_ssh_public_key").(string),
	}
	/*	if v, ok := d.GetOk("system_agent_pool_taints"); ok && len(v.([]interface{})) > 0 {
		for _, taint := range v.([]interface{}) {
			body.SystemAgentPoolTaints = append(body.SystemAgentPoolTaints, taint.(string))
		}
	}*/

	if body.Name == "" {
		body.Name = d.Get("infra_name").(string)
	}

	if v, ok := d.GetOk("active_directory_config"); ok && len(v.([]interface{})) > 0 {
		aadConfig := v.([]interface{})[0].(map[string]interface{})
		body.AadConfig = &duplosdk.DuploAksAadConfig{
			ADTenantId:          aadConfig["ad_tenant_id"].(string),
			IsManagedAadEnabled: aadConfig["enable_ad"].(bool),
			IsAzureRbacEnabled:  aadConfig["enable_rbac"].(bool),
		}
		for _, val := range aadConfig["admin_group_object_ids"].([]interface{}) {
			body.AadConfig.AdminGroupObjectIds = append(body.AadConfig.AdminGroupObjectIds, val.(string))
		}
	}
	return body
}

func flattenAzureK8sCluster(d *schema.ResourceData, duplo *duplosdk.DuploAksConfig) {
	d.Set("name", duplo.Name)
	d.Set("private_cluster_enabled", duplo.PrivateCluster)
	d.Set("kubernetes_version", duplo.K8sVersion)
	d.Set("vm_size", duplo.VmSize)
	d.Set("network_plugin", duplo.NetworkPlugin)
	d.Set("outbound_type", duplo.OutboundType)
	d.Set("resource_group_name", duplo.NodeResourceGroup)
	d.Set("enable_workload_identity", duplo.EnableWorkloadIdentity)
	d.Set("enable_blob_csi_driver", duplo.EnableBlobCsiDriver)
	d.Set("disable_run_command", duplo.DisableRunCommand)
	d.Set("add_critical_taint_to_system_agent_pool", duplo.AddCriticalTaintToSystemAgentPool)
	d.Set("enable_image_cleaner", duplo.EnableImageCleaner)
	d.Set("image_cleaner_interval_in_days", duplo.ImageCleanerIntervalInDays)
	d.Set("pricing_tier", duplo.PricingTier)
	d.Set("linux_admin_username", duplo.LinuxAdminUsername)
	d.Set("linux_ssh_public_key", duplo.LinuxSshPublicKey)
	/*
		if len(duplo.SystemAgentPoolTaints) > 0 {
			s := []interface{}{}
			for _, taint := range duplo.SystemAgentPoolTaints {
				s = append(s, taint)
			}
			d.Set("system_agent_pool_taints", s)
		} else {
			d.Set("system_agent_pool_taints", make([]interface{}, 0))
		}*/
	if duplo.AadConfig != nil {
		m := map[string]interface{}{
			"ad_tenant_id":           duplo.AadConfig.ADTenantId,
			"enable_ad":              duplo.AadConfig.IsManagedAadEnabled,
			"enable_rbac":            duplo.AadConfig.IsAzureRbacEnabled,
			"admin_group_object_ids": duplo.AadConfig.AdminGroupObjectIds,
		}
		if len(duplo.AadConfig.AdminGroupObjectIds) > 0 {
			s := []interface{}{}
			for _, conf := range duplo.AadConfig.AdminGroupObjectIds {
				s = append(s, conf)
			}
			m["system_agent_pool_taints"] = s
		} else {
			m["system_agent_pool_taints"] = make([]interface{}, 0)
		}
		d.Set("active_directory_config", []interface{}{m})

	}

}
