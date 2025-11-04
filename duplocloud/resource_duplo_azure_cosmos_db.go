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

func duploAzureCosmosDBschema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant in which the Cosmos DB database will be created for specified account",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name for Cosmsos DB database.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"account_name": {
			Description: "Indicates on which account database will be created",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"type": {
			Description: "Specifies the Cosmos DB account type.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "databaseAccounts/sqlDatabases",
			ForceNew:    true,
		},
		"namespace": {
			Description: "The Azure resource provider namespace for Cosmos DB. Defaults to 'Microsoft.DocumentDB'. This value is typically not changed and identifies the resource type within Azure.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.DocumentDB",
			ForceNew:    true,
		},
	}
}

func resourceAzureCosmosDB() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db_database` manages cosmos db database resource for azure",

		ReadContext:   resourceAzureCosmosDBRead,
		CreateContext: resourceAzureCosmosDBCreate,
		DeleteContext: resourceAzureCosmosDBDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureCosmosDBschema(),
	}
}

func resourceAzureCosmosDBRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDB(idParts[0], idParts[2], idParts[4])
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db database %s from account %s details for tenantId %s :%s", idParts[4], idParts[2], idParts[0], err.Error())
	}
	if err != nil && err.Status() == 404 {
		log.Printf("[DEBUG] resourceAzureCosmosDBRead: Cosmos DB database %s from account %s for tenantId %s not found, removing from state", idParts[4], idParts[2], idParts[0])
		d.SetId("")
		return nil
	}
	if rp == nil {
		log.Printf("[DEBUG] resourceAzureCosmosDBRead: Cosmos DB database %s from account %s for tenantId %s not found, removing from state", idParts[4], idParts[2], idParts[0])
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", idParts[0])
	d.Set("account_name", idParts[2])
	flattenAzureCosmosDB(d, *rp)
	return nil
}
func resourceAzureCosmosDBCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("resourceAzureCosmosDBCreate started for %s", tenantId)
	account := d.Get("account_name").(string)
	rq := expandAzureCosmosDB(d)
	c := m.(*duplosdk.Client)
	err := c.CreateCosmosDB(tenantId, account, rq)
	if err != nil {
		return diag.Errorf("Error creating cosmos db for tenantId %s : %s", tenantId, err.Error())
	}
	id := fmt.Sprintf("%s/cosmosdb/%s/database/%s", tenantId, account, rq.Name)
	d.SetId(id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure cosmosdb database", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetCosmosDB(tenantId, account, rq.Name)
	})
	if diags != nil {
		return diags
	}

	diag := resourceAzureCosmosDBRead(ctx, d, m)
	if diag != nil {
		return diag
	}
	log.Printf("resourceAzureCosmosDBCreate end for %s", tenantId)

	return nil
}

func resourceAzureCosmosDBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 5 {
		return diag.Errorf("Invalid resource ID format: %s. Expected format: tenantId/cosmosdb/accountName/database/databaseName", id)
	}
	tenantId := idParts[0]
	accountName := idParts[2]
	databaseName := idParts[4]
	log.Printf("resourceAzureCosmosDBDelete started for %s", tenantId)
	c := m.(*duplosdk.Client)
	err := c.DeleteCosmosDB(tenantId, accountName, databaseName)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureCosmosDBDelete: Cosmos DB database %s from account %s for tenantId %s not found, removing from state", databaseName, accountName, tenantId)
			return nil
		}
		return diag.Errorf("Error deleting Cosmos DB database %s from account %s for tenantId %s: %s", databaseName, accountName, tenantId, err.Error())
	}
	log.Printf("resourceAzureCosmosDBDelete end for %s", tenantId)
	// Return nil to indicate successful deletion
	return nil
}
func expandAzureCosmosDB(d *schema.ResourceData) duplosdk.DuploAzureCosmosDB {
	obj := duplosdk.DuploAzureCosmosDB{}

	obj.Resource = duplosdk.DuploAzureCosmosDBResource{
		DatabaseName: d.Get("name").(string),
	}

	obj.ResourceType = duplosdk.DuploAzureCosmosDBResourceType{
		Namespace: d.Get("namespace").(string),
		Type:      d.Get("type").(string),
	}
	obj.Name = d.Get("name").(string)
	return obj
}

func flattenAzureCosmosDB(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDB) {
	d.Set("name", rp.Resource.DatabaseName)
	d.Set("namespace", rp.ResourceType.Namespace)
	d.Set("type", rp.ResourceType.Type)
}
