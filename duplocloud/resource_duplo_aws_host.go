package duplocloud

import (
	"context"
	"log"
	"strconv"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func resourceAwsHost() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceAwsHostRead,
		CreateContext: resourceAwsHostCreate,
		UpdateContext: resourceAwsHostUpdate,
		DeleteContext: resourceAwsHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.AwsHostSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceAwsHost() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsHostRead,
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
					Schema: *duplosdk.AwsHostSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceAwsHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceAwsHostRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.AwsHostGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.AwsHostsFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceAwsHostRead ******** end")

	return diags
}

/// READ resource
func resourceAwsHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceAwsHostRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.AwsHostGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.AwsHostSetId(d)
	log.Printf("[TRACE] duplo-resourceAwsHostRead ******** end")
	return diags
}

/// CREATE resource
func resourceAwsHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceAwsHostCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.AwsHostCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.AwsHostSetId(d)
	resourceAwsHostRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceAwsHostCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceAwsHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceAwsHostUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.AwsHostUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.AwsHostSetId(d)
	resourceAwsHostRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceAwsHostUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceAwsHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceAwsHostDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.AwsHostDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceAwsHostDelete ******** end")

	return diags
}
