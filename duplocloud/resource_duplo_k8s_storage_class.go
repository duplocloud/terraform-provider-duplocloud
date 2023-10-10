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

func K8sStorageClassSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the storage class will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the storage class.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the storage class.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"parameters": {
			Type:        schema.TypeMap,
			Description: "The parameters for the provisioner that should create volumes of this storage class",
			Optional:    true,
			ForceNew:    true,
		},
		"storage_provisioner": {
			Type:        schema.TypeString,
			Description: "Indicates the type of the provisioner",
			Required:    true,
			ForceNew:    true,
		},
		"reclaim_policy": {
			Type:        schema.TypeString,
			Description: "Indicates the type of the reclaim policy",
			Optional:    true,
			Default:     "Delete",
			ValidateFunc: validation.StringInSlice([]string{
				"Delete",
				"Retain",
			}, false),
		},
		"volume_binding_mode": {
			Type:        schema.TypeString,
			Description: "Indicates when volume binding and dynamic provisioning should occur",
			Optional:    true,
			ForceNew:    true,
			Default:     "Immediate",
			ValidateFunc: validation.StringInSlice([]string{
				"WaitForFirstConsumer",
				"Immediate",
			}, false),
		},
		"allow_volume_expansion": {
			Type:        schema.TypeBool,
			Description: "Indicates whether the storage class allow volume expand",
			Optional:    true,
			Default:     false,
		},
		"allowed_topologies": {
			Type:        schema.TypeList,
			Description: "Restrict the node topologies where volumes can be dynamically provisioned.",
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"match_label_expressions": {
						Type:        schema.TypeList,
						Description: "A list of topology selector requirements by labels.",
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:        schema.TypeString,
									Description: "The label key that the selector applies to.",
									Optional:    true,
								},
								"values": {
									Type:        schema.TypeSet,
									Description: "An array of string values. One value must match the label to be selected.",
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
		"annotations": {
			Description: "An unstructured key value map stored with the storage class that may be used to store arbitrary metadata.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the service.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func resourceK8StorageClass() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_storage_class` manages a kubernetes storage class in a Duplo tenant.",

		ReadContext:   resourceK8sStorageClassRead,
		CreateContext: resourceK8sStorageClassCreate,
		UpdateContext: resourceK8sStorageClassUpdate,
		DeleteContext: resourceK8sStorageClassDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: K8sStorageClassSchema(),
	}
}

func resourceK8sStorageClassRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, fullname, err := parseK8sStorageClassIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sStorageClassRead(%s, %s): start", tenantID, fullname)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.K8StorageClassGet(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s storage class %s : %s", tenantID, fullname, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, fullname)
	d.Set("name", name)
	flattenK8sStorageClass(tenantID, d, rp)
	log.Printf("[TRACE] resourceK8sStorageClassRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8sStorageClassCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sStorageClassCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sStorageClass(d)
	if err != nil {
		return diag.FromErr(err)
	}
	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	resp, cerr := c.K8StorageClassCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	// fullname, clientErr := c.GetDuploServicesName(tenantID, name)
	// if clientErr != nil {
	// 	return diag.FromErr(clientErr)
	// }
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/storageclass/%s", tenantID, resp.Name))

	diags := resourceK8sStorageClassRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sStorageClassCreate(%s, %s): end", tenantID, resp.Name)
	return diags
}

// UPDATE resource
func resourceK8sStorageClassUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceK8sStorageClassCreate(ctx, d, m)
}

// DELETE resource
func resourceK8sStorageClassDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, fullname, err := parseK8sStorageClassIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sStorageClassDelete(%s, %s): start", tenantID, fullname)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.K8StorageClassGet(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s storage class %s : %s", tenantID, fullname, clientErr)
	}
	if rp != nil && rp.Name != "" {
		_, clientErr := c.K8StorageClassDelete(tenantID, fullname)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s storage class %s : %s", tenantID, fullname, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sStorageClassDelete(%s, %s): end", tenantID, fullname)
	return nil
}

