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

func duploAzureStorageShareFileSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure storage account share file will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"storage_account_name": {
			Description: "Specifies the storage account in which to create the share. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: "The name (or path) of the File that should be created within this File Share. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"url": {
			Description: "The URL of the File Share.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureStorageShareFile() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storage_share_file` manages an Azure storage share file in Duplo.",

		ReadContext:   resourceAzureStorageShareFileRead,
		CreateContext: resourceAzureStorageShareFileCreate,
		UpdateContext: resourceAzureStorageShareFileUpdate,
		DeleteContext: resourceAzureStorageShareFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureStorageShareFileSchema(),
	}
}

func resourceAzureStorageShareFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageShareFileIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageShareFileRead(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.StorageShareFileGet(tenantID, storageAccountName, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure storage share file %s : %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("storage_account_name", storageAccountName)
	d.Set("name", name)
	d.Set("url", duplo.URI)

	log.Printf("[TRACE] resourceAzureStorageShareFileRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureStorageShareFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureStorageShareFileCreate(%s, %s, %s): start", tenantID, storageAccountName, name)
	c := m.(*duplosdk.Client)

	err = c.StorageShareFileCreate(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage account share file '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, storageAccountName, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure storage account share file", id, func() (interface{}, duplosdk.ClientError) {
		return c.StorageShareFileGet(tenantID, storageAccountName, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureStorageShareFileRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureStorageShareFileCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureStorageShareFileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support update.
}

func resourceAzureStorageShareFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support delete.
}

func parseAzureStorageShareFileIdParts(id string) (tenantID, storageAccountName, name string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, storageAccountName, name = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
