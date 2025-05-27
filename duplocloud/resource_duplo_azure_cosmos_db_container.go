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

func duploAzureCosmosDBContainerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant in which the Cosmos DB database will be created for specified account",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name for Cosmsos DB database container name.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},

		"database_name": {
			Description: "The name for Cosmsos DB database in which container will be created.",
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
			Required:    true,
			Default:     "databaseAccounts/sqlDatabases/containers",
		},
		"namespace": {
			Description: "The Azure resource provider namespace for Cosmos DB. This value is typically not changed and identifies the resource type within Azure.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.DocumentDB",
		},
		"partition_key_path": {
			Description: "The partition key for the Cosmos DB container. This is a required field for creating a Cosmos DB container.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"partition_key_version": {
			Description: "The version of the partition key for the Cosmos DB container. This is a required field for creating a Cosmos DB container.",
			Type:        schema.TypeString,
			Default:     2,
			Optional:    true,
		},
	}
}

func resourceAzureCosmosDBContainer() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db` manages cosmos db resource for azure",

		ReadContext:   resourceAzureCosmosDBContainerRead,
		CreateContext: resourceAzureCosmosDBContainerCreate,
		DeleteContext: resourceAzureCosmosDBContainerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureCosmosDBContainerSchema(),
	}
}

func resourceAzureCosmosDBContainerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	c := m.(*duplosdk.Client)
	rp, err := c.GetCosmosDB(idParts[0], idParts[2], idParts[4])
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db database %s from account %s details for tenantId %s", idParts[0], idParts[2], idParts[0])
	}
	if rp == nil {
		log.Printf("[DEBUG] resourceAzureCosmosDBRead: Cosmos DB database %s from account %s for tenantId %s not found, removing from state", idParts[4], idParts[2], idParts[0])
		d.SetId("")
		return nil
	}
	flattenAzureCosmosDB(d, *rp)
	return nil
}
func resourceAzureCosmosDBContainerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	diag := resourceAzureCosmosDBRead(ctx, d, m)
	if diag != nil {
		return diag
	}
	log.Printf("resourceAzureCosmosDBCreate end for %s", tenantId)

	return nil
}

func resourceAzureCosmosDBContainerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("Error deleting Cosmos DB database %s from account %s for tenantId %s: %s", databaseName, accountName, tenantId, err.Error())
	}
	log.Printf("resourceAzureCosmosDBDelete end for %s", tenantId)
	// Return nil to indicate successful deletion
	return nil
}
func expandAzureCosmosDBContainer(d *schema.ResourceData) duplosdk.DuploAzureCosmosDB {
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

func flattenAzureCosmosDBContainer(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDB) {
	d.Set("name", rp.Resource.DatabaseName)
	d.Set("namespace", rp.ResourceType.Namespace)
	d.Set("type", rp.ResourceType.Type)
}
