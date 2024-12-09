package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureAvailablitySetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the host will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description:  "The name for availability set",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9._-]{0,78}[a-zA-Z0-9_]$"), `The length must be between 1 and 80 characters. The first character must be a letter or number. The last character must be a letter, number, or underscore. The remaining characters must be letters, numbers, periods, underscores, or dashes.`),
		},
		"platform_update_domain_count": {
			Description:  "Specify platform update domain count between 1-20, for availability set. Virtual machines in the same update domain will be restarted together during planned maintenance. Azure never restarts more than one update domain at a time.",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      5,
			ValidateFunc: validation.IntBetween(1, 20),
		},
		"platform_fault_domain_count": {
			Description:  "Specify platform fault domain count betweem 1-3, for availability set. Virtual machines in the same fault domain share a common power source and physical network switch.",
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(1, 3),
			ForceNew:     true,
			Default:      2,
		},
		"use_managed_disk": {
			Description:  "Set this to `Aligned` if you plan to create virtual machines in this availability set with managed disks.",
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Default:      "Classic",
			ValidateFunc: validation.StringInSlice([]string{"Aligned", "Classic"}, false),
		},
		"location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
		},
		"virtual_machines": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"availability_set_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAzureAvailabilitySet() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_availability_set` manages logical groupings of VMs that enhance reliability by placing VMs in different fault domains to minimize correlated failures, offering improved VM-to-VM latency and high availability, with no extra cost beyond the VM instances themselves, though they may still be affected by shared infrastructure failures.",

		ReadContext:   resourceAzureAvailabilitySetRead,
		CreateContext: resourceAzureAvailabilitySetCreate,
		//	UpdateContext: resourceAzureAvailabilitySetUpdate,
		DeleteContext: resourceAzureAvailabilitySetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAzureAvailablitySetSchema(),
		CustomizeDiff: validateAvailabilitySetAttribute,
	}
}

func resourceAzureAvailabilitySetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	token := strings.Split(id, "/")
	tenantID, name := token[0], token[2]
	log.Printf("[TRACE] resourceAzureAvailabilitySetRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	duplo, clientErr := c.AzureAvailabilitySetGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure virtual machine %s : %s", tenantID, name, clientErr)
	}
	d.Set("tenant_id", tenantID)
	flattenAzureAvailabilitySet(d, duplo)
	log.Printf("[TRACE] resourceAzureAvailabilitySetRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureAvailabilitySetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureAvailabilitySetCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureAvailabilitySet(d)
	err = c.AzureAvailabilitySetCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure availability set '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/availability-set/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "Availability Set", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureAvailabilitySetGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	diags = resourceAzureAvailabilitySetRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureAvailabilitySetCreate(%s, %s): end", tenantID, name)
	return diags
}

/*func resourceAzureAvailabilitySetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}*/

func resourceAzureAvailabilitySetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	token := strings.Split(id, "/")
	tenantID, name := token[0], token[2]
	log.Printf("[TRACE] resourceAzureAvailabilitySetDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	clientErr := c.AzureAvailabilitySetDelete(tenantID, name)
	if clientErr != nil {
		return diag.Errorf("Unable to delete tenant %s azure availablity set '%s': %s", tenantID, name, clientErr)

	}
	log.Printf("[TRACE] resourceAzureAvailabilitySetDelete(%s, %s): end", tenantID, name)

	return nil
}

func expandAzureAvailabilitySet(d *schema.ResourceData) *duplosdk.DuploAvailabilitySet {
	req := duplosdk.DuploAvailabilitySet{}
	if v, ok := d.GetOk("platform_update_domain_count"); ok {
		req.PlatformUpdateDomainCount = v.(int)
	}
	if v, ok := d.GetOk("platform_fault_domain_count"); ok {
		req.PlatformFaultDomainCount = v.(int)
	}
	if v, ok := d.GetOk("use_managed_disk"); ok && v != nil && v.(string) != "" {
		req.Sku.Name = v.(string)
	}
	if v, ok := d.GetOk("name"); ok && v != nil && v.(string) != "" {
		req.Name = v.(string)
	}

	return &req
}

func flattenAzureAvailabilitySet(d *schema.ResourceData, duplo *duplosdk.DuploAvailabilitySetResponse) {
	d.Set("name", duplo.Name)
	d.Set("platform_update_domain_count", duplo.PlatformUpdateDomainCount)
	d.Set("platform_fault_domain_count", duplo.PlatformFaultDomainCount)
	d.Set("use_managed_disk", duplo.Sku.Name)
	d.Set("location", duplo.Location)
	d.Set("tags", duplo.Tags)
	d.Set("type", duplo.Type)
	d.Set("virtual_machines", flattenVMIds(&duplo.VirtualMachines))
	d.Set("availability_set_id", duplo.AvailabilitySetId)
}
func flattenVMIds(duplo *[]duplosdk.VMIds) []interface{} {
	if duplo == nil {
		return []interface{}{}
	}

	list := make([]interface{}, 0, len(*duplo))
	for _, item := range *duplo {
		list = append(list, map[string]interface{}{
			"id": item.Id,
		})
	}

	return list
}

func validateAvailabilitySetAttribute(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	uc := diff.Get("platform_update_domain_count").(int)
	fc := diff.Get("platform_fault_domain_count").(int)

	if fc == 1 && uc != 1 {
		return fmt.Errorf("cannot set platform_update_domain_count to 1 if platform_fault_domain_count is set to 1")
	}
	return nil
}
