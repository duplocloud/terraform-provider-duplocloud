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
func resourceDuploServiceParams() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploServiceParamsRead,
		CreateContext: resourceDuploServiceParamsCreate,
		UpdateContext: resourceDuploServiceParamsUpdate,
		DeleteContext: resourceDuploServiceParamsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.DuploServiceParamsSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceDuploServiceParams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceParamsRead,
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
					Schema: *duplosdk.DuploServiceParamsSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceDuploServiceParamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceDuploServiceParamsRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duplo_objs, err := c.DuploServiceParamsGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.DuploServiceParamsFlatten(duplo_objs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceDuploServiceParamsRead ******** end")

	return diags
}

/// READ resource
func resourceDuploServiceParamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.DuploServiceParamsGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceParamsSetId(d)
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsRead ******** end")
	return diags
}

/// CREATE resource
func resourceDuploServiceParamsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceParamsCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceParamsSetId(d)
	resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploServiceParamsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceParamsUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceParamsSetId(d)
	resourceDuploServiceParamsRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceDuploServiceParamsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceParamsDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceParamsDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted

	log.Printf("[TRACE] duplo-resourceDuploServiceParamsDelete ******** end")

	return diags
}
