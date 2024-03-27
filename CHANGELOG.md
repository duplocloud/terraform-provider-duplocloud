<<<<<<< HEAD
## 2023-03-27

### Added
- Added example for `duplocloud_gcp_node_pool` and `duplocloud_gcp_node_pools` for data-source

=======
## 2023-03-26

### Fixed
- Fixed `duplocloud_s3_bucket` resource creation issue
>>>>>>> develop
## 20234-03-22

### Fixed
- Fixed plugin crash issue for user exist case related to `duplocloud_user` resource
## 2024-03-21

### Updated
- Updated documentation and examples to version 0.10.14.

## 2024-03-20

### Added
- Introduced `force_recreate_on_volumes_change` boolean field in `duplocloud_duplo_service` to control resource recreation when volume mappings are modified.
- VM not getting created on `duplocloud_gcp_node_pool` resource creation fixed

### Updated
- Enhanced `volumes` field description in `duplocloud_duplo_service` schema for better clarity.
- Implemented `customDuploServiceDiff` function to enforce resource recreation based on `force_recreate_on_volumes_change` field.

### Fixed
- VM not getting created on `duplocloud_gcp_node_pool` resource creation fixed
 
## 2024-03-19

### Added
- Added datasource `duplocloud_gcp_node_pools`
- Added datasource `duplocloud_gcp_node_pool`
 
## 2024-03-18

### Added
- Added support for specifying the instruction set architecture for AWS Lambda functions in the Terraform provider, including `[x86_64]` and `[arm64]`.

## 2024-03-15

### Updated
- Enhanced `duplocloud_aws_cloudfront_distribution` resource documentation with cache policy descriptions and default TTL clarifications.
- Added support for GCP SQL data source and SQL list data source in DuploCloud provider.

## 2024-03-14

### Added
- Added `duplocloud_gcp_node_pool` resource to the Terraform provider.
- Added `duplocloud_gcp_node_pool` resource example and document.
- Implemented V3 API support for S3 bucket operations, including create, read, update, and delete, with fallback to older API if V3 is not available.
- Added support for specifying the region of an S3 bucket in Terraform resource.
- Extended SDK to support new V3 API endpoints for S3 bucket operations.
- Updated documentation to include the new `region` attribute for the S3 bucket resource.

## 2024-03-13

### Updated
- updated doc for `duplocloud_aws_cloudfront_distribution` resource

## 2024-03-12

### Added
- Extended python version support for `duplocloud_aws_lambda_function` resource
- Added datasource for `duplocloud_gke_credentials`

### Fixed
- `duplocloud_infrastructure` resource dereference bug fix

## 2024-03-11

## Fixed
- fixed changes occuring while planning on already created memcache type for `duplocloud_ecache_instance` resource when no changes done

## 2024-03-07

### Fixed
- `unrestricted_ext_lb` attribute not setting to false fix for `duplocloud_plan_settings` resource

## 2024-03-06

### Changed
- Simplified the `DuploAwsLifecyclePolicyUpdate` function by removing the return value to enhance clarity and efficiency in EFS lifecycle updates.
- Updated function calls and error handling in `resource_duplo_aws_efs_file_system` and `resource_duplo_aws_efs_file_system_lifecycle_policy` to accommodate the changes in lifecycle policy updates.

### Fixed
- Fixed parameter handling in the update lifecycle policy for EFS, ensuring more reliable lifecycle management.
- 
## 2024-03-05

### Added
- Added data source for `duplocloud_gcp_sql_database_instance`
- Added example related to data for `duplocloud_gcp_sql_database_instance`

### Fixed
- Fixed `duplocloud_k8_secret` nil map panic bug

## 2024-02-28

### Added
- Introduced shared test utilities for mocking HTTP responses and refactored HTTP test setup.
- Added acceptance tests for `data.duplocloud_native_hosts` data source.
- Implementd support for `secret_labels` attribute for `duplocloud_k8_secret` resource
- Added support for `cluster_parameter_group_name` for `duplocloud_rds_instance` resource
- Added support for configuring deletion protection and point-in-time recovery for the `duplocloud_aws_dynamodb_table_v2` resource.
- Introduced shared test utilities for mocking HTTP responses and refactored HTTP test setup.
- Added acceptance tests for `data.duplocloud_native_hosts` data source.

