# Example: Importing an existing AWS Glue Schema.
#  - *TENANT_ID* is the tenant GUID
#  - *REGISTRY_NAME* is the short parent registry name (no prefix)
#  - *NAME* is the short schema name (no prefix)
#
# The provider fetches the latest schema version on import so the
# SchemaDefinition is preserved in state — no extra apply needed.
#
terraform import duplocloud_aws_glue_schema.click *TENANT_ID*/*REGISTRY_NAME*/*NAME*
