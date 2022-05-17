# Example: Importing an existing AWS SNS Topic
#  - *TENANT_ID* is the tenant GUID
#  - *ARN* The ARN of the created Amazon SNS Topic.
#
terraform import duplocloud_aws_sns_topic.sns_topic *TENANT_ID*/*ARN*
