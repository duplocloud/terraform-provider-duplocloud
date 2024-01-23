## 2024-01-23

### Updated
- Enhanced the documentation and examples for ECS service and task definition resources.
- Introduced `health_check_url` to the ECS service resource in the documentation and example.
- Added `ContainerMappings` to the ECS task definition resource in the documentation and example.

### Added
- Introduced a new attribute `can_scale_from_zero` to the `duplocloud_asg_profile` resource, allowing an AWS Autoscaling Group to leverage DuploCloud's scale from zero feature on Amazon EKS.

## 2024-01-22

### Fixed
- Added `"ImplementationSpecific"` to the list of acceptable values for ingress rule path types to the `duplocloud_k8_ingress` resource.
