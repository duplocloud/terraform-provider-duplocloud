resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

data "duplocloud_tenant_aws_kms_key" "tenant_kms_key" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
}

resource "duplocloud_aws_mwaa_environment" "my-mwaa" {
  tenant_id                       = duplocloud_tenant.myapp.tenant_id
  name                            = "airflow-test"
  source_bucket_arn               = "arn:aws:s3:::duploservices-demo01-dags-140563923322"
  dag_s3_path                     = "AirflowDags/dag"
  kms_key                         = data.duplocloud_tenant_aws_kms_key.tenant_kms_key.key_arn
  schedulers                      = 2
  max_workers                     = 10
  min_workers                     = 1
  airflow_version                 = "2.2.2"
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
