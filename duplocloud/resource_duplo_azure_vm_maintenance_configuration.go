package duplocloud

import (
	"context"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an infrastructure's settings.
func resourceAzureVmMaintenanceConfig() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_vm_maintenance_configuration` applies maintenance window to an gcp infrastructure",

		ReadContext:   resourceVmMaintenanceConfigRead,
		CreateContext: resourceVmMaintenanceConfigCreate,
		UpdateContext: resourceVmMaintenanceConfigUpdate,
		DeleteContext: resourceVmMaintenanceConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the azure vm feature will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"vm_name": {
				Description: "The name of the virtual machine where maintenance configuration need to be configured.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"visiblity": {
				Description: "he visibility of the Maintenance Configuration. The only allowable value is Custom.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Custom",
			},

			"window": {
				Description: "Window block to schedule maintenance",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Description: "Effective start date of the maintenance window in YYYY-MM-DD hh:mm format.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"expiration_time": {
							Description: "Effective expiration date of the maintenance window in YYYY-MM-DD hh:mm format.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"duration": {
							Description: "The duration of the maintenance window in HH:mm format.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"recur_every": {
							Description: "he rate at which a maintenance window is expected to recur. The rate can be expressed as daily, weekly, or monthly schedules.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							//	ValidateFunc: validation.StringInSlice([]string{"daily", "weekly", "monthly"}, false),
						},

						"time_zone": {
							Description: "The timezone on which maintenance should be scheduled.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceVmMaintenanceConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceVmMaintenanceConfigRead(%s): start", vmName)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.AzureVmMaintenanceConfigurationGet(tenantId, vmName)
	if err != nil {
		return diag.Errorf("Unable to retrieve vm maintenance configuration details for '%s': %s", vmName, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("vm_name", vmName)
	flattenVmMaintenance(d, *duplo)

	log.Printf("[TRACE] resourceVmMaintenanceConfigRead(%s): end", vmName)
	return nil
}

func resourceVmMaintenanceConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	vmName := d.Get("vm_name").(string)
	log.Printf("[TRACE] resourceVmMaintenanceConfigCreate(%s): start", vmName)

	rq := expandVmMaintenance(d)

	c := m.(*duplosdk.Client)

	err := c.AzureVmMaintenanceConfigurationCreate(tenantId, vmName, rq)
	if err != nil {
		return diag.Errorf("resourceVmMaintenanceConfigCreate cannot create maintenance config for vm %s error: %s", vmName, err.Error())
	}
	d.SetId(tenantId + "/" + vmName + "/maintenance-configuration")

	diags := resourceVmMaintenanceConfigRead(ctx, d, m)
	log.Printf("[TRACE] resourceVmMaintenanceConfigCreate(%s): end", vmName)
	return diags
}

func resourceVmMaintenanceConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceVmMaintenanceConfigUpdate(%s): start", vmName)
	rq := expandVmMaintenance(d)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	err := c.AzureVmMaintenanceConfigurationUpdate(tenantId, vmName, rq)
	if err != nil {
		return diag.Errorf("Unable to retrieve vm maintenance configuration details for '%s': %s", vmName, err)
	}
	diags := resourceVmMaintenanceConfigRead(ctx, d, m)

	log.Printf("[TRACE] resourceVmMaintenanceConfigCreate(%s): end", vmName)

	return diags
}

func resourceVmMaintenanceConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceVmMaintenanceConfigDelete(%s): start", vmName)
	c := m.(*duplosdk.Client)
	err := c.AzureVmMaintenanceConfigurationDelete(tenantId, vmName)
	if err != nil {
		return diag.Errorf("Unable to retrieve vm maintenance configuration details for '%s': %s", vmName, err)
	}
	d.SetId("")
	log.Printf("[TRACE] resourceVmMaintenanceConfigDelete(%s): end", vmName)

	return nil
}

func expandVmMaintenance(d *schema.ResourceData) *duplosdk.DuploAzureVmMaintenanceWindow {
	obj := duplosdk.DuploAzureVmMaintenanceWindow{}

	obj.StartDateTime = d.Get("window.0.start_time").(string)
	obj.ExpirationDateTime = d.Get("window.0.expiration_time").(string)
	obj.Duration = d.Get("window.0.duration").(string)
	obj.RecurEvery = d.Get("window.0.recur_every").(string)
	obj.Visibility = d.Get("visiblity").(string)
	obj.TimeZone = d.Get("window.0.time_zone").(string)
	return &obj
}

func flattenVmMaintenance(d *schema.ResourceData, rb duplosdk.DuploAzureVmMaintenanceWindow) {
	mp := map[string]interface{}{
		"start_time":      rb.StartDateTime,
		"expiration_time": rb.ExpirationDateTime,
		"duration":        rb.Duration,
		"recur_every":     rb.RecurEvery,
		"time_zone":       rb.TimeZone,
	}
	i := []interface{}{}
	i = append(i, mp)
	d.Set("window", i)
	d.Set("visiblity", rb.Visibility)
}
