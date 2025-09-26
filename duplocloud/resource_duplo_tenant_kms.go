package duplocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTenantKMS() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant_kms` manages the list of kms avaialble to a tenant in Duplo.\n\n" +
			"This resource allows you take control of individual tenant kms for a specific tenant.",

		ReadContext:   resourceTenantKMSRead,
		CreateContext: resourceTenantKMSCreate,
		UpdateContext: resourceTenantKMSUpdate,
		DeleteContext: resourceTenantKMSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The ID of the tenant to configure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"kms": {
				Description: "A list of KMS key to manage.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        tenantKmsSchema(true),
			},
			"unspecified_kms_keys": {
				Description: "A list of kms keys not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        tenantKmsSchema(false),
			},
			"delete_unspecified_kms_keys": {
				Description: "Whether or not this resource should delete any kms not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func tenantKmsSchema(h bool) *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: h,
				ForceNew: h,
				Computed: !h,
			},
			"id": {
				Type:     schema.TypeString,
				Required: h,
				ForceNew: h,
				Computed: !h,
			},
			"arn": {
				Type:     schema.TypeString,
				Required: h,
				Computed: !h,
			},
		},
	}
}

func resourceTenantKMSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	tenantID := info[0]

	log.Printf("[TRACE] resourceTenantKMSRead(%s): start", tenantID)

	c := m.(*duplosdk.Client)

	duplo, err := c.TenantKMSGetList(tenantID)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed to retrieve tenant kmskeys for '%s': %s", tenantID, err)
	}
	d.Set("tenant_id", tenantID)
	flattenTenantKmsKeys(d, duplo)

	log.Printf("[TRACE] resourceTenantCertificatesRead(%s): end", tenantID)
	return nil

}

func flattenTenantKmsKeys(d *schema.ResourceData, list *[]duplosdk.DuploTenantKmsKeyInfo) []interface{} {

	skms := d.Get("kms").([]interface{})
	kmsInfos := []duplosdk.DuploTenantKmsKeyInfo{}
	for _, i := range skms {
		m := i.(map[string]interface{})
		kmsInfos = append(kmsInfos, duplosdk.DuploTenantKmsKeyInfo{
			KeyName: m["name"].(string),
			KeyId:   m["id"].(string),
			KeyArn:  m["arn"].(string),
		})
	}
	result, unspecified := segregateUnspecifiedKms(&kmsInfos, list)
	d.Set("unspecified_kms_keys", unspecified)
	d.Set("kms_keys", result)

	return result
}

func segregateUnspecifiedKms(skms, list *[]duplosdk.DuploTenantKmsKeyInfo) ([]interface{}, []interface{}) {
	result := make([]interface{}, 0, len(*list))
	unspecified := []interface{}{}
	specified := make(map[string]struct{})
	for _, v := range *skms {
		specified[v.KeyName] = struct{}{}
	}
	for _, kms := range *list {
		if _, ok := specified[kms.KeyName]; ok {

			result = append(result, map[string]interface{}{
				"name": kms.KeyName,
				"id":   kms.KeyId,
				"arn":  kms.KeyArn,
			})
		} else {
			unspecified = append(unspecified, map[string]interface{}{
				"name": kms.KeyName,
				"id":   kms.KeyId,
				"arn":  kms.KeyArn,
			})
		}
	}
	return result, unspecified
}
func resourceTenantKMSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceTenantKMSCreate(%s): start", tenantID)

	c := m.(*duplosdk.Client)
	rq := expandTenantKms(d)
	rp, _ := c.TenantKMSGetList(tenantID)
	_, toDel := segregateUnspecifiedKms(rq, rp)
	if d.Get("delete_unspecified_kms_keys").(bool) {
		// Apply the changes via Duplo
		for _, i := range toDel {
			m := i.(map[string]interface{})
			c.TenantKMSDelete(tenantID, m["id"].(string))
		}
	}
	for _, i := range *rq {
		cerr := c.TenantCreateKMSKey(tenantID, i)
		if cerr != nil {
			return diag.Errorf("Error creating tenant kms %s for '%s': %s", i.KeyName, tenantID, cerr)
		}
	}
	d.SetId(tenantID + "/tenantkms")

	diags := resourceTenantKMSRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantKMSCreate(%s): end", tenantID)
	return diags

}

func resourceTenantKMSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceTenantKMSUpdate(%s): start", tenantID)
	rq := expandTenantKms(d)
	c := m.(*duplosdk.Client)

	// Apply the changes via Duplo
	if d.HasChange("delete_unspecified_kms_keys") && d.Get("delete_unspecified_kms_keys").(bool) {
		usp := d.Get("unspecified_kms_keys").([]interface{})
		for _, i := range usp {
			m := i.(map[string]interface{})
			c.TenantKMSDelete(tenantID, m["id"].(string))
		}
	}
	for _, i := range *rq {
		cerr := c.TenantUpdateKMSKey(tenantID, i)
		if cerr != nil {
			return diag.Errorf("Error updating tenant kms %s for '%s': %s", i.KeyName, tenantID, cerr)
		}
	}
	d.SetId(tenantID + "/kms")

	diags := resourceTenantKMSRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantKMSUpdate(%s): end", tenantID)
	return diags
}

func resourceTenantKMSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	info := strings.SplitN(id, "/", 3)
	tenantID := info[0]
	log.Printf("[TRACE] resourceTenantKMSDelete(%s): start", tenantID)

	c := m.(*duplosdk.Client)

	// Apply the changes via Duplo
	if d.Get("delete_unspecified_kms_keys").(bool) {
		usp := d.Get("unspecified_kms_keys").([]interface{})
		for _, i := range usp {
			m := i.(map[string]interface{})
			c.TenantKMSDelete(tenantID, m["id"].(string))
		}
	}
	for _, i := range d.Get("kms").([]interface{}) {
		m := i.(map[string]interface{})
		c.TenantKMSDelete(tenantID, m["id"].(string))
	}

	log.Printf("[TRACE] resourceTenantKMSDelete(%s): end", tenantID)
	return nil
}

func expandTenantKms(d *schema.ResourceData) *[]duplosdk.DuploTenantKmsKeyInfo {
	var ary []duplosdk.DuploTenantKmsKeyInfo

	if v, ok := d.GetOk("kms"); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		ary = make([]duplosdk.DuploTenantKmsKeyInfo, 0, len(kvs))

		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploTenantKmsKeyInfo{
				KeyId:   kv["id"].(string),
				KeyName: kv["name"].(string),
				KeyArn:  kv["arn"].(string),
			})
		}
	}

	return &ary
}
