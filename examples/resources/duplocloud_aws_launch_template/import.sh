# Example: Importing an existing AWS launch template
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name of the AWS launch template
#  - *VERSION* available version of launch template , it is optional, use if needed to import specific version of available launch template

terraform import duplocloud_aws_launch_template.lt *TENANT_ID*/launch-template/*NAME*/*VERSION*
