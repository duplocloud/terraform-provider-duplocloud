# Example: Importing an existing S3 bucket
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the S3 bucket (without the duploservices prefix)
#
terraform import duplocloud_s3_bucket.mybucket *TENANT_ID*/*SHORTNAME*
