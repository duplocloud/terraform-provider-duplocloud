resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_azure_storageclass_queue" "qu" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  storage_account_name = "test-store"
  name                 = "qaque"
}