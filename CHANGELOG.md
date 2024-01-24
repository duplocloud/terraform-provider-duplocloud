## 2024-01-23

### Fixed
- Corrected a typo in `duplocloud_tenant` data source read function (tenant_id is the id naming for the resource but datasource calls it "id").
- Enhanced schema for `duplocloud_asg_profile` resource by forcing recreation when `is_cluster_autoscaled` and `can_scale_from_zero` attributes are changed.

### Added
- Introduced a new attribute `can_scale_from_zero` to the `duplocloud_asg_profile` resource, allowing an AWS Autoscaling Group to leverage DuploCloud's scale from zero feature on Amazon EKS.

## 2024-01-22

### Fixed
- Added `"ImplementationSpecific"` to the list of acceptable values for ingress rule path types to the `duplocloud_k8_ingress` resource.
