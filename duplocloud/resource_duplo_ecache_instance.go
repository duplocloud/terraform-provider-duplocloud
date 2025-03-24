package duplocloud

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/go-cty/cty"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	TOTALECACHENAMELENGTH = 40
)

func ecacheInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the elasticache instance will be created in.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the elasticache instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 50-MAX_DUPLO_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`), "Invalid AWS Elasticache cluster name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "AWS Elasticache cluster names cannot end with a hyphen"),
				validation.StringNotInSlice([]string{"--"}, true),
			),
		},
		"identifier": {
			Description: "The full name of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"endpoint": {
			Description: "The endpoint of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"host": {
			Description: "The DNS hostname of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the elasticache instance.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"cache_type": {
			Description: "The numerical index of elasticache instance type.\n" +
				"Should be one of:\n\n" +
				"   - `0` : Redis\n" +
				"   - `1` : Memcache\n\n",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      0,
			ValidateFunc: validation.IntBetween(0, 1),
		},
		"engine_version": {
			Description: "The engine version of the elastic instance.\n" +
				"See AWS documentation for the [available Redis instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/supported-engine-versions.html) " +
				"or the [available Memcached instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/supported-engine-versions-mc.html).",
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			DiffSuppressFunc: suppressEnginePatchVersion,
		},
		"size": {
			Description: "The instance type of the elasticache instance.\n" +
				"See AWS documentation for the [available instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html).",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^cache\.`), "Elasticache instance types must start with 'cache.'"),
		},
		"replicas": {
			Description:  "The number of replicas to create. Supported number of replicas is 1 to 6",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      1,
			ValidateFunc: validation.IntBetween(1, 6),
		},
		"encryption_at_rest": {
			Description: "Enables encryption-at-rest.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"automatic_failover_enabled": {
			Description: "Enables automatic failover.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"encryption_in_transit": {
			Description: "Enables encryption-in-transit.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"auth_token": {
			Description: "Set a password for authenticating to the ElastiCache instance.  Only supported if `encryption_in_transit` is to to `true`.\n\n" +
				"See AWS documentation for the [required format](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth.html) of this field.",
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(16, 128),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9!&#$<>^-]*$`), "Invalid AWS Elasticache Redis password"),
			),
		},
		"instance_status": {
			Description: "The status of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kms_key_id": {
			Description: "The globally unique identifier for the key.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"parameter_group_name": {
			Description: "The REDIS parameter group to supply.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"enable_cluster_mode": {
			Description: "Flag to enable/disable redis cluster mode.",
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
		},
		"number_of_shards": {
			Description:      "The number of shards to create. Applicable only if enable_cluster_mode is set to true",
			Type:             schema.TypeInt,
			Optional:         true,
			DiffSuppressFunc: suppressNoOfShardsDiff,
			Computed:         true,
			ValidateFunc:     validation.IntBetween(1, 500),
		},
		"snapshot_arns": {
			Description:   "Specify the ARN of a Redis RDB snapshot file stored in Amazon S3. User should have the access to export snapshot to s3 bucket. One can find steps to provide access to export snapshot to s3 on following link https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/backups-exporting.html",
			Type:          schema.TypeList,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"snapshot_name"},
			Elem:          &schema.Schema{Type: schema.TypeString},
		},
		"snapshot_name": {
			Description:   "Select the snapshot/backup you want to use for creating redis.",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"snapshot_arns"},
		},
		"snapshot_retention_limit": {
			Description:  "Specify retention limit in days. Accepted values - 1-35.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(1, 35),
		},
		"snapshot_window": {
			Description:      "Specify snapshot window limit The daily time range (in UTC) during which ElastiCache begins taking a daily snapshot of your node group (shard). Example: 05:00-09:00. If you do not specify this parameter, ElastiCache automatically chooses an appropriate time range.",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: isValidSnapshotWindow(),
		},
		"log_delivery_configuration": {
			Type:     schema.TypeSet,
			MaxItems: 2,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"destination_type": {
						Description: "destination type : must be cloudwatch-logs.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_CLOUDWATCH_LOGS,
						}, false),
					},
					"log_format": {
						Type:        schema.TypeString,
						Description: "log_format: Value must be one of the ['json', 'text']",
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							duplosdk.REDIS_LOG_DELIVERY_LOG_FORMAT_JSON,
							duplosdk.REDIS_LOG_DELIVERY_LOG_FORMAT_TEXT,
						}, true),
					},
					"log_type": {
						Type:        schema.TypeString,
						Description: "log_type: Value must be one of the ['slow-log', 'engine-log']",
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							duplosdk.REDIS_LOG_DELIVERY_LOG_TYPE_SLOW_LOG,
							duplosdk.REDIS_LOG_DELIVERY_LOG_TYPE_ENGINE_LOG,
						}, false),
					},
					"log_group": {
						Description: "cloudwatch log_group",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcacheInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecache_instance` used to manage Amazon ElastiCache instances within a DuploCloud-managed environment. " +
			"<p>This resource allows you to define and manage Redis or Memcached instances on AWS through Terraform, with DuploCloud handling the underlying infrastructure and integration aspects." +
			"</p>",

		ReadContext:   resourceDuploEcacheInstanceRead,
		CreateContext: resourceDuploEcacheInstanceCreate,
		DeleteContext: resourceDuploEcacheInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(29 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        ecacheInstanceSchema(),
		CustomizeDiff: validateEcacheParameters,
	}
}

// READ resource
func resourceDuploEcacheInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseECacheInstanceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploEcacheInstanceRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcacheInstanceGet(tenantID, name)
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	diag := flattenEcacheInstance(duplo, d)

	log.Printf("[TRACE] resourceDuploEcacheInstanceRead(%s, %s): end", tenantID, name)
	return diag
}

// CREATE resource
func resourceDuploEcacheInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate(%s): start", tenantID)

	duplo, diagErr := expandEcacheInstance(d)
	if diagErr != nil {
		return diagErr
	}
	c := m.(*duplosdk.Client)

	fullName, errname := c.GetResourceName("duploservices", tenantID, duplo.Name, false)
	if errname != nil {
		return diag.Errorf("resourceDuploEcacheInstanceCreate: Unable to retrieve duplo service name (name: %s, error: %s)", duplo.Name, errname.Error())

	}
	if !validateStringLength(fullName, TOTALECACHENAMELENGTH) {
		return diag.Errorf("resourceDuploEcacheInstanceCreate: fullname %s exceeds allowable ecache name length %d)", fullName, TOTALECACHENAMELENGTH)

	}

	duplo.Identifier = duplo.Name
	id := fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/%s", tenantID, duplo.Name)

	// Perform additional validation.
	if !duplo.EncryptionInTransit && duplo.AuthToken != "" {
		return diag.Errorf("Invalid ECache instance '%s': 'auth_token' must not be specified when 'encryption_in_transit' is false", id)
	}

	if duplicateLogType, found := hasDuplicateLogTypes(*duplo.LogDeliveryConfigurations); found {
		return diag.Errorf("log_delivery_configuration: Duplicate log_type are not allowed. Found '%s' log_type repeated.", duplicateLogType)
	}

	errDiag := validateLogDeliveryConfigurations(duplo.EngineVersion, *duplo.LogDeliveryConfigurations)
	if errDiag != nil {
		return errDiag
	}

	if duplo.Replicas < 2 && duplo.AutomaticFailoverEnabled {
		return diag.Errorf("Invalid automatic_failover_enabled '%s': To enable automatic failover, replicas must be 2 or more", id)
	}

	// Post the object to Duplo
	_, err = c.EcacheInstanceCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error updating ECache instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "ECache instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.EcacheInstanceGet(tenantID, duplo.Name)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = ecacheInstanceWaitUntilAvailable(ctx, c, tenantID, duplo.Name)
	if err != nil {
		return diag.Errorf("Error waiting for ECache instance '%s' to be available: %s", id, err)
	}

	// Read the resource state
	diags = resourceDuploEcacheInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate(%s, %s): end", tenantID, duplo.Name)
	return diags
}

func validateLogDeliveryConfigurations(engineVersion string, configs []duplosdk.LogDeliveryConfigurationRequest) diag.Diagnostics {
	if engineVersion == "" || len(configs) == 0 {
		return nil
	}
	for _, config := range configs {
		switch config.LogType {
		case "engine-log":
			if !IsAppVersionEqualOrGreater(engineVersion, "6.2.0") {
				return diag.Errorf("log_delivery_configuration with log_type 'engine-log' cannot be used with engine_version '%s'. Please use engine_version '6.2.0' or above.", engineVersion)
			}
		case "slow-log":
			if !IsAppVersionEqualOrGreater(engineVersion, "6.0.0") {
				return diag.Errorf("log_delivery_configuration with log_type 'slow-log' cannot be used with engine_version '%s'. Please use engine_version '6.0.0' or above.", engineVersion)
			}
		default:
			return nil
		}
	}
	return nil
}

func hasDuplicateLogTypes(configs []duplosdk.LogDeliveryConfigurationRequest) (string, bool) {
	seen := make(map[string]bool)

	for _, config := range configs {
		if _, exists := seen[config.LogType]; exists {
			return config.LogType, true // Duplicate found
		}
		seen[config.LogType] = true
	}

	return "", false // No duplicates
}

// DELETE resource
func resourceDuploEcacheInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseECacheInstanceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploEcacheInstanceDelete(%s, %s): start", tenantID, name)

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	err = c.EcacheInstanceDelete(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to show the object as deleted.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "ECache instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.EcacheInstanceGet(tenantID, name)
	})

	// Wait for some time to deal with consistency issues.
	if diag == nil {
		time.Sleep(time.Duration(90) * time.Second)
	}

	log.Printf("[TRACE] resourceDuploEcacheInstanceDelete(%s, %s): end", tenantID, name)
	return diag
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

func expandLogDeliveryConfigurations(s []interface{}) ([]duplosdk.LogDeliveryConfigurationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(s) == 0 {
		return nil, diags
	}

	items := make([]duplosdk.LogDeliveryConfigurationRequest, 0, len(s))
	for _, i := range s {
		itemMap, ok := i.(map[string]interface{})
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Type assertion failed",
				Detail:   "Expected map[string]interface{} but got something else",
			})
			continue
		}
		item, diagErr := expandLogDeliveryConfiguration(itemMap)
		if diagErr != nil {
			diags = append(diags, diagErr...)
			continue
		}
		items = append(items, *item)
	}
	return items, diags
}

func expandLogDeliveryConfiguration(logDelConfig map[string]interface{}) (*duplosdk.LogDeliveryConfigurationRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	if err := validateLogDeliveryConfiguration(logDelConfig); err != nil {
		diags = append(diags, err...)
		return nil, diags
	}

	duplo := &duplosdk.LogDeliveryConfigurationRequest{
		DestinationType: logDelConfig["destination_type"].(string),
		LogFormat:       logDelConfig["log_format"].(string),
		LogType:         logDelConfig["log_type"].(string),
	}

	switch duplo.DestinationType {
	case duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_CLOUDWATCH_LOGS:
		cloudwatch := &duplosdk.CloudWatchLogsDestinationDetails{
			LogGroup: logDelConfig["log_group"].(string),
		}
		duplo.DestinationDetails = &duplosdk.DestinationDetails{
			CloudWatchLogsDetails: cloudwatch,
		}
	case duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_KINESIS_FIREHOSE:
		firhose := &duplosdk.KinesisFirehoseDetails{
			DeliveryStream: logDelConfig["delivery_stream"].(string),
		}
		duplo.DestinationDetails = &duplosdk.DestinationDetails{
			KinesisFirehoseDetails: firhose,
		}
	default:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unsupported destination type",
			Detail:   "Unsupported destination_type: " + duplo.DestinationType,
		})
	}

	return duplo, diags
}

func validateLogDeliveryConfiguration(m map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	destinationType, ok := m["destination_type"].(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid destination_type",
			Detail:   "destination_type must be a string",
		})
		return diags
	}

	if destinationType == duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_CLOUDWATCH_LOGS {
		if logGroup, ok := m["log_group"].(string); !ok || logGroup == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing log_group",
				Detail:   "log_group must be defined for destination_type=" + duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_CLOUDWATCH_LOGS,
			})
		}
	}

	if destinationType == duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_KINESIS_FIREHOSE {
		if deliveryStream, ok := m["delivery_stream"].(string); !ok || deliveryStream == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing delivery_stream",
				Detail:   "delivery_stream must be defined for destination_type=" + duplosdk.REDIS_LOG_DELIVERYDIST_DEST_TYPE_KINESIS_FIREHOSE,
			})
		}
	}

	return diags
}

// expand Ecache Instance converts resource data respresenting an ECache instance to a Duplo SDK object.
func expandEcacheInstance(d *schema.ResourceData) (*duplosdk.AddDuploEcacheInstanceRequest, diag.Diagnostics) {
	data := &duplosdk.AddDuploEcacheInstanceRequest{
		DuploEcacheInstance: duplosdk.DuploEcacheInstance{
			Name:                     d.Get("name").(string),
			Identifier:               d.Get("identifier").(string),
			Arn:                      d.Get("arn").(string),
			Endpoint:                 d.Get("endpoint").(string),
			CacheType:                d.Get("cache_type").(int),
			EngineVersion:            d.Get("engine_version").(string),
			Size:                     d.Get("size").(string),
			Replicas:                 d.Get("replicas").(int),
			EncryptionAtRest:         d.Get("encryption_at_rest").(bool),
			EncryptionInTransit:      d.Get("encryption_in_transit").(bool),
			AuthToken:                d.Get("auth_token").(string),
			InstanceStatus:           d.Get("instance_status").(string),
			KMSKeyID:                 d.Get("kms_key_id").(string),
			ParameterGroupName:       d.Get("parameter_group_name").(string),
			SnapshotName:             d.Get("snapshot_name").(string),
			SnapshotRetentionLimit:   d.Get("snapshot_retention_limit").(int),
			AutomaticFailoverEnabled: d.Get("automatic_failover_enabled").(bool),
			SnapshotWindow:           d.Get("snapshot_window").(string),
		},
	}
	if ds, ok := d.Get("log_delivery_configuration").(*schema.Set); ok {
		log.Printf("[DEBUG] resourceDuploEcacheInstanceCreate log_delivery_configuration found count: %d", len(ds.List()))
		logDelConfig, diagErr := expandLogDeliveryConfigurations(ds.List())
		if diagErr != nil {
			return nil, diagErr
		}
		data.LogDeliveryConfigurations = &logDelConfig
	}

	for _, val := range d.Get("snapshot_arns").([]interface{}) {
		data.SnapshotArns = append(data.SnapshotArns, val.(string))
	}
	if data.CacheType == 0 {
		if v, ok := d.GetOk("enable_cluster_mode"); ok { //applicable for only redis type
			data.EnableClusterMode = v.(bool)
		}
	}
	if data.EnableClusterMode {
		if v, ok := d.GetOk("number_of_shards"); !ok || (v.(int) < 1 && v.(int) > 500) {
			data.NumberOfShards = 2
		} else {
			data.NumberOfShards = v.(int) //number of shards accepted if cluster mode is enabled
		}
	}
	return data, nil
}

// flattenEcacheInstance converts a Duplo SDK object respresenting an ECache instance to terraform resource data.
func flattenEcacheInstance(duplo *duplosdk.DuploEcacheInstance, d *schema.ResourceData) diag.Diagnostics {
	d.Set("tenant_id", duplo.TenantID)
	d.Set("name", duplo.Name)
	d.Set("identifier", duplo.Identifier)
	d.Set("arn", duplo.Arn)
	d.Set("endpoint", duplo.Endpoint)
	if duplo.Endpoint != "" {
		uriParts := strings.SplitN(duplo.Endpoint, ":", 2)
		d.Set("host", uriParts[0])
		if len(uriParts) == 2 {
			port, _ := strconv.Atoi(uriParts[1])
			d.Set("port", port)
		}
	}
	d.Set("cache_type", duplo.CacheType)
	v, err := normalizeEngineVersion(duplo.EngineVersion)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.Set("engine_version", v.String())
	d.Set("size", duplo.Size)
	d.Set("replicas", duplo.Replicas)
	d.Set("encryption_at_rest", duplo.EncryptionAtRest)
	d.Set("encryption_in_transit", duplo.EncryptionInTransit)
	d.Set("auth_token", duplo.AuthToken)
	d.Set("instance_status", duplo.InstanceStatus)
	d.Set("kms_key_id", duplo.KMSKeyID)
	d.Set("parameter_group_name", duplo.ParameterGroupName)
	d.Set("enable_cluster_mode", duplo.EnableClusterMode)
	d.Set("number_of_shards", duplo.NumberOfShards)
	d.Set("snapshot_name", duplo.SnapshotName)
	d.Set("snapshot_arns", duplo.SnapshotArns)
	d.Set("snapshot_retention_limit", duplo.SnapshotRetentionLimit)
	d.Set("snapshot_window", duplo.SnapshotWindow)
	d.Set("automatic_failover_enabled", duplo.AutomaticFailoverEnabled)
	return nil
}

// ecacheInstanceWaitUntilAvailable waits until an ECache instance is available.
//
// It should be usable both post-creation and post-modification.
func ecacheInstanceWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"processing", "creating", "modifying", "rebooting cluster nodes", "snapshotting"},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.EcacheInstanceGet(tenantID, name)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] EcacheInstanceWaitUntilAvailable (%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func parseECacheInstanceIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func suppressNoOfShardsDiff(k, old, new string, d *schema.ResourceData) bool {
	newValue, err := strconv.Atoi(new)
	if err != nil {
		return false // Unexpected new value type
	}
	if newValue == 0 {
		return true
	}

	oldValue, err := strconv.Atoi(old)
	if err != nil {
		return false // Unexpected old value type
	}

	return newValue == oldValue // Suppress diff if between 1 and 500 (inclusive)
}

func suppressEnginePatchVersion(k, old, new string, d *schema.ResourceData) bool {
	oldVer := removePatchVersion(old)
	newVer := removePatchVersion(new)
	return oldVer == newVer // Suppress diff if patch exist
}

func removePatchVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[:2], ".")
	}
	return version
}
func isValidSnapshotWindow() schema.SchemaValidateDiagFunc {

	return func(i interface{}, path cty.Path) diag.Diagnostics {
		if i == nil {
			return nil
		}

		var diags diag.Diagnostics
		v, ok := i.(string)
		if !ok {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid input type for 'snapshot_window'. Expected a string.",
				Detail:   "The 'snapshot_window' value must be a string formatted as 'HH:MM-HH:MM'.",
			})
		}

		if v == "" {
			return nil
		}

		times := strings.Split(v, "-")
		if len(times) != 2 {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid 'snapshot_window' format.",
				Detail:   fmt.Sprintf("The value '%s' must be in the format 'HH:MM-HH:MM'. For example, '05:00-09:00' is valid.", v),
			})
		}

		startTime, err1 := time.Parse("15:04", times[0])
		endTime, err2 := time.Parse("15:04", times[1])
		if err1 != nil || err2 != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid time format in 'snapshot_window'.",
				Detail:   fmt.Sprintf("Both start and end times in '%s' must follow the 'HH:MM' 24-hour format. For example, '05:00-09:00' is valid.", v),
			})
		}

		timeDiff := endTime.Sub(startTime)
		if timeDiff < 0 {
			timeDiff += 24 * time.Hour
		}

		if timeDiff < time.Hour {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Insufficient time window in 'snapshot_window'.",
				Detail:   fmt.Sprintf("The time difference between the start ('%s') and end ('%s') must be at least 1 Hour. For example, '05:00-05:30' is valid, but '05:00-05:20' is not.", times[0], times[1]),
			})
		}

		return nil
	}
}

func validateEcacheParameters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	ecm := diff.Get("enable_cluster_mode").(bool)
	nshard := diff.Get("number_of_shards").(int)
	if !ecm && nshard > 0 {
		return fmt.Errorf("number_of_shards can be set only if cluster mode is enabled")
	}
	eng := diff.Get("cache_type").(int)
	engVer := diff.Get("engine_version").(string)
	diag := validateClusterEngineVersion(eng, engVer)
	if diag != nil {
		return diagsToError(diag)
	}
	return nil
}

/*func normalizeVersion(version string) (string, error) {
	// Validate input format
	re := regexp.MustCompile(`^\d+(\.\d+){0,2}$`)
	if !re.MatchString(version) {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	// Split version into parts
	parts := strings.Split(version, ".")

	if len(parts) > 0 {
		maj, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", nil
		}
		if maj > 5 && len(parts) == 3 { //if major version is greater than 5 convert version into major.minor format
			return parts[0] + "." + parts[1], nil

		}

	}

	return version, nil
}*/

const (
	redisVersionPreV6RegexpPattern  = `^[1-5](\.[[:digit:]]+){2}$`
	redisVersionPostV6RegexpPattern = `^((6)\.x)|([6-9]\.[[:digit:]]+)$`

	redisVersionRegexpPattern = redisVersionPreV6RegexpPattern + "|" + redisVersionPostV6RegexpPattern
)

var (
	redisVersionRegexp       = regexp.MustCompile(redisVersionRegexpPattern)
	redisVersionPostV6Regexp = regexp.MustCompile(redisVersionPostV6RegexpPattern)
)

func validRedisVersionString(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := v.(string)
	if !ok {
		return diag.Errorf("Invalid type: expected a string.")
	}

	if !redisVersionRegexp.MatchString(value) {
		return diag.Errorf(
			"Invalid Redis version: %q is not valid. For Redis v6 or higher, use <major>.<minor>. "+
				"For Redis v5 or lower, use <major>.<minor>.<patch>.", value,
		)
	}

	return diags
}

const (
	versionStringRegexpInternalPattern = `[[:digit:]]+(\.[[:digit:]]+){2}`
	versionStringRegexpPattern         = "^" + versionStringRegexpInternalPattern + "$"
)

var versionStringRegexp = regexp.MustCompile(versionStringRegexpPattern)

/*
	func validMemcachedVersionString(v any, k string) (ws []string, errors []error) {
		value := v.(string)

		if !versionStringRegexp.MatchString(value) {
			errors = append(errors, fmt.Errorf("%s: must be a version string matching <major>.<minor>.<patch>", k))
		}

		return
	}
*/
func validMemcachedVersionString(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	value, ok := v.(string)
	if !ok {
		return diag.Errorf("Invalid Type", "Expected a string value.")
	}

	if !versionStringRegexp.MatchString(value) {
		diag.Errorf("Invalid Memcached Version value  must be in <major>.<minor>.<patch> format.", value)
	}

	return diags
}
func validateClusterEngineVersion(engine int, engineVersion string) diag.Diagnostics {
	// Memcached: Versions in format <major>.<minor>.<patch>
	// Redis: Starting with version 6, must match <major>.<minor>, prior to version 6, <major>.<minor>.<patch>
	// Valkey: Versions in format <major>.<minor>
	var validator schema.SchemaValidateDiagFunc

	switch engine {
	case 1:
		validator = validMemcachedVersionString
	case 0:
		validator = validRedisVersionString
	}

	diags := validator(engineVersion, cty.Path{
		cty.GetAttrStep{Name: "engine_version"},
	})

	return diags
}

// normalizeEngineVersion returns a github.com/hashicorp/go-version Version from:
// - a regular 1.2.3 version number
// - either the 6.x or 6.0 version number used for ElastiCache Redis version 6. 6.x will sort to 6.<maxint>
// - a 7.0 version number used from version 7
func normalizeEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		if matches[1] != "" {
			version = fmt.Sprintf("%s.%d", matches[2], math.MaxInt)
		}
	}
	return gversion.NewVersion(version)
}

func diagsToError(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}
	var errMsgs []string
	for _, d := range diags {
		errMsgs = append(errMsgs, d.Summary)
	}
	return fmt.Errorf("%s", strings.Join(errMsgs, "; "))
}
