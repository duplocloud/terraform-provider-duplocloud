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

// duploAzureCosmosDBschema returns the Terraform schema definition for an Azure Cosmos DB resource.
// The schema includes the following fields:
//   - tenant_id: The GUID of the tenant in which the Cosmos DB account will be created.
//   - name: The name of the Cosmos DB resource. Must be 1-80 characters, starting with a letter or number, and ending with a letter, number, or underscore.
//   - account_name: The type of database account. Allowed values: GlobalDocumentDB, MongoDB, Parse. This can only be set at account creation.
//   - type: Specifies the Cosmos DB account type. Defaults to "databaseAccounts/sqlDatabases".
//   - namespace:
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
			Description: "Specifies the Cosmos DB account type. Defaults to 'databaseAccounts/sqlDatabases'",
			Type:        schema.TypeString,
			Required:    true,
			Default:     "databaseAccounts/sqlDatabases",
		},
		"namespace": {
			Description: "The Azure resource provider namespace for Cosmos DB. Defaults to 'Microsoft.DocumentDB'. This value is typically not changed and identifies the resource type within Azure.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "Microsoft.DocumentDB"},
	}
}

func resourceAzureCosmosDB() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_cosmos_db` manages cosmos db resource for azure",

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
	rp, err := c.GetCosmosDB(idParts[0], idParts[3])
	if err != nil {
		return diag.Errorf("Error fetching cosmos db account %s details for tenantId %s", idParts[2], idParts[0])
	}
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
	diag := resourceAzureCosmosDBRead(ctx, d, m)
	if diag != nil {
		return diag
	}
	log.Printf("resourceAzureCosmosDBCreate end for %s", tenantId)

	return nil
}

func resourceAzureCosmosDBDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