### Fixed
- Fixed handling of `volume` and `network_interface` in `duplocloud_aws_host` resource to avoid unnecessary diffs.
- Fixed crash in tenant data source on missing `TenantPolicy`.

## 2024-02-27

### Fixed
- Improved the deletion process for AWS Batch Job Definitions to correctly handle all revisions, resolving a timeout issue during resource destruction.
- Fixed resource deletion terraform timeout issue for `duplo_aws_batch_job_definition` resource
 
## 2024-02-26

### Fixed
- `duplocloud_rds_instance` Resolved an issue where `SkipFinalSnapshot` was not included in the JSON during an update in the DuploRdsUpdateInstance serialization.

## 2024-02-27

### Changed
- Enhanced EFS lifecycle policies by adding detailed descriptions for `transition_to_ia` and `transition_to_primary_storage_class` policies in `duplocloud_aws_efs_lifecycle_policy` resource.
- Provided an example Terraform configuration demonstrating how to use the `duplocloud_aws_efs_lifecycle_policy` resource.

### Added
- Added support for configuring a dead letter queue (DLQ) `dead_letter_queue` for AWS Lambda functions `duplocloud_aws_lambda_function`, allowing users to specify an SNS topic or SQS queue for failed invocation notifications. This includes CRUD operations for this new feature and state management in the Terraform provider.
- Added support for Node.js 20.x runtime for the `duplocloud_aws_lambda_function` resource.
- Introduced shared test utilities for mocking HTTP responses and refactored HTTP test setup.
- Added acceptance tests for `data.duplocloud_native_hosts` data source.
- Implementd support for `secret_labels` attribute for `duplocloud_k8_secret` resource

### Fixed
- Resolved an issue in `duplocloud_aws_elasticsearch` resource where OpenSearch could not be created with both warm enable and cold storage set to false.
- Corrected the `FileSystemId` assignment in the EFS update function for `duplocloud_aws_efs_lifecycle_policy`
- Fixed a potential crash in `duplocloud_aws_elasticsearch` creation when `dedicated_master_type` was not defined.
- Fixed `store_details_in_secret_manager` attribute not getting set for `duplocloud_rds_instance` resource in duplo 
- Fixed handling of `volume` and `network_interface` in `duplocloud_aws_host` resource to avoid unnecessary diffs.
- Fixed crash in tenant data source on missing `TenantPolicy`.

## 2024-02-21

### Added 
- Added `duplocloud_gcp_node_pool` resource to the Terraform provider.
- Added `duplocloud_gcp_node_pool` resource example and document

## 2024-02-15

### Added
- Introduced `skip_final_snapshot` attribute to `duplocloud_rds_instance`, allowing control over whether a final snapshot is taken upon RDS instance deletion.

### Fixed
- Fixed plugin crash while creating opensearch resource when warm enabled set to true for `duplocloud_aws_elasticsearch` resource at `awsElasticSearchDomainClusterConfigFromState`..
- Handled nil pointer exceptions in `duplocloud_aws_cloudwatch_event_rule` resources at `resourceAwsCloudWatchEventRuleRead`  
- Fixed the issue of missing Elasticsearch version in Terraform state after apply by correcting the JSON tag name for `ElasticSearchVersion` in `DuploElasticSearchDomain` struct for `duplocloud_aws_elasticsearch` resource.

## 2024-02-13

### Added
- Unit tests for `duplocloud_tenant_config` resource and data source, both happy path and edge cases.
- Updates to `TenantSetConfigKey` and `TenantDeleteConfigKey` methods in the SDK to use the v3 API, improving the handling of tenant configurations.

### Changed
- The emulator now supports dynamic path parameters and added routes for the tenant metadata API.
- Updated the `duplocloud_aws_host` resource test helper function to use the new test helper for generating Terraform resource definitions.

### Breaking Changes
- Duplo releases older than 1/2023 will be missing the v3 tenant metadata APIs.

