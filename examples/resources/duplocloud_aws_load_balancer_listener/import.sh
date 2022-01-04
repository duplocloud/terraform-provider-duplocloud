# Example: Importing an existing AWS load balancer listener
#  - *TENANT_ID* is the tenant GUID
#  - *LB_NAME* is the name of the AWS load balancer
#  - *LISTENER_ARN* is the arn of the AWS load balancer listener
#
terraform import duplocloud_aws_load_balancer_listener.myalb-listener *TENANT_ID*/*LB_NAME*/*LISTENER_ARN*
