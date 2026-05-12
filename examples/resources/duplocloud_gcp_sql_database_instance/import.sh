# Example: Importing an existing GCP SQL database instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the GCP SQL database instance
#
terraform import duplocloud_gcp_sql_database_instance.sql_instance *TENANT_ID*/*SHORT_NAME*

# After import, `terraform plan` may show a diff for `root_password` because the
# GCP Cloud SQL GET API redacts this field. To avoid this, add the following
# lifecycle block to your resource:
#
#   lifecycle {
#     ignore_changes = [root_password]
#   }
