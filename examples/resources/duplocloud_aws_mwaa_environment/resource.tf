resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

data "duplocloud_tenant_aws_kms_key" "tenant_kms_key" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
}

resource "duplocloud_aws_mwaa_environment" "my-mwaa" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
  name                   = "airflow-test"
  source_bucket_arn      = "arn:aws:s3:::xxx-xxx-xx-xxxx"
  dag_s3_path            = "AirflowDags/dag"
  plugins_s3_path        = "AirflowDags/plugins.zip"
  requirements_s3_path   = "AirflowDags/requirements.txt"
  startup_script_s3_path = "AirflowDags/startup-script.sh"

  ## optional: Provide particular s3 version of the file. If empty, latest version will be used (recommended).
  # plugins_s3_object_version = "S3 object version"
  # requirements_s3_object_version = "S3 object version"
  # startup_script_s3_object_version = "S3 object version"

  kms_key                         = data.duplocloud_tenant_aws_kms_key.tenant_kms_key.key_arn
  schedulers                      = 2
  max_workers                     = 10
  min_workers                     = 1
  airflow_version                 = "2.6.3"
  weekly_maintenance_window_start = "SUN:23:30"
  environment_class               = "mw1.small"

  airflow_configuration_options = {
    "core.log_format" : "[%%(asctime)s] {{%%(filename)s:%%(lineno)d}} %%(levelname)s - %%(message)s"
  }

  logging_configuration {
    dag_processing_logs {
      enabled   = false
      log_level = "INFO"
    }

    scheduler_logs {
      enabled   = false
      log_level = "INFO"
    }

    task_logs {
      enabled   = false
      log_level = "INFO"
    }

    webserver_logs {
      enabled   = false
      log_level = "INFO"
    }

    worker_logs {
      enabled   = false
      log_level = "INFO"
    }
  }
}
