# Example: Importing an existing S3 bucket event notification configuration
#  - *TENANT_ID* is the tenant GUID
#  - *SOURCEBUCKETNAME* is the full name of the S3 bucket
#
terraform import duplocloud_s3_event_notification.event *TENANT_ID*/*SOURCEBUCKETNAME*/event_notification
