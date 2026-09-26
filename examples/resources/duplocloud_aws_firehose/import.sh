# Example: Importing an existing AWS Firehose delivery stream
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the delivery stream (without the tenant prefix)
#
terraform import duplocloud_aws_firehose.stream *TENANT_ID*/*SHORT_NAME*
