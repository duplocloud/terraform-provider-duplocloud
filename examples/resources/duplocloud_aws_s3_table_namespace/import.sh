# Example: Importing an existing S3 table for an AWS resource
#  - *TENANT_ID* is the tenant GUID
#  - *S3_TABLE_NAME* The fullname of the s3 table.
#  - *NAMESPACE_NAME* The name of the s3 table namespace.
terraform import duplocloud_aws_s3_table_namespace.namespace *TENANT_ID*/*S3_TABLE_NAME*/*NAMESPACE_NAME*
