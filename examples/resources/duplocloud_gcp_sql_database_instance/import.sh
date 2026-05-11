# Example: Importing an existing GCP SQL database instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the GCP SQL database instance
#
terraform import duplocloud_gcp_sql_database_instance.sql_instance *TENANT_ID*/*SHORT_NAME*

# After import, `terraform plan` may show an in-place change for `root_password`
# (the GCP API never returns the password, so state can't be populated from it).
# Set `root_password` in your config to the existing password, then run
# `terraform apply` once to sync state. No API call is made for this field;
# the apply only updates the local state file.
