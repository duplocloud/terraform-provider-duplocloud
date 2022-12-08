package duplocloud

import (
	"context"
	"fmt"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func K8sPVCSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the persistent volume claim will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the persistent volume claim.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec defines the desired characteristics of a volume requested by a pod author. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: persistentVolumeClaimSpecFields(),
			},
		},

		"annotations": {
			Description: "An unstructured key value map stored with the persistent volume claim that may be used to store arbitrary metadata.",
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

func persistentVolumeClaimSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"access_modes": {
			Type:        schema.TypeSet,
			Description: "A set of the desired access modes the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#access-modes-1",
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"ReadWriteOnce",
					"ReadOnlyMany",
					"ReadWriteMany",
					"ReadWriteOncePod",
				}, false),
			},
			Set: schema.HashString,
		},
		"resources": {
			Type:        schema.TypeList,
			Description: "A list of the minimum resources the volume should have. More info: http://kubernetes.io/docs/user-guide/persistent-volumes#resources",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"limits": {
						Type:        schema.TypeMap,
						Description: "Map describing the maximum amount of compute resources allowed. More info: http://kubernetes.io/docs/user-guide/compute-resources/",
						Optional:    true,
						ForceNew:    true,
					},
					// This is the only field the API will allow modifying in-place, so ForceNew is not used.
					"requests": {
						Type:        schema.TypeMap,
						Description: "Map describing the minimum amount of compute resources required. If this is omitted for a container, it defaults to `limits` if that is explicitly specified, otherwise to an implementation-defined value. More info: http://kubernetes.io/docs/user-guide/compute-resources/",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
		"volume_name": {
			Type:        schema.TypeString,
			Description: "The binding reference to the PersistentVolume backing this claim.",
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},
		"volume_mode": {
			Type:        schema.TypeString,
			Description: "Kubernetes supports two volumeModes of PersistentVolumes: `Filesystem` and `Block`.",
			Optional:    true,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{
					"Filesystem",
					"Block",
				}, false),
			},
		},
		"storage_class_name": {
			Type:        schema.TypeString,
			Description: "Name of the storage class requested by the claim",
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
	}
}

func resourceK8PVC() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_persistent_volume_claim` manages a kubernetes persistent volume claim in a Duplo tenant.",

		ReadContext:   resourceK8sPVCRead,
		CreateContext: resourceK8sPVCCreate,
		UpdateContext: resourceK8sPVCUpdate,
		DeleteContext: resourceK8sPVCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: K8sPVCSchema(),
	}
}

func resourceK8sPVCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sPVCIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sPVCRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.K8PvcGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s persistent volume claim %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] resourceK8sPVCRead(%s, %s): end", tenantID, name)
	flattenK8sPVC(tenantID, d, rp)
	return nil
}

/// CREATE resource
func resourceK8sPVCCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sPVCCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sPVC(d)
	if err != nil {
		return diag.FromErr(err)
	}
	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.K8PvcCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/pvc/%s", tenantID, name))

	diags := resourceK8sPVCRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sPVCCreate(%s, %s): end", tenantID, name)
	return diags
}

/// UPDATE resource
func resourceK8sPVCUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceK8sPVCCreate(ctx, d, m)
}

/// DELETE resource
func resourceK8sPVCDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sPVCIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sPVCDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.K8PvcGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s persistent volume claim %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.K8PvcDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s persistent volume claim %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sPVCDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sPVCIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sPVC(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sPvc) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)
	d.Set("annotations", duplo.Annotations)
	d.Set("labels", duplo.Labels)
	d.Set("spec", flattenPersistentVolumeClaimSpec(duplo.Spec))
}

func expandK8sPVC(d *schema.ResourceData) (*duplosdk.DuploK8sPvc, error) {
	pvc := &duplosdk.DuploK8sPvc{
		Name: d.Get("name").(string),
	}
	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		pvc.Annotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			pvc.Annotations[key] = value.(string)
		}
	}
	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		pvc.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			pvc.Labels[key] = value.(string)
		}
	}
	spec, err := expandPersistentVolumeClaimSpec(d)
	if err != nil {
		return nil, err
	}
	pvc.Spec = spec
	return pvc, nil
}

func expandPersistentVolumeClaimSpec(d *schema.ResourceData) (*duplosdk.DuploK8sPvcSpec, error) {
	obj := &duplosdk.DuploK8sPvcSpec{}

	if v, ok := d.GetOk("spec"); ok && !isInterfaceNil(v) {
		l := v.([]interface{})
		if len(l) == 0 || l[0] == nil {
			return obj, nil
		}
		in := l[0].(map[string]interface{})

		if in["resources"] != nil {
			resourceRequirements, err := expandResourceRequirements(d)
			if err != nil {
				return nil, err
			}
			obj.Resources = resourceRequirements
		}

		obj.AccessModes = expandStringSet(in["access_modes"].(*schema.Set))

		if v, ok := in["volume_name"].(string); ok {
			obj.VolumeName = v
		}
		if v, ok := in["volume_mode"].(string); ok {
			obj.VolumeMode = v
		}
		if v, ok := in["storage_class_name"].(string); ok && v != "" {
			obj.StorageClassName = v
		}
	}

	return obj, nil
}

func expandResourceRequirements(d *schema.ResourceData) (*duplosdk.DuploK8sPvcSpecResources, error) {
	obj := &duplosdk.DuploK8sPvcSpecResources{}
	if v, ok := d.GetOk("resources"); ok && !isInterfaceNil(v) {
		l := v.([]interface{})
		if len(l) == 0 || l[0] == nil {
			return obj, nil
		}
		in := l[0].(map[string]interface{})
		if v, ok := in["limits"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Limits = expandAsStringMap("limits", d)
		}
		if v, ok := in["requests"].(map[string]interface{}); ok && len(v) > 0 {
			obj.Requests = expandAsStringMap("requests", d)
		}
	}

	return obj, nil
}

func flattenPersistentVolumeClaimSpec(spec *duplosdk.DuploK8sPvcSpec) []interface{} {
	att := make(map[string]interface{})
	if len(spec.AccessModes) > 0 {
		att["access_modes"] = flattenStringSet(spec.AccessModes)
	}
	if spec.Resources != nil {
		att["resources"] = flattenResourceRequirements(spec.Resources)
	}
	if len(spec.VolumeName) > 0 {
		att["volume_name"] = spec.VolumeName
	}
	if len(spec.StorageClassName) > 0 {
		att["storage_class_name"] = spec.StorageClassName
	}
	if len(spec.VolumeMode) > 0 {
		att["volume_mode"] = spec.VolumeMode
	}
	return []interface{}{att}
}

func flattenResourceRequirements(resources *duplosdk.DuploK8sPvcSpecResources) []interface{} {
	att := make(map[string]interface{})
	if len(resources.Limits) > 0 {
		att["limits"] = flattenStringMap(resources.Limits)
	}
	if len(resources.Requests) > 0 {
		att["requests"] = flattenStringMap(resources.Requests)
	}
	return []interface{}{att}
}
