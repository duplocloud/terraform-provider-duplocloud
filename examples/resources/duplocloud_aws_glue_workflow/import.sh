# Example: Importing an existing AWS Glue Workflow.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short workflow name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_workflow.pipeline *TENANT_ID*/*NAME*
