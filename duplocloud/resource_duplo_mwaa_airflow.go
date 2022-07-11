package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploMwaaAirflowSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		"tenant_id": {
			Description:  "The GUID of the tenant that the MWAA Airflow event target will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Name MWAA Airflow.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"max_workers": {
			Description: "Airflow max_workers.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"min_workers": {
			Description: "Airflow min_workers.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"schedulers": {
			Description: "Airflow schedulers.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"airflow_version": {
			Description: "Airflow Version.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"environment_class": {
			Description: "The Environment Class for MWAA Airflow. e.g. mw1.small.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"source_bucket_arn": {
			Description: "Source Bucket Arn for DAG.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"dag_s3_path": {
			Description: "DAG S3 Path of this MWAA Airflow.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"weekly_maintenance_window_start": {
			Description: "Weekly Maintenance Window Start eg SUN:23:30 .",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},

		"plugins_s3_path": {
			Description: "Plugins s3 path.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"plugins_s3_object_version": {
			Description: "Plugins S3 object version.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"requirements_s3_path": {
			Description: "Requirements S3 path.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"requirements_s3_object_version": {
			Description: "Requirements S3 path object version.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},

		"kms_key": {
			Description: "Kms Key for encryption.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"webserver_access_mode": {
			Description: "Webserver Access Mode. PUBLIC_ONLY or PRIVATE_ONLY",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     "PUBLIC_ONLY",
		},
		"airflow_configuration_options": {
			Description: "airflow configuration options",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"logging_configuration": {
			Description:      "Logging Configuration for DagProcessingLogs, SchedulerLogs, WebserverLogs, TaskLogs, WorkerLogs.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForAirflowLogConfiguration,
		},

		// Computed
		"mwaa_airflow_id": {
			Description: "The id of the MWAA Airflow.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The arn of this Mwaa Airflow.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"last_update": {
			Description: "last update status.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "status.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"created_at": {
			Description: "created date.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"webserver_url": {
			Description: "Webserver Url.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Description: "Tags.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"execution_role_arn": {
			Description: "Execution Role Arn.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceMwaaAirflow() *schema.Resource {
	return &schema.Resource{
		Description: "Mwaa Airflow.",

		ReadContext:   resourceMwaaAirflowRead,
		CreateContext: resourceMwaaAirflowCreate,
		UpdateContext: resourceMwaaAirflowUpdate,
		DeleteContext: resourceMwaaAirflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploMwaaAirflowSchema(),
	}
}

func resourceMwaaAirflowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseMwaaAirflowIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceMwaaAirflowRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.MwaaAirflowGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s  mwaa airflow detail %s : %s", tenantID, name, clientErr)
	}
	duploWithNodes, err := c.MwaaAirflowDetailsGet(tenantID, duplo.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	flattenMwaaAirflow(d, duploWithNodes)
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceMwaaAirflowRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceMwaaAirflowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceMwaaAirflowCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq, err := expandMwaaAirflow(d)
	if err != nil {
		return diag.Errorf("Error expanding tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}
	err = c.MwaaAirflowCreate(tenantID, name, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	getResp, err := c.MwaaAirflowGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Error getting tenant %s mwaa airflow detail '%s': %s", tenantID, name, err)
	}
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "mwaa airflow detail", id, func() (interface{}, duplosdk.ClientError) {
		return c.MwaaAirflowGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	err = nodePoolWaitUntilReady(ctx, c, tenantID, getResp.Name, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	diags = resourceMwaaAirflowRead(ctx, d, m)
	log.Printf("[TRACE] resourceMwaaAirflowCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceMwaaAirflowUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceMwaaAirflowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	mwaaAirflowId := d.Get("mwaa_airflow_id").(string)
	tenantID, name, err := parseMwaaAirflowIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceMwaaAirflowDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.MwaaAirflowDelete(tenantID, mwaaAirflowId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s mwaa airflow detail '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "mwaa airflow detail", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.MwaaAirflowExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
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
	if webserverAccessMode, ok := d.GetOk("webserver_access_mode"); ok {
		request.WebserverAccessMode = webserverAccessMode.(string)
	}
	if weeklyMaintenanceWindowStart, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		request.WeeklyMaintenanceWindowStart = weeklyMaintenanceWindowStart.(string)
	}
	if loggingConfiguration, ok := d.GetOk("logging_configuration"); ok {
		if loggingConfiguration != nil {
			var o1 map[string]interface{}
			err := json.Unmarshal([]byte(loggingConfiguration.(string)), o1)
			if err != nil {
				request.LoggingConfiguration = &o1
			}
		}
	}
	if airflowConfigurationOptions, ok := d.GetOk("airflow_configuration_options"); ok {
		if airflowConfigurationOptions != nil {
			var o1 map[string]interface{}
			err := json.Unmarshal([]byte(airflowConfigurationOptions.(string)), o1)
			if err != nil {
				request.AirflowConfigurationOptions = &o1
			}
		}
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

func flattenMwaaAirflow(d *schema.ResourceData, duplo *duplosdk.DuploMwaaAirflowDetail) {
	d.Set("name", duplo.Name)
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
	//webserver_access_mode := getAirflowObjectFromString(duplo.WebserverAccessMode)
	//if duplo.WebserverAccessMode != nil {
	//	d.Set("webserver_access_mode", duplo.WebserverAccessMode["Value"].(string))
	//}
	d.Set("status", getAirflowObjectFromString(duplo.Status))
	d.Set("last_update", getAirflowObjectFromString(duplo.LastUpdate))
	d.Set("tags", getAirflowObjectFromString(duplo.Tags))
	d.Set("logging_configuration", getAirflowObjectFromString(duplo.LoggingConfiguration))
	d.Set("airflow_configuration_options", getAirflowObjectFromString(duplo.AirflowConfigurationOptions))
}

func getAirflowObjectFromString(data *map[string]interface{}) string {
	if data != nil {
		data_str, _ := json.Marshal(&data)
		return string(data_str)
	}
	return ""
}

func diffIgnoreForAirflowLogConfiguration(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("logging_configuration", d)
}
