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
func resourceDuploServiceLBConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploServiceLBConfigsRead,
		CreateContext: resourceDuploServiceLBConfigsCreate,
		UpdateContext: resourceDuploServiceLBConfigsUpdate,
		DeleteContext: resourceDuploServiceLBConfigsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.DuploServiceLBConfigsSchema(),
	}
}

// SCHEMA for resource search/data
func dataSourceDuploServiceLBConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceLBConfigsRead,
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
					Schema: *duplosdk.DuploServiceLBConfigsSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceDuploServiceLBConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceDuploServiceLBConfigsRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.DuploServiceLBConfigsGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.DuploServiceLBConfigsListFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceDuploServiceLBConfigsRead ******** end")

	return diags
}

/// READ resource
func resourceDuploServiceLBConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.DuploServiceLBConfigsGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceLBConfigsSetID(d)
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsRead ******** end")
	return diags
}

/// CREATE resource
func resourceDuploServiceLBConfigsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceLBConfigsCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceLBConfigsSetID(d)
	resourceDuploServiceLBConfigsRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploServiceLBConfigsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceLBConfigsUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceLBConfigsSetID(d)
	resourceDuploServiceLBConfigsRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceDuploServiceLBConfigsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceLBConfigsDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceDuploServiceLBConfigsDelete ******** end")

	return diags
}
