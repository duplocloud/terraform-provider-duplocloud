package duplocloud

import (
	"context"

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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"infra_name": {
				Type:     schema.TypeString,
				Optional: false,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				ForceNew: true,
			},
			"cloud": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
				ForceNew: true,
				Default:  0,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"azcount": {
				Type:     schema.TypeInt,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"enable_k8_cluster": {
				Type:     schema.TypeBool,
				Optional: false,
				Required: true,
			},
			"address_prefix": {
				Type:     schema.TypeString,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"subnet_cidr": {
				Type:     schema.TypeInt,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
			},
		},
	}
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
