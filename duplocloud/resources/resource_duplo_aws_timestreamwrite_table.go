package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsTimestreamTableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Timestream Table will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the Timestream Table.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(3, 64),
				// validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
			),
		},
		"database_name": {
			Description: "The full name of the Timestream database.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(3, 64),
				// validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
			),
		},
		"magnetic_store_write_properties": {
			Description: "Contains properties to set on the table when enabling magnetic store writes.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enable_magnetic_store_writes": {
						Description: "A flag to enable magnetic store writes.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"magnetic_store_rejected_data_location": {
						Description: "The location to write error reports for records rejected asynchronously during magnetic store writes.",
						Type:        schema.TypeList,
						Optional:    true,
						Computed:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"s3_configuration": {
									Description: "Configuration of an S3 location to write error reports for records rejected, asynchronously, during magnetic store writes.",
									Type:        schema.TypeList,
									Optional:    true,
									Computed:    true,
									MaxItems:    1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"bucket_name": {
												Description: "Bucket name of the customer S3 bucket.",
												Type:        schema.TypeString,
												Optional:    true,
												Computed:    true,
											},
											"encryption_option": {
												Description:  " Encryption option for the customer s3 location. Options are S3 server side encryption with an S3-managed key or KMS managed key. Valid values are `SSE_KMS` and `SSE_S3`.",
												Type:         schema.TypeString,
												Optional:     true,
												Computed:     true,
												ValidateFunc: validation.StringInSlice([]string{"SSE_KMS", "SSE_S3"}, false),
											},
											"kms_key_id": {
												Description: "KMS key arn for the customer s3 location when encrypting with a KMS managed key.",
												Type:        schema.TypeString,
												Optional:    true,
												Computed:    true,
											},
											"object_key_prefix": {
												Description: "Object key prefix for the customer S3 location.",
												Type:        schema.TypeString,
												Optional:    true,
												Computed:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},

		"retention_properties": {
			Description: "The retention duration for the memory store and magnetic store.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"magnetic_store_retention_period_in_days": {
						Description:  "The duration for which data must be stored in the magnetic store. Minimum value of 1. Maximum value of 73000.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 73000),
					},

					"memory_store_retention_period_in_hours": {
						Description:  "The duration for which data must be stored in the memory store. Minimum value of 1. Maximum value of 8766.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 8766),
					},
				},
			},
		},
		"arn": {
			Description: "The ARN that uniquely identifies this Table.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Description: "Tags in key-value format.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        duplocloud.KeyValueSchema(),
		},
		"all_tags": {
			Description: "A complete list of tags for this time stream database, even ones not being managed by this resource.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        duplocloud.KeyValueSchema(),
		},
		"specified_tags": {
			Description: "A list of tags being managed by this resource.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"wait_for_deployment": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
	}
}

