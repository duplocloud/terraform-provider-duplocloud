# Example: Importing an existing AWS host
#  - *TENANT_ID* is the tenant GUID
#  - *INSTANCE_ID* is the AWS EC2 instance ID
#
terraform import duplocloud_aws_host.myhost v2/subscriptions/*TENANT_ID*/NativeHostV2/*INSTANCE_ID*
