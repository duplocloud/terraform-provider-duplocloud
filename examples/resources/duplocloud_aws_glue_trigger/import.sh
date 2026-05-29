# Example: Importing an existing AWS Glue Trigger.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short trigger name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_trigger.nightly *TENANT_ID*/*NAME*
