## 2024-09-06

### Enhanced
- Renamed the `enable` attribute to `enabled` for performance insights in RDS instances, improving clarity and consistency.
- Updated documentation to reflect changes in performance insights configuration for RDS instances.

## 2024-09-05

### Fixed
- Resolved a regression issue in `duplocloud_azure_k8_node_pool` by adding a length check for the `scale_priority` slice to prevent crashes when the scale set priority is not specified.

### Enhanced
- Added validation to the `enabled_metrics` attribute in the Auto Scaling Group profile to ensure only specified metrics are allowed, improving configuration accuracy.
- Improved log delivery configuration with updated documentation and validation to prevent duplicate log types.
- Implemented version checks for log types to ensure compatibility with specified engine versions.
- Updated GitHub Actions workflows to use `DUPLO_TF_GITHUB_TOKEN` for improved security and consistency.
- Removed duplicate examples and outdated content from AWS host and RDS instance documentation, enhancing clarity and accuracy.
- Introduced support for S3 bucket replication, enabling cross-region replication configurations.
- Added performance insights configuration for RDS instances, allowing enhanced monitoring and tuning capabilities.
- Implemented failover and logging configurations for ElastiCache instances, improving reliability and observability.
- Updated Terraform provider version to 0.10.42 across multiple integration examples, ensuring compatibility with the latest features and improvements.

## 2024-09-04

### Enhanced
- Updated `spot_max_price` data type to float and added validation to ensure it is at least 0.00001 in `duplocloud_azure_k8_node_pool` resource.
- Set default value for `eviction_policy` to "Delete" for Spot priority in `duplocloud_azure_k8_node_pool` resource.
- Added custom diff validator to ensure proper handling of `scale_priority` attributes, preventing unsupported configurations.
- Improved logging for AWS SSM parameters to securely handle sensitive data by masking values for `SecureString` type.
- Added `ForceNew` to the `eviction_policy` attribute in Azure K8s node pool resources to ensure proper resource replacement when the policy changes.
- Updated `duplocloud_azure_k8_node_pool` resource to enforce recreation when the `priority` attribute is modified, improving resource management.
- Implemented validation to ensure `eviction_policy` and `spot_max_price` are not set for `Regular` priority type in `duplocloud_azure_k8_node_pool` resource, enhancing error handling during the plan phase.
- Implemented validation to ensure `spot_max_price` is not set when `scale_priority` is `Regular` in Azure Kubernetes Node Pool configurations, enhancing error handling during the plan phase.

## 2024-09-03

### Enhanced
- Implemented a diff suppression function for Kubernetes secrets to handle JSON objects as string key-value pairs, improving accuracy in change detection.
- Improved descriptions for `duplocloud_ecache_instance`, `duplocloud_rds_instance`, and `duplocloud_duplo_service` resources.
- Updated documentation for ElastiCache, RDS, S3 bucket, and tenant resources with detailed examples and usage instructions.
- Added new templates for resource documentation, including structured examples and import instructions.
- Clarified cloud provider codes in infrastructure documentation and templates.

### Fixed
- Added a nil check in SSM parameter reading to prevent potential nil pointer exceptions, enhancing stability.


## 2024-08-26

### Enhanced
- Added support for automatic failover in Redis configurations, enabling enhanced reliability for instances with multiple replicas.
- Introduced log delivery configurations for Redis, allowing logs to be sent to CloudWatch Logs or Kinesis Firehose.
- Added `availability_zone` field to RDS instance schema, allowing specification of a valid Availability Zone for RDS primary or Aurora writer instances, with validation to prevent conflicts with `multi_az`. Updated documentation with examples for the new field.

### Fixed
- Removed fallback mechanism for error handling in DynamoDB table operations, simplifying code and improving maintainability.

## 2024-08-21

### Enhanced
- Added support for configuring dead letter queues in SQS resources within the DuploCloud Terraform provider, allowing specification of target DLQ and maximum message receive attempts.

## 2024-07-31

### Fixed
- Adjusted URL path construction for force delete ECR operations to reflect changes in the data contract.

## 2024-07-23

### Enhanced
- Added `force_delete` option to ECR resource, allowing forced deletion of ECR repositories in DuploCloud provider.

## 2024-07-22

### Added
- Introduced support for managing AWS Lambda function event invoke configurations in the DuploCloud provider.
- Added a new resource `duplocloud_aws_lambda_function_event_config` to handle asynchronous invocation settings for AWS Lambda functions.

## 2024-07-17

### Enhanced
- Added `force_delete` option to ECR resource, allowing forced deletion of ECR repositories in DuploCloud provider.

## 2024-04-29

### Fixed
- Fixed an issue with handling slashes in `keyType` for `duplocloud_admin_system_setting` resource ID creation and parsing.

### Enhanced
- Added utility functions to encode and decode slashes in identifiers, improving handling of special characters in resource IDs.
- Updated build script to include documentation generation and modified import script to guide users on handling slashes in identifiers.


## 2024-07-12

### Enhanced
- Added support for advanced health check configurations in load balancer configurations for Duplo services, including parameters for thresholds, timeouts, intervals, and success codes.
- Added support for enabling/disabling metrics in ASG resources by introducing the `enabled_metrics` property in the ASG profile schema.

## 2024-04-29

### Fixed
- Fixed an issue with handling slashes in `keyType` for `duplocloud_admin_system_setting` resource ID creation and parsing.

### Enhanced
- Added utility functions to encode and decode slashes in identifiers, improving handling of special characters in resource IDs.
- Updated build script to include documentation generation and modified import script to guide users on handling slashes in identifiers.

## 2024-04-15

