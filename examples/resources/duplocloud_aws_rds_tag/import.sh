# Example: Importing an existing RDS Tag.
#  - *TENANT_ID* is the tenant GUID.
#  - *RESOURCE_TYPE* The type of the RDS resource, Valid vaues are - "cluster" and "instance".
#  - *RESOURCE_ID* The RDS identifier.
#  - *TAG_KEY* The tag name.
#
terraform import duplocloud_aws_rds_tag.tag1 *TENANT_ID*/*RESOURCE_TYPE*/*RESOURCE_ID*/*TAG_KEY*
