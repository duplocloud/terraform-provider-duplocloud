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

func duploAzureRecoveryServicesVaultSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description:  "The name of the infrastructure. Infrastructure names are globally unique and less than 13 characters.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(2, 12),
		},
		"resource_group_name": {
			Description: "The name of the resource group in which to create the Recovery Services Vault. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"name": {
			Description: "Specifies the name of the Recovery Services Vault. Recovery Service Vault name must be 2 - 50 characters long, start with a letter, contain only letters, numbers and hyphens. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until Recovery Services Vault to be ready, after creation.",
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
	}
}

func resourceAzureRecoveryServicesVault() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_recovery_services_vault` manages an Azure Recovery Services Vault in Duplo.",

		ReadContext:   resourceAzureRecoveryServicesVaultRead,
		CreateContext: resourceAzureRecoveryServicesVaultCreate,
		UpdateContext: resourceAzureRecoveryServicesVaultUpdate,
		DeleteContext: resourceAzureRecoveryServicesVaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureRecoveryServicesVaultSchema(),
	}
}

func resourceAzureRecoveryServicesVaultRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureRecoveryServicesVaultIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureRecoveryServicesVaultRead(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureRecoveryServicesVaultGet(infraName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure Recovery Services Vault %s : %s", infraName, name, clientErr)
	}

	d.Set("infra_name", infraName)
	d.Set("resource_group_name", strings.Split(duplo.ID, "/")[4])
	d.Set("name", name)
	d.Set("azure_id", duplo.ID)
	d.Set("sku", duplo.Sku.Name)
	d.Set("location", duplo.Location)

	log.Printf("[TRACE] resourceAzureRecoveryServicesVaultRead(%s, %s): end", infraName, infraName)
	return nil
}

func resourceAzureRecoveryServicesVaultCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureRecoveryServicesVaultCreate(%s, %s): start", infraName, name)
	c := m.(*duplosdk.Client)

	err = c.AzureRecoveryServicesVaultCreate(infraName, duplosdk.DuploAzureRecoveryServicesVaultRq{
		Name:          name,
		ResourceGroup: d.Get("resource_group_name").(string),
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure Recovery Services Vault '%s': %s", infraName, name, err)
	}

	id := fmt.Sprintf("%s/%s", infraName, name)
	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "Azure Recovery Services Vault", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureRecoveryServicesVaultGet(infraName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = recoveryServicesVaultWaitUntilReady(ctx, c, infraName, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureRecoveryServicesVaultRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureRecoveryServicesVaultCreate(%s, %s): end", infraName, name)
	return diags
}

func resourceAzureRecoveryServicesVaultUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support update.
}

func resourceAzureRecoveryServicesVaultDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support delete.
}

func parseAzureRecoveryServicesVaultIdParts(id string) (infraName, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		infraName, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func recoveryServicesVaultWaitUntilReady(ctx context.Context, c *duplosdk.Client, infraName string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AzureRecoveryServicesVaultGet(infraName)
			log.Printf("[TRACE] Recovery Services Vault state is (%s).", rp.Properties.ProvisioningState)
			status := "pending"
			if err == nil {
				if rp.Properties.ProvisioningState == "Succeeded" {
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
	log.Printf("[DEBUG] recoveryServicesVaultWaitUntilReady(%s, %s)", infraName, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
