# Example: Importing an existing AWS API Gateway Event
#  - *TENANT_ID* is the tenant GUID
#  - *API_GATEWAY_ID* The API Gateway ID.
#  - *METHOD* The HTTP Method.
#  - *PATH* The API endpoint path.

terraform import duplocloud_aws_ssm_parameter.ssm_param *TENANT_ID*/*API_GATEWAY_ID*/*METHOD*/*PATH*
