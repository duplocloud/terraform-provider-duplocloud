# Example: Importing an existing GCP SQL database instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the GCP SQL database instance
#
terraform import duplocloud_gcp_sql_database_instance.sql_instance *TENANT_ID*/*SHORT_NAME*

# After import, `terraform plan` will show a diff for `root_password` because the
# GCP Cloud SQL GET API redacts this field. You have two options:
#
# 1. Set `root_password` in your config to the current GCP value and run
#    `terraform apply` once. The provider syncs state to the config value
#    without pushing a password change to GCP. For future rotations, change
#    the password in GCP first, then update `root_password` and re-apply.
#
# 2. Suppress the diff permanently by adding the lifecycle block below.
#    Use this only if you do not want Terraform to track the password value
#    (rotations must then be performed out-of-band):
#
#      lifecycle {
#        ignore_changes = [root_password]
#      }
