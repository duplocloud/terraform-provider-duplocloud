# Example: Importing an existing AWS SNS Topic
#  - *TENANT_ID* is the tenant GUID
#  - *RESOURCE_TYPE* The type of the RDS resource, Valid vaues are - "cluster" and "db".
#  - *RESOURCE_ID* The RDS identifier.
#  - *TAG_KEY* The tag name..
#
terraform import duplocloud_aws_rds_tag.tag1 *TENANT_ID*/*RESOURCE_TYPE*/*RESOURCE_ID*/*TAG_KEY*
