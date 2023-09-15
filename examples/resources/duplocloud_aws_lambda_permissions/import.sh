# Example: Importing an existing AWS lambda function permissions
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS lambda function
#
terraform import duplocloud_aws_lambda_permissions.permission v3/subscriptions/*TENANT_ID*ß/serverless/lambdapermission/*SHORT_NAME*
