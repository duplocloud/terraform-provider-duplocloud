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
func resourceK8Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceK8SecretRead,
		CreateContext: resourceK8SecretCreate,
		UpdateContext: resourceK8SecretUpdate,
		DeleteContext: resourceK8SecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: *duplosdk.K8SecretSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceK8Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8SecretRead,
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
					Schema: *duplosdk.K8SecretSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceK8SecretRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.K8SecretGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.K8SecretsFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceK8SecretRead ******** end")

	return diags
}

/// READ resource
func resourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.K8SecretGet(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8SecretSetID(d)
	log.Printf("[TRACE] duplo-resourceK8SecretRead ******** end")
	return diags
}

/// CREATE resource
func resourceK8SecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8SecretSetID(d)
	resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8SecretCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceK8SecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.K8SecretSetID(d)
	resourceK8SecretRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceK8SecretUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceK8SecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceK8SecretDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.K8SecretDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted
	log.Printf("[TRACE] duplo-resourceK8SecretDelete ******** end")

	return diags
}
