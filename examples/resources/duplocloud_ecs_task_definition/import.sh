# Example: Importing an existing ECS task definition
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the full name of the task definition
#
terraform import duplocloud_ecs_task_definition.myservice subscriptions/*TENANT_ID*/EcsTaskDefinition/*NAME*
