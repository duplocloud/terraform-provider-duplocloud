# Example: Importing an existing AWS MQ configuration
#  - *TENANT_ID* is the tenant GUID
#  - *BROKER_NAME* is the short name of the AWS MQ broker
#  - *BROKER_ID* is the broker ID of the AWS MQ broker

terraform import duplocloud_aws_mq_config.mqconfig *TENANT_ID*/*BROKER_ID*/*BROKER_NAME*
 