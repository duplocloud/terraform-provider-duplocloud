resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_efs_file_system" "efs" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "efs-test"
  performance_mode = "generalPurpose"
  throughput_mode  = "elastic"
  backup           = true
  encrypted        = true
}

resource "duplocloud_aws_efs_lifecycle_policy" "efs_policy" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  file_system_id = duplocloud_aws_efs_file_system.efs.file_system_id
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
  }
  lifecycle_policy {
    transition_to_archive = "AFTER_14_DAYS"
  }
  lifecycle_policy {
    transition_to_primary_storage_class = "AFTER_1_ACCESS"
  }
}
