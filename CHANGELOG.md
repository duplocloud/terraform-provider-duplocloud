## 2024-01-23

### Fixed
- Corrected a typo in the tenant data source read function in `duplocloud/data_source_duplo_tenant.go`.
- Enhanced the autoscaling group schema by forcing a new resource to be created when `is_cluster_autoscaled` and `can_scale_from_zero` attributes are changed.

## 2024-01-23

### Added
- Introduced a new attribute `can_scale_from_zero` to the `duplocloud_asg_profile` resource, allowing an AWS Autoscaling Group to leverage DuploCloud's scale from zero feature on Amazon EKS.

## 2024-01-22

### Fixed
- Added `"ImplementationSpecific"` to the list of acceptable values for ingress rule path types to the `duplocloud_k8_ingress` resource.
