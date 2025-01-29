package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func rdsReadReplicaSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the RDS read replica will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the RDS read replica.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 63-MAX_DUPLO_NO_HYPHEN_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-z0-9-]*$`), "Invalid RDS read replica name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "RDS read replica name cannot end with a hyphen"),
				validation.StringDoesNotMatch(regexp.MustCompile(`--`), "RDS read replica name cannot contain two hyphens"),
				duplosdk.ValidateRdsNoDoubleDuploPrefix,
				// NOTE: some validations are moot, because Duplo provides a prefix and suffix for the name:
				//
				// - First character must be a letter
				//
				// Because Duplo automatically prefixes names, it is impossible to break any of the rules in the above bulleted list.
			),
		},
		"cluster_identifier": {
			Description: "The full name of the RDS Cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"size": {
			Description: "The type of the RDS read replica.\n" +
				"See AWS documentation for the [available instance types](https://aws.amazon.com/rds/instance-types/)." +
				"Size should be set as db.serverless if read replica instamce is created as serverless",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^db\.`), "RDS read replica types must start with 'db.'"),
		},
		"availability_zone": {
			Description: "The AZ for the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},
		"identifier": {
			Description: "The full name of the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"endpoint": {
			Description: "The endpoint of the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"host": {
			Description: "The DNS hostname of the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the RDS read replica.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"engine": {
			Description: "The numerical index of database engine to be used the for the RDS read replica.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"engine_type": {
			Description: "Engine type required to validate applicable parameter group setting for different instance. Should be referred from writer",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"engine_version": {
			Description: "The database engine version to be used the for the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"encrypt_storage": {
			Description: "Whether or not to encrypt the RDS instance storage.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"enable_logging": {
			Description: "Whether or not to enable the RDS instance logging. This setting is not applicable for document db cluster instance.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"multi_az": {
			Description: "Specifies if the RDS instance is multi-AZ.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"kms_key_id": {
			Description: "The globally unique identifier for the key.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"replica_status": {
			Description: "The current status of the RDS read replica.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"parameter_group_name": {
			Description: "A RDS parameter group name to apply to the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 255),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "DB parameter group name cannot end with a hyphen"),
				validation.StringDoesNotMatch(regexp.MustCompile(`--`), "DB parameter group name cannot contain two hyphens"),
			),
			DiffSuppressFunc: diffIgnoreDefaultParamaterGroupName,
		},
		"cluster_parameter_group_name": {
			Description: "Parameter group associated with this instance's DB Cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"performance_insights": {
			Description:      "Amazon RDS Performance Insights is a database performance tuning and monitoring feature that helps you quickly assess the load on your database, and determine when and where to take action. Perfomance Insights get apply when enable is set to true.",
			Type:             schema.TypeList,
			MaxItems:         1,
			Optional:         true,
			Computed:         true,
			DiffSuppressFunc: suppressIfPerformanceInsightsDisabled,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Description: "Turn on or off Performance Insights",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
					},
					"kms_key_id": {
						Description:      "Specify ARN for the KMS key to encrypt Performance Insights data.",
						Type:             schema.TypeString,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressKmsIfPerformanceInsightsDisabled,
					},
					"retention_period": {
						Description: "Specify retention period in Days. Valid values are 7, 731 (2 years) or a multiple of 31. For Document DB retention period is 7",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     7,
						ValidateFunc: validation.Any(
							validation.IntInSlice([]int{7, 731}),
							validation.IntDivisibleBy(31),
						),
						DiffSuppressFunc: suppressRetentionPeriodIfPerformanceInsightsDisabled,
					},
				},
			},
		},
		"enhanced_monitoring": {
			Description:  "Interval to capture metrics in real time for the operating system (OS) that your Amazon RDS DB instance runs on.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntInSlice([]int{0, 1, 5, 10, 15, 30, 60}),
		},
		"v2_scaling_configuration": {
			Description: "Serverless v2_scaling_configuration min and max scalling capacity. Required during creating a servless read replica.",
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
func resourceDuploRdsReadReplica() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_rds_read_replica` manages an AWS RDS read replica in Duplo.",

		ReadContext:   resourceDuploRdsReadReplicaRead,
		CreateContext: resourceDuploRdsReadReplicaCreate,
		UpdateContext: resourceDuploRdsReadReplicaUpdate,
		DeleteContext: resourceDuploRdsReadReplicaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema:        rdsReadReplicaSchema(),
		CustomizeDiff: customdiff.All(validateRDSParameters, validateRDSReplicaUse),
	}
}

// READ resource
func resourceDuploRdsReadReplicaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsReadReplicaRead ******** start")

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
	jo := rdsReadReplicaToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", duplo.TenantID, duplo.Name))

	log.Printf("[TRACE] resourceDuploRdsReadReplicaRead ******** end")
	return nil
}

// CREATE resource
func resourceDuploRdsReadReplicaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsReadReplicaCreate ******** start")
	tenantID := d.Get("tenant_id").(string)
	// Convert the Terraform resource data into a Duplo object
	duplo, err := rdsReadReplicaFromState(d)
	if err != nil {
		return diag.Errorf("Internal error: %s", err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)

	// Get RDS writer instance
	identifier := d.Get("cluster_identifier").(string)
	idParts := strings.SplitN(identifier, "-cluster", 2)
	name := strings.TrimPrefix(idParts[0], "duplo")
	duploWriterInstance, err := c.RdsInstanceGetByName(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	duplo.Identifier = duplo.Name
	duplo.Engine = duploWriterInstance.Engine
	duplo.Cloud = duploWriterInstance.Cloud
	if strings.HasSuffix(identifier, "-cluster") {
		duplo.ClusterIdentifier = identifier
	} else {
		duplo.ReplicationSourceIdentifier = identifier
	}
	id := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, duplo.Name)

	pI := expandPerformanceInsight(d)

	if pI != nil && duplo.Engine != RDS_DOCUMENT_DB_ENGINE && !validateReplicaPerformanceInsightsConfigurationAuroaDB(duplo.Engine, d) {

		period := pI["retention_period"].(int)
		kmsid := pI["kms_key_id"].(string)
		duplo.EnablePerformanceInsights = pI["enabled"].(bool)
		duplo.PerformanceInsightsRetentionPeriod = period
		duplo.PerformanceInsightsKMSKeyId = kmsid

	}

	// Validate the RDS instance.
	errors := validateRdsInstance(duplo)
	if len(errors) > 0 {
		return errorsToDiagnostics(fmt.Sprintf("Cannot create RDS DB read replica: %s: ", id), errors)
	}
	if duplo.SizeEx == "db.serverless" {
		rq := duplosdk.DuploRdsModifyAuroraV2ServerlessInstanceSize{
			Identifier:        duploWriterInstance.Identifier,
			ClusterIdentifier: duplo.ClusterIdentifier,
			ApplyImmediately:  true,
			//	SizeEx:            "db.serverless",
		}
		if v, ok := d.GetOk("v2_scaling_configuration"); ok {
			rq.V2ScalingConfiguration = expandV2ScalingConfiguration(v.([]interface{}))
		} else {
			return diag.Errorf("v2_scaling_configuration: min_capacity and max_capacity must be provided")

		}
		err := c.ReadReplicaServerlessCreate(tenantID, rq.ClusterIdentifier, rq)
		if err != nil {
			return diag.Errorf("%s", err.Error())
		}
		//wrtId := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, duploWriterInstance.Name)
		//err1 := readReplicaInstanceWaitUntilUnavailable(ctx, c, wrtId, d.Timeout("create"))
		//if err1 != nil {
		//	return diag.Errorf("Error waiting for RDS DB read replica '%s' to be available: %s", id, err)
		//}
		//err1 = rdsReadReplicaWaitUntilAvailable(ctx, c, wrtId, d.Timeout("create"))
		//if err1 != nil {
		//	return diag.Errorf("Error waiting for RDS DB read replica '%s' to be available: %s", id, err)
		//}

	}
	instResp, err := c.RdsInstanceCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error creating RDS DB read replica '%s': %s", id, err)
	}

	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS DB Read Replica", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})
	if diags != nil {
		return diags
	}
	// Wait for the instance to become available.
	err1 := rdsReadReplicaWaitUntilAvailable(ctx, c, id, d.Timeout("create"))
	if err1 != nil {
		return diag.Errorf("Error waiting for RDS DB read replica '%s' to be available: %s", id, err)
	}

	if d.HasChange("enhanced_monitoring") {
		val := d.Get("enhanced_monitoring").(int)
		err1 = c.RdsUpdateMonitoringInterval(tenantID, duplosdk.DuploMonitoringInterval{
			DBInstanceIdentifier: instResp.Identifier,
			ApplyImmediately:     true,
			MonitoringInterval:   val,
		})
		if err1 != nil {
			return diag.FromErr(err)
		}

		err1 = rdsInstanceSyncMonitoringInterval(ctx, c, id, d.Timeout("create"), val)
		if err1 != nil {
			return diag.Errorf("Error waiting for RDS read replica DB instance '%s' to update enhanced monitoring level: %s", id, err.Error())

		}

	}
	//performance insights update for document db specific
	if pI != nil && duplo.Engine == RDS_DOCUMENT_DB_ENGINE {
		obj := enablePerformanceInstanceObject(pI)
		obj.DBInstanceIdentifier = instResp.Identifier
		insightErr := c.UpdateDBInstancePerformanceInsight(tenantID, obj)
		if insightErr != nil {
			return diag.FromErr(insightErr)

		}
		err1 = performanceInsightsWaitUntilEnabled(ctx, c, id)
		if err1 != nil {
			return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
		}

	}

	diags = resourceDuploRdsReadReplicaRead(ctx, d, m)

	log.Printf("[TRACE] resourceDuploRdsReadReplicaCreate ******** end")
	return diags
}

// UPDATE resource
func resourceDuploRdsReadReplicaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := d.Id()
	identifier := d.Get("identifier").(string)
	if d.HasChange("parameter_group_name") {
		req := duplosdk.DuploRdsUpdatePayload{}
		if v, ok := d.GetOk("parameter_group_name"); ok {
			req.DbParameterGroupName = v.(string)
		}
		err := c.RdsInstanceUpdateParameterGroupName(tenantID, identifier, &req)
		if err != nil {
			return diag.FromErr(err)
		}

		err = rdsInstanceWaitUntilAvailable(ctx, c, id, 7*time.Minute)
		if err != nil {
			return diag.Errorf("Error waiting for RDS DB instance '%s' to be unavailable: %s", id, err)
		}
		err = rdsInstanceSyncParameterGroup(ctx, c, id, 20*time.Minute, req.DbParameterGroupName, "DBPARAM")
		if err != nil {
			return diag.Errorf("Error waiting for RDS DB instance '%s' to update db parameter group name: %s", id, err.Error())

		}
	}
	obj := duplosdk.DuploRdsUpdatePerformanceInsights{}
	pI := expandPerformanceInsight(d)
	if pI != nil {
		period := pI["retention_period"].(int)
		kmsid := pI["kms_key_id"].(string)
		enable := duplosdk.PerformanceInsightEnable{
			EnablePerformanceInsights:          pI["enabled"].(bool),
			PerformanceInsightsRetentionPeriod: period,
			PerformanceInsightsKMSKeyId:        kmsid,
		}
		obj.Enable = &enable
	} else {
		disable := duplosdk.PerformanceInsightDisable{
			EnablePerformanceInsights: false,
		}
		obj.Disable = &disable
	}
	obj.DBInstanceIdentifier = identifier
	if !isAuroraDB(d) {
		insightErr := c.UpdateDBInstancePerformanceInsight(tenantID, obj)
		if insightErr != nil {
			return diag.FromErr(insightErr)

		}

	}
	// Wait for the instance to become unavailable - but continue on if we timeout, without any errors raised.
	_ = readReplicaInstanceWaitUntilUnavailable(ctx, c, id, 150*time.Second)

	// Wait for the instance to become available.
	err = rdsReadReplicaWaitUntilAvailable(ctx, c, id, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}
	if d.HasChange("enhanced_monitoring") {
		val := d.Get("enhanced_monitoring").(int)
		err = c.RdsUpdateMonitoringInterval(tenantID, duplosdk.DuploMonitoringInterval{
			DBInstanceIdentifier: identifier,
			ApplyImmediately:     true,
			MonitoringInterval:   val,
		})
	}
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceDuploRdsReadReplicaRead(ctx, d, m)

	log.Printf("[TRACE] resourceDuploRdsReadReplicaUpdate ******** end")
	return diags
}

// DELETE resource
func resourceDuploRdsReadReplicaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsReadReplicaDelete ******** start")

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	_, err := c.RdsInstanceDelete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS DB Read Replica", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})

	// Wait 1 more minute to deal with consistency issues.
	if diags == nil {
		time.Sleep(time.Minute)
	}

	log.Printf("[TRACE] resourceDuploRdsReadReplicaDelete ******** end")
	return diags
}

// RdsInstanceWaitUntilAvailable waits until an RDS instance is available.
//
// It should be usable both post-creation and post-modification.
func rdsReadReplicaWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
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

// ReadReplicaInstanceWaitUntilUnavailable waits until an RDS instance is unavailable.
//
// It should be usable post-modification.
func readReplicaInstanceWaitUntilUnavailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
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

// RdsInstanceFromState converts resource data respresenting an RDS read replica to a Duplo SDK object.
func rdsReadReplicaFromState(d *schema.ResourceData) (*duplosdk.DuploRdsInstance, error) {
	duploObject := new(duplosdk.DuploRdsInstance)
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("name").(string)
	duploObject.SizeEx = d.Get("size").(string)
	duploObject.AvailabilityZone = d.Get("availability_zone").(string)
	duploObject.DBParameterGroupName = d.Get("parameter_group_name").(string)
	return duploObject, nil
}

// RdsInstanceToState converts a Duplo SDK object respresenting an RDS instance to terraform resource data.
func rdsReadReplicaToState(duploObject *duplosdk.DuploRdsInstance, d *schema.ResourceData) map[string]interface{} {
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
	if duploObject.Endpoint != "" {
		uriParts := strings.SplitN(duploObject.Endpoint, ":", 2)
		jo["host"] = uriParts[0]
		if len(uriParts) == 2 {
			jo["port"], _ = strconv.Atoi(uriParts[1])
		}
	}
	jo["engine"] = duploObject.Engine
	jo["engine_version"] = duploObject.EngineVersion
	jo["size"] = duploObject.SizeEx
	jo["availability_zone"] = duploObject.AvailabilityZone
	jo["encrypt_storage"] = duploObject.EncryptStorage
	jo["kms_key_id"] = duploObject.EncryptionKmsKeyId
	jo["enable_logging"] = duploObject.EnableLogging
	jo["multi_az"] = duploObject.MultiAZ
	jo["replica_status"] = duploObject.InstanceStatus
	jo["parameter_group_name"] = duploObject.DBParameterGroupName
	jo["cluster_parameter_group_name"] = duploObject.ClusterParameterGroupName
	jo["enhanced_monitoring"] = duploObject.MonitoringInterval

	clusterIdentifier := duploObject.ClusterIdentifier
	if len(clusterIdentifier) == 0 {
		clusterIdentifier = duploObject.ReplicationSourceIdentifier
	}
	jo["cluster_identifier"] = clusterIdentifier
	pis := []interface{}{}
	pi := make(map[string]interface{})
	pi["enabled"] = duploObject.EnablePerformanceInsights
	pi["retention_period"] = duploObject.PerformanceInsightsRetentionPeriod
	pi["kms_key_id"] = duploObject.PerformanceInsightsKMSKeyId
	pis = append(pis, pi)
	jo["performance_insights"] = pis

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}

func validateReplicaPerformanceInsightsConfigurationAuroaDB(engineCode int, tfSpecification *schema.ResourceData) bool {
	if isEngineAuroraType(engineCode) && hasPerformanceInsightConfigurations(tfSpecification) {
		return true
	}
	return false
}

func isEngineAuroraType(engineCode int) bool {
	engineNameByCode := map[int]string{
		8:  "AuroraMySql",
		9:  "AuroraPostgreSql",
		16: "Aurora",
	}

	value, ok := engineNameByCode[engineCode]

	return ok && strings.HasPrefix(value, "Aurora")
}

func hasPerformanceInsightConfigurations(tfRdsReplicaSpecification *schema.ResourceData) bool {
	configuration := tfRdsReplicaSpecification.Get("performance_insights").([]interface{})
	return len(configuration) > 0
}

func validateRDSReplicaUse(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	engines := map[int]string{
		0:  "MySQL",
		1:  "PostgreSQL",
		2:  "MsftSQL-Express",
		3:  "MsftSQL-Standard",
		8:  "Aurora-MySQL",
		9:  "Aurora-PostgreSQL",
		10: "MsftSQL-Web",
		11: "Aurora-Serverless-MySql",
		12: "Aurora-Serverless-PostgreSql",
		13: "DocumentDB",
		14: "MariaDB",
		16: "Aurora",
	}

	eng := diff.Get("engine_type").(int)

	if eng == 13 && diff.Get("parameter_group_name").(string) != "" {
		return fmt.Errorf("parameter group is not applicable for %s engine read replica", engines[eng])

	}
	if eng == 2 || eng == 3 || eng == 10 {
		return fmt.Errorf("resource duplocloud_read_replica is not applicable for %s engine", engines[eng])
	}

	//	if eng == 0 || eng == 1 || eng == 14 {
	//		diff.SetNewComputed("parameter_group_name")
	//		//if diff.HasChange("parameter_group_name") {
	//		//	return fmt.Errorf("cannot update parameter_group_name for %s engine", engines[eng])
	//		//
	//		//}
	//	}
	return nil
}

func diffIgnoreDefaultParamaterGroupName(k, old, new string, d *schema.ResourceData) bool {
	o, n := d.GetChange("parameter_group_name")
	if (o.(string) == "" && strings.Contains(n.(string), "default")) ||
		(n.(string) == "" && strings.Contains(o.(string), "default")) ||
		(strings.Contains(o.(string), "default") && strings.Contains(n.(string), "default")) {
		return true
	}
	return false
}
