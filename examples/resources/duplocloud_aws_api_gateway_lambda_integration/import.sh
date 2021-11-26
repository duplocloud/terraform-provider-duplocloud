# Example: Importing an existing aws api gateway resource
#  - *TENANT_ID* is the tenant GUID
#  - *FRIENDLY_NAME* is the duploservices-<account_name>-<friendly_name>-<aws-account-number>
#
terraform import duplocloud_aws_api_gateway_integration.myApiGateway *TENANT_ID*/*FRIENDLY_NAME*
