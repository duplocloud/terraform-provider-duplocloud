# Example: Importing an existing AWS Glue Schema Registry.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short registry name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_registry.events *TENANT_ID*/*NAME*
