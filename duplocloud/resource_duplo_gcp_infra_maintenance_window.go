package duplocloud

import (
	"context"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Resource for managing an infrastructure's settings.
func resourceGCPInfraMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_infra_maintenance_window` applies maintenance window to an gcp infrastructure",

		ReadContext:   resourceInfrastructureMaintenanceWindowRead,
		CreateContext: resourceInfrastructureMaintenanceWindowCreateOrUpdate,
		UpdateContext: resourceInfrastructureMaintenanceWindowCreateOrUpdate,
		DeleteContext: resourceInfrastructureMaintenanceWindowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"infra_name": {
				Description: "The name of the infrastructure where maintenance windows need to be scheduled.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"exclusions": {
				Description: "Exceptions to maintenance window. Non-emergency maintenance should not occur in these windows. A cluster can have up to 20 maintenance exclusions at a time",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validateDateTimeFormat,
						},
						"end_time": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validateDateTimeFormat,
						},
						"scope": {
							Description: "The scope of automatic upgrades to restrict in the exclusion window. One of: NO_UPGRADES | NO_MINOR_UPGRADES | NO_MINOR_OR_NODE_UPGRADES",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"daily_maintenance_start_time": {
				Description:      "Time window specified for daily maintenance operations. Specify 'start_time 'in RFC3339 format HH:MM, where HH : [00-23] and MM : [00-59] GMT",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"recurring_window"},
				ValidateDiagFunc: validateDateTimeFormat,
			},
			"recurring_window": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"daily_maintenance_start_time"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type: schema.TypeString,

							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: validateDateTimeFormat,
						},
						"end_time": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: validateDateTimeFormat,
						},
						"recurrence": {
							Description: "Specify recurrence in RFC5545 RRULE format, to specify when this recurs.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
		},
	}
}

func resourceInfrastructureMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	token := d.Id()
	infraName := strings.Split(token, "/")[1]
	log.Printf("[TRACE] resourceInfrastructureMaintenanceWindowRead(%s): start", infraName)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GetGCPInfraMaintenanceWindow(infraName)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure maintenance window details for '%s': %s", infraName, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("infra_name", infraName)
	flattenWindowMaintenance(d, *duplo)

	log.Printf("[TRACE] resourceInfrastructureMaintenanceWindowRead(%s): end", infraName)
	return nil
}

func resourceInfrastructureMaintenanceWindowCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName := d.Get("infra_name").(string)
	rq, err := expandWindowsMaintenance(d)
	if err != nil {
		return diag.Errorf("resourceInfrastructureMaintenanceWindowCreateOrUpdate cannot create maintenance window for infra %s error: %s", infraName, err.Error())
	}

	c := m.(*duplosdk.Client)

	err = c.CreateGCPInfraMaintenanceWindow(infraName, rq)
	if err != nil {
		return diag.Errorf("resourceInfrastructureMaintenanceWindowCreateOrUpdate cannot create maintenance window for infra %s error: %s", infraName, err.Error())
	}
	d.SetId("maintenance-window/" + infraName)

	diags := resourceInfrastructureMaintenanceWindowRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureSettingCreateOrUpdate(%s): end", infraName)
	return diags
}

func resourceInfrastructureMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandWindowsMaintenance(d *schema.ResourceData) (*duplosdk.DuploGcpInfraMaintenanceWindow, error) {

	dst := d.Get("daily_maintenance_start_time").(string)

	obj := &duplosdk.DuploGcpInfraMaintenanceWindow{
		DailyMaintenanceStartTime: dst,
	}

	exclusions := []duplosdk.Exclusion{}
	exc := d.Get("exclusions").([]interface{})
	for _, v := range exc {
		mp := v.(map[string]interface{})
		exclusions = append(exclusions, duplosdk.Exclusion{
			Scope:     mp["scope"].(string),
			StartTime: mp["start_time"].(string),
			EndTime:   mp["end_time"].(string),
		})
	}

	recW := d.Get("recurring_window").([]interface{})
	if len(recW) > 0 {
		mp := recW[0].(map[string]interface{})
		recurring := &duplosdk.Recurring{
			Recurrence: mp["recurrence"].(string),
			StartTime:  mp["start_time"].(string),
			EndTime:    mp["end_time"].(string),
		}
		obj.RecurringWindow = recurring
	}

	if len(exclusions) > 0 {
		obj.Exclusions = &exclusions
	}
	return obj, nil
}

func flattenWindowMaintenance(d *schema.ResourceData, rb duplosdk.DuploGcpInfraMaintenanceWindow) {
	d.Set("daily_maintenance_start_time", rb.DailyMaintenanceStartTime)
	if rb.Exclusions != nil {
		i := make([]interface{}, 0, len(*rb.Exclusions))
		for _, v := range *rb.Exclusions {
			mp := map[string]interface{}{
				"start_time": v.StartTime,
				"end_time":   v.EndTime,
				"scope":      v.Scope,
			}
			i = append(i, mp)
		}
		d.Set("exclusions", i)
	}
	if rb.RecurringWindow != nil {
		ri := make([]interface{}, 0, 1)
		mpr := map[string]interface{}{
			"start_time": rb.RecurringWindow.StartTime,
			"end_time":   rb.RecurringWindow.EndTime,
			"recurrence": rb.RecurringWindow.Recurrence,
		}
		ri = append(ri, mpr)
		d.Set("recurring_window", ri)
	}
}
