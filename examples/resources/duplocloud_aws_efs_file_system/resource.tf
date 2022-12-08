resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_efs_file_system" "efs" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "efs-test"
  performance_mode = "generalPurpose"
  throughput_mode  = "bursting"
  backup           = false
  encrypted        = false
}
