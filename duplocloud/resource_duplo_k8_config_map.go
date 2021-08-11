package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sConfigMapSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description:  "The name of the configmap.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: ValidateDnsSubdomainRFC1123(),
		},
		"tenant_id": {
			Description:  "The GUID of the tenant that the configmap will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"data": {
			Description: "A JSON encoded string representing the configmap data. " +
				"You can use the `jsonencode()` function to build this from JSON.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ValidateFunc: validation.StringIsJSON,
		},
		"metadata": {
			Description: "A JSON encoded string representing the configmap metadata. " +
				"You can use the `jsondecode()` function to parse this, if needed.",
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// SCHEMA for resource crud
func resourceK8ConfigMap() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_config_map` manages a kubernetes configmap in a Duplo tenant.",

		ReadContext:   resourceK8ConfigMapRead,
		CreateContext: resourceK8ConfigMapCreate,
		UpdateContext: resourceK8ConfigMapUpdate,
		DeleteContext: resourceK8ConfigMapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: k8sConfigMapSchema(),
	}
}

/// READ resource
func resourceK8ConfigMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sConfigMapIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8ConfigMapRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8ConfigMapGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	flattenK8sConfigMap(d, rp)
	log.Printf("[TRACE] resourceK8ConfigMapRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceK8ConfigMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8ConfigMapCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sConfigMap(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.K8ConfigMapCreate(tenantID, rq)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2/%s", tenantID, name))

	diags := resourceK8ConfigMapRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8ConfigMapCreate(%s, %s): end", tenantID, name)
	return diags
}

/// UPDATE resource
func resourceK8ConfigMapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sConfigMapIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8ConfigMapUpdate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sConfigMap(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.K8ConfigMapUpdate(tenantID, rq)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceK8ConfigMapRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8ConfigMapUpdate(%s, %s): end", tenantID, name)
	return diags
}

/// DELETE resource
func resourceK8ConfigMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sConfigMapIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8ConfigMapDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8ConfigMapGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp != nil && rp.Name != "" {
		err := c.K8ConfigMapDelete(tenantID, name)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] resourceK8ConfigMapDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandK8sConfigMap(d *schema.ResourceData) (*duplosdk.DuploK8sConfigMap, error) {
	duplo := duplosdk.DuploK8sConfigMap{}

	// The name field is passed through metadata.
	duplo.Metadata = map[string]interface{}{
		"name": d.Get("name").(string),
	}

	// The data must be decoded as JSON.
	data := d.Get("data").(string)
	if data != "" {
		err := json.Unmarshal([]byte(data), &duplo.Data)
		if err != nil {
			return nil, err
		}
	}

	return &duplo, nil
}

func flattenK8sConfigMap(d *schema.ResourceData, duplo *duplosdk.DuploK8sConfigMap) {
	// First, set the simple fields.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("name", duplo.Name)

	// Next, set the JSON encoded strings.
	toJsonStringState("data", duplo.Data, d)
	toJsonStringState("metadata", duplo.Metadata, d)
}

func parseK8sConfigMapIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
