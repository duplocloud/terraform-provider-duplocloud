# Example: Importing an existing S3 table for an AWS resource
#  - *TENANT_ID* is the tenant GUID
#  - *S3_TABLE_NAME* is the non-prefixed name of the S3 table bucket.
#
terraform import duplocloud_aws_s3_table.table *TENANT_ID*/*S3_TABLE_NAME*
