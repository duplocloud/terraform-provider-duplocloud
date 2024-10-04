package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureStorageQueueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure storage class queue will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"storage_account_name": {
			Description: "Specifies the storage class in which to create the queue. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the Queue. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"url": {
			Description: "The URL of the Queue.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
	}
}

func resourceAzureStorageQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storageclass_queue` manages an Azure storage class queue in Duplo.",

		ReadContext:   resourceAzureStorageQueueRead,
		CreateContext: resourceAzureStorageQueueCreate,
		UpdateContext: resourceAzureStorageQueueUpdate,
		DeleteContext: resourceAzureStorageQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureStorageQueueSchema(),
	}
}

func resourceAzureStorageQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageQueueRead(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureStorageAccountQueueGet(tenantID, storageAccountName, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure storage class queue %s : %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("storage_account_name", storageAccountName)
	d.Set("name", name)
	d.Set("url", duplo.URI)

	log.Printf("[TRACE] resourceAzureStorageQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureStorageQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureStorageQueueCreate(%s, %s, %s): start", tenantID, storageAccountName, name)
	c := m.(*duplosdk.Client)

	err = c.AzureStorageAccountQueueCreate(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage class queue '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, storageAccountName, "queue", name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure storage class queue ", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureStorageAccountQueueGet(tenantID, storageAccountName, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureStorageQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureStorageQueueCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureStorageQueueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil // Backend doesn't support update.
}

func resourceAzureStorageQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageQueueDelete(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	err = c.AzureStorageAccountQueueDelete(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage class queue '%s': %s", tenantID, name, err)
	}
	log.Printf("[TRACE] resourceAzureStorageQueueDelete(%s, %s): end", tenantID, name)

	return nil // Backend doesn't support delete.
}
