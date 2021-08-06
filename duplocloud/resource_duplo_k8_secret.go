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
	log.Printf("[TRACE] duplo-resourceK8SecretCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8SecretSetID(d)
	resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8SecretCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceK8SecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8SecretSetID(d)
	resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8SecretUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceK8SecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceK8SecretDelete ******** end")

	return diags
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

	// Next, set the JSON encoded strings.
	toJsonStringState("secret_data", duplo.SecretData, d)

	// Finally, set the map
	d.Set("secret_annotations", duplo.SecretAnnotations)
}
