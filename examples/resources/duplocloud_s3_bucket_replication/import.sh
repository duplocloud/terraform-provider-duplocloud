# Example: Importing an existing S3 bucket replication
#  - *TENANT_ID* is the tenant GUID
#  - *SOURCEBUCKETNAME* is the full name of the S3 bucket
#
terraform import duplocloud_s3_bucket_replication.mybucket *TENANT_ID*/*SOURCEBUCKETNAME*
