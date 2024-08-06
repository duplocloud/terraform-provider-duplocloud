package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

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
			Description: "The name of the aks.",
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

func expandAzureK8sCluster(d *schema.ResourceData) *duplosdk.AksConfig {
	body := &duplosdk.AksConfig{
		Name:              d.Get("name").(string),
		PrivateCluster:    d.Get("private_cluster_enabled").(bool),
		K8sVersion:        d.Get("kubernetes_version").(string),
		VmSize:            d.Get("vm_size").(string),
		NetworkPlugin:     d.Get("network_plugin").(string),
		OutboundType:      d.Get("outbound_type").(string),
		NodeResourceGroup: d.Get("resource_group_name").(string),
		CreateAndManage:   true,
	}
	if body.Name == "" {
		body.Name = d.Get("infra_name").(string)
	}

	if body.NodeResourceGroup == "" {
		body.Name = d.Get("infra_name").(string)
	}
	return body
}

func flattenAzureK8sCluster(d *schema.ResourceData, duplo *duplosdk.AksConfig) {
	d.Set("name", duplo.Name)
	d.Set("private_cluster_enabled", duplo.PrivateCluster)
	d.Set("kubernetes_version", duplo.K8sVersion)
	d.Set("vm_size", duplo.VmSize)
	d.Set("network_plugin", duplo.NetworkPlugin)
	d.Set("outbound_type", duplo.OutboundType)
	d.Set("resource_group_name", duplo.NodeResourceGroup)
}
