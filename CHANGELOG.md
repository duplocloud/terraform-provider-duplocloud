## 2024-02-06

### Fixed
- Fixed an issue where changes to allocation tags in the ASG profile were not being detected and updated correctly.

### Updated
- Enhanced the ASG profile update process to detect and update changes in allocation tags.
- Refactored the update logic in ASG profiles for better clarity and maintainability.
- Introduced a new function to check for differences in allocation tags, ensuring that updates are only applied when necessary.

## 2024-02-01

### Added
- Introduced a new optional field `is_any_host_allowed` to the CronJob and Job resources, enhancing the control over host selection.


## 2024-01-29

### Updated
- Upgraded the `terraform-plugin-docs` version to fix an issue with generated docs incorrectly marking some fields as read-only.
- Updated multiple indirect dependencies to newer versions for bug fixes, performance improvements, and new features.

## 2024-01-29

### Added
- Introduced a new resource `duplocloud_gcp_sql_database_instance` for managing a GCP SQL Database Instance in Duplo.

## 2024-01-25

### Added
- Introduced support for backup retention period in RDS instances.
- Added support for spot instances and maximum spot price in ASG profiles.
- Introduced a new resource for managing tenant cleanup timers.

### Updated
- Refactored deletion protection settings update in RDS instances.
- Consolidated size and logging updates into a single function in RDS instances.
- Enhanced validation for update requests in tenant cleanup timers.

## 2024-01-24

### Added
- Introduced a new feature for validating the length of the lambda function name based on whether tag-based resource management is enabled or not in the `duplocloud/resource_duplo_aws_lambda_function.go` file.
- Added a new field `IsTagsBasedResourceMgmtEnabled` to the `DuploSystemFeatures` struct in the `duplosdk/admin.go` file.

### Updated
- Enhanced the documentation for `cloud` and `agent_platform` fields in the `duplocloud_duplo_service` resource. The description now includes a list of numeric IDs representing different cloud providers and container agents respectively.


## 2024-01-23

### Fixed
- Corrected a typo in `duplocloud_tenant` data source read function (tenant_id is the id naming for the resource but datasource calls it "id").
- Enhanced schema for `duplocloud_asg_profile` resource by forcing recreation when `is_cluster_autoscaled` and `can_scale_from_zero` attributes are changed.

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
