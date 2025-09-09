# Example: Importing an existing AWS MQ broker
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS load balancer
#
terraform import duplocloud_aws_mq_broker.mq *TENANT_ID*/*SHORT_NAME*
