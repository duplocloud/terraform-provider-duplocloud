# Example: Importing an existing kubernetes gateway
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the gateway name
#
terraform import duplocloud_k8s_gateway.mygateway v3/subscriptions/*TENANT_ID*/k8s/gateway/*NAME*
