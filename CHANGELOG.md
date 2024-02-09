## 2024-02-09

### Fixed
- Resolved nil pointer and index out of bound exceptions in `resource_duplo_aws_elasticsearch.go` by properly initializing `ColdStorageOptions` and `WarmType` only if applicable for `aws_elasticsearch` resource.
- Updated `DuploElasticSearchDomainClusterConfig` struct to use pointers for `WarmType` and `ColdStorageOptions` to prevent nil pointer dereference issues for `aws_eleasticsearch` resource.

## 2024-02-08

### Added
- Introduced a new resource `duplocloud_aws_efs_lifecycle_policy` for managing AWS EFS lifecycle policies.
- Added `lifecycle_policy` attribute to `duplocloud_aws_efs_file_system` resource to support lifecycle policy configurations.
- Implemented CRUD operations for EFS lifecycle policy management in both Terraform resource and Duplo SDK.
- Added support for `port_name` attribute in `duplocloud_k8_ingress` allowing service port specification by name.
- Made `port` in `duplocloud_k8_ingress` attribute optional and enforced port range validation for ingress rules.

## 2024-02-06

### Fixed
- Fixed an issue where changes to allocation tags in the ASG profile were not being detected and updated correctly.
- assigns `CustomDataTags` to AsgProfile's minion_tags fields as this field receives the tag edits in the beckend

### Added
- Introduced comprehensive unit tests for the `getAPI`, `putAPI`, and `deleteAPI` methods in the DuploCloud SDK client, enhancing the test coverage for various scenarios including successful API calls, error handling, and response parsing.

## 2024-02-03

### Fixed
- Resolved a bug where an empty subnet was being created in all infrastructures, which could lead to errors. Now, a subnet is only created when the infrastructure is Azure, where it is needed.

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
