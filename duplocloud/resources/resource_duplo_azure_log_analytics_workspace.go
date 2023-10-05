package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureLogAnalyticsWorkspaceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description:  "The name of the infrastructure. Infrastructure names are globally unique and less than 13 characters.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(2, 12),
		},
		"resource_group_name": {
			Description: "The name of the resource group in which the Log Analytics workspace is created. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"name": {
			Description: "Specifies the name of the Log Analytics Workspace. Workspace name should include 4-63 letters, digits or '-'. The '-' shouldn't be the first or the last symbol. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until Log Analytics Workspace to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"azure_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"sku": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"retention_in_days": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"public_network_access_for_ingestion": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"public_network_access_for_query": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAzureLogAnalyticsWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_log_analytics_workspace` manages an Azure Log Analytics Workspace in Duplo.",

		ReadContext:   resourceAzureLogAnalyticsWorkspaceRead,
		CreateContext: resourceAzureLogAnalyticsWorkspaceCreate,
		UpdateContext: resourceAzureLogAnalyticsWorkspaceUpdate,
		DeleteContext: resourceAzureLogAnalyticsWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureLogAnalyticsWorkspaceSchema(),
	}
}

func resourceAzureLogAnalyticsWorkspaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureLogAnalyticsWorkspaceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureLogAnalyticsWorkspaceRead(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureLogAnalyticsWorkspaceGet(infraName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure Log Analytics Workspace %s : %s", infraName, name, clientErr)
	}

	d.Set("infra_name", infraName)
	d.Set("resource_group_name", strings.Split(duplo.ID, "/")[4])
	d.Set("name", name)
	d.Set("azure_id", duplo.ID)
	d.Set("sku", duplo.PropertiesSku.Name)
	d.Set("location", duplo.Location)
	d.Set("retention_in_days", duplo.PropertiesRetentionInDays)
	d.Set("public_network_access_for_ingestion", duplo.PropertiesPublicNetworkAccessForIngestion)
	d.Set("public_network_access_for_query", duplo.PropertiesPublicNetworkAccessForQuery)

	log.Printf("[TRACE] resourceAzureLogAnalyticsWorkspaceRead(%s, %s): end", infraName, infraName)
	return nil
}

func resourceAzureLogAnalyticsWorkspaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureLogAnalyticsWorkspaceCreate(%s, %s): start", infraName, name)
	c := m.(*duplosdk.Client)

	err = c.AzureLogAnalyticsWorkspaceCreate(infraName, duplosdk.DuploAzureLogAnalyticsWorkspaceRq{
		Name:          name,
		ResourceGroup: d.Get("resource_group_name").(string),
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure Log Analytics Workspace '%s': %s", infraName, name, err)
	}

	id := fmt.Sprintf("%s/%s", infraName, name)
	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "Azure Log Analytics Workspace", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureLogAnalyticsWorkspaceGet(infraName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = logAnalyticsWorkspaceWaitUntilReady(ctx, c, infraName, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureLogAnalyticsWorkspaceRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureLogAnalyticsWorkspaceCreate(%s, %s): end", infraName, name)
	return diags
}

func resourceAzureLogAnalyticsWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support update.
}

func resourceAzureLogAnalyticsWorkspaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support delete.
}

func parseAzureLogAnalyticsWorkspaceIdParts(id string) (infraName, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		infraName, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func logAnalyticsWorkspaceWaitUntilReady(ctx context.Context, c *duplosdk.Client, infraName string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AzureLogAnalyticsWorkspaceGet(infraName)
			log.Printf("[TRACE] Log Analytics Workspace state is (%s).", rp.PropertiesProvisioningState)
			status := "pending"
			if err == nil {
				if rp.PropertiesProvisioningState == "Succeeded" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] logAnalyticsWorkspaceWaitUntilReady(%s, %s)", infraName, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
