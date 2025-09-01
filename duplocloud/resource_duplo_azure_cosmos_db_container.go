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
			Optional:    true,
			Default:     "databaseAccounts/sqlDatabases/containers",
			ForceNew:    true,
		},
		"namespace": {
			Description: "The Azure resource provider namespace for Cosmos DB. This value is typically not changed and identifies the resource type within Azure.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.DocumentDB",
			ForceNew:    true,
		},
		"partition_key_path": {
			Description: "The partition key for the Cosmos DB container. This is a required field for creating a Cosmos DB container.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

func resourceAzureCosmosDBContainer() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db_container` manages cosmos db container resource for azure",

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
	if len(idParts) != 6 {
		return diag.Errorf("Invalid resource ID format: %s. Expected format: tenantId/cosmosdb/accountName/databaseName/container/containerName", id)
	}
	c := m.(*duplosdk.Client)
	tenantId := idParts[0]
	account := idParts[2]
	dbName := idParts[3]
	container := idParts[5]
	rp, err := c.GetCosmosDBDatabaseContainer(tenantId, account, dbName, container)
	if err != nil && err.Status() != 404 {
		return diag.Errorf("Error fetching cosmos db container %s associated to database %s from account %s of tenantId %s : %s", container, dbName, account, tenantId, err.Error())
	}
	if err != nil && err.Status() == 404 {
		log.Printf("[DEBUG] resourceAzureCosmosDBRead: Cosmos DB container %s of database %s from account %s for tenantId %s not found, removing from state", container, dbName, account, tenantId)
		d.SetId("")
		return nil
	}
	if rp == nil {
		log.Printf("[DEBUG] resourceAzureCosmosDBRead: Cosmos DB container %s of database %s from account %s for tenantId %s not found, removing from state", container, dbName, account, tenantId)
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantId)
	d.Set("account_name", account)
	d.Set("database_name", dbName)
	flattenAzureCosmosDBContainer(d, *rp)
	return nil
}
func resourceAzureCosmosDBContainerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("resourceAzureCosmosDBContainerCreate started for %s", tenantId)
	account := d.Get("account_name").(string)
	database := d.Get("database_name").(string)
	rq := expandAzureCosmosDBContainer(d)
	c := m.(*duplosdk.Client)
	err := c.CreateCosmosDBDatabaseContainer(tenantId, account, database, rq)
	if err != nil {
		return diag.Errorf("Error creating cosmos db for tenantId %s : %s", tenantId, err.Error())
	}
	id := fmt.Sprintf("%s/cosmosdb/%s/%s/container/%s", tenantId, account, database, rq.Name)
	d.SetId(id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure cosmosdb container", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetCosmosDBDatabaseContainer(tenantId, account, database, rq.Name)
	})
	if diags != nil {
		return diags
	}
	diags = resourceAzureCosmosDBContainerRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	log.Printf("resourceAzureCosmosDBContainerCreate end for %s", tenantId)

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
	databaseName := idParts[3]
	containerName := idParts[5]
	log.Printf("resourceAzureCosmosDBDelete started for %s", tenantId)
	c := m.(*duplosdk.Client)
	err := c.DeleteCosmosDBDatabaseContainer(tenantId, accountName, databaseName, containerName)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("[DEBUG] resourceAzureCosmosDBDelete: Cosmos DB container %s of database %s from account %s for tenantId %s not found, removing from state", containerName, databaseName, accountName, tenantId)
			return nil
		}
		return diag.Errorf("Error deleting Cosmos DB container %s of database %s from account %s for tenantId %s: %s", containerName, databaseName, accountName, tenantId, err.Error())
	}
	log.Printf("resourceAzureCosmosDBDelete end for %s", tenantId)
	// Return nil to indicate successful deletion
	return nil
}
func expandAzureCosmosDBContainer(d *schema.ResourceData) duplosdk.DuploAzureCosmosDBContainer {
	obj := duplosdk.DuploAzureCosmosDBContainer{}
	paths := []string{}
	paths = append(paths, d.Get("partition_key_path").(string))
	obj.Resource = &duplosdk.DuploAzureCosmosDBContainerResource{
		ContainerName: d.Get("name").(string),
		PartitionKey: &duplosdk.DuploAzureCosmosDBContainerPartitionKey{
			Paths: paths,
		},
	}

	obj.ResourceType = &duplosdk.DuploAzureCosmosDBResourceType{
		Namespace: d.Get("namespace").(string),
		Type:      d.Get("type").(string),
	}
	obj.Name = d.Get("name").(string)
	return obj
}

func flattenAzureCosmosDBContainer(d *schema.ResourceData, rp duplosdk.DuploAzureCosmosDBContainer) {
	d.Set("name", rp.Resource.ContainerName)
	d.Set("partition_key_path", rp.Resource.PartitionKey.Paths[0])
	d.Set("namespace", rp.ResourceType.Namespace)
	d.Set("type", rp.ResourceType.Type)
}
