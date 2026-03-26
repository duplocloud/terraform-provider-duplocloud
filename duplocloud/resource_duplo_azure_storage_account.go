package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureStorageAccountSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the storage account will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the storage account. Changing this forces a new resource to be created. This must be unique across the entire Azure service, not just within the resource group.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		// "account_tier": {
		// 	Type: schema.TypeString,
		// 	//Required: true,
		// 	Computed: true,
		// 	Optional: true,
		// 	ValidateFunc: validation.StringInSlice([]string{
		// 		"Standard",
		// 		"Premium",
		// 	}, true),
		// },
		// "access_tier": {
		// 	Description: "Defines the access tier for `BlobStorage`, `FileStorage` and `StorageV2` accounts. Valid options are `Hot` and `Cool`, defaults to Hot.",
		// 	Type:        schema.TypeString,
		// 	Optional:    true,
		// 	Computed:    true,
		// 	ValidateFunc: validation.StringInSlice([]string{
		// 		"Hot",
		// 		"Cool",
		// 	}, true),
		// },
		// "enable_https_traffic_only": {
		// 	Description: "Boolean flag which forces HTTPS if enabled.",
		// 	Type:        schema.TypeBool,
		// 	Optional:    true,
		// 	Default:     true,
		// },
		"allow_blob_public_access": {
			Description: "Whether or not to allow public access to all blobs or containers in the storage account.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true,
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until azure storage account to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceAzureStorageAccount() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_storage_account` manages an Azure storage account in Duplo.",

		ReadContext:   resourceAzureStorageAccountRead,
		CreateContext: resourceAzureStorageAccountCreate,
		UpdateContext: resourceAzureStorageAccountUpdate,
		DeleteContext: resourceAzureStorageAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureStorageAccountSchema(),
	}
}

func resourceAzureStorageAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureStorageAccountIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageAccountRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.StorageAccountGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureStorageAccountRead: Azure storage account %s not found for tenantId %s, removing from state", name, tenantID)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure storage account %s : %s", tenantID, name, clientErr)
	}

	// flattenAzureStorageAccount(d, duplo)
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	log.Printf("[TRACE] resourceAzureStorageAccountRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureStorageAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureStorageAccountCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := &duplosdk.DuploAzureStorageAccountCreateRequest{
		Name:                  name,
		AllowBlobPublicAccess: d.Get("allow_blob_public_access").(bool),
	}
	createErr := c.StorageAccountCreate(tenantID, rq)
	if createErr != nil {
		log.Printf("[WARN] resourceAzureStorageAccountCreate(%s, %s): API returned error, will check if resource was created: %s", tenantID, name, createErr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		time.Sleep(time.Duration(60) * time.Second)
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure storage account", id, func() (interface{}, duplosdk.ClientError) {
			return c.StorageAccountGet(tenantID, name)
		})
		if diags != nil {
			return diags
		}
	} else if createErr != nil {
		var found bool
		for i := 0; i < 3; i++ {
			time.Sleep(time.Duration(30) * time.Second)
			resp, getErr := c.StorageAccountGet(tenantID, name)
			if getErr == nil && resp != nil && resp.Name != "" {
				found = true
				break
			}
			log.Printf("[WARN] resourceAzureStorageAccountCreate(%s, %s): retry %d - resource not yet available", tenantID, name, i+1)
		}
		if !found {
			return diag.Errorf("Error creating tenant %s azure storage account '%s': %s", tenantID, name, createErr)
		}
	}

	diags := resourceAzureStorageAccountRead(ctx, d, m)
	for i := 0; i < 2 && diags.HasError(); i++ {
		log.Printf("[WARN] resourceAzureStorageAccountCreate(%s, %s): read failed (attempt %d), retrying...", tenantID, name, i+1)
		time.Sleep(time.Duration(10) * time.Second)
		diags = resourceAzureStorageAccountRead(ctx, d, m)
	}
	log.Printf("[TRACE] resourceAzureStorageAccountCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureStorageAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureStorageAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureStorageAccountIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureStorageAccountDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.StorageAccountDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureStorageAccountDelete: Azure storage account %s not found for tenantId %s, removing from state", name, tenantID)
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure storage account '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure storage account", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.StorageAccountExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureStorageAccountDelete(%s, %s): end", tenantID, name)
	return nil
}

// func expandAzureStorageAccount(d *schema.ResourceData) *duplosdk.DuploAzureStorageAccount {
// 	//TODO implement here
// 	return nil
// }

func parseAzureStorageAccountIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

// func flattenAzureStorageAccount(d *schema.ResourceData, duplo *duplosdk.DuploAzureStorageAccount) {
// 	d.Set("account_tier", duplo.Sku.Tier)
// 	d.Set("access_tier", duplo.PropertiesAccessTier)
// 	d.Set("enable_https_traffic_only", duplo.PropertiesSupportsHTTPSTrafficOnly)
// }

