package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsDynamoDBTableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the dynamodb table will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the dynamodb table.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the dynamodb table.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the dynamodb table.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The status of the dynamodb table.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"primary_key_name": {
			Description: "The attribute name of the primary key attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"attribute_type": {
			Description: "The attribute type of the primary key attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"key_type": {
			Description: "The key type of the primary key.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

// Resource for managing an AWS dynamodb table
func resourceAwsDynamoDBTable() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_dynamodb_table` manages an AWS dynamodb table in Duplo.",

		ReadContext:   resourceAwsDynamoDBTableRead,
		CreateContext: resourceAwsDynamoDBTableCreate,
		DeleteContext: resourceAwsDynamoDBTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(25 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(25 * time.Minute),
		},
		Schema: awsDynamoDBTableSchema(),
	}
}

// READ resource
func resourceAwsDynamoDBTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsDynamoDBTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsDynamoDBTableRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	fullname, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("Error creating fullname %s dynamodb table '%s': %s", tenantID, name, err)
	}
	duplo, clientErr := c.DynamoDBTableGet(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s dynamodb table '%s': %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)
	d.Set("status", duplo.Status)
	expandDynamoDBTablePrimaryKey(duplo, d)

	log.Printf("[TRACE] resourceAwsDynamoDBTableRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceAwsDynamoDBTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreate(%s, %s): start", tenantID, name)

	// Create the request object.
	rq := duplosdk.DuploDynamoDBTableRequest{
		Name:           name,
		PrimaryKeyName: d.Get("primary_key_name").(string),
		AttributeType:  d.Get("attribute_type").(string),
		KeyType:        d.Get("key_type").(string),
	}

	c := m.(*duplosdk.Client)
	fullname, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("Error creating fullname %s dynamodb table '%s': %s", tenantID, name, err)
	}
	// Post the object to Duplo
	_, err = c.DynamoDBTableCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s dynamodb table '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGet(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsDynamoDBTableRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceAwsDynamoDBTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsDynamoDBTableIdParts(id)
	fullname := d.Get("fullname").(string)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsDynamoDBTableDelete(%s, %s): start", tenantID, name)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.DynamoDBTableDelete(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s dynamodb table '%s': %s", tenantID, name, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGet(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsDynamoDBTableDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandDynamoDBTablePrimaryKey(duplo *duplosdk.DuploDynamoDBTable, d *schema.ResourceData) {
	if duplo.KeySchema == nil || len(*duplo.KeySchema) == 0 {
		return
	}

	keySchema := (*duplo.KeySchema)[0]
	primaryKeyName := keySchema.AttributeName

	d.Set("primary_key_name", primaryKeyName)
	if keySchema.KeyType != "" {
		d.Set("key_type", keySchema.KeyType)
	}

	if duplo.AttributeDefinitions != nil {
		for _, attr := range *duplo.AttributeDefinitions {
			if attr.AttributeName == primaryKeyName && attr.AttributeType != nil {
				d.Set("attribute_type", attr.AttributeType.Value)
			}
		}
	}
}

func parseAwsDynamoDBTableIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
