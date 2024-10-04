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

func duploAzureStorageBlobSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure storage account share file will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"storage_account_name": {
			Description: "Specifies the storage account in which to create the blob. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the Blob. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"url": {
			Description: "The URL of the Blob.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureStorageBlob() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storageclass_blob` manages an Azure storage class blob in Duplo.",

		ReadContext:   resourceAzureStorageBlobRead,
		CreateContext: resourceAzureStorageBlobCreate,
		UpdateContext: resourceAzureStorageBlobUpdate,
		DeleteContext: resourceAzureStorageBlobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureStorageBlobSchema(),
	}
}

func resourceAzureStorageBlobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageBlobRead(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureStorageAccountBlobGet(tenantID, storageAccountName, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
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

	log.Printf("[TRACE] resourceAzureStorageBlobRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureStorageBlobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureStorageBlobCreate(%s, %s, %s): start", tenantID, storageAccountName, name)
	c := m.(*duplosdk.Client)

	err = c.AzureStorageAccountBlobCreate(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage account blob '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, storageAccountName, "queue", name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure storage account blob ", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureStorageAccountBlobGet(tenantID, storageAccountName, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureStorageBlobRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureStorageBlobCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureStorageBlobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support update.
}

func resourceAzureStorageBlobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageBlobDelete(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	err = c.AzureStorageAccountBlobDelete(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage account share file '%s': %s", tenantID, name, err)
	}
	log.Printf("[TRACE] resourceAzureStorageBlobDelete(%s, %s): end", tenantID, name)

	return nil // Backend doesn't support delete.
}

func parseAzureStorageResourceIdParts(id string) (tenantID, storageAccountName, name string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, storageAccountName, name = idParts[0], idParts[1], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
