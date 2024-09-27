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
		Description: "`duplocloud_infrastructure_setting` manages a infrastructure's configuration in Duplo.\n\n" +
			"Infrastructure settings are initially populated by Duplo when an infrastructure is created.  This resource " +
			"allows you take control of individual configuration settings for a specific infrastructure.",

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
				Description: "The name of the infrastructure to configure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"exclusions": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"scope": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"daily_maintenance_start_time": {
				Type:     schema.TypeString,
				Required: true,
			},
			"recurring_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"recurrence": {
							Type:     schema.TypeString,
							Required: true,
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

	_, err = c.CreateGCPInfraMaintenanceWindow(infraName, rq)
	if err != nil {
		return diag.Errorf("resourceInfrastructureMaintenanceWindowCreateOrUpdate cannot create maintenance window for infra %s error: %s", infraName, err.Error())
	}
	d.SetId("maintenance/" + infraName)

	diags := resourceInfrastructureMaintenanceWindowRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureSettingCreateOrUpdate(%s): end", infraName)
	return diags
}

func resourceInfrastructureMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandWindowsMaintenance(d *schema.ResourceData) (*duplosdk.DuploGcpInfraMaintenanceWindow, error) {
	layout := "15:04"

	v := d.Get("daily_maintenance_start_time").(string)
	dst := &time.Time{}
	if v == "" {
		dst = nil
	} else {
		t, err := time.Parse(layout, v)
		if err != nil {
			return nil, err
		}
		dst = &t
	}

	obj := &duplosdk.DuploGcpInfraMaintenanceWindow{
		DailyMaintenanceStartTime: dst,
	}
	layout = time.RFC3339

	exclusions := []duplosdk.Exclusion{}
	exc := d.Get("exclusions").([]interface{})
	for _, v := range exc {
		mp := v.(map[string]interface{})
		st, err := time.Parse(layout, mp["start_time"].(string))
		if err != nil {
			return nil, err
		}
		et, err := time.Parse(layout, mp["end_time"].(string))
		if err != nil {
			return nil, err
		}
		exclusions = append(exclusions, duplosdk.Exclusion{
			Scope:     mp["scope"].(string),
			StartTime: st,
			EndTime:   et,
		})
	}

	recW := d.Get("recurring_window").([]interface{})
	if recW != nil {
		mp := recW[0].(map[string]interface{})
		st, err := time.Parse(layout, mp["start_time"].(string))
		if err != nil {
			return nil, err
		}
		et, err := time.Parse(layout, mp["end_time"].(string))
		if err != nil {
			return nil, err
		}
		recurring := &duplosdk.Recurring{
			Recurrence: mp["recurrence"].(string),
			StartTime:  st,
			EndTime:    et,
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
