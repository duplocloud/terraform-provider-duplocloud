package duplocloud

import (
	"context"
	"log"
	"regexp"
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
		Description: "`duplocloud_azure_vm_maintenance_configuration` manages maintenance window to an azure vm",

		ReadContext:   resourceAzureVmMaintenanceConfigRead,
		CreateContext: resourceAzureVmMaintenanceConfigCreate,
		UpdateContext: resourceAzureVmMaintenanceConfigUpdate,
		DeleteContext: resourceAzureVmMaintenanceConfigDelete,
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
				Description:  "The visibility of the Maintenance Configuration. The only allowable value is Custom.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Custom",
				ValidateFunc: validation.StringInSlice([]string{"Custom"}, false),
			},

			"window": {
				Description: "Block to configure maintenance window",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Description:  "Effective start date of the maintenance window in YYYY-MM-DD HH:MM format.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$`), "valid datetime format is YYYY-MM-DD HH:MM"),
						},
						"expiration_time": {
							Description:  "Effective expiration date of the maintenance window in YYYY-MM-DD hh:mm format.",
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$`), "valid datetime format is YYYY-MM-DD HH:MM"),
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

func resourceAzureVmMaintenanceConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigRead(%s): start", vmName)

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
	flattenAzureVmMaintenance(d, *duplo)

	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigRead(%s): end", vmName)
	return nil
}

func resourceAzureVmMaintenanceConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	vmName := d.Get("vm_name").(string)
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigCreate(%s): start", vmName)

	rq := expandAzureVmMaintenance(d)

	c := m.(*duplosdk.Client)

	err := c.AzureVmMaintenanceConfigurationCreate(tenantId, vmName, rq)
	if err != nil {
		return diag.Errorf("resourceAzureVmMaintenanceConfigCreate cannot create maintenance config for vm %s error: %s", vmName, err.Error())
	}
	d.SetId(tenantId + "/" + vmName + "/maintenance-configuration")

	diags := resourceAzureVmMaintenanceConfigRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigCreate(%s): end", vmName)
	return diags
}

func resourceAzureVmMaintenanceConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigUpdate(%s): start", vmName)
	rq := expandAzureVmMaintenance(d)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	err := c.AzureVmMaintenanceConfigurationUpdate(tenantId, vmName, rq)
	if err != nil {
		return diag.Errorf("Unable to retrieve vm maintenance configuration details for '%s': %s", vmName, err)
	}
	diags := resourceAzureVmMaintenanceConfigRead(ctx, d, m)

	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigUpdate(%s): end", vmName)

	return diags
}

func resourceAzureVmMaintenanceConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tokens := strings.Split(d.Id(), "/")
	tenantId := tokens[0]
	vmName := tokens[1]
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigDelete(%s): start", vmName)
	c := m.(*duplosdk.Client)
	err := c.AzureVmMaintenanceConfigurationDelete(tenantId, vmName)
	if err != nil {
		return diag.Errorf("Unable to retrieve vm maintenance configuration details for '%s': %s", vmName, err)
	}
	d.SetId("")
	log.Printf("[TRACE] resourceAzureVmMaintenanceConfigDelete(%s): end", vmName)

	return nil
}

func expandAzureVmMaintenance(d *schema.ResourceData) *duplosdk.DuploAzureVmMaintenanceWindow {
	obj := duplosdk.DuploAzureVmMaintenanceWindow{}

	obj.StartDateTime = d.Get("window.0.start_time").(string)
	obj.ExpirationDateTime = d.Get("window.0.expiration_time").(string)
	obj.Duration = d.Get("window.0.duration").(string)
	obj.RecurEvery = d.Get("window.0.recur_every").(string)
	obj.Visibility = d.Get("visiblity").(string)
	obj.TimeZone = d.Get("window.0.time_zone").(string)
	return &obj
}

func flattenAzureVmMaintenance(d *schema.ResourceData, rb duplosdk.DuploAzureVmMaintenanceWindow) {
	d.Set("window.0.start_time", rb.StartDateTime)
	d.Set("window.0.expiration_time", rb.ExpirationDateTime)
	d.Set("window.0.duration", rb.Duration)
	d.Set("window.0.recur_every", rb.RecurEvery)
	d.Set("window.0.time_zone", rb.TimeZone)
}
