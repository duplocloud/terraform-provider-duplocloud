# Example: Importing an existing service's load balancer configurations
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name of the service
#
terraform import duplocloud_duplo_service_lbconfigs.myservice v2/subscriptions/*TENANT_ID*/ServiceLBConfigsV2/*NAME*
