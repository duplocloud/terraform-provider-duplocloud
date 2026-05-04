# Example: Importing an existing ASG warm pool
#  - *TENANT_ID* is the tenant GUID
#  - *ASG_NAME*  is the full ASG name (e.g. duploservices-<tenant>-<shortname>)
#
terraform import duplocloud_aws_asg_warm_pool.mywarmpool *TENANT_ID*/*ASG_NAME*/warmpool
