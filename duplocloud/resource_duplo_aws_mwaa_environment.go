package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func environmentModuleLoggingConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// "cloud_watch_log_group_arn": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"log_level": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"CRITICAL", "DEBUG", "ERROR", "INFO", "WARNING",
				}, false),
			},
		},
	}
}

func duploMwaaAirflowSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Managed Workflows Apache Airflow will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Apache Airflow Environment.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name provided by duplo for Apache Airflow Environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"max_workers": {
			Description:  "The maximum number of workers that can be automatically scaled up. Value need to be between `1` and `25`.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},
		"min_workers": {
			Description:  "The minimum number of workers that you want to run in your environment.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntAtLeast(1),
		},
		"schedulers": {
			Description: "The number of schedulers that you want to run in your environment.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"airflow_version": {
			Description: "Airflow version of your environment, will be set by default to the latest version that MWAA supports.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"2.4.3", "2.5.1", "2.6.3", "2.7.2", "2.8.1", "2.9.2", "2.10.1", "2.10.3",
			}, false),
		},
		"environment_class": {
			Description: "Environment class for the cluster. Possible options are `mw1.small`, `mw1.medium`, `mw1.large`, `mw1.xlarge`, `mw1.2xlarge`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"mw1.small", "mw1.medium", "mw1.large", "mw1.xlarge", "mw1.2xlarge",
			}, false),
		},
		"source_bucket_arn": {
			Description: "The Amazon Resource Name (ARN) of your Amazon S3 storage bucket. For example, arn:aws:s3:::airflow-mybucketname.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"dag_s3_path": {
			Description: "The relative path to the DAG folder on your Amazon S3 storage bucket.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"weekly_maintenance_window_start": {
			Description: "Specifies the start date for the weekly maintenance window.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"plugins_s3_path": {
			Description: "The relative path to the plugins.zip file on your Amazon S3 storage bucket. For example, plugins.zip. If a relative path is provided in the request, then `plugins_s3_object_version` is required.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"plugins_s3_object_version": {
			Description: "The plugins.zip file version you want to use. If not set, latest s3 file version will be used.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"requirements_s3_path": {
			Description: "The relative path to the requirements.txt file on your Amazon S3 storage bucket. For example, requirements.txt. If a relative path is provided in the request, then requirements_s3_object_version is required.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"requirements_s3_object_version": {
			Description: "The requirements.txt file version you want to use. If not set, latest s3 file version will be used.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"startup_script_s3_path": {
			Description: "The relative path to the startup script file on your Amazon S3 storage bucket. For example, startup_script.sh.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"startup_script_s3_object_version": {
			Description: "The startup script file version you want to use. If not set, latest s3 file version will be used.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"kms_key": {
			Description: "The Amazon Resource Name (ARN) of your KMS key that you want to use for encryption. Will be set to the ARN of the managed KMS key aws/airflow by default.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"webserver_access_mode": {
			Description: "Specifies whether the webserver should be accessible over the internet or via your specified VPC. ",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     "PUBLIC_ONLY",
			ValidateFunc: validation.StringInSlice([]string{
				"PRIVATE_ONLY", "PUBLIC_ONLY",
			}, false),
		},
		"airflow_configuration_options": {
			Description: "The `airflow_configuration_options` parameter specifies airflow override options",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"logging_configuration": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"dag_processing_logs": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem:     environmentModuleLoggingConfigurationSchema(),
					},
					"scheduler_logs": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem:     environmentModuleLoggingConfigurationSchema(),
					},
					"task_logs": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem:     environmentModuleLoggingConfigurationSchema(),
					},
					"webserver_logs": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem:     environmentModuleLoggingConfigurationSchema(),
					},
					"worker_logs": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem:     environmentModuleLoggingConfigurationSchema(),
					},
				},
			},
		},
		"wait_until_ready": {
			Description:      "Whether or not to wait until Amazon MWAA Environment to be ready, after creation.",
			Type:             schema.TypeBool,
			Optional:         true,
			Default:          true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"arn": {
			Description: "The ARN of the Managed Workflows Apache Airflow.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The status of the Amazon MWAA Environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"webserver_url": {
			Description: "The webserver URL of the MWAA Environment.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Description: "Tags.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"execution_role_arn": {
			Description: "The Execution Role ARN of the Amazon MWAA Environment",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"last_updated": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"created_at": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"error": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"error_code": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"error_message": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
					"status": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func resourceMwaaAirflow() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_aws_mwaa_environment` manages an AWS MWAA Environment resource in Duplo.",
		ReadContext:   resourceMwaaAirflowRead,
		CreateContext: resourceMwaaAirflowCreate,
		UpdateContext: resourceMwaaAirflowUpdate,
		DeleteContext: resourceMwaaAirflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: duploMwaaAirflowSchema(),
	}
}

func resourceMwaaAirflowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullname, err := parseMwaaAirflowIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceMwaaAirflowRead(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MwaaAirflowGet(tenantID, fullname)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s  mwaa airflow detail %s : %s", tenantID, fullname, clientErr)
	}
	duploDetails, err := c.MwaaAirflowDetailsGet(tenantID, duplo.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenMwaaAirflow(d, duploDetails)
	if err != nil {
		return diag.FromErr(err)
	}
	prefix, err := c.GetResourcePrefix("duploservices", tenantID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("name", strings.SplitN(fullname, prefix+"-", 2)[1])
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceMwaaAirflowRead(%s, %s): end", tenantID, duplo.Name)
	return nil
}

func resourceMwaaAirflowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceMwaaAirflowCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullname, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.Errorf("Error getting duplo service prefix tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}
	fullname += "-" + name
	rq, err := expandMwaaAirflow(d)
	if err != nil {
		return diag.Errorf("Error expanding tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}
	err = c.MwaaAirflowCreate(tenantID, name, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullname)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "mwaa airflow detail", id, func() (interface{}, duplosdk.ClientError) {
		return c.MwaaAirflowGet(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mwaaWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceMwaaAirflowRead(ctx, d, m)
	log.Printf("[TRACE] resourceMwaaAirflowCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceMwaaAirflowUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	fullname := d.Get("fullname").(string)
	log.Printf("[TRACE] resourceMwaaAirflowUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	updated := false
	input := duplosdk.DuploMwaaAirflowCreateRequest{
		Name: fullname,
	}
	if d.HasChange("airflow_configuration_options") {
		updated = true
		options, ok := d.GetOk("airflow_configuration_options")
		if !ok {
			options = map[string]interface{}{}
		}

		input.AirflowConfigurationOptions = expandAirflowConfigurationOptions(options.(map[string]interface{}))
	}

	if d.HasChange("airflow_version") {
		updated = true
		input.AirflowVersion = d.Get("airflow_version").(string)
	}

	if d.HasChange("dag_s3_path") {
		updated = true
		input.DagS3Path = d.Get("dag_s3_path").(string)
	}

	if d.HasChange("environment_class") {
		updated = true
		input.EnvironmentClass = d.Get("environment_class").(string)
	}

	if d.HasChange("logging_configuration") {
		updated = true
		input.LoggingConfiguration = expandEnvironmentLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
	}

	if d.HasChange("max_workers") {
		updated = true
		input.MaxWorkers = d.Get("max_workers").(int)
	}

	if d.HasChange("min_workers") {
		updated = true
		input.MinWorkers = d.Get("min_workers").(int)
	}

	if d.HasChange("plugins_s3_object_version") {
		updated = true
		input.PluginsS3ObjectVersion = d.Get("plugins_s3_object_version").(string)
	}

	if d.HasChange("plugins_s3_path") {
		updated = true
		input.PluginsS3Path = d.Get("plugins_s3_path").(string)
	}

	if d.HasChange("requirements_s3_object_version") {
		updated = true
		input.RequirementsS3ObjectVersion = d.Get("requirements_s3_object_version").(string)
	}

	if d.HasChange("requirements_s3_path") {
		updated = true
		input.RequirementsS3Path = d.Get("requirements_s3_path").(string)
	}

	if d.HasChange("startup_script_s3_object_version") {
		updated = true
		input.StartupScriptS3ObjectVersion = d.Get("startup_script_s3_object_version").(string)
	}

	if d.HasChange("startup_script_s3_path") {
		updated = true
		input.StartupScriptS3Path = d.Get("startup_script_s3_path").(string)
	}

	if d.HasChange("schedulers") {
		updated = true
		input.Schedulers = d.Get("schedulers").(int)
	}

	if d.HasChange("source_bucket_arn") {
		updated = true
		input.SourceBucketArn = d.Get("source_bucket_arn").(string)
	}

	if d.HasChange("webserver_access_mode") {
		updated = true
		input.WebserverAccessMode = &duplosdk.DuploStringValue{
			Value: d.Get("webserver_access_mode").(string),
		}
	}

	if d.HasChange("weekly_maintenance_window_start") {
		updated = true
		input.WeeklyMaintenanceWindowStart = d.Get("weekly_maintenance_window_start").(string)
	}

	if !updated {
		return nil
	}
	err = c.MwaaAirflowUpdate(tenantID, fullname, &input)
	if err != nil {
		return diag.Errorf("Error updating tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = mwaaWaitUntilReady(ctx, c, tenantID, fullname, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := resourceMwaaAirflowRead(ctx, d, m)
	log.Printf("[TRACE] resourceMwaaAirflowUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceMwaaAirflowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	fullname := d.Get("fullname").(string)
	tenantID, name, err := parseMwaaAirflowIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceMwaaAirflowDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.MwaaAirflowDelete(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s mwaa airflow detail '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "mwaa airflow detail", id, func() (interface{}, duplosdk.ClientError) {
		return c.MwaaAirflowGet(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceMwaaAirflowDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandMwaaAirflow(d *schema.ResourceData) (*duplosdk.DuploMwaaAirflowCreateRequest, error) {
	request := duplosdk.DuploMwaaAirflowCreateRequest{}
	if name, ok := d.GetOk("name"); ok {
		request.Name = name.(string)
	}
	if airflowVersion, ok := d.GetOk("airflow_version"); ok {
		request.AirflowVersion = airflowVersion.(string)
	}
	if environmentClass, ok := d.GetOk("environment_class"); ok {
		request.EnvironmentClass = environmentClass.(string)
	}
	if kmsKey, ok := d.GetOk("kms_key"); ok {
		request.KmsKey = kmsKey.(string)
	}
	if sourceBucketArn, ok := d.GetOk("source_bucket_arn"); ok {
		request.SourceBucketArn = sourceBucketArn.(string)
	}
	if dagS3Path, ok := d.GetOk("dag_s3_path"); ok {
		request.DagS3Path = dagS3Path.(string)
	}
	if maxWorkers, ok := d.GetOk("max_workers"); ok {
		request.MaxWorkers = maxWorkers.(int)
	}
	if minWorkers, ok := d.GetOk("min_workers"); ok {
		request.MinWorkers = minWorkers.(int)
	}
	if schedulers, ok := d.GetOk("schedulers"); ok {
		request.Schedulers = schedulers.(int)
	}
	if executionRoleArn, ok := d.GetOk("execution_role_arn"); ok {
		request.ExecutionRoleArn = executionRoleArn.(string)
	}
	if pluginsS3ObjectVersion, ok := d.GetOk("plugins_s3_object_version"); ok {
		request.PluginsS3ObjectVersion = pluginsS3ObjectVersion.(string)
	}
	if pluginsS3Path, ok := d.GetOk("plugins_s3_path"); ok {
		request.PluginsS3Path = pluginsS3Path.(string)
	}
	if requirementsS3ObjectVersion, ok := d.GetOk("requirements_s3_object_version"); ok {
		request.RequirementsS3ObjectVersion = requirementsS3ObjectVersion.(string)
	}
	if requirementsS3Path, ok := d.GetOk("requirements_s3_path"); ok {
		request.RequirementsS3Path = requirementsS3Path.(string)
	}
	if startupScriptS3ObjectVersion, ok := d.GetOk("startup_script_s3_object_version"); ok {
		request.StartupScriptS3ObjectVersion = startupScriptS3ObjectVersion.(string)
	}
	if startupScriptS3Path, ok := d.GetOk("startup_script_s3_path"); ok {
		request.StartupScriptS3Path = startupScriptS3Path.(string)
	}
	if webserverAccessMode, ok := d.GetOk("webserver_access_mode"); ok {
		request.WebserverAccessMode = &duplosdk.DuploStringValue{
			Value: webserverAccessMode.(string),
		}
	}
	if weeklyMaintenanceWindowStart, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		request.WeeklyMaintenanceWindowStart = weeklyMaintenanceWindowStart.(string)
	}

	if v, ok := d.GetOk("logging_configuration"); ok {
		request.LoggingConfiguration = expandEnvironmentLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("airflow_configuration_options"); ok {
		request.AirflowConfigurationOptions = expandAirflowConfigurationOptions(v.(map[string]interface{}))
	}

	return &request, nil
}

func parseMwaaAirflowIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenMwaaAirflow(d *schema.ResourceData, duplo *duplosdk.DuploMwaaAirflowDetail) error {
	d.Set("fullname", duplo.Name)
	d.Set("mwaa_airflow_id", duplo.Name)
	d.Set("created_at", duplo.CreatedAt)
	d.Set("airflow_version", duplo.AirflowVersion)
	d.Set("arn", duplo.Arn)
	d.Set("environment_class", duplo.EnvironmentClass)
	d.Set("execution_role_arn", duplo.ExecutionRoleArn)
	d.Set("kms_key", duplo.KmsKey)
	d.Set("source_bucket_arn", duplo.SourceBucketArn)
	d.Set("dag_s3_path", duplo.DagS3Path)
	d.Set("max_workers", duplo.MaxWorkers)
	d.Set("min_workers", duplo.MinWorkers)
	d.Set("schedulers", duplo.Schedulers)
	d.Set("service_role_arn", duplo.ServiceRoleArn)
	d.Set("webserver_url", duplo.WebserverUrl)
	d.Set("weekly_maintenance_window_start", duplo.WeeklyMaintenanceWindowStart)
	d.Set("status", duplo.Status.Value)
	d.Set("tags", duplo.Tags)
	d.Set("webserver_access_mode", duplo.WebserverAccessMode.Value)
	d.Set("plugins_s3_path", duplo.PluginsS3Path)
	d.Set("plugins_s3_object_version", duplo.PluginsS3ObjectVersion)
	d.Set("requirements_s3_path", duplo.RequirementsS3Path)
	d.Set("requirements_s3_object_version", duplo.RequirementsS3ObjectVersion)
	d.Set("startup_script_s3_path", duplo.StartupScriptS3Path)
	d.Set("startup_script_s3_object_version", duplo.StartupScriptS3ObjectVersion)

	if err := d.Set("last_updated", flattenLastUpdate(duplo.LastUpdate)); err != nil {
		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}
	if err := d.Set("logging_configuration", flattenLoggingConfiguration(duplo.LoggingConfiguration)); err != nil {
		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}

	d.Set("airflow_configuration_options", duplo.AirflowConfigurationOptions)

	return nil
}

func flattenLastUpdate(lastUpdate *duplosdk.DuploMwaaLastUpdate) []interface{} {
	if lastUpdate == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if len(lastUpdate.CreatedAt) > 0 {
		m["created_at"] = lastUpdate.CreatedAt
	}

	if lastUpdate.Error != nil {
		m["error"] = flattenLastUpdateError(lastUpdate.Error)
	}

	if lastUpdate.Status != nil {
		m["status"] = lastUpdate.Status.Value
	}

	return []interface{}{m}
}

func flattenLastUpdateError(error *duplosdk.DuploMwaaLastUpdateError) []interface{} {
	if error == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if len(error.ErrorCode) > 0 {
		m["error_code"] = error.ErrorCode
	}

	if len(error.ErrorMessage) > 0 {
		m["error_message"] = error.ErrorMessage
	}

	return []interface{}{m}
}

func flattenLoggingConfiguration(loggingConfiguration *duplosdk.DuploMwaaLoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if loggingConfiguration.DagProcessingLogs != nil {
		m["dag_processing_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.DagProcessingLogs)
	}

	if loggingConfiguration.SchedulerLogs != nil {
		m["scheduler_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.SchedulerLogs)
	}

	if loggingConfiguration.TaskLogs != nil {
		m["task_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.TaskLogs)
	}

	if loggingConfiguration.WebserverLogs != nil {
		m["webserver_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.WebserverLogs)
	}

	if loggingConfiguration.WorkerLogs != nil {
		m["worker_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.WorkerLogs)
	}

	return []interface{}{m}
}

func flattenModuleLoggingConfiguration(moduleLoggingConfiguration *duplosdk.DuploLoggingConfigurationInput) []interface{} {
	if moduleLoggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		// "cloud_watch_log_group_arn": moduleLoggingConfiguration.CloudWatchLogGroupArn,
		"enabled":   moduleLoggingConfiguration.Enabled,
		"log_level": moduleLoggingConfiguration.LogLevel.Value,
	}

	return []interface{}{m}
}

func expandEnvironmentLoggingConfiguration(l []interface{}) *duplosdk.DuploMwaaLoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &duplosdk.DuploMwaaLoggingConfiguration{}

	m := l[0].(map[string]interface{})

	if v, ok := m["dag_processing_logs"]; ok {
		input.DagProcessingLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["scheduler_logs"]; ok {
		input.SchedulerLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["task_logs"]; ok {
		input.TaskLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["webserver_logs"]; ok {
		input.WebserverLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["worker_logs"]; ok {
		input.WorkerLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	return input
}

func expandEnvironmentModuleLoggingConfiguration(l []interface{}) *duplosdk.DuploLoggingConfigurationInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &duplosdk.DuploLoggingConfigurationInput{}
	m := l[0].(map[string]interface{})

	input.Enabled = m["enabled"].(bool)
	input.LogLevel = &duplosdk.DuploStringValue{
		Value: m["log_level"].(string),
	}

	return input
}

func expandAirflowConfigurationOptions(m map[string]interface{}) map[string]string {
	stringMap := make(map[string]string, len(m))
	for k, v := range m {
		stringMap[k] = v.(string)
	}
	return stringMap
}

func mwaaWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, mwaaFullname string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.MwaaAirflowDetailsGet(tenantID, mwaaFullname)
			status := "pending"
			if rp != nil {
				if "AVAILABLE" == rp.Status.Value || "CREATE_FAILED" == rp.Status.Value || "UPDATE_FAILED" == rp.Status.Value {
					status = "ready"
				} else {
					status = "pending"
				}
				log.Printf("[TRACE] Status of the Amazon MWAA Environment (%s).", rp.Status.Value)
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] mwaaWaitUntilReady(%s, %s)", tenantID, mwaaFullname)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
