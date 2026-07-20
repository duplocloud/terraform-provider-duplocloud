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

func k8sReferenceGrantSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the ReferenceGrant will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the ReferenceGrant.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"from": {
			Description: "The namespaces and resource kinds that are permitted to make cross-namespace references.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group": {
						Description: "The API group of the referencing resource. Use `gateway.networking.k8s.io` for HTTPRoute and Gateway.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"kind": {
						Description: "The kind of the referencing resource (e.g. `HTTPRoute`, `Gateway`).",
						Type:        schema.TypeString,
						Required:    true,
					},
					"namespace": {
						Description: "The namespace of the referencing resource that is allowed to make cross-namespace references.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"to": {
			Description: "The resource kinds in the ReferenceGrant's namespace that may be referenced.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group": {
						Description: "The API group of the target resource. Empty string for core Kubernetes resources (Services, Secrets).",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"kind": {
						Description: "The kind of the target resource (e.g. `Service`, `Secret`).",
						Type:        schema.TypeString,
						Required:    true,
					},
					"name": {
						Description: "Optional name of a specific target resource. Omit to allow any resource of this kind.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceK8sReferenceGrant() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8s_reference_grant` manages a Kubernetes ReferenceGrant (gateway.networking.k8s.io/v1beta1) in a Duplo tenant.",

		ReadContext:   resourceK8sReferenceGrantRead,
		CreateContext: resourceK8sReferenceGrantCreate,
		UpdateContext: resourceK8sReferenceGrantUpdate,
		DeleteContext: resourceK8sReferenceGrantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sReferenceGrantSchema(),
	}
}

func resourceK8sReferenceGrantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sReferenceGrantIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sReferenceGrantRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sReferenceGrantGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sReferenceGrantRead(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s reference grant %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	flattenK8sReferenceGrant(tenantID, d, rp)
	log.Printf("[TRACE] resourceK8sReferenceGrantRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8sReferenceGrantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sReferenceGrantCreate(%s, %s): start", tenantID, name)

	rq := expandK8sReferenceGrant(d)

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sReferenceGrantCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/referencegrant/%s", tenantID, name))

	diags := resourceK8sReferenceGrantRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sReferenceGrantCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceK8sReferenceGrantUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sReferenceGrantIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sReferenceGrantUpdate(%s, %s): start", tenantID, name)

	rq := expandK8sReferenceGrant(d)

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sReferenceGrantUpdate(tenantID, name, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8sReferenceGrantRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sReferenceGrantUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceK8sReferenceGrantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sReferenceGrantIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sReferenceGrantDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sReferenceGrantGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sReferenceGrantDelete(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s reference grant %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.DuploK8sReferenceGrantDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s reference grant %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sReferenceGrantDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sReferenceGrantIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandK8sReferenceGrant(d *schema.ResourceData) *duplosdk.DuploK8sReferenceGrant {
	duplo := duplosdk.DuploK8sReferenceGrant{
		Name: d.Get("name").(string),
	}

	fromList := d.Get("from").([]interface{})
	froms := make([]duplosdk.DuploK8sReferenceGrantFrom, 0, len(fromList))
	for _, v := range fromList {
		m := v.(map[string]interface{})
		froms = append(froms, duplosdk.DuploK8sReferenceGrantFrom{
			Group:     m["group"].(string),
			Kind:      m["kind"].(string),
			Namespace: m["namespace"].(string),
		})
	}
	duplo.From = &froms

	toList := d.Get("to").([]interface{})
	tos := make([]duplosdk.DuploK8sReferenceGrantTo, 0, len(toList))
	for _, v := range toList {
		m := v.(map[string]interface{})
		tos = append(tos, duplosdk.DuploK8sReferenceGrantTo{
			Group: m["group"].(string),
			Kind:  m["kind"].(string),
			Name:  m["name"].(string),
		})
	}
	duplo.To = &tos

	return &duplo
}

func flattenK8sReferenceGrant(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sReferenceGrant) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)

	if duplo.From != nil {
		froms := make([]interface{}, 0, len(*duplo.From))
		for _, v := range *duplo.From {
			froms = append(froms, map[string]interface{}{
				"group":     v.Group,
				"kind":      v.Kind,
				"namespace": v.Namespace,
			})
		}
		d.Set("from", froms)
	}

	if duplo.To != nil {
		tos := make([]interface{}, 0, len(*duplo.To))
		for _, v := range *duplo.To {
			tos = append(tos, map[string]interface{}{
				"group": v.Group,
				"kind":  v.Kind,
				"name":  v.Name,
			})
		}
		d.Set("to", tos)
	}
}
