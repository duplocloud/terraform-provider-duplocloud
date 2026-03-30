package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// firehoseProcessingConfigurationSchema returns the schema for processing_configuration.
// Maps to AWS SDK ProcessingConfiguration (Enabled + Processors[]).
func firehoseProcessingConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Configuration for data processing.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Description: "Whether data processing is enabled.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"processors": {
					Description: "List of processors.",
					Type:        schema.TypeList,
					Optional:    true,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"type": {
								Description:  "Processor type. Valid values: `RecordDeAggregation`, `Decompression`, `CloudWatchLogProcessing`, `Lambda`, `MetadataExtraction`, `AppendDelimiterToRecord`.",
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"RecordDeAggregation", "Decompression", "CloudWatchLogProcessing", "Lambda", "MetadataExtraction", "AppendDelimiterToRecord"}, false),
							},
							"parameters": {
								Description: "List of processor parameters.",
								Type:        schema.TypeSet,
								Optional:    true,
								Computed:    true,
								Set: func(v interface{}) int {
									m := v.(map[string]interface{})
									return schema.HashString(m["parameter_name"].(string))
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"parameter_name": {
											Description:      "Parameter name (e.g. `LambdaArn`, `NumberOfRetries`, `RoleArn`, `BufferSizeInMBs`, `BufferIntervalInSeconds`).",
											Type:             schema.TypeString,
											Required:         true,
											DiffSuppressFunc: suppressFirehoseParamRemoval,
										},
										"parameter_value": {
											Description:      "Parameter value.",
											Type:             schema.TypeString,
											Required:         true,
											DiffSuppressFunc: suppressFirehoseParamRemoval,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// firehoseCloudWatchLoggingSchema returns the schema for cloudwatch_logging_options.
func firehoseCloudWatchLoggingSchema() *schema.Schema {
	return &schema.Schema{
		Description: "CloudWatch logging configuration.",
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"log_group_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"log_stream_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

// firehoseBufferingHintsSchema returns the schema for buffering_hints.
func firehoseBufferingHintsSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Buffering options.",
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"size_in_mbs": {
					Description:  "Buffer size in MB before delivery (1–128). Default 5.",
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      5,
					ValidateFunc: validation.IntBetween(1, 128),
				},
				"interval_in_seconds": {
					Description:  "Buffer interval in seconds (60–900). Default 300.",
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      300,
					ValidateFunc: validation.IntBetween(60, 900),
				},
			},
		},
	}
}

// firehoseNestedS3ConfigElem returns the schema.Resource for a nested S3 configuration block
// used inside non-S3 destination configs (Redshift, OpenSearch, Splunk, etc.).
// Maps to AWS SDK S3DestinationConfiguration (key: "S3Configuration" in request).
func firehoseNestedS3ConfigElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket_arn": {
				Description: "ARN of the S3 bucket.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role_arn": {
				Description: "ARN of the IAM role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"prefix": {
				Description: "S3 key prefix.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"compression_format": {
				Description:  "Compression format. Default: `UNCOMPRESSED`.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "UNCOMPRESSED",
				ValidateFunc: validation.StringInSlice([]string{"UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY"}, false),
			},
			"buffering_hints":            firehoseBufferingHintsSchema(),
			"cloudwatch_logging_options": firehoseCloudWatchLoggingSchema(),
		},
	}
}

func duploAwsFirehoseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Firehose delivery stream will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the Firehose delivery stream. Duplo prepends a tenant prefix automatically.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"delivery_stream_type": {
			Description:  "Source type. `DirectPut` (default) or `KinesisStreamAsSource`.",
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Default:      "DirectPut",
			ValidateFunc: validation.StringInSlice([]string{"DirectPut", "KinesisStreamAsSource"}, false),
		},

		// Structured kinesis source configuration.
		"kinesis_source_configuration": {
			Description: "Kinesis stream source. Required when `delivery_stream_type` is `KinesisStreamAsSource`.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"kinesis_stream_arn": {
						Description: "ARN of the source Kinesis stream.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"role_arn": {
						Description: "ARN of the IAM role that grants Firehose access to the Kinesis stream.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},

		// Structured extended S3 destination configuration.
		"extended_s3_destination_configuration": {
			Description: "Extended S3 destination configuration (recommended over `s3_destination_configuration`).",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"bucket_arn": {
						Description: "ARN of the S3 bucket.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"role_arn": {
						Description: "ARN of the IAM role used to access the S3 bucket.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"prefix": {
						Description: "YYYY/MM/DD/HH prefix pattern for S3 object keys.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"error_output_prefix": {
						Description: "Prefix pattern for failed records.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"compression_format": {
						Description:  "Compression format. Default: `UNCOMPRESSED`.",
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "UNCOMPRESSED",
						ValidateFunc: validation.StringInSlice([]string{"UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY"}, false),
					},
					"buffering_hints":            firehoseBufferingHintsSchema(),
					"processing_configuration":   firehoseProcessingConfigurationSchema(),
					"cloudwatch_logging_options": firehoseCloudWatchLoggingSchema(),
					"s3_backup_mode": {
						Description:  "S3 backup mode. `Disabled` (default) or `Enabled`.",
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "Disabled",
						ValidateFunc: validation.StringInSlice([]string{"Disabled", "Enabled"}, false),
					},
				},
			},
		},

		// Redshift destination configuration.
		"redshift_destination_configuration": {
			Description: "Redshift destination configuration.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster_jdbcurl": {
						Description: "The database connection string.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"role_arn": {
						Description: "ARN of the IAM role.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"username": {
						Description: "The Redshift username.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"password": {
						Description: "The Redshift password.",
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
					},
					"copy_command": {
						Description: "COPY command configuration.",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"data_table_name": {
									Description: "The name of the target table.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"data_table_columns": {
									Description: "Comma-separated list of column names.",
									Type:        schema.TypeString,
									Optional:    true,
								},
								"copy_options": {
									Description: "Optional parameters for the COPY command.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"s3_configuration": {
						Description: "Intermediate S3 configuration (mandatory for Redshift).",
						Type:        schema.TypeList,
						Required:    true,
						MaxItems:    1,
						Elem:        firehoseNestedS3ConfigElem(),
					},
					"s3_backup_mode": {
						Description:  "S3 backup mode. `Disabled` (default) or `Enabled`.",
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "Disabled",
						ValidateFunc: validation.StringInSlice([]string{"Disabled", "Enabled"}, false),
					},
					"processing_configuration":   firehoseProcessingConfigurationSchema(),
					"cloudwatch_logging_options": firehoseCloudWatchLoggingSchema(),
				},
			},
		},

		"tags": {
			Description: "Key-value map of resource tags.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"arn": {
			Description: "The ARN of the Firehose delivery stream.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The current status of the delivery stream.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"full_name": {
			Description: "The full name of the delivery stream including the Duplo tenant prefix.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceDuploAwsFirehose() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_aws_firehose` manages an AWS Data Firehose delivery stream in Duplo.",
		ReadContext:   resourceDuploAwsFirehoseRead,
		CreateContext: resourceDuploAwsFirehoseCreate,
		UpdateContext: resourceDuploAwsFirehoseUpdate,
		DeleteContext: resourceDuploAwsFirehoseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsFirehoseSchema(),
	}
}

func resourceDuploAwsFirehoseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsFirehoseIDParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsFirehoseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	stream, clientErr := c.DuploFirehoseGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceDuploAwsFirehoseRead(%s, %s): object missing", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if stream == nil {
		log.Printf("[TRACE] resourceDuploAwsFirehoseRead(%s, %s): object missing", tenantID, name)
		d.SetId("")
		return nil
	}

	flattenAwsFirehose(d, tenantID, stream)
	log.Printf("[TRACE] resourceDuploAwsFirehoseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceDuploAwsFirehoseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsFirehoseCreate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq := expandAwsFirehose(d)

	clientErr := c.DuploFirehoseCreate(tenantID, rq)
	if clientErr != nil {
		return diag.Errorf("Error creating tenant %s Firehose delivery stream '%s': %s", tenantID, name, clientErr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	d.SetId(id)

	if diags := waitForFirehoseActive(ctx, d, c, tenantID, name); diags != nil {
		return diags
	}

	diags := resourceDuploAwsFirehoseRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploAwsFirehoseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceDuploAwsFirehoseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsFirehoseIDParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsFirehoseUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq := expandAwsFirehose(d)

	clientErr := c.DuploFirehoseUpdate(tenantID, name, rq)
	if clientErr != nil {
		return diag.Errorf("Error updating tenant %s Firehose delivery stream '%s': %s", tenantID, name, clientErr)
	}

	diags := resourceDuploAwsFirehoseRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploAwsFirehoseUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceDuploAwsFirehoseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsFirehoseIDParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsFirehoseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploFirehoseDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceDuploAwsFirehoseDelete(%s, %s): object already missing", tenantID, name)
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s Firehose delivery stream '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "Firehose delivery stream", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploFirehoseGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceDuploAwsFirehoseDelete(%s, %s): end", tenantID, name)
	return nil
}

func waitForFirehoseActive(ctx context.Context, d *schema.ResourceData, c *duplosdk.Client, tenantID, name string) diag.Diagnostics {
	id := fmt.Sprintf("%s/%s", tenantID, name)
	err := retry.RetryContext(ctx, d.Timeout("create"), func() *retry.RetryError {
		stream, clientErr := c.DuploFirehoseGet(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				return retry.RetryableError(fmt.Errorf("Firehose delivery stream '%s' not yet available", id))
			}
			return retry.NonRetryableError(fmt.Errorf("error getting Firehose delivery stream '%s': %s", id, clientErr))
		}
		if stream == nil || duplosdk.FirehoseStringValue(stream.DeliveryStreamStatus) != "ACTIVE" {
			status := "unknown"
			if stream != nil {
				status = duplosdk.FirehoseStringValue(stream.DeliveryStreamStatus)
			}
			return retry.RetryableError(fmt.Errorf("Firehose delivery stream '%s' is not yet ACTIVE (current: %s)", id, status))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("error waiting for Firehose delivery stream '%s' to become ACTIVE: %s", id, err)
	}
	return nil
}

// expandAwsFirehose builds the DuploFirehoseRequest from Terraform state.
// Destination config blocks use Terraform snake_case; this function translates to
// AWS SDK PascalCase field names expected by the Duplo backend.
func expandAwsFirehose(d *schema.ResourceData) *duplosdk.DuploFirehoseRequest {
	rq := &duplosdk.DuploFirehoseRequest{
		DeliveryStreamName: d.Get("name").(string),
		DeliveryStreamType: d.Get("delivery_stream_type").(string),
		Tags:               expandAsStringMap("tags", d),
	}

	if v, ok := d.GetOk("kinesis_source_configuration"); ok {
		rq.KinesisStreamSourceConfiguration = expandFirehoseKinesisSource(v.([]interface{}))
	}

	if v, ok := d.GetOk("extended_s3_destination_configuration"); ok {
		rq.ExtendedS3DestinationConfiguration = expandFirehoseExtendedS3(v.([]interface{}))
	}

	if v, ok := d.GetOk("redshift_destination_configuration"); ok {
		rq.RedshiftDestinationConfiguration = expandFirehoseRedshift(v.([]interface{}))
	}

	return rq
}

// expandFirehoseKinesisSource maps kinesis_source_configuration → KinesisStreamSourceConfiguration (PascalCase).
func expandFirehoseKinesisSource(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	return map[string]interface{}{
		"KinesisStreamARN": m["kinesis_stream_arn"].(string),
		"RoleARN":          m["role_arn"].(string),
	}
}

// expandFirehoseExtendedS3 maps extended_s3_destination_configuration → ExtendedS3DestinationConfiguration (PascalCase).
func expandFirehoseExtendedS3(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	result := map[string]interface{}{
		"BucketARN": m["bucket_arn"].(string),
		"RoleARN":   m["role_arn"].(string),
	}

	if s, ok := m["prefix"].(string); ok && s != "" {
		result["Prefix"] = s
	}
	if s, ok := m["error_output_prefix"].(string); ok && s != "" {
		result["ErrorOutputPrefix"] = s
	}
	if s, ok := m["compression_format"].(string); ok && s != "" && s != "UNCOMPRESSED" {
		result["CompressionFormat"] = s
	}
	if s, ok := m["s3_backup_mode"].(string); ok && s != "" && s != "Disabled" {
		result["S3BackupMode"] = s
	}

	if bh := m["buffering_hints"].([]interface{}); len(bh) > 0 && bh[0] != nil {
		bhMap := bh[0].(map[string]interface{})
		result["BufferingHints"] = map[string]interface{}{
			"SizeInMBs":         bhMap["size_in_mbs"].(int),
			"IntervalInSeconds": bhMap["interval_in_seconds"].(int),
		}
	}

	if pc := m["processing_configuration"].([]interface{}); len(pc) > 0 && pc[0] != nil {
		result["ProcessingConfiguration"] = expandFirehoseProcessingConfig(pc)
	}

	if cw := m["cloudwatch_logging_options"].([]interface{}); len(cw) > 0 && cw[0] != nil {
		result["CloudWatchLoggingOptions"] = expandFirehoseCloudWatchLogging(cw)
	}

	return result
}

// expandFirehoseNestedS3Config maps a nested s3_configuration block → S3Configuration (PascalCase).
func expandFirehoseNestedS3Config(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	result := map[string]interface{}{
		"BucketARN": m["bucket_arn"].(string),
		"RoleARN":   m["role_arn"].(string),
	}
	if s, ok := m["prefix"].(string); ok && s != "" {
		result["Prefix"] = s
	}
	if s, ok := m["compression_format"].(string); ok && s != "" && s != "UNCOMPRESSED" {
		result["CompressionFormat"] = s
	}
	if bh := m["buffering_hints"].([]interface{}); len(bh) > 0 && bh[0] != nil {
		bhMap := bh[0].(map[string]interface{})
		result["BufferingHints"] = map[string]interface{}{
			"SizeInMBs":         bhMap["size_in_mbs"].(int),
			"IntervalInSeconds": bhMap["interval_in_seconds"].(int),
		}
	}
	if cw := m["cloudwatch_logging_options"].([]interface{}); len(cw) > 0 && cw[0] != nil {
		result["CloudWatchLoggingOptions"] = expandFirehoseCloudWatchLogging(cw)
	}
	return result
}

// expandFirehoseRedshift maps redshift_destination_configuration → RedshiftDestinationConfiguration (PascalCase).
func expandFirehoseRedshift(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	result := map[string]interface{}{
		"ClusterJDBCURL": m["cluster_jdbcurl"].(string),
		"RoleARN":        m["role_arn"].(string),
		"Username":       m["username"].(string),
		"Password":       m["password"].(string),
	}
	if cc := m["copy_command"].([]interface{}); len(cc) > 0 && cc[0] != nil {
		ccMap := cc[0].(map[string]interface{})
		copyCmd := map[string]interface{}{}
		if s, ok := ccMap["data_table_name"].(string); ok && s != "" {
			copyCmd["DataTableName"] = s
		}
		if s, ok := ccMap["data_table_columns"].(string); ok && s != "" {
			copyCmd["DataTableColumns"] = s
		}
		if s, ok := ccMap["copy_options"].(string); ok && s != "" {
			copyCmd["CopyOptions"] = s
		}
		result["CopyCommand"] = copyCmd
	}
	if s3 := m["s3_configuration"].([]interface{}); len(s3) > 0 {
		result["S3Configuration"] = expandFirehoseNestedS3Config(s3)
	}
	if s, ok := m["s3_backup_mode"].(string); ok && s != "" && s != "Disabled" {
		result["S3BackupMode"] = s
	}
	if pc := m["processing_configuration"].([]interface{}); len(pc) > 0 && pc[0] != nil {
		result["ProcessingConfiguration"] = expandFirehoseProcessingConfig(pc)
	}
	if cw := m["cloudwatch_logging_options"].([]interface{}); len(cw) > 0 && cw[0] != nil {
		result["CloudWatchLoggingOptions"] = expandFirehoseCloudWatchLogging(cw)
	}
	return result
}

// expandFirehoseProcessingConfig maps processing_configuration → ProcessingConfiguration (PascalCase).
func expandFirehoseProcessingConfig(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	result := map[string]interface{}{
		"Enabled": m["enabled"].(bool),
	}

	if procs, ok := m["processors"].([]interface{}); ok && len(procs) > 0 {
		processors := make([]map[string]interface{}, 0, len(procs))
		for _, p := range procs {
			pm := p.(map[string]interface{})
			proc := map[string]interface{}{
				"Type": pm["type"].(string),
			}
			var paramsList []interface{}
			if ps, ok := pm["parameters"].(*schema.Set); ok {
				paramsList = ps.List()
			} else if pl, ok := pm["parameters"].([]interface{}); ok {
				paramsList = pl
			}
			if len(paramsList) > 0 {
				parameters := make([]map[string]interface{}, 0, len(paramsList))
				for _, param := range paramsList {
					paramMap := param.(map[string]interface{})
					parameters = append(parameters, map[string]interface{}{
						"ParameterName":  paramMap["parameter_name"].(string),
						"ParameterValue": paramMap["parameter_value"].(string),
					})
				}
				proc["Parameters"] = parameters
			}
			processors = append(processors, proc)
		}
		result["Processors"] = processors
	}

	return result
}

// expandFirehoseCloudWatchLogging maps cloudwatch_logging_options → CloudWatchLoggingOptions (PascalCase).
func expandFirehoseCloudWatchLogging(v []interface{}) map[string]interface{} {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	return map[string]interface{}{
		"Enabled":       m["enabled"].(bool),
		"LogGroupName":  m["log_group_name"].(string),
		"LogStreamName": m["log_stream_name"].(string),
	}
}

// flattenAwsFirehose populates Terraform state from a DuploFirehoseDeliveryStream response.
func flattenAwsFirehose(d *schema.ResourceData, tenantID string, stream *duplosdk.DuploFirehoseDeliveryStream) {
	arn := stream.DeliveryStreamARN
	status := duplosdk.FirehoseStringValue(stream.DeliveryStreamStatus)
	streamType := duplosdk.FirehoseStringValue(stream.DeliveryStreamType)

	// Use short name from resource ID — DeliveryStreamName from AWS is the full prefixed name.
	_, shortName, _ := parseAwsFirehoseIDParts(d.Id())

	d.Set("tenant_id", tenantID)
	d.Set("name", shortName)
	d.Set("full_name", stream.DeliveryStreamName)
	d.Set("arn", arn)
	d.Set("status", status)
	d.Set("delivery_stream_type", streamType)

	// Parse Destinations[0] to reconstruct destination config blocks.
	if dests, ok := stream.Destinations.([]interface{}); ok && len(dests) > 0 {
		if destMap, ok := dests[0].(map[string]interface{}); ok {
			if extS3, ok := destMap["ExtendedS3DestinationDescription"].(map[string]interface{}); ok {
				flattened := flattenFirehoseExtendedS3(extS3)
				preserveFirehoseProcessingConfig(d, "extended_s3_destination_configuration", flattened)
				d.Set("extended_s3_destination_configuration", flattened)
			}
			if rs, ok := destMap["RedshiftDestinationDescription"].(map[string]interface{}); ok {
				flattened := flattenFirehoseRedshift(rs)
				preserveFirehoseProcessingConfig(d, "redshift_destination_configuration", flattened)
				d.Set("redshift_destination_configuration", flattened)
			}
		}
	}

	// Kinesis source is at the top-level Source field, not inside Destinations.
	if src, ok := stream.Source.(map[string]interface{}); ok {
		if ksis, ok := src["KinesisStreamSourceDescription"].(map[string]interface{}); ok {
			d.Set("kinesis_source_configuration", flattenFirehoseKinesisSource(ksis))
		}
	}

	// Tags are not returned by the Firehose describe endpoint.
	// Preserve whatever is already in state so they don't phantom-diff on every plan.
	d.Set("tags", d.Get("tags"))
}

// preserveFirehoseProcessingConfig filters the processing_configuration in the
// flattened AWS response to only include parameters the user actually configured.
// AWS auto-adds default parameters (e.g. NumberOfRetries, RoleArn) that the user
// didn't configure, causing phantom diffs on every plan.
// Uses d.GetRawConfig() (the actual .tf file) not d.GetOk() (stale state).
func preserveFirehoseProcessingConfig(d *schema.ResourceData, blockKey string, flattened []interface{}) {
	if len(flattened) == 0 || flattened[0] == nil {
		return
	}
	flatMap, ok := flattened[0].(map[string]interface{})
	if !ok {
		return
	}

	configuredParams := firehoseConfiguredParamNames(d, blockKey)
	if len(configuredParams) == 0 {
		return
	}

	// Filter AWS-returned parameters to only include user-configured ones.
	pc, ok := flatMap["processing_configuration"].([]interface{})
	if !ok || len(pc) == 0 || pc[0] == nil {
		return
	}
	pcMap, ok := pc[0].(map[string]interface{})
	if !ok {
		return
	}
	procs, ok := pcMap["processors"].([]interface{})
	if !ok {
		return
	}
	for i, p := range procs {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		params, ok := pm["parameters"].([]interface{})
		if !ok {
			continue
		}
		filtered := make([]interface{}, 0, len(params))
		for _, param := range params {
			paramMap, ok := param.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := paramMap["parameter_name"].(string)
			if configuredParams[name] {
				filtered = append(filtered, param)
			}
		}
		pm["parameters"] = filtered
		procs[i] = pm
	}
	pcMap["processors"] = procs
	pc[0] = pcMap
	flatMap["processing_configuration"] = pc
}

// suppressFirehoseParamRemoval is a DiffSuppressFunc for parameter_name and parameter_value.
// It suppresses removal of parameters that AWS auto-adds as defaults (e.g. NumberOfRetries,
// RoleArn) which the user never configured, preventing phantom diffs on every plan.
// Runs during PlanResourceChange where d.GetRawConfig() is populated.
func suppressFirehoseParamRemoval(k, old, new string, d *schema.ResourceData) bool {
	// Only suppress when a value is disappearing from state (old → "")
	if old == "" || new != "" {
		return false
	}
	// Extract the top-level block key (first segment of path)
	dotIdx := strings.Index(k, ".")
	if dotIdx < 0 {
		return false
	}
	blockKey := k[:dotIdx]
	// Determine the parameter_name of the element being removed
	var paramName string
	switch {
	case strings.HasSuffix(k, ".parameter_name"):
		paramName = old
	case strings.HasSuffix(k, ".parameter_value"):
		// d.Get returns new/planned value (empty for removed elements).
		// Use GetChange to get the OLD state value of parameter_name.
		paramNameKey := strings.TrimSuffix(k, ".parameter_value") + ".parameter_name"
		oldParamName, _ := d.GetChange(paramNameKey)
		paramName, _ = oldParamName.(string)
	default:
		return false
	}
	if paramName == "" {
		return false
	}
	// Suppress only if this parameter is NOT in the user's .tf config
	configuredParams := firehoseConfiguredParamNames(d, blockKey)
	return len(configuredParams) > 0 && !configuredParams[paramName]
}

// firehoseConfiguredParamNames reads the actual .tf config (via GetRawConfig, not state)
// and returns the set of parameter_name values the user configured for the first processor
// inside processing_configuration of the given destination block.
func firehoseConfiguredParamNames(d *schema.ResourceData, blockKey string) map[string]bool {
	names := make(map[string]bool)
	raw := d.GetRawConfig()
	if raw == cty.NilVal || raw.IsNull() || !raw.IsKnown() {
		return names
	}
	ty := raw.Type()
	if !ty.IsObjectType() || !ty.HasAttribute(blockKey) {
		return names
	}
	blockList := raw.GetAttr(blockKey)
	if blockList.IsNull() || !blockList.IsKnown() || blockList.LengthInt() == 0 {
		return names
	}
	first := blockList.Index(cty.NumberIntVal(0))
	if first.IsNull() || !first.IsKnown() || !first.Type().IsObjectType() {
		return names
	}
	if !first.Type().HasAttribute("processing_configuration") {
		return names
	}
	pcList := first.GetAttr("processing_configuration")
	if pcList.IsNull() || !pcList.IsKnown() || pcList.LengthInt() == 0 {
		return names
	}
	pcFirst := pcList.Index(cty.NumberIntVal(0))
	if pcFirst.IsNull() || !pcFirst.IsKnown() || !pcFirst.Type().IsObjectType() {
		return names
	}
	if !pcFirst.Type().HasAttribute("processors") {
		return names
	}
	procList := pcFirst.GetAttr("processors")
	if procList.IsNull() || !procList.IsKnown() || procList.LengthInt() == 0 {
		return names
	}
	procFirst := procList.Index(cty.NumberIntVal(0))
	if procFirst.IsNull() || !procFirst.IsKnown() || !procFirst.Type().IsObjectType() {
		return names
	}
	if !procFirst.Type().HasAttribute("parameters") {
		return names
	}
	params := procFirst.GetAttr("parameters")
	if params.IsNull() || !params.IsKnown() {
		return names
	}
	for it := params.ElementIterator(); it.Next(); {
		_, elem := it.Element()
		if elem.IsNull() || !elem.IsKnown() || !elem.Type().IsObjectType() {
			continue
		}
		if !elem.Type().HasAttribute("parameter_name") {
			continue
		}
		pnVal := elem.GetAttr("parameter_name")
		if !pnVal.IsNull() && pnVal.IsKnown() && pnVal.Type() == cty.String {
			names[pnVal.AsString()] = true
		}
	}
	return names
}

func flattenFirehoseExtendedS3(m map[string]interface{}) []interface{} {
	cfg := map[string]interface{}{
		"bucket_arn":          strFromMap(m, "BucketARN"),
		"role_arn":            strFromMap(m, "RoleARN"),
		"prefix":              strFromMap(m, "Prefix"),
		"error_output_prefix": strFromMap(m, "ErrorOutputPrefix"),
		"compression_format":  firehoseEnumFromMap(m, "CompressionFormat"),
		"s3_backup_mode":      firehoseEnumFromMap(m, "S3BackupMode"),
	}

	if bh, ok := m["BufferingHints"].(map[string]interface{}); ok {
		cfg["buffering_hints"] = []interface{}{map[string]interface{}{
			"size_in_mbs":         intFromMap(bh, "SizeInMBs"),
			"interval_in_seconds": intFromMap(bh, "IntervalInSeconds"),
		}}
	}

	if pc, ok := m["ProcessingConfiguration"].(map[string]interface{}); ok {
		cfg["processing_configuration"] = flattenFirehoseProcessingConfig(pc)
	}

	if cw, ok := m["CloudWatchLoggingOptions"].(map[string]interface{}); ok {
		cfg["cloudwatch_logging_options"] = []interface{}{map[string]interface{}{
			"enabled":         boolFromMap(cw, "Enabled"),
			"log_group_name":  strFromMap(cw, "LogGroupName"),
			"log_stream_name": strFromMap(cw, "LogStreamName"),
		}}
	}

	return []interface{}{cfg}
}

func flattenFirehoseProcessingConfig(m map[string]interface{}) []interface{} {
	cfg := map[string]interface{}{
		"enabled": boolFromMap(m, "Enabled"),
	}

	if procs, ok := m["Processors"].([]interface{}); ok {
		processors := make([]interface{}, 0, len(procs))
		for _, p := range procs {
			if pm, ok := p.(map[string]interface{}); ok {
				proc := map[string]interface{}{
					"type": firehoseEnumFromMap(pm, "Type"),
				}
				if params, ok := pm["Parameters"].([]interface{}); ok {
					parameters := make([]interface{}, 0, len(params))
					for _, param := range params {
						if paramMap, ok := param.(map[string]interface{}); ok {
							parameters = append(parameters, map[string]interface{}{
								"parameter_name":  firehoseEnumFromMap(paramMap, "ParameterName"),
								"parameter_value": strFromMap(paramMap, "ParameterValue"),
							})
						}
					}
					proc["parameters"] = parameters
				}
				processors = append(processors, proc)
			}
		}
		cfg["processors"] = processors
	}

	return []interface{}{cfg}
}

// flattenFirehoseNestedS3Config flattens a nested S3 config description (PascalCase → snake_case).
func flattenFirehoseNestedS3Config(m map[string]interface{}) []interface{} {
	cfg := map[string]interface{}{
		"bucket_arn":         strFromMap(m, "BucketARN"),
		"role_arn":           strFromMap(m, "RoleARN"),
		"prefix":             strFromMap(m, "Prefix"),
		"compression_format": firehoseEnumFromMap(m, "CompressionFormat"),
	}
	if bh, ok := m["BufferingHints"].(map[string]interface{}); ok {
		cfg["buffering_hints"] = []interface{}{map[string]interface{}{
			"size_in_mbs":         intFromMap(bh, "SizeInMBs"),
			"interval_in_seconds": intFromMap(bh, "IntervalInSeconds"),
		}}
	}
	if cw, ok := m["CloudWatchLoggingOptions"].(map[string]interface{}); ok {
		cfg["cloudwatch_logging_options"] = []interface{}{map[string]interface{}{
			"enabled":         boolFromMap(cw, "Enabled"),
			"log_group_name":  strFromMap(cw, "LogGroupName"),
			"log_stream_name": strFromMap(cw, "LogStreamName"),
		}}
	}
	return []interface{}{cfg}
}

func flattenFirehoseRedshift(m map[string]interface{}) []interface{} {
	cfg := map[string]interface{}{
		"cluster_jdbcurl": strFromMap(m, "ClusterJDBCURL"),
		"role_arn":        strFromMap(m, "RoleARN"),
		"username":        strFromMap(m, "Username"),
		"password":        strFromMap(m, "Password"),
		"s3_backup_mode":  firehoseEnumFromMap(m, "S3BackupMode"),
	}
	if cc, ok := m["CopyCommand"].(map[string]interface{}); ok {
		cfg["copy_command"] = []interface{}{map[string]interface{}{
			"data_table_name":    strFromMap(cc, "DataTableName"),
			"data_table_columns": strFromMap(cc, "DataTableColumns"),
			"copy_options":       strFromMap(cc, "CopyOptions"),
		}}
	}
	if s3, ok := m["S3DestinationDescription"].(map[string]interface{}); ok {
		cfg["s3_configuration"] = flattenFirehoseNestedS3Config(s3)
	}
	if pc, ok := m["ProcessingConfiguration"].(map[string]interface{}); ok {
		cfg["processing_configuration"] = flattenFirehoseProcessingConfig(pc)
	}
	if cw, ok := m["CloudWatchLoggingOptions"].(map[string]interface{}); ok {
		cfg["cloudwatch_logging_options"] = []interface{}{map[string]interface{}{
			"enabled":         boolFromMap(cw, "Enabled"),
			"log_group_name":  strFromMap(cw, "LogGroupName"),
			"log_stream_name": strFromMap(cw, "LogStreamName"),
		}}
	}
	return []interface{}{cfg}
}

func flattenFirehoseKinesisSource(m map[string]interface{}) []interface{} {
	return []interface{}{map[string]interface{}{
		"kinesis_stream_arn": strFromMap(m, "KinesisStreamARN"),
		"role_arn":           strFromMap(m, "RoleARN"),
	}}
}

// firehoseEnumFromMap extracts a string from a map value that may be either a plain string
// or a C# ConstantClass object {"Value":"..."} as returned by the Duplo backend.
func firehoseEnumFromMap(m map[string]interface{}, key string) string {
	return duplosdk.FirehoseStringValue(m[key])
}

// Helpers for safely extracting typed values from map[string]interface{}.
func strFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolFromMap(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func intFromMap(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

func parseAwsFirehoseIDParts(id string) (tenantID, name string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) == 2 {
		tenantID, name = parts[0], parts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
