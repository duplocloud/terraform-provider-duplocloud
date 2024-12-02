package duplocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAgentK8NodePoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure node pool will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"identifier": {
			Description:   "Node pool identifier. This field is deprecated, Please use `name` instead.",
			Type:          schema.TypeInt,
			Computed:      true,
			Optional:      true,
			ValidateFunc:  validation.IntInSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			Deprecated:    "identifier has been deprecated instead use name",
			ConflictsWith: []string{"name"},
		},
		"name": {
			Description:   "The name of the node pool.",
			Type:          schema.TypeString,
			Computed:      true,
			Optional:      true,
			ConflictsWith: []string{"identifier"},
		},
		"min_capacity": {
			Description: "The minimum number of nodes which should exist within this Node Pool.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"max_capacity": {
			Description: "The maximum number of nodes which should exist within this Node Pool.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"desired_capacity": {
			Description: "The initial number of nodes which should exist within this Node.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"enable_auto_scaling": {
			Description: "Whether to enable auto-scaler.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"vm_size": {
			Description: "The SKU which should be used for the Virtual Machines used in this Node Pool. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"allocation_tag": {
			Description: "Allocation tags for this node pool.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"scale_priority": {
			Description: "specify the priority for scaling operations,supported priority Regular or Spot",
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"priority": {
						Description: "priority levels Regular/Spot",
						Optional:    true,
						Computed:    true,
						Type:        schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{
							"Regular",
							"Spot",
						}, false),
						ForceNew: true,
					},
					"eviction_policy": {
						Description: "eviction policies Delete/Deallocate. Default value is Delete",
						Optional:    true,
						Computed:    true,
						Type:        schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{
							"Delete",
							"Deallocate",
						}, false),
						ForceNew: true,
					},
					"spot_max_price": {
						Description:  " for spot VMs sets the maximum price you're willing to pay, controlling costs, while priority.spot determines the scaling order of spot VM pools.",
						Optional:     true,
						Computed:     true,
						Type:         schema.TypeFloat,
						ValidateFunc: validation.FloatAtLeast(0.00001),
					},
				},
			},
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until node pool to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"availability_zones": {
			Description: "availability zones of node pool",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			Elem: &schema.Schema{Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"1", "2", "3"}, true),
			},
		},
		"os_type": {
			Description: "Specifies the OS used by the agent pool. Possible values are `Linux`, `Windows`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Linux",
			ValidateFunc: validation.StringInSlice([]string{
				"Linux",
				"Windows",
			}, false),
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"node_taints": {
			Description:      "A list of Kubernetes taints which should be applied to nodes in the agent pool.",
			Type:             schema.TypeList,
			Optional:         true,
			Computed:         true,
			Elem:             &schema.Schema{Type: schema.TypeString},
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"node_labels": {
			Description:      "Kubernetes labels which should be applied to nodes in this Node Pool.",
			Type:             schema.TypeList,
			Optional:         true,
			Computed:         true,
			Elem:             KeyValueSchema(),
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
	}
}

func resourceAzureK8NodePool() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_k8_node_pool` manages a Node Pool of Kubernetes Cluster in Duplo.",

		ReadContext:   resourceAgentK8NodePoolRead,
		CreateContext: resourceAgentK8NodePoolCreate,
		UpdateContext: resourceAgentK8NodePoolUpdate,
		DeleteContext: resourceAgentK8NodePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAgentK8NodePoolSchema(),
		CustomizeDiff: validateScalePriorityAttribute,
	}
}

func resourceAgentK8NodePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, friendlyName, err := parseAgentK8NodePoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAgentK8NodePoolRead(%s, %s): start", tenantID, friendlyName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureK8NodePoolGet(tenantID, friendlyName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure node pool %s : %s", tenantID, friendlyName, clientErr)
	}
	d.Set("tenant_id", tenantID)
	flattenAgentK8NodePool(d, duplo)

	log.Printf("[TRACE] resourceAgentK8NodePoolRead(%s, %s): end", tenantID, friendlyName)
	return nil
}

func resourceAgentK8NodePoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	identifier := useNameElseIdentifier(d)

	log.Printf("[TRACE] resourceAgentK8NodePoolCreate(%s, %s): start", tenantID, identifier)
	c := m.(*duplosdk.Client)

	rq, err := expandAgentK8NodePool(d, identifier)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure node pool '%s': %s", tenantID, identifier, err)
	}
	friendlyName, err := c.AzureK8NodePoolCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure node pool '%s': %s", tenantID, identifier, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, *friendlyName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureK8NodePoolGet(tenantID, *friendlyName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = azureK8NodePoolWaitUntilReady(ctx, c, tenantID, *friendlyName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAgentK8NodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceAgentK8NodePoolCreate(%s, %s): end", tenantID, *friendlyName)
	return diags
}

func resourceAgentK8NodePoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	id := d.Id()
	// Parse ID parts
	tenantID, name, err := parseAgentK8NodePoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAgentK8NodePoolUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	nodepool := &duplosdk.DuploAzureK8NodePoolRequest{
		FriendlyName:      name,
		MinSize:           d.Get("min_capacity").(int),
		MaxSize:           d.Get("max_capacity").(int),
		DesiredCapacity:   d.Get("desired_capacity").(int),
		EnableAutoScaling: d.Get("enable_auto_scaling").(bool),
		Capacity:          d.Get("vm_size").(string),
	}
	if d.HasChanges("min_capacity", "max_capacity", "desired_capacity", "enable_auto_scaling", "vm_size") {
		friendlyName, err := c.AzureK8NodePoolCreate(tenantID, nodepool)
		if err != nil {
			return diag.Errorf("Error updating tenant %s azure node pool '%v': %s", tenantID, friendlyName, err)
		}

		d.SetId(fmt.Sprintf("%s/%s", tenantID, *friendlyName))

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure node pool", id, func() (interface{}, duplosdk.ClientError) {
			return c.AzureK8NodePoolGet(tenantID, *friendlyName)
		})
		if diags != nil {
			return diags
		}
		if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
			err := azureK8NodePoolWaitUntilReady(ctx, c, tenantID, *friendlyName, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	diags := resourceAgentK8NodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceAgentK8NodePoolUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAgentK8NodePoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAgentK8NodePoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAgentK8NodePoolDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.AzureK8NodePoolDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure node pool '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure node pool", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.AzureK8NodePoolExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAgentK8NodePoolDelete(%s, %s): end", tenantID, name)
	return nil
}

func useNameElseIdentifier(d *schema.ResourceData) string {
	identifier := d.Get("identifier").(int)
	name := d.Get("name").(string)
	if name != "" {
		return name
	}
	return strconv.Itoa(identifier)
}

func expandAgentK8NodePool(d *schema.ResourceData, identifier string) (*duplosdk.DuploAzureK8NodePoolRequest, error) {
	nodePool := &duplosdk.DuploAzureK8NodePoolRequest{
		MinSize:           d.Get("min_capacity").(int),
		MaxSize:           d.Get("max_capacity").(int),
		DesiredCapacity:   d.Get("desired_capacity").(int),
		EnableAutoScaling: d.Get("enable_auto_scaling").(bool),
		FriendlyName:      identifier,
		Capacity:          d.Get("vm_size").(string),
	}

	if v, ok := d.GetOk("allocation_tag"); ok {
		nodePool.CustomDataTags = &[]duplosdk.DuploKeyStringValue{
			{
				Key:   "AllocationTags",
				Value: v.(string),
			},
		}
	}
	if v, ok := d.GetOk("scale_priority"); ok {
		scalePriority := v.([]interface{})
		if len(scalePriority) > 0 {
			for _, mp := range scalePriority {
				data := mp.(map[string]interface{})
				nodePool.ScaleSetPriority = data["priority"].(string)
				nodePool.ScaleSetEvictionPolicy = data["eviction_policy"].(string)
				spotPrice := data["spot_max_price"].(float64)
				if spotPrice > 0 {
					nodePool.SpotMaxPrice = float32(spotPrice)
				}
			}
		}
	}
	if v, ok := d.GetOk("availability_zones"); ok && v != nil {
		azs := v.(*schema.Set)
		for _, v := range azs.List() {
			nodePool.AvailabilityZones = append(nodePool.AvailabilityZones, v.(string))
		}
	}
	if v, ok := d.GetOk("os_type"); ok && v != nil {
		if v.(string) == "Windows" {
			nodePool.K8sWorkerOs = 1
		} else {
			nodePool.K8sWorkerOs = 0
		}
	}
	if v, ok := d.GetOk("node_taints"); ok && v != nil {
		nodetaints := v.([]interface{})
		if len(nodetaints) > 0 {
			for _, v := range nodetaints {
				nodePool.NodeTaints = append(nodePool.NodeTaints, v.(string))
			}
		}
	}
	if v, ok := d.GetOk("node_labels"); ok && v != nil {
		labels := v.([]interface{})
		if len(labels) > 0 {
			nodePool.NodeLabels = keyValueFromState("node_labels", d)
		}
	}

	return nodePool, nil
}

func parseAgentK8NodePoolIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAgentK8NodePool(d *schema.ResourceData, duplo *duplosdk.DuploAzureK8NodePool) {
	d.Set("min_capacity", duplo.MinSize)
	d.Set("max_capacity", duplo.MaxSize)
	d.Set("desired_capacity", duplo.DesiredCapacity)
	d.Set("enable_auto_scaling", duplo.EnableAutoScaling)
	d.Set("vm_size", duplo.Capacity)
	i := 0
	i, _ = strconv.Atoi(string(duplo.FriendlyName[len(duplo.FriendlyName)-1:]))
	if i > 0 {
		d.Set("identifier", i)
	}

	d.Set("name", duplo.FriendlyName)
	if len(duplo.CustomDataTags) > 0 {
		for _, kv := range duplo.CustomDataTags {
			if kv.Key == "AllocationTags" {
				d.Set("allocation_tag", kv.Value)
				break
			}
		}
	}
	if duplo.ScaleSetPriority != "" {
		mp := map[string]interface{}{}
		mp["priority"] = duplo.ScaleSetPriority
		mp["eviction_policy"] = duplo.ScaleSetEvictionPolicy
		mp["spot_max_price"] = duplo.SpotMaxPrice
	}
	if len(duplo.AvailabilityZones) > 0 {
		d.Set("availability_zones", duplo.AvailabilityZones)
	}
	if duplo.K8sWorkerOs == 1 {
		d.Set("os_type", "Windows")
	} else {
		d.Set("os_type", "Linux")
	}
	if len(duplo.NodeTaints) > 0 {
		d.Set("node_taints", duplo.NodeTaints)
	}
	if duplo.NodeLabels != nil && len(*duplo.NodeLabels) > 0 {
		d.Set("node_labels", keyValueToState("node_labels", duplo.NodeLabels))
	}

}

func azureK8NodePoolWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AzureK8NodePoolGet(tenantID, name)
			status := "pending"
			if err == nil {
				if rp != nil && rp.ProvisioningState == "Succeeded" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] azureK8NodePoolWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func validateScalePriorityAttribute(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	scalePriority := diff.Get("scale_priority").([]interface{})
	if len(scalePriority) == 0 {
		return nil
	}
	mp := scalePriority[0].(map[string]interface{})
	if mp["priority"].(string) == "Regular" && mp["spot_max_price"].(float64) != 0 {
		return errors.New("Scale Priority of type Regular does not support Spot max price")
	}
	if mp["priority"].(string) == "Regular" && mp["eviction_policy"].(string) != "" {
		return errors.New("Scale Priority of type Regular does not support Eviction Policy")
	}
	if mp["priority"].(string) == "Spot" && mp["eviction_policy"].(string) == "" {
		smp := make(map[string]interface{})
		smp["eviction_policy"] = "Delete"
		smp["priority"] = "Spot"
		smp["spot_max_price"] = mp["spot_max_price"]
		p := make([]interface{}, 0, 1)
		p = append(p, smp)
		err := diff.SetNew("scale_priority", p)
		if err != nil {
			return err
		}
	}

	return nil
}
