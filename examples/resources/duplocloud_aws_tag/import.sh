# Example: Importing an existing AWS tag
#  - *TENANT_ID* is the tenant GUID
#  - *ARN* The resource arn. Should be encoded thrice with URL encoding.
#  - *TAGKEY* Key of the tag
#
terraform import duplocloud_aws_tag.custom *TENANT_ID*/*ARN*/*TAGKEY*
