# Example: Importing an existing AWS MQ broker
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS MQ broker
# -  *BROKER_ID* is the id of an AWS MQ instance
#
terraform import duplocloud_aws_mq_broker.mq *TENANT_ID*/*BROKER_ID*/*SHORT_NAME*
 