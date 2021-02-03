package duplocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strconv"
	"terraform-provider-duplocloud/duplosdk"
	"time"
)

// SCHEMA for resource crud
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
		Schema: *duplosdk.TenantSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTenantRead,
		Schema: map[string]*schema.Schema{
			"data": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: *duplosdk.TenantSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceTenantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceTenantRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duplo_objs, err := c.TenantGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.TenantsFlatten(duplo_objs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceTenantRead ******** end")

	return diags
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

	c.TenantSetId(d)
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

	c.TenantSetId(d)
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

	c.TenantSetId(d)
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
