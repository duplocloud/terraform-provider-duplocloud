# Example: Importing an existing cloudwatch metric alarm
#  - *TENANT_ID* is the tenant GUID
#  - *FRIENDLY_NAME* is the hypen separated alarm dimension values and metric name

terraform import duplocloud_aws_cloudwatch_metric_alarm.myMetricAlarm *TENANT_ID*/*FRIENDLY_NAME*