## 2024-02-12

### Added
- Unit tests for `duplocloud_aws_host`
  - Added comprehensive test cases for AWS host resource, covering basic configuration, public/private subnets, and zone selection.
  - Enhanced emulator response for AWS hosts to include `UserAccount` and `IdentityRole`.
  - Improved handling of `volume` and `network_interface` in AWS host resource to avoid unnecessary diffs.
  - Extended the emulator to support AWS host creation and deletion APIs, and to list external and internal subnets per tenant.
  - Introduced helper functions for generating Terraform resource definitions in tests.
  - Updated and added new fixtures for AWS hosts and subnets to support testing.

## 2024-02-10

### Added
- Added Terraform acceptance tests for `data.duplocloud_native_hosts`.
- Introduced a placeholder for future `duplocloud_aws_host` resource tests.
- Extended the emulator to handle dynamic path parameters and added routes for the AWS host API.

### Fixed
- Fixed bugs in `data.duplocloud_native_hosts` and `duplocloud_aws_host` where `volume` and `network_interface` fields were not parsed.

## 2024-02-09

### Added
- Introduced a new attribute `prepend_user_data` to the `duplocloud_aws_host` and `duplocloud_asg_profile` resources, allowing for prepending user data on AWS hosts.
- Enhanced the `duplocloud_aws_host` and `duplocloud_asg_profile` resources to avoid triggering an unnecessary "force replacement" on hosts and ASGs that prepend Duplo's user data.
- Updated the data source types for `duplocloud_aws_host` and `duplocloud_asg_profile` resources to include the new `prepend_user_data` attribute.
- Introduced acceptance tests for Terraform provider with mock data.
- Implemented a mock server setup for testing HTTP requests.
- Added a basic acceptance test for tenant data source.
- Refactored existing SDK client tests to utilize shared test utilities.
- Created shared test utilities for mocking HTTP responses.
- Added a JSON fixture for tenant data for testing purposes.

## Updated
- Updated `duplocloud_ecs_task_definition` examples and documentation to use the correct `PortMappings` property instead of `ContainerMappings` in `container_definitions` field.
- Fixed property typos in `duplocloud_ecs_task_definition`'s `container_definitions` example and docs

### Fixed
- Resolved nil pointer and index out of bound exceptions in `resource_duplo_aws_elasticsearch.go` by properly initializing `ColdStorageOptions` and `WarmType` only if applicable for `duplocloud_aws_elasticsearch` resource.
- Updated `DuploElasticSearchDomainClusterConfig` struct to use pointers for `WarmType` and `ColdStorageOptions` to prevent nil pointer dereference issues for `duplocloud_aws_elasticsearch` resource.
- Fixed a bug that prevented the creation of `duplocloud_aws_elasticsearch` with both `warm_enabled` and `cold_storage_options` are set to `false`. Now, these options are properly handled, preventing errors during the creation of `duplocloud_aws_elasticsearch`  with these options disabled.
- Fixed a regression in `duplocloud_k8_ingress` validation where `port` and `port_name` were not correctly validated as mutually exclusive.

## 2024-02-08

### Added
- Introduced a new resource `duplocloud_aws_efs_lifecycle_policy` for managing AWS EFS lifecycle policies.
- Added `lifecycle_policy` attribute to `duplocloud_aws_efs_file_system` resource to support lifecycle policy configurations.
- Implemented CRUD operations for EFS lifecycle policy management in both Terraform resource and Duplo SDK.
- Added support for `port_name` attribute in `duplocloud_k8_ingress` allowing service port specification by name.
- Made `port` in `duplocloud_k8_ingress` attribute optional and enforced port range validation for ingress rules.
- Added the `delay_seconds` attribute to the `aws_sqs_queue` resource, enabling the postponing of delivery for new messages in seconds.

### Update 
- Updated documentation for `aws_sqs_queue` for new attribute `delay_seconds`.
- Updated example for `aws_sqs_queue` resource added attribute `delay_seconds` to resource type `duplocloud_aws_sqs_queue`.

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
