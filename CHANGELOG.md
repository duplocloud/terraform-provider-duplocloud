## 2024-01-24

### Added
- Introduced a new feature for validating the length of the lambda function name based on whether tag-based resource management is enabled or not in the `duplocloud/resource_duplo_aws_lambda_function.go` file.
- Added a new field `IsTagsBasedResourceMgmtEnabled` to the `DuploSystemFeatures` struct in the `duplosdk/admin.go` file.

## 2024-01-23

### Updated
- Enhanced the documentation and examples for `duplocloud_ecs_service` and `duplocloud_ecs_task_definition` resources.
- Introduced `health_check_url` to the `duplocloud_ecs_service` resource in the documentation and example.
- Added `ContainerMappings` to the `duplocloud_ecs_task_definition` resource in the documentation and example.

### Added
- Introduced a new attribute `can_scale_from_zero` to the `duplocloud_asg_profile` resource, allowing an AWS Autoscaling Group to leverage DuploCloud's scale from zero feature on Amazon EKS.
- Introduced support for serverless Kubernetes in the `duplocloud_infrastructure` resource. A new property `is_serverless_kubernetes` has been added to the resource schema.

## 2024-01-22

### Fixed
- Added `"ImplementationSpecific"` to the list of acceptable values for ingress rule path types to the `duplocloud_k8_ingress` resource.
