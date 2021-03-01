package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"account_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"plan_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"infra_owner": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"policy": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allow_volume_mapping": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"block_external_ep": {
						Type:     schema.TypeBool,
						Computed: true,
					},
				},
			},
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

func resourceTenant() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceTenantRead,
		CreateContext: resourceTenantCreate,
		UpdateContext: resourceTenantUpdate,
		DeleteContext: resourceTenantDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: tenantSchema(),
	}
}

/// READ resource
func resourceTenantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceTenantRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.TenantGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.TenantSetID(d)
	log.Printf("[TRACE] duplo-resourceTenantRead ******** end")
	return diags
}

/// CREATE resource
func resourceTenantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceTenantCreate ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics

	err := resource.Retry(4*time.Minute, func() *resource.RetryError {
		_, err := c.TenantCreate(d, m)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		time.Sleep(time.Duration(3) * time.Minute)
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	c.TenantSetID(d)
	resourceTenantRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceTenantCreate ******** end")

	return diags
}

/// UPDATE resource
func resourceTenantUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceTenantUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.TenantUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.TenantSetID(d)
	resourceTenantRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceTenantUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceTenantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceTenantDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.TenantDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] duplo-resourceTenantDelete ******** end")

	return diags
}
