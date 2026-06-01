# Example: Importing an existing AWS Glue Crawler.
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short crawler name (without the duploservices-<tenant>- prefix)
#
terraform import duplocloud_aws_glue_crawler.events *TENANT_ID*/*NAME*
