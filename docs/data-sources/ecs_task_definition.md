---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_ecs_task_definition Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  
---

# duplocloud_ecs_task_definition (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `arn` (String) The ARN of the task definition.
- `tenant_id` (String) The GUID of the tenant that the task definition will be created in.

### Read-Only

- `container_definitions` (String)
- `container_definitions_updates` (String) container_definitions updates in backend
- `cpu` (String)
- `execution_role_arn` (String)
- `family` (String) The name of the task definition to create.
- `full_family_name` (String) The name of the task definition to create.
- `id` (String) The ID of this resource.
- `inference_accelerator` (Set of Object) (see [below for nested schema](#nestedatt--inference_accelerator))
- `ipc_mode` (String)
- `memory` (String)
- `network_mode` (String)
- `pid_mode` (String)
- `placement_constraints` (Set of Object) (see [below for nested schema](#nestedatt--placement_constraints))
- `prevent_tf_destroy` (Boolean) Prevent this resource to be deleted from terraform destroy. Default value is `true`.
- `proxy_configuration` (List of Object) (see [below for nested schema](#nestedatt--proxy_configuration))
- `requires_attributes` (Set of Object) (see [below for nested schema](#nestedatt--requires_attributes))
- `requires_compatibilities` (Set of String) Requires compatibilities for running jobs. Valid values are [FARGATE]
- `revision` (Number) The current revision of the task definition.
- `runtime_platform` (List of Object) Configuration block for runtime_platform that containers in your task may use. Required on ecs tasks that are hosted on Fargate. (see [below for nested schema](#nestedatt--runtime_platform))
- `status` (String) The status of the task definition.
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))
- `task_role_arn` (String)
- `volumes` (String) A JSON-encoded string containing a list of volumes that are used by the ECS task definition.

<a id="nestedatt--inference_accelerator"></a>
### Nested Schema for `inference_accelerator`

Read-Only:

- `device_name` (String)
- `device_type` (String)


<a id="nestedatt--placement_constraints"></a>
### Nested Schema for `placement_constraints`

Read-Only:

- `expression` (String)
- `type` (String)


<a id="nestedatt--proxy_configuration"></a>
### Nested Schema for `proxy_configuration`

Read-Only:

- `container_name` (String)
- `properties` (Map of String)
- `type` (String)


<a id="nestedatt--requires_attributes"></a>
### Nested Schema for `requires_attributes`

Read-Only:

- `name` (String)


<a id="nestedatt--runtime_platform"></a>
### Nested Schema for `runtime_platform`

Read-Only:

- `cpu_architecture` (String)
- `operating_system_family` (String)


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)
