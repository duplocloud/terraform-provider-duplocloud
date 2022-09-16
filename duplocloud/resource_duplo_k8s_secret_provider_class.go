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

func K8sSecretProviderClassSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Secret Provider Class will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Secret Provider Class.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"secret_provider": {
			Description: "Provider to be used.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"annotations": {
			Description: "An unstructured key value map stored with the secret provider class that may be used to store arbitrary metadata.",
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
		"secret_object": {
			Description: "You may want to create a Kubernetes Secret to mirror the mounted content. Use the optional secretObjects field to define the desired state of the synced Kubernetes secret objects",
			Type:        schema.TypeList,
			Computed:    true,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Name of the secret object.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"type": {
						Description: "Type of the secret object.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"annotations": {
						Description: "An unstructured key value map stored with the secret object that may be used to store arbitrary metadata.",
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
					"data": {
						Type:     schema.TypeList,
						Computed: true,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:     schema.TypeString,
									Required: true,
								},
								"object_name": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"parameters": {
			Description: "The parameters section contains the details of the mount request.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceK8SecretProviderClass() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_secret_provider_class` manages a kubernetes Secret Provider Class in a Duplo tenant.",

		ReadContext:   resourceK8sSecretProviderClassRead,
		CreateContext: resourceK8sSecretProviderClassCreate,
		UpdateContext: resourceK8sSecretProviderClassUpdate,
		DeleteContext: resourceK8sSecretProviderClassDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: K8sSecretProviderClassSchema(),
	}
}

func resourceK8sSecretProviderClassRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretProviderClassIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sSecretProviderClassRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sSecretProviderClassGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s secret provider class %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	log.Printf("[TRACE] resourceK8sSecretProviderClassRead(%s, %s): end", tenantID, name)
	flattenK8sSecretProviderClass(tenantID, d, rp)
	return nil
}

/// CREATE resource
func resourceK8sSecretProviderClassCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sSecretProviderClassCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecretProviderClass(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sSecretProviderClassCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/secretproviderclass/%s", tenantID, name))

	diags := resourceK8sSecretProviderClassRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sSecretProviderClassCreate(%s, %s): end", tenantID, name)
	return diags
}

/// UPDATE resource
func resourceK8sSecretProviderClassUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretProviderClassIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sSecretProviderClassUpdate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecretProviderClass(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sSecretProviderClassUpdate(tenantID, name, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8sSecretProviderClassRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sSecretProviderClassUpdate(%s, %s): end", tenantID, name)
	return diags
}

/// DELETE resource
func resourceK8sSecretProviderClassDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretProviderClassIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sSecretProviderClassDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sSecretProviderClassGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s secret provider class %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.DuploK8sSecretProviderClassDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s secret provider class %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sSecretProviderClassDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sSecretProviderClassIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sSecretProviderClass(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sSecretProviderClass) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)
	d.Set("secret_provider", duplo.Provider)
	d.Set("annotations", duplo.Annotations)
	d.Set("labels", duplo.Labels)
	d.Set("secret_object", flattenProvderClassSecretObjects(duplo.SecretObjects))
	d.Set("parameters", duplo.Parameters.Objects)

}

func expandK8sSecretProviderClass(d *schema.ResourceData) (*duplosdk.DuploK8sSecretProviderClass, error) {
	duplo := duplosdk.DuploK8sSecretProviderClass{
		Name:     d.Get("name").(string),
		Provider: d.Get("secret_provider").(string),
		Parameters: &duplosdk.DuploK8sSecretProviderClassParameters{
			Objects: d.Get("parameters").(string),
		},
	}

	// The annotations must be converted to a map of strings.
	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		duplo.Annotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.Annotations[key] = value.(string)
		}
	}

	// The labels must be converted to a map of strings.
	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		duplo.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.Labels[key] = value.(string)
		}
	}

	if v, ok := d.GetOk("secret_object"); ok && !isInterfaceNil(v) {
		duplo.SecretObjects = expandProvderClassSecretObjects(v.([]interface{}))
	}

	return &duplo, nil
}

func flattenProvderClassSecretObjects(duplo *[]duplosdk.DuploK8sSecretProviderClassSecretObject) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *duplo {
		s = append(s, flattenProvderClassSecretObject(v))
	}
	return s
}

func flattenProvderClassSecretObject(duplo duplosdk.DuploK8sSecretProviderClassSecretObject) map[string]interface{} {
	m := make(map[string]interface{})
	m["name"] = duplo.SecretName
	m["type"] = duplo.Type
	m["annotations"] = duplo.Annotations
	m["labels"] = duplo.Labels
	s := []map[string]interface{}{}
	for _, d := range *duplo.Data {
		mm := make(map[string]interface{})
		mm["key"] = d.Key
		mm["object_name"] = d.ObjectName
		s = append(s, mm)
	}
	m["data"] = s
	return m
}

func expandProvderClassSecretObjects(lst []interface{}) *[]duplosdk.DuploK8sSecretProviderClassSecretObject {
	items := make([]duplosdk.DuploK8sSecretProviderClassSecretObject, 0, len(lst))
	for _, v := range lst {
		items = append(items, expandProvderClassSecretObject(v.(map[string]interface{})))
	}
	return &items
}

func expandProvderClassSecretObject(m map[string]interface{}) duplosdk.DuploK8sSecretProviderClassSecretObject {
	obj := duplosdk.DuploK8sSecretProviderClassSecretObject{
		SecretName: m["name"].(string),
		Type:       m["type"].(string),
	}

	// The annotations must be converted to a map of strings.
	if v, ok := m["annotations"]; ok && !isInterfaceNil(v) {
		obj.Annotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			obj.Annotations[key] = value.(string)
		}
	}

	// The labels must be converted to a map of strings.
	if v, ok := m["labels"]; ok && !isInterfaceNil(v) {
		obj.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			obj.Labels[key] = value.(string)
		}
	}

	if v, ok := m["data"]; ok && !isInterfaceNil(v) {
		lst := v.([]interface{})
		items := make([]duplosdk.DuploK8sSecretProviderClassSecretObjectData, 0, len(lst))
		for _, vv := range lst {
			vvv := vv.(map[string]interface{})
			items = append(items, duplosdk.DuploK8sSecretProviderClassSecretObjectData{
				Key:        vvv["key"].(string),
				ObjectName: vvv["object_name"].(string),
			})
		}
		obj.Data = &items
	}

	return obj
}
