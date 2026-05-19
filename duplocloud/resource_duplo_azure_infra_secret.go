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

func duploAzureInfraSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the key vault secret will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Key Vault Secret. Must start with the tenant's account name (case-insensitive); the Duplo backend filters list results by this prefix, so secrets without it cannot be retrieved after creation. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"fullname": {
			Description: "Duplo will generate name of the Key Vault Secret.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"value": {
			Description: "Specifies the value of the Key vault secret.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
		},
		"type": {
			Description: "Specifies the content type for the Key Vault Secret.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "duplo_container_env",
		},
		"key_vault_id": {
			Description: "The ID of the Key Vault where the Secret should be created.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"vault_base_url": {
			Description: "Base URL of the Azure Key Vault",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version": {
			Description: "The current version of the Key Vault Secret.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"recovery_level": {
			Description: "Reflects the deletion recovery level currently in effect for secrets in the current vault.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enabled": {
			Description: "Determines whether the object is enabled.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

func resourceAzureInfraSecret() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_infra_secret` manages an infrastructure-level Azure Key Vault Secret in Duplo.",

		ReadContext:   resourceAzureInfraSecretRead,
		CreateContext: resourceAzureInfraSecretCreate,
		UpdateContext: resourceAzureInfraSecretUpdate,
		DeleteContext: resourceAzureInfraSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureInfraSecretSchema(),
	}
}

// resourceAzureKeyVaultSecret is the deprecated alias for resourceAzureInfraSecret.
// Retained so existing state using `duplocloud_azure_key_vault_secret` continues to work;
// emits a deprecation warning pointing users to `duplocloud_azure_infra_secret`.
func resourceAzureKeyVaultSecret() *schema.Resource {
	r := resourceAzureInfraSecret()
	r.DeprecationMessage = "`duplocloud_azure_key_vault_secret` has been renamed to `duplocloud_azure_infra_secret`. " +
		"Please update your configuration to use the new resource name. The old name will be removed in a future release."
	return r
}

func resourceAzureInfraSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureInfraSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureInfraSecretRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	secretItem, clientErr := c.KeyVaultSecretGet(tenantID, name)
	if secretItem == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureInfraSecretRead: Azure key vault secret %s not found for tenantId %s, removing from state", name, tenantID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure key vault secret %s : %s", tenantID, name, clientErr)
	}

	flattenAzureInfraSecret(d, secretItem)

	log.Printf("[TRACE] resourceAzureInfraSecretRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureInfraSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureInfraSecretCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	if diags := validateKeyVaultSecretNamePrefix(c, tenantID, name); diags != nil {
		return diags
	}

	rq := expandAzureInfraSecret(d)
	err = c.KeyVaultSecretCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure key vault secret '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure key vault secret", id, func() (interface{}, duplosdk.ClientError) {
		return c.KeyVaultSecretGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAzureInfraSecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureInfraSecretCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureInfraSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureInfraSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureInfraSecretUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq := expandAzureInfraSecret(d)
	if clientErr := c.KeyVaultSecretCreate(tenantID, rq); clientErr != nil {
		return diag.Errorf("Error updating tenant %s azure key vault secret '%s': %s", tenantID, name, clientErr)
	}

	diags := resourceAzureInfraSecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureInfraSecretUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureInfraSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureInfraSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureInfraSecretDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.KeyVaultSecretDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureInfraSecretDelete: Azure key vault secret %s not found for tenantId %s, removing from state", name, tenantID)
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure key vault secret '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure key vault secret", id, func() (interface{}, duplosdk.ClientError) {
		return c.KeyVaultSecretGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureInfraSecretDelete(%s, %s): end", tenantID, name)
	return nil
}

// validateKeyVaultSecretNamePrefix ensures the secret name starts with the tenant's
// account name (case-insensitive). The Duplo backend's ListKeyVaultSecretsForTenant
// filters list results by this prefix, so a secret without it would be created in
// Azure but never returned by the post-create lookup, causing the provider to hang
// until the create timeout expires.
func validateKeyVaultSecretNamePrefix(c *duplosdk.Client, tenantID, name string) diag.Diagnostics {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return diag.Errorf("Unable to fetch tenant %s: %s", tenantID, err)
	}
	if tenant == nil {
		return diag.Errorf("tenant %s not found", tenantID)
	}
	if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(tenant.AccountName)) {
		return diag.Errorf(
			"azure key vault secret name %q must start with the tenant's account name %q (case-insensitive); the Duplo backend filters list results by this prefix and secrets without it cannot be retrieved after creation",
			name, tenant.AccountName,
		)
	}
	return nil
}

func expandAzureInfraSecret(d *schema.ResourceData) *duplosdk.DuploAzureKeyVaultRequest {
	return &duplosdk.DuploAzureKeyVaultRequest{
		SecretName:  d.Get("name").(string),
		SecretValue: d.Get("value").(string),
		ContentType: d.Get("type").(string),
	}
}

func parseAzureInfraSecretIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureInfraSecret(d *schema.ResourceData, duplo *duplosdk.DuploAzureSecretItem) {
	d.Set("fullname", duplo.Identifier.Name)
	d.Set("vault_base_url", duplo.Identifier.BaseIdentifier)
	d.Set("version", duplo.Identifier.Version)
	d.Set("recovery_level", duplo.Attributes.RecoveryLevel)
	d.Set("enabled", duplo.Attributes.Enabled)
	if duplo.ContentType != "" {
		d.Set("type", duplo.ContentType)
	}
}
