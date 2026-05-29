# Example: Importing an existing AWS Glue Job.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short job name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_job.etl *TENANT_ID*/*NAME*
