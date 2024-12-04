package duplocloud

import (
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAzureAvailabilitySet() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_availability_set` manages a azure availability set in Duplo.",

		Read: dataSourceAzureAvailabilitySetRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"platform_update_domain_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"platform_fault_domain_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"sku_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
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
			"tags": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAzureAvailabilitySetRead(d *schema.ResourceData, m interface{}) error {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourceAzureAvailabilitySetRead(%s, %s): start", tenantID, name)

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
		return fmt.Errorf("Unable to retrieve tenant %s azure virtual machine %s : %s", tenantID, name, clientErr)
	}
	d.SetId(fmt.Sprintf("%s/availability-set/%s", tenantID, name))

	d.Set("tenant_id", tenantID)
	d.Set("name", duplo.Name)
	d.Set("platform_update_domain_count", duplo.PlatformUpdateDomainCount)
	d.Set("platform_fault_domain_count", duplo.PlatformFaultDomainCount)
	d.Set("sku_name", duplo.Sku.Name)
	d.Set("location", duplo.Location)
	d.Set("tags", duplo.Tags)
	d.Set("type", duplo.Type)
	d.Set("virtual_machines", flattenVMIds(&duplo.VirtualMachines))
	d.Set("availability_set_id", duplo.AvailabilitySetId)
	log.Printf("[TRACE] dataSourceAzureAvailabilitySetRead(%s, %s): end", tenantID, name)
	return nil

}
