resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_storage_account" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "storagacctest"
}
