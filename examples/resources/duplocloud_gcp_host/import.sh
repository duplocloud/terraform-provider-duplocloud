# Example: Importing an existing AWS host
#  - *TENANT_ID* is the tenant GUID
#  - *INSTANCE_ID* is the GCP VM instance ID
#
terraform import duplocloud_gcp_host.myhost v2/subscriptions/*TENANT_ID*/gcpHost/*INSTANCE_ID*


