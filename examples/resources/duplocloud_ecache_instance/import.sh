# Example: Importing an existing AWS ElastiCache cluster
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS ElastiCache cluster
#
terraform import duplocloud_ecache_instance.mycluster v2/subscriptions/*TENANT_ID*/ECacheDBInstance/*SHORT_NAME*
