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

func awsDynamoDBTableSchemaV2() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the dynamodb table will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the table, this needs to be unique within a region.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
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
		"billing_mode": {
			Description: "Controls how you are charged for read and write throughput and how you manage capacity. The valid values are `PROVISIONED` and `PAY_PER_REQUEST`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "PROVISIONED",
			ValidateFunc: validation.StringInSlice([]string{
				"PROVISIONED ",
				"PAY_PER_REQUEST",
			}, false),
		},
		"read_capacity": {
			Description: "The number of read units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"write_capacity": {
			Description: "The number of write units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"tag": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},

		"attribute": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The name of the attribute",
						Type:        schema.TypeString,
						Required:    true,
					},
					"type": {
						Description: "Attribute type, which must be a scalar type: `S`, `N`, or `B` for (S)tring, (N)umber or (B)inary data",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},

		"key_schema": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"attribute_name": {
						Description: "The name of the attribute",
						Type:        schema.TypeString,
						Required:    true,
					},
					"key_type": {
						Description: "Attribute type, which must be a scalar type: `S`, `N`, or `B` for (S)tring, (N)umber or (B)inary data",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
	}
}

func resourceAwsDynamoDBTableV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_dynamodb_table_v2` manages an AWS dynamodb table in Duplo.",

		ReadContext:   resourceAwsDynamoDBTableReadV2,
		CreateContext: resourceAwsDynamoDBTableCreateV2,
		UpdateContext: resourceAwsDynamoDBTableUpdateV2,
		DeleteContext: resourceAwsDynamoDBTableDeleteV2,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsDynamoDBTableSchemaV2(),
	}
}

func resourceAwsDynamoDBTableReadV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsDynamoDBTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsDynamoDBTableReadV2(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DynamoDBTableGetV2(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s dynamodb table '%s': %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("arn", duplo.TableArn)
	d.Set("status", duplo.TableStatus.Value)

	log.Printf("[TRACE] resourceAwsDynamoDBTableReadV2(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsDynamoDBTableCreateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreateV2(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.DynamoDBTableCreateV2(tenantID, expandDynamoDBTable(d))
	if err != nil {
		return diag.Errorf("Error creating tenant %s dynamodb table '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGetV2(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsDynamoDBTableReadV2(ctx, d, m)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreateV2(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsDynamoDBTableUpdateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsDynamoDBTableDeleteV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsDynamoDBTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsDynamoDBTableDeleteV2(%s, %s): start", tenantID, name)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.DynamoDBTableDeleteV2(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s dynamodb table '%s': %s", tenantID, name, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGetV2(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsDynamoDBTableDeleteV2(%s, %s): end", tenantID, name)
	return nil
}

func expandDynamoDBTable(d *schema.ResourceData) *duplosdk.DuploDynamoDBTableRequestV2 {
	return &duplosdk.DuploDynamoDBTableRequestV2{
		TableName:            d.Get("name").(string),
		BillingMode:          d.Get("billing_mode").(string),
		Tags:                 keyValueFromState("tag", d),
		AttributeDefinitions: expandDynamoDBAttributes(d),
		KeySchema:            expandDynamoDBKeySchema(d),
	}
}

func expandDynamoDBAttributes(d *schema.ResourceData) *[]duplosdk.DuploDynamoDBAttributeDefinionV2 {
	var ary []duplosdk.DuploDynamoDBAttributeDefinionV2

	if v, ok := d.GetOk("attribute"); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		ary = make([]duplosdk.DuploDynamoDBAttributeDefinionV2, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploDynamoDBAttributeDefinionV2{
				AttributeName: kv["name"].(string),
				AttributeType: kv["type"].(string),
			})
		}
	}
	return &ary
}

func expandDynamoDBKeySchema(d *schema.ResourceData) *[]duplosdk.DuploDynamoDBKeySchemaV2 {
	var ary []duplosdk.DuploDynamoDBKeySchemaV2
	if v, ok := d.GetOk("key_schema"); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		ary = make([]duplosdk.DuploDynamoDBKeySchemaV2, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploDynamoDBKeySchemaV2{
				AttributeName: kv["attribute_name"].(string),
				KeyType:       kv["key_type"].(string),
			})
		}
	}
	return &ary
}
