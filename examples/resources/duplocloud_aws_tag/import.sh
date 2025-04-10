# Example: Importing an existing AWS SSM Parameter
#  - *TENANT_ID* is the tenant GUID
#  - *ARN* The resource arn.
#  - *TAGKEY* Key of the tag
#
terraform import duplocloud_aws_tag.custom *TENANT_ID*/*ARN*/*TAGKEY*
