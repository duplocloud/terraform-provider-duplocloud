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

func duploAzureKeyVaultSecretSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the key vault secret will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Key Vault Secret. Changing this forces a new resource to be created.",
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

func resourceAzureKeyVaultSecret() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_key_vault_secret` manages a Key Vault Secret in Duplo.",

		ReadContext:   resourceAzureKeyVaultSecretRead,
		CreateContext: resourceAzureKeyVaultSecretCreate,
		UpdateContext: resourceAzureKeyVaultSecretUpdate,
		DeleteContext: resourceAzureKeyVaultSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureKeyVaultSecretSchema(),
	}
}

func resourceAzureKeyVaultSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureKeyVaultSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureKeyVaultSecretRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	secretItem, clientErr := c.KeyVaultSecretGet(tenantID, name)
	if secretItem == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure key vault secret %s : %s", tenantID, name, clientErr)
	}

	flattenAzureKeyVaultSecret(d, secretItem)

	log.Printf("[TRACE] resourceAzureKeyVaultSecretRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureKeyVaultSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureKeyVaultSecretCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureKeyVaultSecret(d)
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

	diags = resourceAzureKeyVaultSecretRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureKeyVaultSecretCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureKeyVaultSecretUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureKeyVaultSecretDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureKeyVaultSecretIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureKeyVaultSecretDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.KeyVaultSecretDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
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

	log.Printf("[TRACE] resourceAzureKeyVaultSecretDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureKeyVaultSecret(d *schema.ResourceData) *duplosdk.DuploAzureKeyVaultRequest {
	return &duplosdk.DuploAzureKeyVaultRequest{
		SecretName:  d.Get("name").(string),
		SecretValue: d.Get("value").(string),
		SecretType:  d.Get("type").(string),
	}
}

func parseAzureKeyVaultSecretIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureKeyVaultSecret(d *schema.ResourceData, duplo *duplosdk.DuploAzureSecretItem) {
	d.Set("fullname", duplo.Identifier.Name)
	d.Set("vault_base_url", duplo.Identifier.BaseIdentifier)
	d.Set("version", duplo.Identifier.Version)
	d.Set("recovery_level", duplo.Attributes.RecoveryLevel)
	d.Set("enabled", duplo.Attributes.Enabled)
}
