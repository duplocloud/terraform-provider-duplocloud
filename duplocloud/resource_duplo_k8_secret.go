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
)

func k8sSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"secret_name": {
			Description: "The name of the secret.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"secret_type": {
			Description: "The type of the secret.  Usually `\"Opaque\"`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"tenant_id": {
			Description: "The GUID of the tenant that the secret will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"client_secret_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"secret_data": {
			Description: "A JSON encoded string representing the secret metadata. " +
				"You can use the `jsonencode()` function to convert map or object data, if needed.",
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
			//DiffSuppressFunc: diffIgnoreIfSameHash,
			DiffSuppressFunc: diffIgnoreForSecretMap,
		},
		"secret_annotations": {
			Description: "Annotations for the secret",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// SCHEMA for resource crud
func resourceK8Secret() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_secret` manages a kubernetes secret in a Duplo tenant.",

		ReadContext:   resourceK8SecretRead,
		CreateContext: resourceK8SecretCreate,
		UpdateContext: resourceK8SecretUpdate,
		DeleteContext: resourceK8SecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sSecretSchema(),
	}
}

/// READ resource
func resourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8SecretGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.SecretName == "" {
		d.SetId("")
		return nil
	}

	flattenK8sSecret(d, rp)
	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceK8SecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("secret_name").(string)

	log.Printf("[TRACE] resourceK8SecretCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.K8SecretCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantID, name))

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8ConfigMapCreate(%s, %s): end", tenantID, name)
	return diags
}

/// UPDATE resource
func resourceK8SecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretUpdate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.K8SecretUpdate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8ConfigMapUpdate(%s, %s): end", tenantID, name)
	return diags
}

/// DELETE resource
func resourceK8SecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8SecretGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp != nil && rp.SecretName != "" {
		err := c.K8SecretDelete(tenantID, name)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): end", tenantID, name)
	return nil
}

func diffIgnoreForSecretMap(k, old, new string, d *schema.ResourceData) bool {
	mapFieldName := "client_secret_version"
	hashFieldName := "secret_data"
	_, dataNew := d.GetChange(hashFieldName)
	hashOld := d.Get(mapFieldName).(string)
	hashNew := hashForData(dataNew.(string))
	log.Printf("[TRACE] duplo-diffIgnoreForSecretMap ******** 1: hash_old %s hash_new %s", hashNew, hashOld)
	return hashOld == hashNew
}

func parseK8sSecretIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sSecret(d *schema.ResourceData, duplo *duplosdk.DuploK8sSecret) {
	// First, set the simple fields.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("secret_name", duplo.SecretName)
	d.Set("secret_type", duplo.SecretType)
	d.Set("secret_version", duplo.SecretVersion)

	// Next, set the JSON encoded strings.
	toJsonStringState("secret_data", duplo.SecretData, d)

	// Finally, set the map
	d.Set("secret_annotations", duplo.SecretAnnotations)
}

func expandK8sSecret(d *schema.ResourceData) (*duplosdk.DuploK8sSecret, error) {
	duplo := duplosdk.DuploK8sSecret{
		SecretName: d.Get("secret_name").(string),
		SecretType: d.Get("secret_type").(string),
	}

	// The annotations must be converted to a map of strings.
	if v, ok := d.GetOk("secret_annotations"); ok && !isInterfaceNil(v) {
		duplo.SecretAnnotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.SecretAnnotations[key] = value.(string)
		}
	}

	// The data must be decoded as JSON.
	data := d.Get("secret_data").(string)
	if data != "" {
		err := json.Unmarshal([]byte(data), &duplo.SecretData)
		if err != nil {
			return nil, err
		}
	}
	if duplo.SecretData == nil {
		duplo.SecretData = map[string]interface{}{}
	}

	return &duplo, nil
}
