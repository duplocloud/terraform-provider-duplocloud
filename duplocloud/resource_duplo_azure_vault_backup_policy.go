package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureVaultBackupPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"infra_name": {
			Description:  "The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(2, 12),
		},
		"name": {
			Description: "Specifies the name of the vault backup policy.",
			Type:        schema.TypeString,
			//ForceNew:    true,
			Required: true,
		},
		"azure_id": {
			Description: "Azure id for vault backup policy.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureVaultBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_vault_backup_policy` manages a Vault Backup Policy in Duplo.",

		ReadContext:   resourceAzureVaultBackupPolicyRead,
		CreateContext: resourceAzureVaultBackupPolicyCreate,
		UpdateContext: resourceAzureVaultBackupPolicyUpdate,
		DeleteContext: resourceAzureVaultBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureVaultBackupPolicySchema(),
	}
}

func resourceAzureVaultBackupPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureVaultBackupPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyRead(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	policy, clientErr := c.VaultBackupPolicyGet(infraName, name)
	if policy == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve infra %s azure vault backup policy %s : %s", infraName, name, clientErr)
	}

	flattenAzureVaultBackupPolicy(infraName, d, policy)

	log.Printf("[TRACE] resourceAzureVaultBackupPolicyRead(%s, %s): end", infraName, name)
	return nil
}

func resourceAzureVaultBackupPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	infraName := d.Get("infra_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyCreate(%s, %s): start", infraName, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureVaultBackupPolicy(d)
	err = c.VaultBackupPolicyCreate(infraName, rq)
	if err != nil {
		return diag.Errorf("Error creating infra %s azure vault backup policy '%s': %s", infraName, name, err)
	}

	id := fmt.Sprintf("%s/%s", infraName, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure vault backup policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.VaultBackupPolicyGet(infraName, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAzureVaultBackupPolicyRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyCreate(%s, %s): end", infraName, name)
	return diags
}

func resourceAzureVaultBackupPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureVaultBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	infraName, name, err := parseAzureVaultBackupPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureVaultBackupPolicyDelete(%s, %s): start", infraName, name)

	c := m.(*duplosdk.Client)
	rq := expandAzureVaultBackupPolicy(d)
	clientErr := c.VaultBackupPolicyDelete(infraName, rq)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure vault backup policy '%s': %s", infraName, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure vault backup policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.VaultBackupPolicyGet(infraName, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureVaultBackupPolicyDelete(%s, %s): end", infraName, name)
	return nil
}

func expandAzureVaultBackupPolicy(d *schema.ResourceData) *duplosdk.DuploAzureVaultBackupPolicyPostReq {
	return &duplosdk.DuploAzureVaultBackupPolicyPostReq{
		Name: d.Get("name").(string),
	}
}

func parseAzureVaultBackupPolicyIdParts(id string) (infraName, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		infraName, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureVaultBackupPolicy(infraName string, d *schema.ResourceData, duplo *duplosdk.DuploAzureVaultBackupPolicy) {
	d.Set("name", duplo.Name)
	d.Set("infra_name", infraName)
	d.Set("azure_id", duplo.ID)
}