### Fixed
- Fixed an issue in `duplocloud_ecs_task_definition` update process by marking the `task_definition` field with `ForceNew` in `duplocloud_ecs_service.go`, ensuring the ECS service resource is recreated when the task definition is updated.

## 2024-04-12

### Enhanced
- Introduced logic for in-place updates during GCP Cloud SQL database version changes, eliminating the need for resource replacement.
- Implemented custom diff logic to selectively force new resource creation based on specific conditions for GCP Cloud SQL.
- Added a separate update request for `DeletionProtectionEnabled` attribute to ensure its proper handling.
- Made `DeletionProtectionEnabled` field non-omitempty to ensure it's always included in update requests.
- Enhanced the update process for Global Secondary Indexes and Throughput when changes are detected.

## 2024-04-11

### Added
- Introduced GCP Firestore resource and data sources for managing GCP Firestore in DuploCloud.
- Added CRUD operations and data retrieval for GCP Firestore resources.

### Enhanced
- Enhanced DynamoDB table logic to support delete controller updates.
- Enhanced Kubernetes Just-In-Time (JIT) access and cluster availability validation for Azure in the `duplocloud_eks_credentials` data source.
- Introduced `AksConfig` struct for managing Azure Kubernetes Service (AKS) configurations, including cluster management and privacy settings.

### Fixed
- Fixed error handling for `duplocloud_eks_credentials` data-source related to Azure infrastructure, ensuring proper Kubernetes cluster availability checks.


## 2024-04-09

### Fixed
- Fixed an issue where Terraform resources would not be recreated when volume mappings were modified for Duplo services.
- Implemented exponential backoff in `TenantUpdateLbSettings` and `TenantGetLbSettings` methods to handle AWS API rate limits more gracefully.
- Fixed rate limiting exceptions when creating and managing `duplocloud_duplo_service_params`s with terraform

### Enhanced
- Improved the deletion process for Duplo services by efficiently checking the existence of replication controllers before proceeding.
- Increased the wait time after service deletion from 40 seconds to 240 seconds to better accommodate backend cleanup processes, especially for GCP environments.
- Added `RetryWithExponentialBackoff` utility function for general use in retrying operations with exponential backoff, configurable delay, jitter, and total timeout.

markdown
## 2024-04-08

### Fixed
- Fixed state not being set for GCP Node Pools data source.

### Enhanced
- Enhanced logging in GCP Node Pools data source by adding `fmt` import.
- Changed `cgroup_mode` from `TypeString` to `TypeList` in GCP Node Pools to handle multiple values correctly.
- Updated documentation and examples to reflect changes in data source and resource configurations for GCP Node Pools and GCP SQL Database Instances.


### Documentation
- Corrected GKE credentials documentation and Terraform example, updating references from EKS to GKE and ensuring output values accurately reflect GKE credentials.

markdown

## 2024-04-05

### Added
- Introduced support for `allow_global_access` attribute `duplocloud_duplo_service_lbconfigs` resource.
- Added a new API endpoint `TenantGetLbDetailsInServiceNew` in `duplocloud_duplo_service_lbconfigs` for enhanced load balancer details retrieval.

### Enhanced
- Updated SDK and Terraform provider to support the new `allow_global_access` attribute in load balancer configurations.

### Documentation
- Updated documentation to reflect changes in load balancer configurations, including the addition of the `allow_global_access` attribute.

### Fixed
- Fixed nil pointer exception while error handling for `duplocloud_eks_credentials` and `duplocloud_gke_credentials` datasource
## 2024-04-04

### Added
- Introduced support for custom prefixes in Azure VM names for flexible naming conventions.
- Added a `fullname` attribute to Azure VM resources for enhanced traceability and management.
- Added support to pass `root_password` for `duplocloud_gcp_sql_database_instance` resource

### Enhanced
- Updated Azure VM resource documentation to include the `fullname` attribute.
- Improved logging and resource management to handle the `fullname` of Azure VMs across operations.

### Fixed
- Resolved issues with Azure VM operations (create, read, update, delete) to correctly handle the `fullname`.

## 2024-04-02

### Enhanced
- Enhanced functionality to update single zone cluster to multizone cluster for `duplocloud_aws_elasticsearch` opensearch resource.

## 2024-04-01

### Fixed
- fixed `duplocloud_plan_settings` resource diff issue on no change for gcp

## 2024-03-29

### Enhanced
- Disabled handling for `account_tier`, `access_tier`, and `enable_https_traffic_only` attributes in Azure Storage Account resource to align with API changes.
- Commented out the `flattenAzureStorageAccount` function call, indicating a shift in handling Azure Storage Account data.

### Documentation
- Updated Azure Storage Account and Infrastructure documentation to reflect changes and added `subnet_fullname` attribute documentation for Azure infrastructure.

## 2024-03-28

### Enhanced
- Enhanced handling of subnet names in Azure infrastructure resources to support custom prefixes.
- Removed redundant code that incorrectly set `subnet_fullname` without considering Azure custom prefixes.

## 2023-03-27

## 2024-03-27

### Added
- Added support for Azure custom prefixes in various resources and SDK enhancements.
- Implemented Azure tenant creation logic with specific handling for Azure environments.
- Introduced test infrastructure and fixtures for new Azure-related features.
- Added example for `duplocloud_gcp_node_pool` and `duplocloud_gcp_node_pools` for data-source


### Enhanced
- Enhanced Azure storage account creation with a delay and adjusted `account_tier` attribute for better reliability.

### Fixed
- Error output fix for data-source of `duplocloud_eks_credentials`/`duplocloud_gke_credentials`


## 2023-03-26

### Fixed
- Fixed `secret_data` diff issue for `duplocloud_k8_secret`
- Fixed `duplocloud_s3_bucket` resource creation issue

## 2023-03-22

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
