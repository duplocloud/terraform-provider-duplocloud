# Example: Importing an existing ECS task definition
#  - *TENANT_ID* is the tenant GUID
#  - *ARN* is the full ARN of the task definition
#
terraform import duplocloud_ecs_task_definition.myservice subscriptions/*TENANT_ID*/EcsTaskDefinition/*ARN*
