package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sQuotaSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the K8's quota will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Name of the K8's quota.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"scope_selector": {
			Description: "Applies quota only to resources matching specified workload scopes.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"resource_quota": {
			Description: "Limits the total amount of compute, storage, and object resources that a namespace can use to prevent over-consumption and ensure fair resource allocation in a Kubernetes cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func resourceTenantK8sQuota() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant_k8s_resource_quota` manages a resource quota of k8s in the tenant.",

		ReadContext:   resourceK8QuotaRead,
		CreateContext: resourceK8QuotaCreate,
		UpdateContext: resourceK8QuotaUpdate,
		DeleteContext: resourceK8QuotaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sQuotaSchema(),
	}
}

func resourceK8QuotaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8QuotaIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8QuotaRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploTenantK8QuotaGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8QuotaRead(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s K8 quota %s : %s", tenantID, name, clientErr)
	}
	err = flattenK8ResourceQuota(d, rp)

	if err != nil {
		return diag.Errorf("Unable to flatten tenant %s k8 resource quota %s : %s", tenantID, name, err)
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", rp.Metadata.Name)

	log.Printf("[TRACE] resourceK8QuotaRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8QuotaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8QuotaCreate(%s, %s): start", tenantID, name)

	rq, err := expandK8ResourceQuota(d)
	if err != nil {
		return diag.FromErr(err)
	}
	rq.Metadata = &duplosdk.DuploTenantK8QuotaMetadata{Name: name}

	c := m.(*duplosdk.Client)
	cerr := c.DuploTenantK8QuotaCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(tenantID + "/k8-quota/" + name)
	diags := resourceK8QuotaRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8QuotaCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceK8QuotaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8QuotaIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8QuotaUpdate(%s, %s): start", tenantID, name)
	rq, err := expandK8ResourceQuota(d)
	if err != nil {
		return diag.FromErr(err)
	}
	rq.Metadata = &duplosdk.DuploTenantK8QuotaMetadata{Name: name}

	c := m.(*duplosdk.Client)
	cerr := c.DuploTenantK8QuotaUpdate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8QuotaRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8QuotaUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceK8QuotaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8QuotaIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8QuotaDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	clientErr := c.DuploTenantK8QuotaDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8QuotaDelete(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8 resource quota %s : %s", tenantID, name, clientErr)
	}

	log.Printf("[TRACE] resourceK8QuotaDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8QuotaIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, name = idParts[0], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandK8ResourceQuota(d *schema.ResourceData) (*duplosdk.DuploTenantK8Quota, error) {
	body := duplosdk.DuploTenantK8Quota{}
	quota := d.Get("resource_quota").(string)
	var hard map[string]interface{}
	spec := &duplosdk.DuploTenantK8QuotaSpec{}
	if err := json.Unmarshal([]byte(quota), &hard); err != nil {
		return nil, err
	}
	spec.Hard = hard
	if scope, ok := d.GetOk("scope_selector"); ok {
		mp := map[string]interface{}{}
		if err := json.Unmarshal([]byte(scope.(string)), &mp); err != nil {
			return nil, err
		}
		spec.ScopeSelector = mp
	}
	body.Spec = spec

	return &body, nil
}

func flattenK8ResourceQuota(d *schema.ResourceData, quota *duplosdk.DuploTenantK8Quota) error {

	if quota == nil || quota.Spec == nil {
		return nil
	}

	if quota.Spec.Hard != nil {
		b, err := json.Marshal(quota.Spec.Hard)
		if err != nil {
			return err
		}
		if err := d.Set("resource_quota", string(b)); err != nil {
			return err
		}
	}

	if quota.Spec.ScopeSelector != nil {
		b, err := json.Marshal(quota.Spec.ScopeSelector)
		if err != nil {
			return err
		}
		if err := d.Set("scope_selector", string(b)); err != nil {
			return err
		}
	}

	return nil
}
