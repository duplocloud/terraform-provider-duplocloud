resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_storage_bucket" "mybucket" {

  tenant_id         = duplocloud_tenant.myapp.tenant_id
  name              = "mybucket"
  enable_versioning = false
}
