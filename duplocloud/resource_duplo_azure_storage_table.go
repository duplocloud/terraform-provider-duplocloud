package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureStorageTableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure storage class table will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"storage_account_name": {
			Description: "Specifies the storage class in which to create the table. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description:  "The name of the Table. Changing this forces a new resource to be created.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]{3,63}$`), "Invalid table name. Name should be between 3 to 63 character, Can contain alphanumeric and hypen character."),
		},
		"url": {
			Description: "The URL of the Table.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureStorageTable() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storageclass_table` manages an Azure storage class table in Duplo.",

		ReadContext:   resourceAzureStorageTableRead,
		CreateContext: resourceAzureStorageTableCreate,
		DeleteContext: resourceAzureStorageTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureStorageTableSchema(),
	}
}

func resourceAzureStorageTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageTableRead(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AzureStorageAccountTableGet(tenantID, storageAccountName, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure storage class table %s : %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("storage_account_name", storageAccountName)
	d.Set("name", name)
	d.Set("url", duplo.URI)

	log.Printf("[TRACE] resourceAzureStorageTableRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureStorageTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	storageAccountName := d.Get("storage_account_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureStorageTableCreate(%s, %s, %s): start", tenantID, storageAccountName, name)
	c := m.(*duplosdk.Client)

	err = c.AzureStorageAccountTableCreate(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage class table '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, storageAccountName, "table", name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure storage class table ", id, func() (interface{}, duplosdk.ClientError) {
		return c.AzureStorageAccountTableGet(tenantID, storageAccountName, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzureStorageTableRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureStorageTableCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureStorageTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	tenantID, storageAccountName, name, err := parseAzureStorageResourceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageTableDelete(%s, %s, %s): start", tenantID, storageAccountName, name)

	c := m.(*duplosdk.Client)
	err = c.AzureStorageAccountTableDelete(tenantID, storageAccountName, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure storage class table '%s': %s", tenantID, name, err)
	}
	log.Printf("[TRACE] resourceAzureStorageTableDelete(%s, %s): end", tenantID, name)

	return nil // Backend doesn't support delete.
}
