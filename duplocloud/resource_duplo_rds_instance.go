package duplocloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// DuploRdsInstanceSchema returns a Terraform resource schema for an RDS instance.
func rdsInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the RDS instance will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the RDS instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 63-MAX_DUPLO_NO_HYPHEN_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]*$`), "Invalid RDS instance name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "RDS instance name cannot end with a hyphen"),
				validation.StringDoesNotMatch(regexp.MustCompile(`--`), "RDS instance name cannot contain two hyphens"),

				// NOTE: some validations are moot, because Duplo provides a prefix and suffix for the name:
				//
				// - First character must be a letter
				//
				// Because Duplo automatically prefixes names, it is impossible to break any of the rules in the above bulleted list.
			),
		},
		"identifier": {
			Description: "The full name of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"endpoint": {
			Description: "The endpoint of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"host": {
			Description: "The DNS hostname of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the RDS instance.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"master_username": {
			Description:  "The master username of the RDS instance.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(1, 128), // NOTE: further restrictions must wait until creation time
		},
		"master_password": {
			Description: "The master password of the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(8, 128), // NOTE: further restrictions must wait until creation time
				validation.StringMatch(regexp.MustCompile(`[ -~]`), "RDS passwords must only include printable ASCII characters"),
				validation.StringDoesNotMatch(regexp.MustCompile(`[/"@]`), "RDS passwords must not include '/', '\"', or '@'"),
			),
		},
		"engine": {
			Description: "The numerical index of database engine to use the for the RDS instance.\n" +
				"Should be one of:\n\n" +
				"   - `0` : MySQL\n" +
				"   - `1` : PostgreSQL\n" +
				"   - `2` : MsftSQL-Express\n" +
				"   - `3` : MsftSQL-Standard\n" +
				"   - `8` : Aurora-MySQL\n" +
				"   - `9` : Aurora-PostgreSQL\n" +
				"   - `10` : MsftSQL-Web\n" +
				"   - `11` : Aurora-Serverless-MySql\n" +
				"   - `12` : Aurora-Serverless-PostgreSql\n" +
				"   - `13` : DocumentDB\n" +
				"   - `14` : MariaDB\n" +
				"   - `16` : Aurora\n",
			Type:         schema.TypeInt,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IntInSlice([]int{0, 1, 2, 3, 8, 9, 10, 11, 12, 13, 14, 16}),
		},
		"engine_version": {
			Description: "The database engine version to use the for the RDS instance.\n" +
				"If you don't know the available engine versions for your RDS instance, you can use the [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-engine-versions.html) to retrieve a list.",
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"snapshot_id": {
			Description:   "A database snapshot to initialize the RDS instance from, at launch.",
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"master_username"},
		},
		"db_subnet_group_name": {
			Description: "Name of DB subnet group. DB instance will be created in the VPC associated with the DB subnet group.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"parameter_group_name": {
			Description: "A RDS parameter group name to apply to the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9-]*$`), "Invalid DB parameter group name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "DB parameter group name cannot end with a hyphen"),
				validation.StringDoesNotMatch(regexp.MustCompile(`--`), "DB parameter group name cannot contain two hyphens"),
			),
		},
		"store_details_in_secret_manager": {
			Description: "Whether or not to store RDS details in the AWS secrets manager.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"size": {
			Description: "The instance type of the RDS instance.\n" +
				"See AWS documentation for the [available instance types](https://aws.amazon.com/rds/instance-types/).",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^db\.`), "RDS instance types must start with 'db.'"),
		},
		"storage_type": {
			Description: "Valid values: gp2 | gp3 | io1 | standard. Storage type to be used for RDS instance storage.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateFunc: validation.StringInSlice(
				[]string{"gp2", "gp3", "io1", "standard"},
				false,
			),
		},
		"iops": {
			Description: "The IOPS (Input/Output Operations Per Second) value. Should be specified only if `storage_type` is either io1 or gp3.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"allocated_storage": {
			Description: "(Required unless a `snapshot_id` is provided) The allocated storage in gigabytes.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"encrypt_storage": {
			Description: "Whether or not to encrypt the RDS instance storage.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"enable_logging": {
			Description: "Whether or not to enable the RDS instance logging. This setting is not applicable for document db cluster instance.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"multi_az": {
			Description: "Specifies if the RDS instance is multi-AZ.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"instance_status": {
			Description: "The current status of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"deletion_protection": {
			Description: "If the DB instance should have deletion protection enabled." +
				"The database can't be deleted when this value is set to `true`. This setting is not applicable for document db cluster instance.",
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"kms_key_id": {
			Description: "The globally unique identifier for the key.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"cluster_identifier": {
			Description: "The RDS Cluster Identifier",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"v2_scaling_configuration": {
			Description: "Serverless v2_scaling_configuration min and max scalling capacity.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"min_capacity": {
						Description: "Specifies min scalling capacity.",
						Type:        schema.TypeFloat,
						Required:    true,
					},
					"max_capacity": {
						Description: "Specifies max scalling capacity.",
						Type:        schema.TypeFloat,
						Required:    true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceDuploRdsInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_rds_instance` manages an AWS RDS instance in Duplo.",

		ReadContext:   resourceDuploRdsInstanceRead,
		CreateContext: resourceDuploRdsInstanceCreate,
		UpdateContext: resourceDuploRdsInstanceUpdate,
		DeleteContext: resourceDuploRdsInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: rdsInstanceSchema(),
	}
}

