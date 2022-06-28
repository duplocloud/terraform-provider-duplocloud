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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploOciNodePoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the cloudwatch event target will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"node_pool_id": {
			Description: "The OCID of the node pool.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cluster_id": {
			Description: "The OCID of the cluster to which this node pool is attached.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"compartment_id": {
			Description: "The OCID of the compartment in which the node pool exists.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kubernetes_version": {
			Description: "The version of Kubernetes to install on the nodes in the node pool.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"name": {
			Description: "The name of the node pool.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"node_shape": {
			Description: "The name of the node shape of the nodes in the node pool.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"defined_tags": {
			Description:      "Defined tags for this resource. Each key is predefined and scoped to a namespace.",
			Type:             schema.TypeMap,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem:             schema.TypeString,
		},
		"freeform_tags": {
			Description:      "Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.",
			Type:             schema.TypeMap,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem:             schema.TypeString,
		},
		"wait_until_ready": {
			Description:      "Whether or not to wait until oci node pool to be ready, after creation.",
			Type:             schema.TypeBool,
			Optional:         true,
			Default:          true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
		},
		"initial_node_labels": {
			Type:             schema.TypeList,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// Required

					// Optional
					"key": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},

					// Computed
				},
			},
		},
		"node_config_details": {
			Type:          schema.TypeList,
			Optional:      true,
			Computed:      true,
			MaxItems:      1,
			MinItems:      1,
			ConflictsWith: []string{"quantity_per_subnet", "subnet_ids"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// Required
					"placement_configs": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								// Required
								"availability_domain": {
									Type:     schema.TypeString,
									Required: true,
									ForceNew: true,
								},
								"subnet_id": {
									Type:     schema.TypeString,
									Required: true,
									ForceNew: true,
								},

								// Optional
								"capacity_reservation_id": {
									Type:             schema.TypeString,
									Optional:         true,
									Computed:         true,
									DiffSuppressFunc: diffSuppressFuncIgnore,
								},

								// Computed
							},
						},
					},
					"size": {
						Type:     schema.TypeInt,
						Required: true,
						ValidateFunc: validation.Any(
							validation.IntAtLeast(0),
						),
					},

					// Optional
					"is_pv_encryption_in_transit_enabled": {
						Type:             schema.TypeBool,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: diffSuppressFuncIgnore,
					},
					"kms_key_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"defined_tags": {
						Type:             schema.TypeMap,
						Optional:         true,
						Elem:             schema.TypeString,
						DiffSuppressFunc: diffSuppressFuncIgnore,
					},
					"freeform_tags": {
						Type:             schema.TypeMap,
						Optional:         true,
						Elem:             schema.TypeString,
						DiffSuppressFunc: diffSuppressFuncIgnore,
					},
					"nsg_ids": {
						Type:             schema.TypeSet,
						Optional:         true,
						DiffSuppressFunc: diffSuppressFuncIgnore,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},

					// Computed
				},
			},
		},
		"node_image_id": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"node_image_name", "node_source_details"},
		},
		"node_image_name": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"node_image_id", "node_source_details"},
		},
		"node_metadata": {
			Type:             schema.TypeMap,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem:             schema.TypeString,
		},
		"node_shape_config": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// Required

					// Optional
					"memory_in_gbs": {
						Type:     schema.TypeFloat,
						Optional: true,
						ForceNew: true,
					},
					"ocpus": {
						Type:     schema.TypeFloat,
						Optional: true,
						ForceNew: true,
					},

					// Computed
				},
			},
		},
		"node_source_details": {
			Type:          schema.TypeList,
			Optional:      true,
			Computed:      true,
			MaxItems:      1,
			MinItems:      1,
			ConflictsWith: []string{"node_image_id", "node_image_name"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// Required
					"image_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"source_type": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"IMAGE",
						}, true),
					},

					// Optional
					"boot_volume_size_in_gbs": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},

					// Computed
				},
			},
		},
		"quantity_per_subnet": {
			Type:          schema.TypeInt,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"node_config_details"},
		},
		"ssh_public_key": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"subnet_ids": {
			Type:          schema.TypeSet,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"node_config_details"},
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"system_tags": {
			Type:             schema.TypeMap,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem:             schema.TypeString,
		},
		"nodes": {
			Type:             schema.TypeList,
			Computed:         true,
			Optional:         true,
			DiffSuppressFunc: diffSuppressFuncIgnore,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					// Required

					// Optional

					// Computed
					"availability_domain": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"defined_tags": {
						Type:     schema.TypeMap,
						Computed: true,
						Elem:     schema.TypeString,
					},
					"fault_domain": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"freeform_tags": {
						Type:     schema.TypeMap,
						Computed: true,
						Elem:     schema.TypeString,
					},
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"kubernetes_version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"lifecycle_details": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"node_pool_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"private_ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"public_ip": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"state": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"subnet_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"system_tags": {
						Type:     schema.TypeMap,
						Computed: true,
						Elem:     schema.TypeString,
					},
				},
			},
		},
	}
}

