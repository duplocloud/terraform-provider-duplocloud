package duplocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"
)

func resourceTenantCleanUpTimers() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage tenant expiry in DuploCloud",
		ReadContext:   resourceTenantExpiryRead,
		CreateContext: resourceTenantCleanUpTimersCreate,
		UpdateContext: resourceTenantCleanUpTimersUpdate,
		DeleteContext: resourceTenantCleanUpTimersDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant that the expiry will be created in.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"expiry_time": {
				Description: "The expiry time of the tenant.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"pause_time": {
				Description: "The time to pause the tenant.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"remove_expiry_time": {
				Description: "Whether to remove the expiry time.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"remove_pause_time": {
				Description: "Whether to remove the pause time.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func resourceTenantExpiryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Id()
	if tenantId == "" {
		return diag.Errorf("Invalid resource ID: %s", tenantId)
	}

	log.Printf("[TRACE] resourceTenantExpiryRead(%s): start", tenantId)

	c := m.(*duplosdk.Client)
	tenantCleanUpTimers, err := c.GetTenantCleanUpTimers(tenantId)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant '%s': %s", tenantId, err)
	}

	if tenantCleanUpTimers == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("expiry_time", tenantCleanUpTimers.ExpiryTime)
	d.Set("pause_time", tenantCleanUpTimers.PauseTime)

	log.Printf("[TRACE] resourceTenantExpiryRead(%s): end", tenantId)
	return nil
}

func resourceTenantCleanUpTimersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTenantCleanUpTimersCreateOrUpdate(ctx, d, m, false)
}

func resourceTenantCleanUpTimersUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTenantCleanUpTimersCreateOrUpdate(ctx, d, m, true)
}

func resourceTenantCleanUpTimersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duplosdk.DuploTenantCleanUpTimersUpdateRequest{
		TenantId:         d.Get("tenant_id").(string),
		RemoveExpiryTime: true,
		RemovePauseTime:  true,
	}
	log.Printf("[TRACE] resourceTenantCleanUpTimersDelete(%s): start", rq.TenantId)

	c := m.(*duplosdk.Client)
	err := c.UpdateTenantCleanUpTimers(&rq)
	if err != nil {
		return diag.Errorf("Error deleting tenant cleanup timers '%s': %s", rq.TenantId, err)
	}

	log.Printf("[TRACE] resourceTenantCleanUpTimersDelete(%s): end", rq.TenantId)
	return nil
}

func resourceTenantCleanUpTimersCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, isUpdate bool) diag.Diagnostics {
	rq := duplosdk.DuploTenantCleanUpTimersUpdateRequest{
		TenantId:         d.Get("tenant_id").(string),
		ExpiryTime:       d.Get("expiry_time").(string),
		PauseTime:        d.Get("pause_time").(string),
		RemoveExpiryTime: d.Get("remove_expiry_time").(bool),
		RemovePauseTime:  d.Get("remove_pause_time").(bool),
	}

	log.Printf("[TRACE] resourceTenantCleanUpTimersCreateOrUpdate(%s): start", rq.TenantId)

	diags := validateTenantCleanUpTimersUpdateRequest(&rq)
	if diags != nil {
		return diags
	}

	c := m.(*duplosdk.Client)
	err := c.UpdateTenantCleanUpTimers(&rq)
	if err != nil {
		return diag.Errorf("Error updating tenant cleanup timers '%s': %s", rq.TenantId, err)
	}

	if !isUpdate {
		d.SetId(rq.TenantId)
	}

	return nil
}

func validateTenantCleanUpTimersUpdateRequest(rq *duplosdk.DuploTenantCleanUpTimersUpdateRequest) diag.Diagnostics {
	var diags diag.Diagnostics
	const layout = "2006-01-02T15:04:05Z"

	// Validate ExpiryTime
	if !rq.RemoveExpiryTime && rq.ExpiryTime != "" {
		if _, err := time.Parse(layout, rq.ExpiryTime); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Expiry Time",
				Detail:   fmt.Sprintf("Expiry Time is not in the expected format (YYYY-MM-DDTHH:MM:SS): %v", err),
			})
		}
	}

	// Validate PauseTime
	if !rq.RemovePauseTime && rq.PauseTime != "" {
		if _, err := time.Parse(layout, rq.PauseTime); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Pause Time",
				Detail:   fmt.Sprintf("Pause Time is not in the expected format (YYYY-MM-DDTHH:MM:SS): %v", err),
			})
		}
	}

	// Validate TenantId
	if rq.TenantId == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid Tenant ID",
			Detail:   "Tenant ID cannot be empty.",
		})
	}

	// Additional logic checks, if both RemoveExpiryTime and ExpiryTime are set
	if rq.RemoveExpiryTime && rq.ExpiryTime != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Conflicting Expiry Time settings",
			Detail:   "Both RemoveExpiryTime and ExpiryTime are set, which is conflicting.",
		})
	}

	// Similar check for PauseTime
	if rq.RemovePauseTime && rq.PauseTime != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Conflicting Pause Time settings",
			Detail:   "Both RemovePauseTime and PauseTime are set, which is conflicting.",
		})
	}

	return diags
}