// READ resource
func resourceDuploRdsInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.RdsInstanceGet(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}

	// Convert the object into Terraform resource data
	jo := rdsInstanceToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", duplo.TenantID, duplo.Name))

	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** end")
	return nil
}

// CREATE resource
func resourceDuploRdsInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duplo, err := rdsInstanceFromState(d)
	if err != nil {
		return diag.Errorf("Internal error: %s", err)
	}

	// Populate the identifier field, and determine some other fields
	duplo.Identifier = duplo.Name
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, duplo.Name)

	// Validate the RDS instance.
	errors := validateRdsInstance(duplo)
	if len(errors) > 0 {
		return errorsToDiagnostics(fmt.Sprintf("Cannot create RDS DB instance: %s: ", id), errors)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.RdsInstanceCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error creating RDS DB instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS DB instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = rdsInstanceWaitUntilAvailable(ctx, c, id, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

	diags = resourceDuploRdsInstanceRead(ctx, d, m)

	if isDeleteProtectionSupported(d) {
		identifier := d.Get("identifier").(string)
		deleteProtection := new(bool)
		*deleteProtection = d.Get("deletion_protection").(bool)
		// Update delete protection settings.
		log.Printf("[DEBUG] Updating delete protection settings to '%t' for db instance '%s'.", d.Get("deletion_protection").(bool), identifier)
		if isAuroraDB(d) {
			err = c.RdsClusterChangeDeleteProtection(tenantID, duplosdk.DuploRdsClusterDeleteProtection{
				DBClusterIdentifier: identifier + "-cluster",
				DeletionProtection:  deleteProtection,
				ApplyImmediately:    true,
			})
		} else {
			err = c.RdsInstanceChangeDeleteProtection(tenantID, duplosdk.DuploRdsInstanceDeleteProtection{
				DBInstanceIdentifier: identifier,
				DeletionProtection:   deleteProtection,
			})
		}

		if err != nil {
			return diag.Errorf("Error while setting deletion_protection for RDS DB instance '%s' : %s", id, err)
		}
	}

	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** end")
	return diags
}

// UPDATE resource
func resourceDuploRdsInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := d.Id()

	size := d.Get("size").(string)
	if d.HasChange("v2_scaling_configuration") && size == "db.serverless" {
		// Request Aurora serverless V2 instance-size change
		if v, ok := d.GetOk("v2_scaling_configuration"); ok {
			log.Printf("[TRACE] DuploRdsModifyAuroraV2ServerlessInstanceSize ******** start")
			err = c.RdsModifyAuroraV2ServerlessInstanceSize(tenantID, duplosdk.DuploRdsModifyAuroraV2ServerlessInstanceSize{
				Identifier:             d.Get("identifier").(string),
				ClusterIdentifier:      d.Get("identifier").(string) + "-cluster",
				SizeEx:                 size,
				ApplyImmediately:       true,
				V2ScalingConfiguration: expandV2ScalingConfiguration(v.([]interface{})),
			})
		}

		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for the instance to become available.
		err = rdsInstanceWaitUntilAvailable(ctx, c, id, 7*time.Minute)
		if err != nil {
			return diag.Errorf("Error waiting for RDS DB instance '%s' to be unavailable: %s", id, err)
		}

		// in-case timed out. check one more time .. aurora cluster takes long time to update and backup
		err = rdsInstanceWaitUntilAvailable(ctx, c, id, 3*time.Minute)
		if err != nil {
			return diag.Errorf("Error waiting for RDS DB instance '%s' to be unavailable: %s", id, err)
		}
	}

	// Request the password change in Duplo
	if d.HasChange("master_password") {
		err = c.RdsInstanceChangePassword(tenantID, duplosdk.DuploRdsInstancePasswordChange{
			Identifier:     d.Get("identifier").(string),
			MasterPassword: d.Get("master_password").(string),
			StorePassword:  true,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("enable_logging") {
		identifier := d.Get("identifier").(string)
		enableLogging := new(bool)
		*enableLogging = d.Get("enable_logging").(bool)
		log.Printf("[TRACE] Updating enable_logging to: '%v' for db instance '%s'.", d.Get("enable_logging").(bool), d.Get("identifier").(string))
		err = c.RdsInstanceChangeSizeOrEnableLogging(tenantID, identifier, duplosdk.DuploRdsUpdatePayload{
			EnableLogging: enableLogging,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		identifier := d.Get("identifier").(string)
		size := d.Get("size").(string)
		log.Printf("[TRACE] Updating size to: '%s' for db instance '%s'.", d.Get("size").(string), d.Get("identifier").(string))
		err = c.RdsInstanceChangeSizeOrEnableLogging(tenantID, identifier, duplosdk.DuploRdsUpdatePayload{
			SizeEx: size,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Wait for the instance to become unavailable - but continue on if we timeout, without any errors raised.
	_ = rdsInstanceWaitUntilUnavailable(ctx, c, id, 150*time.Second)

	// Wait for the instance to become available.
	err = rdsInstanceWaitUntilAvailable(ctx, c, id, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

	diags := resourceDuploRdsInstanceRead(ctx, d, m)

	if isDeleteProtectionSupported(d) && d.HasChange("deletion_protection") {
		log.Printf("[DEBUG] Updating delete protection settings to '%t' for db instance '%s'.", d.Get("deletion_protection").(bool), d.Get("identifier").(string))
		deleteProtection := new(bool)
		*deleteProtection = d.Get("deletion_protection").(bool)

		if isAuroraDB(d) {
			err = c.RdsClusterChangeDeleteProtection(tenantID, duplosdk.DuploRdsClusterDeleteProtection{
				DBClusterIdentifier: d.Get("identifier").(string) + "-cluster",
				DeletionProtection:  deleteProtection,
				ApplyImmediately:    true,
			})
		} else {
			err = c.RdsInstanceChangeDeleteProtection(tenantID, duplosdk.DuploRdsInstanceDeleteProtection{
				DBInstanceIdentifier: d.Get("identifier").(string),
				DeletionProtection:   deleteProtection,
			})
		}

		if err != nil {
			return diag.Errorf("Error while setting deletion_protection for RDS DB instance '%s' : %s", id, err)
		}
	}

	log.Printf("[TRACE] resourceDuploRdsInstanceUpdate ******** end")
	return diags
}

// DELETE resource
func resourceDuploRdsInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** start")

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	_, err := c.RdsInstanceDelete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS DB instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})

	// Wait 1 more minute to deal with consistency issues.
	if diags == nil {
		time.Sleep(time.Minute)
	}

	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** end")
	return diags
}

// RdsInstanceWaitUntilAvailable waits until an RDS instance is available.
//
// It should be usable both post-creation and post-modification.
func rdsInstanceWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading", "failed", "submitted",
		},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilAvailable (%s)", id)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// RdsInstanceWaitUntilUnavailable waits until an RDS instance is unavailable.
//
// It should be usable post-modification.
func rdsInstanceWaitUntilUnavailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Target: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
		},
		Pending:      []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "available"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilUnavailable (%s)", id)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// RdsInstanceFromState converts resource data respresenting an RDS instance to a Duplo SDK object.
func rdsInstanceFromState(d *schema.ResourceData) (*duplosdk.DuploRdsInstance, error) {
	duploObject := new(duplosdk.DuploRdsInstance)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("identifier").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Endpoint = d.Get("endpoint").(string)
	duploObject.MasterUsername = d.Get("master_username").(string)
	duploObject.MasterPassword = d.Get("master_password").(string)
	duploObject.Engine = d.Get("engine").(int)
	duploObject.EngineVersion = d.Get("engine_version").(string)
	duploObject.SnapshotID = d.Get("snapshot_id").(string)
	duploObject.DBParameterGroupName = d.Get("parameter_group_name").(string)
	duploObject.DBSubnetGroupName = d.Get("db_subnet_group_name").(string)
	duploObject.Cloud = 0 // AWS
	duploObject.SizeEx = d.Get("size").(string)
	duploObject.EncryptStorage = d.Get("encrypt_storage").(bool)
	duploObject.StorageType = d.Get("storage_type").(string)
	duploObject.Iops = d.Get("iops").(int)
	duploObject.AllocatedStorage = d.Get("allocated_storage").(int)
	duploObject.EncryptionKmsKeyId = d.Get("kms_key_id").(string)
	duploObject.EnableLogging = d.Get("enable_logging").(bool)
	duploObject.MultiAZ = d.Get("multi_az").(bool)
	duploObject.InstanceStatus = d.Get("instance_status").(string)
	if v, ok := d.GetOk("v2_scaling_configuration"); ok {
		duploObject.V2ScalingConfiguration = expandV2ScalingConfiguration(v.([]interface{}))
	}
	if duploObject.SizeEx == "db.serverless" && duploObject.V2ScalingConfiguration == nil {
		return nil, errors.New("v2_scaling_configuration: min_capacity and max_capacity must be provided")
	}

	return duploObject, nil
}

func expandV2ScalingConfiguration(cfg []interface{}) *duplosdk.V2ScalingConfiguration {
	if len(cfg) < 1 {
		return nil
	}
	out := &duplosdk.V2ScalingConfiguration{}
	m := cfg[0].(map[string]interface{})
	if v, ok := m["min_capacity"]; ok {
		out.MinCapacity = v.(float64)
	}
	if v, ok := m["max_capacity"]; ok {
		out.MaxCapacity = v.(float64)
	}
	if out.MinCapacity == 0 || out.MaxCapacity == 0 {
		return nil
	}
	return out
}

// RdsInstanceToState converts a Duplo SDK object respresenting an RDS instance to terraform resource data.
func rdsInstanceToState(duploObject *duplosdk.DuploRdsInstance, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["name"] = duploObject.Name
	jo["identifier"] = duploObject.Identifier
	jo["arn"] = duploObject.Arn
	jo["endpoint"] = duploObject.Endpoint
	clusterIdentifier := duploObject.ClusterIdentifier
	if len(clusterIdentifier) == 0 {
		clusterIdentifier = duploObject.Identifier
	}
	jo["cluster_identifier"] = clusterIdentifier

	if duploObject.Endpoint != "" {
		uriParts := strings.SplitN(duploObject.Endpoint, ":", 2)
		jo["host"] = uriParts[0]
		if len(uriParts) == 2 {
			jo["port"], _ = strconv.Atoi(uriParts[1])
		}
	}
	jo["master_username"] = duploObject.MasterUsername
	jo["master_password"] = duploObject.MasterPassword
	jo["engine"] = duploObject.Engine
	jo["engine_version"] = duploObject.EngineVersion
	jo["snapshot_id"] = duploObject.SnapshotID
	jo["parameter_group_name"] = duploObject.DBParameterGroupName
	jo["db_subnet_group_name"] = duploObject.DBSubnetGroupName
	jo["size"] = duploObject.SizeEx
	jo["encrypt_storage"] = duploObject.EncryptStorage
	jo["storage_type"] = duploObject.StorageType
	jo["iops"] = duploObject.Iops
	jo["allocated_storage"] = duploObject.AllocatedStorage
	jo["kms_key_id"] = duploObject.EncryptionKmsKeyId
	jo["enable_logging"] = duploObject.EnableLogging
	jo["multi_az"] = duploObject.MultiAZ
	jo["instance_status"] = duploObject.InstanceStatus
	if duploObject.V2ScalingConfiguration != nil && duploObject.V2ScalingConfiguration.MinCapacity != 0 {
		d.Set("v2_scaling_configuration", []map[string]interface{}{{
			"min_capacity": duploObject.V2ScalingConfiguration.MinCapacity,
			"max_capacity": duploObject.V2ScalingConfiguration.MaxCapacity,
		}})
	}

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}

func validateRdsInstance(duplo *duplosdk.DuploRdsInstance) (errors []error) {
	if duplo.Engine == duplosdk.DUPLO_RDS_ENGINE_POSTGRESQL {

		// PostgreSQL requires shorter usernames.
		if duplo.MasterUsername != "" {
			length := utf8.RuneCountInString(duplo.MasterUsername)
			if length < 1 || length > 63 {
				errors = append(errors, fmt.Errorf("PostgreSQL requires the 'master_username' to be between 1 and 63 characters"))
			}
		}

	} else if duplo.Engine == duplosdk.DUPLO_RDS_ENGINE_MYSQL {

		// MySQL requires shorter usernames and passwords.
		if duplo.MasterUsername != "" {
			length := utf8.RuneCountInString(duplo.MasterUsername)
			if length < 1 || length > 16 {
				errors = append(errors, fmt.Errorf("MySQL requires the 'master_username' to be between 1 and 16 characters"))
			}
		}
		if duplo.MasterPassword != "" {
			length := utf8.RuneCountInString(duplo.MasterPassword)
			if length < 1 || length > 41 {
				errors = append(errors, fmt.Errorf("MySQL requires the 'master_password' to be between 1 and 41 characters"))
			}
		}
	}

	// Multiple databases require the first username character to be a letter.
	if duplosdk.RdsIsMsSQL(duplo.Engine) ||
		duplo.Engine == duplosdk.DUPLO_RDS_ENGINE_MYSQL ||
		duplo.Engine == duplosdk.DUPLO_RDS_ENGINE_POSTGRESQL {
		if duplo.MasterUsername != "" {
			first, _ := utf8.DecodeRuneInString(duplo.MasterUsername)
			if !unicode.IsLetter(first) {
				errors = append(errors, fmt.Errorf("first character of 'master_password' must be a letter for this RDS engine"))
			}
		}
	}

	return
}

func isAuroraDB(d *schema.ResourceData) bool {
	return d.Get("engine").(int) == 8 || d.Get("engine").(int) == 9 ||
		d.Get("engine").(int) == 11 || d.Get("engine").(int) == 12 || d.Get("engine").(int) == 16
}

func isDeleteProtectionSupported(d *schema.ResourceData) bool {
	// Avoid setting delete protection for document DB
	return d.Get("engine").(int) != 13
}