func resourceOciContainerEngineNodePool() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_oci_containerengine_node_pool` manages an OCI container node pool in Duplo.",

		ReadContext:   resourceOciNodePoolRead,
		CreateContext: resourceOciNodePoolCreate,
		UpdateContext: resourceOciNodePoolUpdate,
		DeleteContext: resourceOciNodePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploOciNodePoolSchema(),
	}
}

func resourceOciNodePoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseOciNodePoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceOciNodePoolRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.OciNodePoolGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s oci node pool %s : %s", tenantID, name, clientErr)
	}
	duploWithNodes, err := c.OciNodePoolGetWithNodes(tenantID, duplo.Id)
	if err != nil {
		return diag.FromErr(err)
	}
	flattenOciNodePool(d, duploWithNodes.NodePool)
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceOciNodePoolRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceOciNodePoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceOciNodePoolCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq, err := expandOciNodePool(d)
	if err != nil {
		return diag.Errorf("Error expanding tenant %s oci node pool '%s': %s", tenantID, name, err)
	}
	err = c.OciNodePoolCreate(tenantID, name, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s oci node pool '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	getResp, err := c.OciNodePoolGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Error getting tenant %s oci node pool '%s': %s", tenantID, name, err)
	}
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "oci node pool", id, func() (interface{}, duplosdk.ClientError) {
		return c.OciNodePoolGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = nodePoolWaitUntilReady(ctx, c, tenantID, getResp.Id, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceOciNodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceOciNodePoolCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceOciNodePoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// ideally for now only size updates are required for OCI nodepool scaling manually....
	// todo : (BE suports only scaling) handle more fields if required later

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceOciNodePoolUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	request := duplosdk.DuploOciNodePoolUpdateReq{}
	request.NodePoolId = d.Get("node_pool_id").(string)
	if nodeConfigDetails, ok := d.GetOk("node_config_details"); ok {
		if tmpList := nodeConfigDetails.([]interface{}); len(tmpList) > 0 {
			fieldKeyFormat := fmt.Sprintf("%s.%d.%%s", "node_config_details", 0)
			if size, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "size")); ok {
				request.Size = size.(int)
			}
		}
	}

	msg, clientErr := c.OciNodePoolUpdate(tenantID, request.NodePoolId, &request)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		log.Printf("[ERROR] resourceOciNodePoolUpdate(%s, %s, %s, %s): end", tenantID, name, msg, clientErr)
	}

	diags := resourceOciNodePoolRead(ctx, d, m)
	log.Printf("[TRACE] resourceOciNodePoolUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceOciNodePoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	nodePoolId := d.Get("node_pool_id").(string)
	tenantID, name, err := parseOciNodePoolIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceOciNodePoolDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.OciNodePoolDelete(tenantID, nodePoolId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s oci node pool '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "oci node pool", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.OciNodePoolExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceOciNodePoolDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandOciNodePool(d *schema.ResourceData) (*duplosdk.DuploOciNodePoolDetailsCreateReq, error) {
	request := duplosdk.DuploOciNodePoolDetailsCreateReq{}

	if definedTags, ok := d.GetOk("defined_tags"); ok {
		convertedDefinedTags, err := mapToDefinedTags(definedTags.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		request.DefinedTags = convertedDefinedTags
	}

	if freeformTags, ok := d.GetOk("freeform_tags"); ok {
		request.FreeformTags = objectMapToStringMap(freeformTags.(map[string]interface{}))
	}

	if initialNodeLabels, ok := d.GetOk("initial_node_labels"); ok {
		interfaces := initialNodeLabels.([]interface{})
		tmp := make([]duplosdk.DuploOciNodePoolKeyValue, len(interfaces))
		for i := range interfaces {
			stateDataIndex := i
			fieldKeyFormat := fmt.Sprintf("%s.%d.%%s", "initial_node_labels", stateDataIndex)
			converted, err := nodePoolMapToKeyValue(fieldKeyFormat, d)
			if err != nil {
				return nil, err
			}
			tmp[i] = converted
		}
		if len(tmp) != 0 || d.HasChange("initial_node_labels") {
			request.InitialNodeLabels = &tmp
		}
	}

	if kubernetesVersion, ok := d.GetOk("kubernetes_version"); ok {
		request.KubernetesVersion = kubernetesVersion.(string)
	}

	if name, ok := d.GetOk("name"); ok {
		request.Name = name.(string)
	}

	if nodeConfigDetails, ok := d.GetOk("node_config_details"); ok {
		if tmpList := nodeConfigDetails.([]interface{}); len(tmpList) > 0 {
			fieldKeyFormat := fmt.Sprintf("%s.%d.%%s", "node_config_details", 0)
			tmp, err := mapToCreateNodePoolNodeConfigDetails(fieldKeyFormat, d)
			if err != nil {
				return nil, err
			}
			request.NodeConfigDetails = &tmp
		}
	}

	if nodeImageId, ok := d.GetOk("node_image_id"); ok {
		request.NodeImageName = nodeImageId.(string)
	}

	if nodeImageName, ok := d.GetOk("node_image_name"); ok {
		request.NodeImageName = nodeImageName.(string)
	}

	if nodeMetadata, ok := d.GetOk("node_metadata"); ok {
		request.NodeMetadata = objectMapToStringMap(nodeMetadata.(map[string]interface{}))
	}

	if nodeShape, ok := d.GetOk("node_shape"); ok {
		request.NodeShape = nodeShape.(string)
	}

	if nodeShapeConfig, ok := d.GetOk("node_shape_config"); ok {
		if tmpList := nodeShapeConfig.([]interface{}); len(tmpList) > 0 {
			fieldKeyFormat := fmt.Sprintf("%s.%d.%%s", "node_shape_config", 0)
			tmp, err := mapToCreateNodeShapeConfigDetails(fieldKeyFormat, d)
			if err != nil {
				return nil, err
			}
			request.NodeShapeConfig = &tmp
		}
	}

	if nodeSourceDetails, ok := d.GetOk("node_source_details"); ok {
		if tmpList := nodeSourceDetails.([]interface{}); len(tmpList) > 0 {
			fieldKeyFormat := fmt.Sprintf("%s.%d.%%s", "node_source_details", 0)
			tmp, err := mapToNodeSourceDetails(fieldKeyFormat, d)
			if err != nil {
				return nil, err
			}
			request.NodeSourceDetails = &tmp
		}
	}

	if quantityPerSubnet, ok := d.GetOk("quantity_per_subnet"); ok {
		request.QuantityPerSubnet = quantityPerSubnet.(int)
	}

	if sshPublicKey, ok := d.GetOk("ssh_public_key"); ok && sshPublicKey != nil {
		request.SshPublicKey = sshPublicKey.(string)
	}

	if subnetIds, ok := d.GetOk("subnet_ids"); ok {
		set := subnetIds.(*schema.Set)
		request.SubnetIds = expandStringSet(set)
	}
	return &request, nil
}

func parseOciNodePoolIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenOciNodePool(d *schema.ResourceData, duplo *duplosdk.DuploOciNodePool) {
	d.Set("cluster_id", duplo.ClusterId)
	d.Set("compartment_id", duplo.CompartmentId)
	d.Set("node_pool_id", duplo.Id)

	if duplo.DefinedTags != nil && len(duplo.DefinedTags) > 0 {
		d.Set("defined_tags", definedTagsToMap(duplo.DefinedTags))
	}
	if duplo.FreeformTags != nil && len(duplo.FreeformTags) > 0 {
		d.Set("freeform_tags", duplo.FreeformTags)
	}

	initialNodeLabels := []interface{}{}
	for _, item := range *duplo.InitialNodeLabels {
		initialNodeLabels = append(initialNodeLabels, nodePoolkeyValueToMap(item))
	}
	if len(initialNodeLabels) > 0 {
		d.Set("initial_node_labels", initialNodeLabels)
	}
	if len(duplo.KubernetesVersion) > 0 {
		d.Set("kubernetes_version", duplo.KubernetesVersion)
	}

	if len(duplo.Name) > 0 {
		d.Set("name", duplo.Name)
	}

	if duplo.NodeConfigDetails != nil {
		d.Set("node_config_details", []interface{}{nodePoolNodeConfigDetailsToMap(duplo.NodeConfigDetails)})
	} else {
		d.Set("node_config_details", nil)
	}

	if len(duplo.NodeImageId) > 0 {
		d.Set("node_image_id", duplo.NodeImageId)
	}

	if len(duplo.NodeImageName) > 0 {
		d.Set("node_image_name", duplo.NodeImageName)
	}
	if duplo.NodeMetadata != nil && len(duplo.NodeMetadata) > 0 {
		d.Set("node_metadata", duplo.NodeMetadata)
	}
	if len(duplo.NodeShape) > 0 {
		d.Set("node_shape", duplo.NodeShape)
	}

	if duplo.NodeShapeConfig != nil {
		d.Set("node_shape_config", []interface{}{nodeShapeConfigToMap(duplo.NodeShapeConfig)})
	} else {
		d.Set("node_shape_config", nil)
	}

	if duplo.NodeSourceDetails != nil {
		nodeSourceDetailsArray := []interface{}{}
		if nodeSourceDetailsMap := nodeSourceDetailsToMap(duplo.NodeSourceDetails); nodeSourceDetailsMap != nil {
			nodeSourceDetailsArray = append(nodeSourceDetailsArray, nodeSourceDetailsMap)
		}
		d.Set("node_source_details", nodeSourceDetailsArray)
	} else {
		d.Set("node_source_details", nil)
	}

	d.Set("quantity_per_subnet", duplo.QuantityPerSubnet)

	if len(duplo.SshPublicKey) > 0 {
		d.Set("ssh_public_key", duplo.SshPublicKey)
	}

	if duplo.SubnetIds != nil {
		d.Set("subnet_ids", flattenStringSet(duplo.SubnetIds))
	} else {
		d.Set("subnet_ids", nil)
	}

	if duplo.SystemTags != nil && len(duplo.SystemTags) > 0 {
		d.Set("system_tags", systemTagsToMap(duplo.SystemTags))
	}

	nodes := []interface{}{}
	for _, item := range *duplo.Nodes {
		nodes = append(nodes, nodeToMap(item))
	}

	if len(nodes) > 0 {
		d.Set("nodes", nodes)
	}

}

func definedTagsToMap(definedTags map[string]map[string]interface{}) map[string]interface{} {
	var tags = make(map[string]interface{})
	if len(definedTags) > 0 {
		for namespace, keys := range definedTags {
			for key, value := range keys {
				tags[namespace+"."+key] = value
			}
		}
	}
	return tags
}

func systemTagsToMap(systemTags map[string]map[string]interface{}) map[string]interface{} {
	return definedTagsToMap(systemTags)
}

func nodePoolNodeConfigDetailsToMap(duplo *duplosdk.NodePoolNodeConfigDetails) map[string]interface{} {
	result := map[string]interface{}{}

	result["is_pv_encryption_in_transit_enabled"] = duplo.IsPvEncryptionInTransitEnabled

	if len(duplo.KmsKeyId) > 0 {
		result["kms_key_id"] = duplo.KmsKeyId
	}
	if duplo.DefinedTags != nil && len(duplo.DefinedTags) > 0 {
		result["defined_tags"] = definedTagsToMap(duplo.DefinedTags)
	}
	if duplo.FreeformTags != nil && len(duplo.FreeformTags) > 0 {
		result["freeform_tags"] = duplo.FreeformTags
	}

	if duplo.NsgIds != nil && len(duplo.NsgIds) > 0 {
		result["nsg_ids"] = flattenStringSet(duplo.NsgIds)
	}

	if len(*duplo.PlacementConfigs) > 0 {
		placementConfigs := []interface{}{}
		for _, item := range *duplo.PlacementConfigs {
			placementConfigs = append(placementConfigs, nodePoolPlacementConfigDetailsToMap(item))
		}
		result["placement_configs"] = placementConfigs
	}

	result["size"] = duplo.Size

	return result
}

func nodePoolPlacementConfigDetailsToMap(duplo duplosdk.NodePoolPlacementConfigDetails) map[string]interface{} {
	result := map[string]interface{}{}
	if len(duplo.AvailabilityDomain) > 0 {
		result["availability_domain"] = duplo.AvailabilityDomain
	}

	if len(duplo.CapacityReservationId) > 0 {
		result["capacity_reservation_id"] = duplo.CapacityReservationId
	}

	if len(duplo.SubnetId) > 0 {
		result["subnet_id"] = duplo.SubnetId
	}
	return result
}

func nodeShapeConfigToMap(duplo *duplosdk.NodeShapeConfig) map[string]interface{} {
	result := map[string]interface{}{}
	result["memory_in_gbs"] = duplo.MemoryInGBs
	result["ocpus"] = duplo.Ocpus
	return result
}

func nodeSourceDetailsToMap(duplo *duplosdk.NodeSourceDetails) map[string]interface{} {
	result := map[string]interface{}{}
	if len(duplo.SourceType) > 0 {
		result["source_type"] = duplo.SourceType
	}
	if len(duplo.ImageId) > 0 {
		result["image_id"] = duplo.ImageId
	}
	if duplo.BootVolumeSizeInGBs > 0 {
		result["boot_volume_size_in_gbs"] = strconv.FormatInt(duplo.BootVolumeSizeInGBs, 10)
	}
	return result
}

func mapToDefinedTags(rawMap map[string]interface{}) (map[string]map[string]interface{}, error) {
	definedTags := make(map[string]map[string]interface{})
	if len(rawMap) > 0 {
		for key, value := range rawMap {
			var keyComponents = strings.Split(key, ".")
			if len(keyComponents) != 2 {
				return nil, fmt.Errorf("invalid key structure found %s", key)
			}
			var namespace = keyComponents[0]
			if _, ok := definedTags[namespace]; !ok {
				definedTags[namespace] = make(map[string]interface{})
			}
			definedTags[namespace][keyComponents[1]] = value
		}
	}
	return definedTags, nil
}

func mapToCreateNodePoolNodeConfigDetails(fieldKeyFormat string, d *schema.ResourceData) (duplosdk.NodePoolNodeConfigDetails, error) {
	result := duplosdk.NodePoolNodeConfigDetails{}

	if isPvEncryptionInTransitEnabled, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "is_pv_encryption_in_transit_enabled")); ok {
		result.IsPvEncryptionInTransitEnabled = isPvEncryptionInTransitEnabled.(bool)
	}

	if kmsKeyId, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "kms_key_id")); ok {
		result.KmsKeyId = kmsKeyId.(string)
	}
	if definedTags, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "defined_tags")); ok {
		tmp, err := mapToDefinedTags(definedTags.(map[string]interface{}))
		if err != nil {
			return result, fmt.Errorf("unable to convert defined_tags, encountered error: %v", err)
		}
		result.DefinedTags = tmp
	}

	if freeformTags, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "freeform_tags")); ok {
		result.FreeformTags = objectMapToStringMap(freeformTags.(map[string]interface{}))
	}

	if nsgIds, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "nsg_ids")); ok {
		set := nsgIds.(*schema.Set)
		result.NsgIds = expandStringSet(set)
	}

	if placementConfigs, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "placement_configs")); ok {
		interfaces := placementConfigs.([]interface{})
		tmp := make([]duplosdk.NodePoolPlacementConfigDetails, len(interfaces))
		for i := range interfaces {
			fieldKeyFormatNextLevel := fmt.Sprintf("%s.%d.%%s", fmt.Sprintf(fieldKeyFormat, "placement_configs"), i)

			converted, err := mapToNodePoolPlacementConfigDetails(fieldKeyFormatNextLevel, d)
			if err != nil {
				return result, err
			}
			tmp[i] = converted
		}
		if len(tmp) > 0 {
			result.PlacementConfigs = &tmp
		}
	}

	if size, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "size")); ok {
		result.Size = size.(int)
	}

	return result, nil
}

func mapToNodePoolPlacementConfigDetails(fieldKeyFormat string, d *schema.ResourceData) (duplosdk.NodePoolPlacementConfigDetails, error) {
	result := duplosdk.NodePoolPlacementConfigDetails{}

	if availabilityDomain, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "availability_domain")); ok {
		result.AvailabilityDomain = availabilityDomain.(string)
	}

	if capacityReservationId, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "capacity_reservation_id")); ok {
		result.CapacityReservationId = capacityReservationId.(string)
	}

	if subnetId, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "subnet_id")); ok {
		result.SubnetId = subnetId.(string)
	}
	return result, nil
}

func mapToCreateNodeShapeConfigDetails(fieldKeyFormat string, d *schema.ResourceData) (duplosdk.NodeShapeConfig, error) {
	result := duplosdk.NodeShapeConfig{}

	if memoryInGBs, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "memory_in_gbs")); ok {
		result.MemoryInGBs = float32(memoryInGBs.(float64))
	}

	if ocpus, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "ocpus")); ok {
		result.Ocpus = float32(ocpus.(float64))
	}

	return result, nil
}

func mapToNodeSourceDetails(fieldKeyFormat string, d *schema.ResourceData) (duplosdk.NodeSourceDetails, error) {
	var details duplosdk.NodeSourceDetails
	sourceType, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "source_type"))
	if ok {
		details.SourceType = sourceType.(string)
	}
	if bootVolumeSizeInGBs, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "boot_volume_size_in_gbs")); ok {
		tmp := bootVolumeSizeInGBs.(string)
		if tmp != "" {
			tmpInt64, err := strconv.ParseInt(tmp, 10, 64)
			if err != nil {
				return details, fmt.Errorf("unable to convert bootVolumeSizeInGBs string: %s to an int64 and encountered error: %v", tmp, err)
			}
			details.BootVolumeSizeInGBs = tmpInt64
		}
	}
	if imageId, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "image_id")); ok {
		details.ImageId = imageId.(string)
	}
	return details, nil
}

func nodePoolMapToKeyValue(fieldKeyFormat string, d *schema.ResourceData) (duplosdk.DuploOciNodePoolKeyValue, error) {
	result := duplosdk.DuploOciNodePoolKeyValue{}
	if key, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "key")); ok {
		result.Key = key.(string)
	}
	if value, ok := d.GetOk(fmt.Sprintf(fieldKeyFormat, "value")); ok {
		result.Value = value.(string)
	}
	return result, nil
}

func nodePoolkeyValueToMap(duplo duplosdk.DuploOciNodePoolKeyValue) map[string]interface{} {
	result := map[string]interface{}{}

	if len(duplo.Key) > 0 {
		result["key"] = duplo.Key
	}

	if len(duplo.Value) > 0 {
		result["value"] = duplo.Value
	}

	return result
}

func nodeToMap(duplo duplosdk.Node) map[string]interface{} {
	result := map[string]interface{}{}

	if len(duplo.AvailabilityDomain) > 0 {
		result["availability_domain"] = duplo.AvailabilityDomain

	}
	if duplo.DefinedTags != nil && len(duplo.DefinedTags) > 0 {
		result["defined_tags"] = definedTagsToMap(duplo.DefinedTags)
	}

	if len(duplo.FaultDomain) > 0 {
		result["fault_domain"] = duplo.FaultDomain
	}
	if duplo.FreeformTags != nil && len(duplo.FreeformTags) > 0 {
		result["freeform_tags"] = duplo.FreeformTags
	}
	if len(duplo.Id) > 0 {
		result["id"] = duplo.Id
	}

	if len(duplo.KubernetesVersion) > 0 {
		result["kubernetes_version"] = duplo.KubernetesVersion
	}

	if len(duplo.LifecycleDetails) > 0 {
		result["lifecycle_details"] = duplo.LifecycleDetails
	}

	if len(duplo.Name) > 0 {
		result["name"] = duplo.Name
	}

	if len(duplo.NodePoolId) > 0 {
		result["node_pool_id"] = duplo.NodePoolId
	}

	if len(duplo.PrivateIp) > 0 {
		result["private_ip"] = duplo.PrivateIp
	}

	if len(duplo.PublicIp) > 0 {
		result["public_ip"] = duplo.PublicIp
	}

	result["state"] = duplo.LifecycleState

	if len(duplo.SubnetId) > 0 {
		result["subnet_id"] = duplo.SubnetId
	}

	if duplo.SystemTags != nil && len(duplo.SystemTags) > 0 {
		result["system_tags"] = systemTagsToMap(duplo.SystemTags)
	}

	return result
}

func nodePoolWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, ociId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.OciNodePoolGetWithNodes(tenantID, ociId)
			status := "pending"
			if rp != nil {
				for _, node := range *rp.NodePool.Nodes {
					if node.LifecycleState == duplosdk.NodeLifecycleStateActive {
						status = "ready"
					} else {
						status = "pending"
					}
					log.Printf("[TRACE] Node pool lifecycle state is (%s).", node.LifecycleState)
				}
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] nodePoolWaitUntilReady(%s, %s)", tenantID, ociId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
