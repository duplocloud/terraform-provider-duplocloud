# Example: Importing an existing AWS Cloudwatch Event Rule
#  - *TENANT_ID* is the tenant GUID
#  - *FRIENDLY_NAME* is the duploservices-<account_name>-<friendly_name>
#
terraform import duplocloud_aws_cloudwatch_event_rule.myEventRule *TENANT_ID*/*FRIENDLY_NAME*
