# Example: Importing an existing AWS Cloudfront Distribution
#  - *TENANT_ID* is the tenant GUID
#  - *CLOUDFRONT_ID* is the cloudfront id
#  - *NAME* is the name of the cloudfront distribution
terraform import duplocloud_aws_cloudfront_distribution_v2.myCFD *TENANT_ID*/*CLOUDFRONT_ID*/*NAME*
