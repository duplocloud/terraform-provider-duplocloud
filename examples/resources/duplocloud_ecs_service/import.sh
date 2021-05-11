# Example: Importing an existing service
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the short name of the service
#
terraform import duplocloud_ecs_service.myservice v2/subscriptions/*TENANT_ID*/EcsServiceApiV2/*NAME*
