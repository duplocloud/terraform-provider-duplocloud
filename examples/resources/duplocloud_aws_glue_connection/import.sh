# Example: Importing an existing AWS Glue Connection.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short connection name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_connection.warehouse *TENANT_ID*/*NAME*
