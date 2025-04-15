package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureTenantKeyVaultSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the DuploCloud tenant that the key vault secret will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Key Vault Secret.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"vault_name": {
			Description: "Name of the Key Vault where the Secret should be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"value": {
			Description: "Specifies the value of the Key Vault Secret. Changing this will create a new version of the Key Vault Secret.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"content_type": {
			Description:      "Specifies the content type for the Key Vault Secret.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"azure_id": {
			Description: "The azure ID of the Key Vault secret.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version": {
			Description: "The current version of the Key Vault Secret.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"vault_base_url": {
			Description: "Base URL of the Azure Key Vault",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"recovery_level": {
			Description: "Reflects the deletion recovery level currently in effect for secrets in the current vault.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureTenantKeyVaultSecret() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_tenant_key_vault_secret` manages a azure Key Vault secret in DuploCloud.",

		ReadContext:   resourceAzureTenantKeyVaultSecretRead,
		CreateContext: resourceAzureTenantKeyVaultSecretCreate,
		UpdateContext: resourceAzureTenantKeyVaultSecretUpdate,
		DeleteContext: resourceAzureTenantKeyVaultSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureTenantKeyVaultSecretSchema(),
	}
}

func resourceAzureTenantKeyVaultSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, vaultName, name, err := parseAzureTenantKeyVaultSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretRead(%s, %s, %s): start", tenantID, vaultName, name)

	c := m.(*duplosdk.Client)
	resp, clientErr := c.TenantKeyVaultSecretGet(tenantID, vaultName, name)
	if resp == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure tenant key vault secret %s : %s", tenantID, name, clientErr)
	}

	flattenAzureTenantKeyVaultSecret(tenantID, d, resp)

	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureTenantKeyVaultSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	vaultName := d.Get("vault_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretCreate(%s, %s, %s): start", tenantID, vaultName, name)
	c := m.(*duplosdk.Client)
	rq := expandAzureTenantKeyVaultSecret(d)
	err = c.TenantKeyVaultSecretCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure key vault secret '%s', '%s': %s", tenantID, vaultName, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, vaultName, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure tenant key vault secret", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantKeyVaultSecretGet(tenantID, vaultName, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAzureTenantKeyVaultSecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureTenantKeyVaultSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureTenantKeyVaultSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, vaultName, name, err := parseAzureTenantKeyVaultSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretDelete(%s,%s,%s): start", tenantID, vaultName, name)

	c := m.(*duplosdk.Client)
	clientErr := c.TenantKeyVaultSecretDelete(tenantID, vaultName, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure tenant key vault secret'%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure tenant key vault secret", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantKeyVaultSecretGet(tenantID, vaultName, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureTenantKeyVaultSecretDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureTenantKeyVaultSecret(d *schema.ResourceData) *duplosdk.DuploAzureTenantKeyVaultSecretRequest {
	request := duplosdk.DuploAzureTenantKeyVaultSecretRequest{
		VaultName:   d.Get("vault_name").(string),
		SecretName:  d.Get("name").(string),
		SecretValue: d.Get("value").(string),
	}
	if v, ok := d.GetOk("content_type"); ok && v != nil && v.(string) != "" {
		request.ContentType = v.(string)
	}
	return &request
}

func parseAzureTenantKeyVaultSecretIdParts(id string) (tenantID, vaultName, name string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, vaultName, name = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureTenantKeyVaultSecret(tenantID string, d *schema.ResourceData, duplo *duplosdk.DuploAzureTenantKeyVaultSecret) {
	d.Set("tenant_id", tenantID)
	d.Set("name", duplo.SecretIdentifier.Name)
	d.Set("vault_name", getVaultName(duplo.SecretIdentifier.Vault))
	d.Set("value", duplo.Value)
	if len(duplo.ContentType) > 0 {
		d.Set("content_type", duplo.ContentType)
	}
	d.Set("azure_id", duplo.ID)
	d.Set("version", duplo.SecretIdentifier.Version)
	d.Set("vault_base_url", duplo.SecretIdentifier.BaseIdentifier)
	d.Set("recovery_level", duplo.Attributes.RecoveryLevel)
}

func getVaultName(vaultURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(vaultURL)
	if err != nil {
		return ""
	}

	host := parsedURL.Host

	// Split the hostname to get the vault name
	parts := strings.Split(host, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
