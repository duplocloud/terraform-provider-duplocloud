package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureVmFeatureSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure vm feature will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"component_id": {
			Description: "Specifies the name of the VM created in duplo. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"feature_name": {
			Description: "The name of the VM feature to be enabled.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"enabled": {
			Description: "The status of the VM feature. By default, this is set to false.",
			Type:        schema.TypeBool,
			Required:    true,
		},
	}
}

func resourceAzureVmFeature() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_vm_feature` manages an Azure VM Feature in Duplo.",

		ReadContext:   resourceAzureVmFeatureRead,
		CreateContext: resourceAzureVmFeatureCreate,
		UpdateContext: resourceAzureVmFeatureUpdate,
		DeleteContext: resourceAzureVmFeatureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureVmFeatureSchema(),
	}
}

func resourceAzureVmFeatureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, vmName, featureName, err := parseAzureVmFeatureIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVmFeatureRead(%s, %s, %s): start", tenantID, vmName, featureName)

	c := m.(*duplosdk.Client)

	duploVm, clientErr := c.AzureVirtualMachineGet(tenantID, vmName)
	if duploVm == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure vm feature %s : %s", tenantID, vmName, clientErr)
	}

	tags := duploVm.Tags

	d.Set("tenant_id", tenantID)
	d.Set("component_id", vmName)
	d.Set("feature_name", featureName)
	if tags != nil {
		for k, v := range tags {
			log.Printf("[TRACE] (feature_tag_key, feature_name)(%s, %s)", k, duplosdk.AzureVmFeatures[featureName])
			if k == duplosdk.AzureVmFeatures[featureName] {
				val, err := strconv.ParseBool(v.(string))
				log.Printf("[TRACE] (feature_tag_key, feature_tag_value)(%s, %v)", k, val)
				if err == nil {
					d.Set("enabled", val)
					break
				}
			}
		}
	}

	log.Printf("[TRACE] resourceAzureVmFeatureRead(%s, %s, %s): end", tenantID, vmName, featureName)
	return nil
}

func resourceAzureVmFeatureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	componentId := d.Get("component_id").(string)
	featureName := d.Get("feature_name").(string)
	enabled := d.Get("enabled").(bool)
	log.Printf("[TRACE] resourceAzureVmFeatureCreate(%s, %s, %s): start", tenantID, componentId, featureName)
	c := m.(*duplosdk.Client)
	err = c.UpdateAzureVmFeature(tenantID, duplosdk.DuploAzureVmFeature{
		ComponentId: componentId,
		FeatureName: featureName,
		Disable:     !enabled,
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure vm feature '%s': %s", tenantID, featureName, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, componentId, featureName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure vm feature", id, func() (interface{}, duplosdk.ClientError) {
		resp, err := c.AzureVirtualMachineTagExists(tenantID, componentId, duplosdk.AzureVmFeatures[featureName])
		if !resp {
			return nil, err
		}
		return resp, err
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureVmFeatureRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVmFeatureCreate(%s, %s, %s): end", tenantID, componentId, featureName)
	return diags
}

func resourceAzureVmFeatureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAzureVmFeatureCreate(ctx, d, m)
}

func resourceAzureVmFeatureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, vmName, featureName, err := parseAzureVmFeatureIdParts(id)
	log.Printf("[TRACE] resourceAzureVmFeatureDelete(%s, %s, %s): start", tenantID, vmName, featureName)
	c := m.(*duplosdk.Client)

	err = c.UpdateAzureVmFeature(tenantID, duplosdk.DuploAzureVmFeature{
		ComponentId: vmName,
		FeatureName: featureName,
		Disable:     true,
	})
	if err != nil {
		return diag.Errorf("Error deleting tenant %s azure vm feature '%s': %s", tenantID, featureName, err)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure virtual machine feature", id, func() (interface{}, duplosdk.ClientError) {
		resp, err := c.AzureVirtualMachineTagExists(tenantID, vmName, duplosdk.AzureVmFeatures[featureName])
		if resp {
			return resp, err
		}
		return nil, err
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVmFeatureDelete(%s, %s, %s): end", tenantID, vmName, featureName)
	return nil
}

func parseAzureVmFeatureIdParts(id string) (tenantID, vmName, featureName string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, vmName, featureName = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