func parseK8sStorageClassIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sStorageClass(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sStorageClass) {
	d.Set("tenant_id", tenantId)
	d.Set("fullname", duplo.Name)
	d.Set("storage_provisioner", duplo.Provisioner)
	d.Set("reclaim_policy", duplo.ReclaimPolicy)
	d.Set("volume_binding_mode", duplo.VolumeBindingMode)
	d.Set("allow_volume_expansion", duplo.AllowVolumeExpansion)
	d.Set("parameters", duplo.Parameters)
	d.Set("annotations", duplo.Annotations)
	d.Set("labels", duplo.Labels)
	if duplo.AllowedTopologies != nil && len(*duplo.AllowedTopologies) > 0 {
		d.Set("allowed_topologies", flattenStorageClassAllowedTopologies(duplo.AllowedTopologies))
	}
}

func flattenStorageClassAllowedTopologies(in *[]duplosdk.DuploK8sAllowedTopologiesMatchLabelExpressions) []interface{} {
	att := make(map[string]interface{})
	for _, n := range *in {
		if n.MatchLabelExpressions != nil && len(*n.MatchLabelExpressions) > 0 {
			att["match_label_expressions"] = flattenStorageClassMatchLabelExpressions(n.MatchLabelExpressions)
		}
	}
	return []interface{}{att}
}

func flattenStorageClassMatchLabelExpressions(in *[]duplosdk.DuploK8sStorageClassAllowedTopologies) []interface{} {
	att := make([]interface{}, len(*in))
	for i, n := range *in {
		m := make(map[string]interface{})
		m["key"] = n.Key
		m["values"] = flattenStringSet(n.Values)
		att[i] = m
	}
	return att
}

func expandK8sStorageClass(d *schema.ResourceData) (*duplosdk.DuploK8sStorageClass, error) {
	storageClass := &duplosdk.DuploK8sStorageClass{
		Name:                 d.Get("name").(string),
		ReclaimPolicy:        d.Get("reclaim_policy").(string),
		Provisioner:          d.Get("storage_provisioner").(string),
		VolumeBindingMode:    d.Get("volume_binding_mode").(string),
		AllowVolumeExpansion: d.Get("allow_volume_expansion").(bool),
	}
	if v, ok := d.GetOk("parameters"); ok && !isInterfaceNil(v) {
		storageClass.Parameters = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			storageClass.Parameters[key] = value.(string)
		}
	}
	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		storageClass.Annotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			storageClass.Annotations[key] = value.(string)
		}
	}
	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		storageClass.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			storageClass.Labels[key] = value.(string)
		}
	}
	if v, ok := d.GetOk("allowed_topologies"); ok && !isInterfaceNil(v) {
		storageClass.AllowedTopologies = expandStorageClassAllowedTopologies(v.([]interface{}))
	}
	return storageClass, nil
}

func expandStorageClassAllowedTopologies(l []interface{}) *[]duplosdk.DuploK8sAllowedTopologiesMatchLabelExpressions {
	if len(l) == 0 || l[0] == nil {
		return &[]duplosdk.DuploK8sAllowedTopologiesMatchLabelExpressions{}
	}

	in := l[0].(map[string]interface{})
	topologies := make([]duplosdk.DuploK8sAllowedTopologiesMatchLabelExpressions, 0)
	obj := duplosdk.DuploK8sAllowedTopologiesMatchLabelExpressions{}

	if v, ok := in["match_label_expressions"].([]interface{}); ok && len(v) > 0 {
		obj.MatchLabelExpressions = expandStorageClassMatchLabelExpressions(v)
	}

	topologies = append(topologies, obj)

	return &topologies
}

func expandStorageClassMatchLabelExpressions(l []interface{}) *[]duplosdk.DuploK8sStorageClassAllowedTopologies {
	if len(l) == 0 || l[0] == nil {
		return &[]duplosdk.DuploK8sStorageClassAllowedTopologies{}
	}
	obj := make([]duplosdk.DuploK8sStorageClassAllowedTopologies, len(l))
	for i, n := range l {
		in := n.(map[string]interface{})
		obj[i] = duplosdk.DuploK8sStorageClassAllowedTopologies{
			Key:    in["key"].(string),
			Values: expandStringSet(in["values"].(*schema.Set)),
		}
	}
	return &obj
}
