# Example: Importing an existing AWS Glue Database.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short database name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_database.analytics *TENANT_ID*/*NAME*
