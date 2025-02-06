package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureTenantKeyVaultSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the DuploCloud tenant that the key vault will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Key Vault.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"sku_name": {
			Description: "The Name of the SKU used for this Key Vault. Possible values are `standard` and `premium`.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"standard",
				"premium",
			}, false),
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"purge_protection_enabled": {
			Description:      "Is Purge Protection enabled for this Key Vault?",
			Type:             schema.TypeBool,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"soft_delete_retention_days": {
			Description:      "The number of days that items should be retained for once soft-deleted. This value can be between `7` and `90` (the default) days.",
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          90,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"azure_id": {
			Description: "The azure ID of the Key Vault.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"vault_uri": {
			Description: "The URI of the Key Vault, used for performing operations on keys and secrets.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enabled_for_disk_encryption": {
			Description: "Azure Disk Encryption is permitted to retrieve secrets from the vault and unwrap keys.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

func resourceAzureTenantKeyVault() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_tenant_key_vault` manages a azure Key Vault in DuploCloud.",

		ReadContext:   resourceAzureTenantKeyVaultRead,
		CreateContext: resourceAzureTenantKeyVaultCreate,
		UpdateContext: resourceAzureTenantKeyVaultUpdate,
		DeleteContext: resourceAzureTenantKeyVaultDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureTenantKeyVaultSchema(),
	}
}

func resourceAzureTenantKeyVaultRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureTenantKeyVaultIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureTenantKeyVaultRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	resp, clientErr := c.TenantKeyVaultGet(tenantID, name)
	if resp == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure tenant key vault %s : %s", tenantID, name, clientErr)
	}

	flattenAzureTenantKeyVault(tenantID, d, resp)

	log.Printf("[TRACE] resourceAzureTenantKeyVaultRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureTenantKeyVaultCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureTenantKeyVaultCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	rq := expandAzureTenantKeyVault(d)
	err = c.TenantKeyVaultCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure key vault '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure tenant key vault", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantKeyVaultGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAzureTenantKeyVaultRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureTenantKeyVaultCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureTenantKeyVaultUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureTenantKeyVaultDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureTenantKeyVaultIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureTenantKeyVaultDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.TenantKeyVaultDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure tenant key vault '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure tenant key vault", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantKeyVaultGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureTenantKeyVaultDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureTenantKeyVault(d *schema.ResourceData) *duplosdk.DuploAzureTenantKeyVaultRequest {
	request := duplosdk.DuploAzureTenantKeyVaultRequest{
		Name: d.Get("name").(string),
	}
	request.Properties.Sku.Name = d.Get("sku_name").(string)
	request.Properties.SoftDeleteRetentionInDays = d.Get("soft_delete_retention_days").(int)
	request.Properties.EnablePurgeProtection = d.Get("purge_protection_enabled").(bool)
	return &request
}

func parseAzureTenantKeyVaultIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureTenantKeyVault(tenantID string, d *schema.ResourceData, duplo *duplosdk.DuploAzureTenantKeyVault) {
	d.Set("tenant_id", tenantID)
	d.Set("name", duplo.Name)
	d.Set("sku_name", duplo.Properties.Sku.Name)
	d.Set("soft_delete_retention_days", duplo.Properties.SoftDeleteRetentionInDays)
	d.Set("azure_id", duplo.ID)
	d.Set("vault_uri", duplo.Properties.VaultURI)
	d.Set("enabled_for_disk_encryption", duplo.Properties.EnabledForDiskEncryption)
	d.Set("purge_protection_enabled", duplo.Properties.EnablePurgeProtection)
}
