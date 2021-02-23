package duplocloud

import (
	"context"
	"strconv"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func resourceK8ConfigMap() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceK8ConfigMapRead,
		CreateContext: resourceK8ConfigMapCreate,
		UpdateContext: resourceK8ConfigMapUpdate,
		DeleteContext: resourceK8ConfigMapDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.K8ConfigMapSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceK8ConfigMap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8ConfigMapRead,
		Schema: map[string]*schema.Schema{
			"filter": FilterSchema(), // todo: search specific to this object... may be api should support filter?
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: *duplosdk.K8ConfigMapSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceK8ConfigMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceK8ConfigMapRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.K8ConfigMapGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.K8ConfigMapsFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceK8ConfigMapRead ******** end")

	return diags
}

/// READ resource
func resourceK8ConfigMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8ConfigMapRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.K8ConfigMapGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8ConfigMapSetID(d)
	log.Printf("[TRACE] duplo-resourceK8ConfigMapRead ******** end")
	return diags
}

/// CREATE resource
func resourceK8ConfigMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8ConfigMapCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8ConfigMapCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8ConfigMapSetID(d)
	resourceK8ConfigMapRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8ConfigMapCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceK8ConfigMapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8ConfigMapUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8ConfigMapUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8ConfigMapSetID(d)
	resourceK8ConfigMapRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8ConfigMapUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceK8ConfigMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8ConfigMapDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8ConfigMapDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceK8ConfigMapDelete ******** end")

	return diags
}