func resourceAwsTimestreamTable() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_timestreamwrite_table` manages an aws Timestream Table resource in Duplo.",

		ReadContext:   resourceAwsTimestreamTableRead,
		CreateContext: resourceAwsTimestreamTableCreate,
		UpdateContext: resourceAwsTimestreamTableUpdate,
		DeleteContext: resourceAwsTimestreamTableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsTimestreamTableSchema(),
	}
}

func resourceAwsTimestreamTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, dbName, name, err := parseAwsTimestreamTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTimestreamTableRead(%s, %s, %s): start", tenantID, dbName, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	table, clientErr := c.DuploTimestreamDBTableGet(tenantID, dbName, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if table == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenTimestreamTable(d, c, table, tenantID)
	log.Printf("[TRACE] resourceAwsTimestreamTableRead(%s, %s, %s): end", tenantID, dbName, name)
	return nil
}

func resourceAwsTimestreamTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsTimestreamTableCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	rq := expandAwsTimestreamTable(d)
	_, clientErr := c.DuploTimestreamDBTableCreate(tenantID, rq)
	if clientErr != nil {
		return diag.Errorf("Error creating tenant %s aws timestream Table '%s': %s", tenantID, name, clientErr)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", tenantID, rq.DatabaseName, name))

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = timestreamTableWaitUntilActive(ctx, c, tenantID, name, rq.DatabaseName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := resourceAwsTimestreamTableRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTimestreamTableCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsTimestreamTableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, dbName, name, err := parseAwsTimestreamTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTimestreamTableUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	rq := &duplosdk.DuploTimestreamDBTableUpdateRequest{
		TableName:    d.Get("name").(string),
		DatabaseName: d.Get("database_name").(string),
	}

	if d.HasChange("tags") {
		duplo, _ := c.DuploTimestreamDBTableGet(tenantID, dbName, name)
		if err != nil {
			return diag.Errorf("failed to retrieve tags  for '%s': %s", "tags", err)
		}
		newTags, deletedKeys := duplocloud.getTfManagedChangesDuploKeyStringValue("tags", duplo.Tags, d)
		rq.UpdatedTags = newTags
		rq.DeletedTags = deletedKeys
	}

	if v, ok := d.GetOk("retention_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		rq.RetentionProperties = expandRetentionProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("magnetic_store_write_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		rq.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(v.([]interface{}))
	}

	_, clientErr := c.DuploTimestreamDBTableUpdate(tenantID, dbName, name, rq)
	if clientErr != nil {
		return diag.Errorf("Error updating tenant %s aws timestream Table '%s': %s", tenantID, name, clientErr)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", tenantID, rq.DatabaseName, name))

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = timestreamTableWaitUntilActive(ctx, c, tenantID, name, rq.DatabaseName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := resourceAwsTimestreamTableRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTimestreamTableUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsTimestreamTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, dbName, name, err := parseAwsTimestreamTableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTimestreamTableDelete(%s, %s, %s): start", tenantID, dbName, name)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploTimestreamDBTableDelete(tenantID, dbName, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws timestream Table '%s': %s", tenantID, name, clientErr)
	}

	diag := duplocloud.waitForResourceToBeMissingAfterDelete(ctx, d, "aws timestream Table", id, func() (interface{}, duplosdk.ClientError) {
		rp, err := c.DuploTimestreamDBTableGet(tenantID, dbName, name)
		if rp == nil || err.Status() == 404 || err.Status() == 400 || err.Status() == 400 {
			return nil, nil
		}
		return rp, err
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsTimestreamTableDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsTimestreamTable(d *schema.ResourceData) *duplosdk.DuploTimestreamDBTableCreateRequest {
	rq := &duplosdk.DuploTimestreamDBTableCreateRequest{
		TableName:    d.Get("name").(string),
		DatabaseName: d.Get("database_name").(string),
		Tags:         duplocloud.keyValueFromState("tags", d),
	}
	if v, ok := d.GetOk("retention_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		rq.RetentionProperties = expandRetentionProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("magnetic_store_write_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		rq.MagneticStoreWriteProperties = expandMagneticStoreWriteProperties(v.([]interface{}))
	}

	return rq
}

func parseAwsTimestreamTableIdParts(id string) (tenantID, dbName, name string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, dbName, name = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenTimestreamTable(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploTimestreamDBTableDetails, tenantId string) diag.Diagnostics {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.TableName)
	d.Set("arn", duplo.Arn)

	duplocloud.flattenTfManagedDuploKeyStringValues("tags", d, duplo.Tags)

	if err := d.Set("retention_properties", flattenRetentionProperties(duplo.RetentionProperties)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting retention_properties: %w", err))
	}

	if err := d.Set("magnetic_store_write_properties", flattenMagneticStoreWriteProperties(duplo.MagneticStoreWriteProperties)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting magnetic_store_write_properties: %w", err))
	}

	return nil
}

func flattenRetentionProperties(rp *duplosdk.DuploTimestreamDBTableRetentionProperties) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"magnetic_store_retention_period_in_days": rp.MagneticStoreRetentionPeriodInDays,
		"memory_store_retention_period_in_hours":  rp.MemoryStoreRetentionPeriodInHours,
	}

	return []interface{}{m}
}

func flattenMagneticStoreWriteProperties(rp *duplosdk.DuploTimestreamDBTableMagneticStoreWriteProperties) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enable_magnetic_store_writes":          rp.EnableMagneticStoreWrites,
		"magnetic_store_rejected_data_location": flattenMagneticStoreRejectedDataLocation(rp.MagneticStoreRejectedDataLocation),
	}

	return []interface{}{m}
}

func flattenMagneticStoreRejectedDataLocation(rp *duplosdk.MagneticStoreRejectedDataLocation) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"s3_configuration": flattenS3Configuration(rp.S3Configuration),
	}

	return []interface{}{m}
}

func flattenS3Configuration(rp *duplosdk.MagneticStoreRejectedDataS3Configuration) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket_name":       rp.BucketName,
		"object_key_prefix": rp.ObjectKeyPrefix,
		"kms_key_id":        rp.KmsKeyId,
	}
	if rp.EncryptionOption != nil {
		m["encryption_option"] = rp.EncryptionOption.Value
	}

	return []interface{}{m}
}

func expandRetentionProperties(l []interface{}) *duplosdk.DuploTimestreamDBTableRetentionProperties {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &duplosdk.DuploTimestreamDBTableRetentionProperties{}

	if v, ok := tfMap["magnetic_store_retention_period_in_days"].(int); ok {
		rp.MagneticStoreRetentionPeriodInDays = v
	}

	if v, ok := tfMap["memory_store_retention_period_in_hours"].(int); ok {
		rp.MemoryStoreRetentionPeriodInHours = v
	}

	return rp
}

func expandMagneticStoreWriteProperties(l []interface{}) *duplosdk.DuploTimestreamDBTableMagneticStoreWriteProperties {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &duplosdk.DuploTimestreamDBTableMagneticStoreWriteProperties{
		EnableMagneticStoreWrites: tfMap["enable_magnetic_store_writes"].(bool),
	}

	if v, ok := tfMap["magnetic_store_rejected_data_location"].([]interface{}); ok && len(v) > 0 {
		rp.MagneticStoreRejectedDataLocation = expandMagneticStoreRejectedDataLocation(v)
	}

	return rp
}

func expandMagneticStoreRejectedDataLocation(l []interface{}) *duplosdk.MagneticStoreRejectedDataLocation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &duplosdk.MagneticStoreRejectedDataLocation{}

	if v, ok := tfMap["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		rp.S3Configuration = expandS3Configuration(v)
	}

	return rp
}

func expandS3Configuration(l []interface{}) *duplosdk.MagneticStoreRejectedDataS3Configuration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &duplosdk.MagneticStoreRejectedDataS3Configuration{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		rp.BucketName = v
	}

	if v, ok := tfMap["object_key_prefix"].(string); ok && v != "" {
		rp.ObjectKeyPrefix = v
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		rp.KmsKeyId = v
	}

	if v, ok := tfMap["encryption_option"].(string); ok && v != "" {
		rp.EncryptionOption = &duplosdk.DuploStringValue{Value: v}
	}

	return rp
}

func timestreamTableWaitUntilActive(ctx context.Context, c *duplosdk.Client, tenantID string, name string, dbName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploTimestreamDBTableGet(tenantID, dbName, name)
			log.Printf("[TRACE] Timestream table status is (%s).", rp.TableStatus.Value)
			status := "pending"
			if err == nil {
				if rp.TableStatus != nil && rp.TableStatus.Value == "ACTIVE" {
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
	log.Printf("[DEBUG] timestreamTableWaitUntilActive(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
