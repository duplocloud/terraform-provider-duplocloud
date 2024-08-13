# Example: Importing an existing AWS lambda function event invoke config
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS lambda function
#
terraform import duplocloud_aws_lambda_function_event_config.event-invoke-config *TENANT_ID*/*SHORT_NAME*/eventInvokeConfig
