# Example: Importing an existing AWS Glue Table.
#  - *TENANT_ID* is the tenant GUID
#  - *DATABASE_NAME* is the short parent database name (no prefix)
#  - *NAME* is the table name (tables are not prefixed)
#
terraform import duplocloud_aws_glue_table.events *TENANT_ID*/*DATABASE_NAME*/*NAME*
