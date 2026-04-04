# Example: Importing an existing S3 table for an AWS resource
#  - *TENANT_ID* is the tenant GUID
#  - *S3_TABLE_NAME* The non prefix name of the s3 table.
#
terraform import duplocloud_aws_s3_table.table *TENANT_ID*/*S3_TABLE_NAME*
