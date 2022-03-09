# Example: Importing an existing AWS target group attributes.
#  - *TENANT_ID* is the tenant GUID
#  - *TARGET_GROUP_ARN* is the ARN of target group.
#
terraform import duplocloud_aws_target_group_attributes.tgAttrs *TENANT_ID*/*TARGET_GROUP_ARN*
