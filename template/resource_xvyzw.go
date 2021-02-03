package duplocloud

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"
)

// SCHEMA for resource crud
func resourceXvyzw() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceXvyzwRead,
		CreateContext: resourceXvyzwCreate,
		UpdateContext: resourceXvyzwUpdate,
		DeleteContext: resourceXvyzwDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.XvyzwSchema(),
	}
}
// SCHEMA for resource data/search
func dataSourceXvyzw() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceXvyzwRead,
		Schema: map[string]*schema.Schema{
			"filter": FilterSchema(), // todo: search specific to this object... may be api should support filter?
			"tenant_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"data": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: *duplosdk.XvyzwSchema(),
				},
			},
		},
	}
}
/// READ/SEARCH resources
func dataSourceXvyzwRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceXvyzwRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duplo_objs, err := c.XvyzwGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.XvyzwsFlatten(duplo_objs,d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceXvyzwRead ******** end")

	return diags
}
/// READ resource
func resourceXvyzwRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceXvyzwRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.XvyzwGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.XvyzwSetId(d)
	log.Printf("[TRACE] duplo-resourceXvyzwRead ******** end")
	return diags
}
/// CREATE resource
func resourceXvyzwCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceXvyzwCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.XvyzwCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.XvyzwSetId(d)
	resourceXvyzwRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceXvyzwCreate ******** end")
	return diags
}
/// UPDATE resource
func resourceXvyzwUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceXvyzwUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.XvyzwUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.XvyzwSetId(d)
	resourceXvyzwRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceXvyzwUpdate ******** end")

	return diags
}
/// DELETE resource
func resourceXvyzwDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceXvyzwDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.XvyzwDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceXvyzwDelete ******** end")

	return diags
}


