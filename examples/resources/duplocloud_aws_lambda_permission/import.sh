# Example: Importing an existing AWS lambda function permission
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS lambda function
#  - *STATEMENT_ID* is the statement ID of the permission
#
terraform import duplocloud_aws_lambda_permission.permission *TENANT_ID*/*SHORT_NAME*/*STATEMENT_ID*
