package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"account_name": {
			Type:     schema.TypeString,
			ForceNew: true, // Change tenant name
			Required: true,
		},
		"plan_id": {
			Type:     schema.TypeString,
			ForceNew: true, // Change plan (infrastructure)
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
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] resourceTenantRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGet(tenantID)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant '%s': %s", tenantID, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.Set("account_name", duplo.AccountName)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("plan_id", duplo.PlanID)
	d.Set("infra_owner", duplo.InfraOwner)

	// Next, set nested fields.
	if duplo.TenantPolicy != nil {
		d.Set("policy", []map[string]interface{}{{
			"allow_volume_mapping": true,
			"block_external_ep":    true,
		}})
	}
	d.Set("tags", duplosdk.KeyValueToState("tags", duplo.Tags))

	log.Printf("[TRACE] resourceTenantRead(%s): end", tenantID)
	return nil
}

/// CREATE resource
func resourceTenantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duplosdk.DuploTenant{
		AccountName: d.Get("account_name").(string),
		PlanID:      d.Get("plan_id").(string),
	}

	log.Printf("[TRACE] resourceTenantCreate(%s): start", rq.AccountName)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	rp, err := c.TenantCreate(rq)
	if err != nil {
		return diag.Errorf("Unable to create tenant '%s': %s", rq.AccountName, err)
	}

	d.SetId(fmt.Sprintf("v2/admin/TenantV2/%s", rp.TenantID))

	// Wait for 3 minutes to allow infrastructure creation.
	time.Sleep(time.Duration(3) * time.Minute)

	diags := resourceTenantRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantCreate(%s): end", rq.AccountName)
	return diags
}

/// DELETE resource
func resourceTenantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] resourceTenantDelete(%s): start", tenantID)

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	_, err := c.TenantDelete(tenantID)
	if err != nil {
		return diag.Errorf("Error deleting tenant '%s': %s", tenantID, err)
	}

	log.Printf("[TRACE] resourceTenantDelete(%s): end", tenantID)
	return nil
}
