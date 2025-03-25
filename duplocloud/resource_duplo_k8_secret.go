package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the secret will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"secret_name": {
			Description:  "The name of the secret.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: ValidateDnsSubdomainRFC1123(),
		},
		"secret_type": {
			Description: "The type of the secret.  Usually `\"Opaque\"`.",
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
				"You can use the `jsonencode()` function to convert map or object data, if needed. You can use the `jsondecode()` function to read data.",
			Type:             schema.TypeString,
			Optional:         true,
			Sensitive:        true,
			ValidateFunc:     ValidateJSONObjectString,
			DiffSuppressFunc: secretDataDiff,
		},
		"secret_annotations": {
			Description: "Annotations for the secret",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"secret_labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the secret",
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

// / READ resource
func resourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8SecretGet(tenantId, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.SecretName == "" {
		d.SetId("")
		return nil
	}

	flattenK8sSecret(d, rp, false)
	log.Printf("[TRACE] resourceK8SecretRead(%s, %s): end", tenantId, name)
	return nil
}

// / CREATE resource
func resourceK8SecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	name := d.Get("secret_name").(string)

	log.Printf("[TRACE] resourceK8SecretCreate(%s, %s): start", tenantId, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.K8SecretCreate(tenantId, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/K8SecretApiV2/%s", tenantId, name))

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8SecretCreate(%s, %s): end", tenantId, name)
	return diags
}

// / UPDATE resource
func resourceK8SecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretUpdate(%s, %s): start", tenantId, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sSecret(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.K8SecretUpdate(tenantId, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8SecretUpdate(%s, %s): end", tenantId, name)
	return diags
}

// / DELETE resource
func resourceK8SecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sSecretIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.K8SecretGet(tenantId, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp != nil && rp.SecretName != "" {
		err := c.K8SecretDelete(tenantId, name)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] resourceK8SecretDelete(%s, %s): end", tenantId, name)
	return nil
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

func flattenK8sSecret(d *schema.ResourceData, duplo *duplosdk.DuploK8sSecret, readOnly bool) {
	// First, set the simple fields.
	d.Set("tenant_id", duplo.TenantID)
	d.Set("secret_name", duplo.SecretName)
	d.Set("secret_type", duplo.SecretType)
	d.Set("secret_version", duplo.SecretVersion)
	if readOnly {
		for key := range duplo.SecretData {
			duplo.SecretData[key] = "**********"
		}
	}
	// Next, set the JSON encoded strings.
	toJsonStringState("secret_data", duplo.SecretData, d)

	// Finally, set the map
	d.Set("secret_annotations", duplo.SecretAnnotations)
	filter := map[string]struct{}{
		"app":        {},
		"owner":      {},
		"tenantid":   {},
		"tenantname": {},
	}
	m := d.Get("secret_labels")
	if m != nil {
		for k := range m.(map[string]interface{}) {
			if _, ok := filter[k]; ok {
				delete(filter, k)
			}
		}
	}
	op := make(map[string]interface{})

	if duplo.SecretLabels != nil {
		for k, v := range duplo.SecretLabels {
			if _, ok := filter[k]; !ok {
				op[k] = v
			}
		}
	}
	d.Set("secret_labels", op)
	log.Printf("[TRACE] K8SecretGetList(%s): received response: %s", duplo.TenantID, duplo)

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

	if v, ok := d.GetOk("secret_labels"); ok && !isInterfaceNil(v) {
		duplo.SecretLabels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			if !isStringValid(regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"), key) {
				return nil, secretLabelValidationError(duplo.SecretName, key)
			}
			v := value.(string)
			if !isStringValid(regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"), v) {
				return nil, secretLabelValidationError(duplo.SecretName, v)
			}
			duplo.SecretLabels[key] = v
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
	d.Set("client_secret_version", hashForData(data))
	if duplo.SecretData == nil {
		duplo.SecretData = map[string]interface{}{}
	}

	return &duplo, nil
}

func secretLabelValidationError(name, value string) error {
	return fmt.Errorf(`Secret '%s' is invalid: metadata.labels: Invalid value: '%s,': a valid label must 
	be an empty string or consist of alphanumeric characters, '-', '', or '.', and must start and end with an alphanumeric 
	character (e.g. 'MyValue', or 'my_value', or '12345',
	 regex used for validation is '((A-Za-z0-9][-A-Za-z0-9.]*)?[A-Za-z0-9])?').`, name, value)
}

func secretDataDiff(k, old, new string, d *schema.ResourceData) bool {
	state, err := secretDataCompare(old, new)
	if err != nil {
		log.Printf("TRACE secretDataCompare : %s", err.Error())
		return state
	}
	return state
}
func secretDataCompare(old, new string) (bool, error) {
	var obj1, obj2 map[string]interface{}

	// Unmarshal the first JSON string into a map
	if err := json.Unmarshal([]byte(old), &obj1); err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 1: %v", err)
	}

	// Unmarshal the second JSON string into a map
	if err := json.Unmarshal([]byte(new), &obj2); err != nil {
		return false, fmt.Errorf("error unmarshalling JSON 2: %v", err)
	}
	if len(obj1) != len(obj2) {
		return false, nil
	}
	for k, v := range obj2 {
		if v1, ok := obj1[k]; !ok {
			return false, nil
		} else {
			s := fmt.Sprintf("%v", v)
			if v1 != s {
				return false, nil
			}
		}
	}
	return true, nil
}
