package duplocloud

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
//creating conflicts with name set up

	func BeforeHook(fn func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics) func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			c := m.(*duplosdk.Client)

			err := prefixName(c, d)
			if err != nil {
				return diag.FromErr(err)
			}

			return fn(ctx, d, m)
		}
	}

	func prefixName(c *duplosdk.Client, d *schema.ResourceData) duplosdk.ClientError {
		tenantId, name := d.Get("tenant_id").(string), d.Get("name").(string)

		prefix, err := c.GetDuploServicesPrefix(tenantId)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(name, prefix) {
			name = fmt.Sprintf("%s-%s", prefix, name)
			d.Set("name", name)
		}

		return nil
	}
*/
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
		"fullname": {
			Description: "The name of the table, this needs to be unique within a region.",
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
		"deletion_protection_enabled": {
			Description: "Deletion protection keeps the tables from being deleted unintentionally. While this setting is on, you can't delete the table.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"is_point_in_time_recovery": {
			Description: "The point in time recovery status of the dynamodb table. Enabled if true.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"billing_mode": {
			Description: "Controls how you are charged for read and write throughput and how you manage capacity. The valid values are `PROVISIONED` and `PAY_PER_REQUEST`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     duplosdk.DynamoDBBillingModeProvisioned,
			ValidateFunc: validation.StringInSlice([]string{
				duplosdk.DynamoDBBillingModeProvisioned,
				duplosdk.DynamoDBBillingModePerRequest,
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
			Type:     schema.TypeSet,
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
						Description: "Applicable key types are `HASH` or `RANGE`.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"global_secondary_index": {
			Description: "Describe a GSI for the table; subject to the normal limits on the number of GSIs, projected attributes, etc.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hash_key": {
						Description: "The name of the hash key in the index; must be defined as an attribute in the resource.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"name": {
						Description: "The name of the index.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"non_key_attributes": {
						Description: "Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.",
						Type:        schema.TypeSet,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"projection_type": {
						Description: "One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects just the hash and range key into the index, and `INCLUDE` projects only the keys specified in the `non_key_attributes` parameter.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"ALL",
							"INCLUDE",
							"KEYS_ONLY",
						}, false),
					},
					"range_key": {
						Description: "The name of the range key; must be defined.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"read_capacity": {
						Description: "The number of read units for this index. Must be set if `billing_mode` is set to `PROVISIONED`.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"write_capacity": {
						Description: "The number of write units for this index. Must be set if `billing_mode` is set to `PROVISIONED`.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
				},
			},
		},
		"local_secondary_index": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The name of the index.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"hash_key": {
						Description: "The name of the hash key in the index; must be defined as an attribute in the resource.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"non_key_attributes": {
						Description: "Only required with `INCLUDE` as a projection type; a list of attributes to project into the index. These do not need to be defined as attributes on the table.",
						Type:        schema.TypeList,
						Optional:    true,
						ForceNew:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"projection_type": {
						Description: "One of `ALL`, `INCLUDE` or `KEYS_ONLY` where `ALL` projects every attribute into the index, `KEYS_ONLY` projects just the hash and range key into the index, and `INCLUDE` projects only the keys specified in the `non_key_attributes` parameter.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"ALL",
							"INCLUDE",
							"KEYS_ONLY",
						}, false),
					},
					"range_key": {
						Description: "The name of the range key; must be defined.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
				},
			},
		},
		"server_side_encryption": {
			Description: "Encryption at rest options. AWS DynamoDB tables are automatically encrypted at rest with an AWS owned Customer Master Key if this argument isn't specified.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Description: "Whether or not to enable encryption at rest using an AWS managed KMS customer master key (CMK).",
						Type:        schema.TypeBool,
						Required:    true,
					},
					"kms_key_arn": {
						Description: "The ARN of the CMK that should be used for the AWS KMS encryption.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
		"stream_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"stream_enabled": {
			Description: "Indicates whether Streams are to be enabled (true) or disabled (false).",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"stream_label": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"stream_view_type": {
			Description: "When an item in the table is modified, StreamViewType determines what information is written to the table's stream. Valid values are `KEYS_ONLY`, `NEW_IMAGE`, `OLD_IMAGE`, `NEW_AND_OLD_IMAGES`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			StateFunc: func(v interface{}) string {
				value := v.(string)
				return strings.ToUpper(value)
			},
			ValidateFunc: validation.StringInSlice([]string{
				"",
				"KEYS_ONLY",
				"NEW_IMAGE",
				"OLD_IMAGE",
				"NEW_AND_OLD_IMAGES",
			}, false),
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until dynamodb instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceAwsDynamoDBTableV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_dynamodb_table_v2` manages an AWS dynamodb table in Duplo.",

		ReadContext:   resourceAwsDynamoDBTableReadV2,   //BeforeHook(resourceAwsDynamoDBTableReadV2),
		CreateContext: resourceAwsDynamoDBTableCreateV2, //BeforeHook(resourceAwsDynamoDBTableCreateV2),
		UpdateContext: resourceAwsDynamoDBTableUpdateV2, //BeforeHook(resourceAwsDynamoDBTableUpdateV2),
		DeleteContext: resourceAwsDynamoDBTableDeleteV2, //BeforeHook(resourceAwsDynamoDBTableDeleteV2),
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
	reservedFullname := d.Get("fullname").(string)
	c := m.(*duplosdk.Client)
	fullName, errname := c.GetDuploServicesNameWithAwsDynamoDbV2(tenantID, name)
	if errname != nil {
		return diag.Errorf("resourceAwsDynamoDBTableReadV2: Unable to retrieve duplo service name (name: %s, error: %s)", name, errname.Error())
	}
	if fullName != reservedFullname {
		fullName = reservedFullname
	}
	duplo, clientErr := c.DynamoDBTableGetV2(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s dynamodb table '%s': %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", fullName)
	d.Set("arn", duplo.TableArn)
	d.Set("status", duplo.TableStatus.Value)
	d.Set("is_point_in_time_recovery", duplo.PointInTimeRecoveryStatus == "ENABLED")
	d.Set("deletion_protection_enabled", duplo.DeletionProtectionEnabled)

	if duplo.BillingModeSummary != nil {
		d.Set("billing_mode", duplo.BillingModeSummary.BillingMode)
	} else {
		d.Set("billing_mode", duplosdk.DynamoDBBillingModeProvisioned)
	}

	if duplo.ProvisionedThroughput != nil {
		d.Set("write_capacity", duplo.ProvisionedThroughput.WriteCapacityUnits)
		d.Set("read_capacity", duplo.ProvisionedThroughput.ReadCapacityUnits)
	}
	if duplo.StreamSpecification != nil {
		d.Set("stream_view_type", duplo.StreamSpecification.StreamViewType)
		d.Set("stream_enabled", duplo.StreamSpecification.StreamEnabled)
	} else {
		d.Set("stream_view_type", "")
		d.Set("stream_enabled", false)
	}
	d.Set("stream_arn", duplo.LatestStreamArn)
	d.Set("stream_label", duplo.LatestStreamLabel)

	if err := d.Set("attribute", flattenTableAttributeDefinitions(duplo.AttributeDefinitions)); err != nil {
		return diag.FromErr(err)
	}

	for _, attribute := range *duplo.KeySchema {
		if attribute.KeyType == duplosdk.DynamoDBKeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
		}

		if attribute.KeyType == duplosdk.DynamoDBKeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
		}
	}

	if err := d.Set("local_secondary_index", flattenTableLocalSecondaryIndex(duplo.LocalSecondaryIndexes)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("global_secondary_index", flattenTableGlobalSecondaryIndex(duplo.GlobalSecondaryIndexes)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("server_side_encryption", flattenDynamoDBTableServerSideEncryption(duplo.SSEDescription)); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceAwsDynamoDBTableReadV2(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsDynamoDBTableCreateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreateV2(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq, err := expandDynamoDBTable(d)
	if err != nil {
		return diag.FromErr(err)
	}

	rp, err := c.DynamoDBTableCreateV2(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s dynamodb table '%s': %s", tenantID, name, err)
	}
	d.Set("fullname", rp.TableName)
	time.Sleep(time.Duration(10) * time.Second)

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGetV2(tenantID, rp.TableName)
	})

	if diags != nil {
		diags = waitForResourceToBePresentAfterCreate(ctx, d, "dynamodb table", id, func() (interface{}, duplosdk.ClientError) {
			return c.DynamoDBTableGetV2(tenantID, name)
		})
		if diags != nil {
			return diags
		}
	}
	d.SetId(id)

	//By default, wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = dynamodbWaitUntilReady(ctx, c, tenantID, rp.TableName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = updateDynamoDBTableV2PointInRecovery(ctx, d, m)
	if diags != nil {
		return diags
	}

	diags = resourceAwsDynamoDBTableReadV2(ctx, d, m)
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreateV2(%s, %s): end", tenantID, name)
	return diags
}

func updateDynamoDBTableV2PointInRecovery(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if v, ok := d.GetOk("is_point_in_time_recovery"); ok && v.(bool) {
		id := d.Id()
		tenantID, name, err := parseAwsDynamoDBTableIdParts(id)
		fname := d.Get("fullname").(string)
		c := m.(*duplosdk.Client)
		_, errPir := c.DynamoDBTableV2PointInRecovery(tenantID, fname, v.(bool))
		if errPir != nil {
			return diag.Errorf("Error while setting point in recovery tenant %s dynamodb table '%s': %s", tenantID, name, err)
		}
	}
	return nil
}

func tagDynamoDBtTableV2(
	tenantId string,
	rq *duplosdk.DuploDynamoDBTagResourceRequest,
	m interface{},
) (*duplosdk.DuploDynamoDBTagResourceResponse, duplosdk.ClientError) {
	c := m.(*duplosdk.Client)
	resp, err := c.DynamoDBTableUpdateTagsV2(tenantId, rq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func resourceAwsDynamoDBTableUpdateV2(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	fullname := d.Get("fullname").(string)
	_, name, err := parseAwsDynamoDBTableIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsDynamoDBTableCreateOrUpdateV2(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	// Expand the resource data into the SDK's DynamoDB table request struct.
	rq, err := expandDynamoDBTable(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fetch the existing table for compairson
	existing, err := c.DynamoDBTableGetV2(tenantID, fullname)
	if err != nil {
		return diag.FromErr(err)
	}

	tagReq := &duplosdk.DuploDynamoDBTagResourceRequest{
		ResourceArn: existing.TableArn,
		Tags:        rq.Tags,
	}
	_, err = tagDynamoDBtTableV2(tenantID, tagReq, m)
	if err != nil {
		return diag.FromErr(err)
	}
	diagErr := updatePointInTimeRecovery(c, d, tenantID, fullname)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateDeleteProtection(c, d, tenantID, fullname)
	if diagErr != nil {
		return diagErr
	}

	if shouldUpdateGSI(existing, rq) {
		rq.SSESpecification, rq.DeletionProtectionEnabled = nil, nil

		log.Printf("[INFO] Updating DynamoDB table '%s' in tenant '%s'", name, tenantID)
		_, err = c.DynamoDBTableUpdateGSIV2(tenantID, rq)
		if err != nil {
			e := "Error updating tenant %s DynamoDB table '%s': %s"
			return diag.Errorf(e, tenantID, name, err)
		}

	}
	if shouldUpdateThroughput(existing, rq) {
		rq.SSESpecification, rq.DeletionProtectionEnabled = nil, nil
		//
		log.Printf("[INFO] Updating DynamoDB table '%s' in tenant '%s'", name, tenantID)
		_, err = c.DynamoDBTableUpdateGSIV2(tenantID, rq)
		if err != nil {
			e := "Error updating tenant %s DynamoDB table '%s': %s"
			return diag.Errorf(e, tenantID, name, err)
		}
	}
	//switch {
	//case isPITREnabled != targetPITRStatus:
	//	log.Printf("[INFO] Updating Point In Recovery for DynamoDB table '%s' in tenant '%s'", name, tenantID)
	//	_, err = c.DynamoDBTableV2PointInRecovery(tenantID, fullname, targetPITRStatus)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//	fallthrough
	//case rq.DeletionProtectionEnabled != nil && existing.DeletionProtectionEnabled != *rq.DeletionProtectionEnabled:
	//	log.Printf("[INFO] Updating Deletion Protection for DynamoDB table '%s' in tenant '%s'", name, tenantID)
	//	_, err := c.DuploDynamoDBTableV2UpdateDeletionProtection(tenantID, rq)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//	fallthrough
	//	//	case !reflect.DeepEqual(existing.SSEDescription, rq.SSESpecification): //shouldUpdateSSESepecification(existing, rq):
	//	//		log.Printf("[INFO] Updating SSE Specification for DynamoDB table '%s' in tenant '%s'", name, tenantID)
	//	//		_, err := c.DuploDynamoDBTableV2UpdateSSESpecification(tenantID, rq)
	//	//		if err != nil {
	//	//			return diag.FromErr(err)
	//	//		}
	//	//		fallthrough
	//case shouldUpdateGSI(existing, rq) || shouldUpdateThroughput(existing, rq):
	//	   SSESpecification & DeletionProtectionEnabled must be updated alone.
	//	   Passing these values with the rest of the update table request willcause
	//	   cause a error. (Per .NET AWS SDK@3.7)
	//	   rq.SSESpecification, rq.DeletionProtectionEnabled = nil, nil
	//	//   	log.Printf("[INFO] Updating DynamoDB table '%s' in tenant '%s'", name, tenantID)
	//   	_, err = c.DynamoDBTableUpdateV2(tenantID, rq)
	//   	if err != nil {
	//   		e := "Error updating tenant %s DynamoDB table '%s': %s"
	//   		return diag.Errorf(e, tenantID, name, err)
	//   	}
	//}
	isSSESUpdatable := shouldUpdateSSESepecification(existing, rq)
	if isSSESUpdatable {
		log.Printf("[INFO] Updating SSE Specification for DynamoDB table '%s' in tenant '%s'", name, tenantID)
		_, err := c.DuploDynamoDBTableV2UpdateSSESpecification(tenantID, rq)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	//// Generate the ID for the resource and set it.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	getResource := func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGet(tenantID, fullname)
	}
	diags := waitForResourceToBePresentAfterUpdate(ctx, d, "dynamodb table", id, getResource)

	// If there are diagnostics from waiting, return them.
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = dynamodbWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Perform a read after update to sync state.
	diags = resourceAwsDynamoDBTableReadV2(ctx, d, m)
	log.Printf("[TRACE] resourceAwsDynamoDBTableUpdateV2(%s, %s): end", tenantID, name)

	return diags
}

func updatePointInTimeRecovery(c *duplosdk.Client, d *schema.ResourceData, tenantID, fullname string) diag.Diagnostics {
	if d.HasChange("is_point_in_time_recovery") {
		targetPITRStatus := d.Get("is_point_in_time_recovery").(bool)
		_, err := c.DynamoDBTableV2PointInRecovery(tenantID, fullname, targetPITRStatus)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
func updateDeleteProtection(c *duplosdk.Client, d *schema.ResourceData, tenantID, fullname string) diag.Diagnostics {
	r := duplosdk.DuploDynamoDBTableRequestV2{}
	if d.HasChange("deletion_protection_enabled") {
		state := d.Get("deletion_protection_enabled").(bool)
		r.DeletionProtectionEnabled = &state

		r.TableName = fullname

		_, err := c.DuploDynamoDBTableV2UpdateDeletionProtection(tenantID, &r)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func updateGlobalSecondaryIndex(c *duplosdk.Client, d *schema.ResourceData, tenantID, fullname string, existing *duplosdk.DuploDynamoDBTableV2) diag.Diagnostics {

	if d.HasChange("global_secondary_index") {
		r := duplosdk.DuploDynamoDBTableRequestV2{}
		billingMode := d.Get("billing_mode").(string)
		r.TableName = fullname
		globalSecondaryIndexes := []duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex{}
		gsiSet := d.Get("global_secondary_index").(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]interface{})
			if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
				return diag.Errorf("failed to create GSI: %w", err)
			}

			gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
			globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
		}
		r.GlobalSecondaryIndexes = &[]duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex{}
		r.GlobalSecondaryIndexes = &globalSecondaryIndexes

		if shouldUpdateGSI(existing, &r) {
			_, diagErr := c.DynamoDBTableUpdateV2(tenantID, &r)
			if diagErr != nil {
				e := "Error updating tenant %s DynamoDB table '%s': %s"
				return diag.Errorf(e, tenantID, fullname, diagErr)
			}
		}
	}

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
		return c.DynamoDBTableGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsDynamoDBTableDeleteV2(%s, %s): end", tenantID, name)
	return nil
}

func expandDynamoDBTable(d *schema.ResourceData) (*duplosdk.DuploDynamoDBTableRequestV2, error) {

	req := &duplosdk.DuploDynamoDBTableRequestV2{
		TableName:   d.Get("name").(string),
		BillingMode: d.Get("billing_mode").(string),
		Tags:        keyValueFromState("tag", d),
		KeySchema:   expandDynamoDBKeySchema(d),
	}

	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandAttributes(aSet.List())
	}

	if v, ok := d.GetOk("stream_enabled"); ok {
		req.StreamSpecification = &duplosdk.DuploDynamoDBTableV2StreamSpecification{
			StreamEnabled:  v.(bool),
			StreamViewType: &duplosdk.DuploStringValue{Value: d.Get("stream_view_type").(string)},
		}
	}

	billingMode := d.Get("billing_mode").(string)

	capacityMap := map[string]interface{}{
		"write_capacity": d.Get("write_capacity"),
		"read_capacity":  d.Get("read_capacity"),
	}

	if err := validateCapacityForBillingMode(capacityMap, billingMode); err != nil {
		return nil, fmt.Errorf("failed to create : %w", err)
	}

	req.ProvisionedThroughput = expandProvisionedThroughput(capacityMap, billingMode)

	if v, ok := d.GetOk("local_secondary_index"); ok {
		lsiSet := v.(*schema.Set)
		req.LocalSecondaryIndexes = expandLocalSecondaryIndexes(lsiSet.List())
	}

	if v, ok := d.GetOk("global_secondary_index"); ok {
		globalSecondaryIndexes := []duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex{}
		gsiSet := v.(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]interface{})
			if err := validateGSIProvisionedThroughput(gsi, billingMode); err != nil {
				return nil, fmt.Errorf("failed to create GSI: %w", err)
			}

			gsiObject := expandGlobalSecondaryIndex(gsi, billingMode)
			globalSecondaryIndexes = append(globalSecondaryIndexes, *gsiObject)
		}
		req.GlobalSecondaryIndexes = &globalSecondaryIndexes
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		req.SSESpecification = expandEncryptAtRestOptions(v.([]interface{}))
	}

	return req, nil
}

func expandAttributes(cfg []interface{}) *[]duplosdk.DuploDynamoDBAttributeDefinionV2 {
	attributes := make([]duplosdk.DuploDynamoDBAttributeDefinionV2, len(cfg))
	for i, attribute := range cfg {
		attr := attribute.(map[string]interface{})
		attributes[i] = duplosdk.DuploDynamoDBAttributeDefinionV2{
			AttributeName: attr["name"].(string),
			AttributeType: attr["type"].(string),
		}
	}
	return &attributes
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

func flattenTableGlobalSecondaryIndex(gsi *[]duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex) []interface{} {
	if len(*gsi) == 0 {
		return []interface{}{}
	}

	var output []interface{}

	for _, g := range *gsi {
		gsi := make(map[string]interface{})

		if g.ProvisionedThroughput != nil {
			gsi["write_capacity"] = g.ProvisionedThroughput.WriteCapacityUnits
			gsi["read_capacity"] = g.ProvisionedThroughput.ReadCapacityUnits
			gsi["name"] = g.IndexName
		}

		for _, attribute := range *g.KeySchema {

			if attribute.KeyType == "HASH" {
				gsi["hash_key"] = attribute.AttributeName
			}

			if attribute.KeyType == "RANGE" {
				gsi["range_key"] = attribute.AttributeName
			}
		}

		if g.Projection != nil {
			gsi["projection_type"] = g.Projection.ProjectionType
			gsi["non_key_attributes"] = g.Projection.NonKeyAttributes
		}

		output = append(output, gsi)
	}

	return output
}

func expandGlobalSecondaryIndex(data map[string]interface{}, billingMode string) *duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex {
	return &duplosdk.DuploDynamoDBTableV2GlobalSecondaryIndex{
		IndexName:             data["name"].(string),
		KeySchema:             expandKeySchema(data),
		Projection:            expandProjection(data),
		ProvisionedThroughput: expandProvisionedThroughput(data, billingMode),
	}
}

func expandKeySchema(data map[string]interface{}) *[]duplosdk.DuploDynamoDBKeySchema {
	keySchema := []duplosdk.DuploDynamoDBKeySchema{}

	if v, ok := data["hash_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, duplosdk.DuploDynamoDBKeySchema{
			AttributeName: v.(string),
			KeyType:       "HASH",
		})
	}

	if v, ok := data["range_key"]; ok && v != nil && v != "" {
		keySchema = append(keySchema, duplosdk.DuploDynamoDBKeySchema{
			AttributeName: v.(string),
			KeyType:       "RANGE",
		})
	}
	return &keySchema
}

func expandProjection(data map[string]interface{}) *duplosdk.DuploDynamoDBTableV2Projection {
	projection := &duplosdk.DuploDynamoDBTableV2Projection{
		ProjectionType: data["projection_type"].(string),
	}

	if v, ok := data["non_key_attributes"].([]interface{}); ok && len(v) > 0 {
		projection.NonKeyAttributes = expandStringList(v)
	}

	if v, ok := data["non_key_attributes"].(*schema.Set); ok && v.Len() > 0 {
		projection.NonKeyAttributes = expandStringSet(v)
	}

	return projection
}

func expandProvisionedThroughput(data map[string]interface{}, billingMode string) *duplosdk.DuploDynamoDBProvisionedThroughput {
	return expandProvisionedThroughputUpdate("", data, billingMode, "")
}

func expandProvisionedThroughputUpdate(id string, data map[string]interface{}, billingMode, oldBillingMode string) *duplosdk.DuploDynamoDBProvisionedThroughput {
	if billingMode == "PAY_PER_REQUEST" {
		return nil
	}

	return &duplosdk.DuploDynamoDBProvisionedThroughput{
		ReadCapacityUnits:  expandProvisionedThroughputField(id, data, "read_capacity", billingMode, oldBillingMode),
		WriteCapacityUnits: expandProvisionedThroughputField(id, data, "write_capacity", billingMode, oldBillingMode),
	}
}

func expandProvisionedThroughputField(id string, data map[string]interface{}, key, billingMode, oldBillingMode string) int {
	v := data[key].(int)
	if v == 0 && billingMode == "PROVISIONED" && oldBillingMode == "PAY_PER_REQUEST" {
		log.Printf("[WARN] Overriding %[1]s on DynamoDB Table (%[2]s) to %[3]d. Switching from billing mode %[4]q to %[5]q without value for %[1]s. Assuming changes are being ignored.",
			key, id, duplosdk.DynamoDBProvisionedThroughputMinValue, oldBillingMode, billingMode)
		v = duplosdk.DynamoDBProvisionedThroughputMinValue
	}
	return v
}

func flattenTableLocalSecondaryIndex(lsi *[]duplosdk.DuploDynamoDBTableV2LocalSecondaryIndex) []interface{} {
	if len(*lsi) == 0 {
		return []interface{}{}
	}

	var output []interface{}

	for _, l := range *lsi {

		m := map[string]interface{}{
			"name": l.IndexName,
		}

		if l.Projection != nil {
			m["projection_type"] = l.Projection.ProjectionType
			m["non_key_attributes"] = l.Projection.NonKeyAttributes
		}

		for _, attribute := range *l.KeySchema {
			if attribute.KeyType == "RANGE" {
				m["range_key"] = attribute.AttributeName
			}
		}

		output = append(output, m)
	}

	return output
}

func expandLocalSecondaryIndexes(cfg []interface{}) *[]duplosdk.DuploDynamoDBTableV2LocalSecondaryIndex {
	indexes := make([]duplosdk.DuploDynamoDBTableV2LocalSecondaryIndex, len(cfg))
	for i, lsi := range cfg {
		m := lsi.(map[string]interface{})
		idxName := m["name"].(string)
		indexes[i] = duplosdk.DuploDynamoDBTableV2LocalSecondaryIndex{
			IndexName:  idxName,
			KeySchema:  expandKeySchema(m),
			Projection: expandProjection(m),
		}
	}
	return &indexes
}

func flattenDynamoDBTableServerSideEncryption(spec *duplosdk.DuploDynamoDBTableV2SSESpecification) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":     spec.Enabled,
		"kms_key_arn": spec.KMSMasterKeyId,
	}

	return []interface{}{m}
}

func expandEncryptAtRestOptions(vOptions []interface{}) *duplosdk.DuploDynamoDBTableV2SSESpecification {
	options := &duplosdk.DuploDynamoDBTableV2SSESpecification{}

	enabled := false
	if len(vOptions) > 0 {
		mOptions := vOptions[0].(map[string]interface{})

		enabled = mOptions["enabled"].(bool)
		if enabled {
			if vKmsKeyArn, ok := mOptions["kms_key_arn"].(string); ok && vKmsKeyArn != "" {
				options.KMSMasterKeyId = vKmsKeyArn
				options.SSEType = &duplosdk.DuploStringValue{Value: "KMS"}
			}
		}
	}
	options.Enabled = enabled

	return options
}

func flattenTableAttributeDefinitions(definitions *[]duplosdk.DuploDynamoDBAttributeDefinion) []interface{} {
	if len(*definitions) == 0 {
		return []interface{}{}
	}

	var attributes []interface{}

	for _, d := range *definitions {

		m := map[string]string{
			"name": d.AttributeName,
			"type": d.AttributeType.Value,
		}

		attributes = append(attributes, m)
	}

	return attributes
}

func validateCapacityForBillingMode(data map[string]interface{}, billingMode string) error {
	if billingMode == duplosdk.DynamoDBBillingModePerRequest {
		writeCapacity, _ := data["write_capacity"].(int)
		readCapacity, _ := data["read_capacity"].(int)
		if writeCapacity > 0 || readCapacity > 0 {
			return fmt.Errorf("write_capacity and read_capacity cannot be set when billing mode is %s", duplosdk.DynamoDBBillingModePerRequest)
		}
	}
	return nil
}

func validateGSIProvisionedThroughput(data map[string]interface{}, billingMode string) error {
	if billingMode == duplosdk.DynamoDBBillingModePerRequest {
		if err := validateCapacityForBillingMode(data, billingMode); err != nil {
			return fmt.Errorf("failed to create GSI: %w", err)
		}
		return nil
	}

	writeCapacity, writeCapacitySet := data["write_capacity"].(int)
	readCapacity, readCapacitySet := data["read_capacity"].(int)

	if !writeCapacitySet || !readCapacitySet {
		return fmt.Errorf("read and write capacity should be set when billing mode is %s", duplosdk.DynamoDBBillingModeProvisioned)
	}

	if writeCapacity < 1 {
		return fmt.Errorf("write capacity must be > 0 when billing mode is %s", duplosdk.DynamoDBBillingModeProvisioned)
	}

	if readCapacity < 1 {
		return fmt.Errorf("read capacity must be > 0 when billing mode is %s", duplosdk.DynamoDBBillingModeProvisioned)
	}

	return nil
}

func dynamodbWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DynamoDBTableGetV2(tenantID, name)
			//			log.Printf("[TRACE] Dynamodb status is (%s).", rp.TableStatus.Value)
			status := "pending"
			if err == nil {
				if rp.TableStatus.Value == "ACTIVE" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] dynamodbWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// shouldUpdateGSI compares the DuploDynamoDBTableV2LocalSecondaryIndex  of the
// existing table with the updated table. Returns true if a change is detected.
func shouldUpdateGSI(
	table *duplosdk.DuploDynamoDBTableV2,
	request *duplosdk.DuploDynamoDBTableRequestV2,
) bool {
	if table.GlobalSecondaryIndexes == nil || request.GlobalSecondaryIndexes == nil {
		return true
	}
	if len(*table.GlobalSecondaryIndexes) != len(*request.GlobalSecondaryIndexes) {
		return true
	}

	for i, aIndex := range *table.GlobalSecondaryIndexes {
		bIndex := (*request.GlobalSecondaryIndexes)[i]

		isDeepEqual := reflect.DeepEqual(
			aIndex.ProvisionedThroughput,
			bIndex.ProvisionedThroughput,
		)
		if aIndex.IndexName != bIndex.IndexName || !isDeepEqual {
			return true
		}
	}

	return false
}

// shouldUpdateThroughput compares the DuploDynamoDBProvisionedThroughput of
// the existing table and updated table. Returns true if a change is detected.
func shouldUpdateThroughput(
	table *duplosdk.DuploDynamoDBTableV2,
	request *duplosdk.DuploDynamoDBTableRequestV2,
) bool {
	return !reflect.DeepEqual(table.ProvisionedThroughput, request.ProvisionedThroughput)
}

// shouldUpdateSSESepecification compares the DuploDynamoDBTableV2SSESpecification of
// the existing table and updated table. Returns true if a change is detected.
func shouldUpdateSSESepecification(
	table *duplosdk.DuploDynamoDBTableV2,
	request *duplosdk.DuploDynamoDBTableRequestV2,
) bool {
	if table.SSEDescription == nil && request.SSESpecification == nil {
		return false
	}
	return !reflect.DeepEqual(table.SSEDescription, request.SSESpecification)
}
