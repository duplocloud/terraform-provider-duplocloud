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
func resourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceInfrastructureRead,
		CreateContext: resourceInfrastructureCreate,
		UpdateContext: resourceInfrastructureUpdate,
		DeleteContext: resourceInfrastructureDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.InfrastructureSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceInfrastructureRead,
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
					Schema: *duplosdk.InfrastructureSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceInfrastructureRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.InfrastructureGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.InfrastructuresFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceInfrastructureRead ******** end")

	return diags
}

/// READ resource
func resourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceInfrastructureRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.InfrastructureGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.InfrastructureSetId(d)
	log.Printf("[TRACE] duplo-resourceInfrastructureRead ******** end")
	return diags
}

/// CREATE resource
func resourceInfrastructureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceInfrastructureCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.InfrastructureCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.InfrastructureSetId(d)
	resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceInfrastructureCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceInfrastructureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceInfrastructureUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.InfrastructureUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.InfrastructureSetId(d)
	resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceInfrastructureUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceInfrastructureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceInfrastructureDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.InfrastructureDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceInfrastructureDelete ******** end")

	return diags
}
