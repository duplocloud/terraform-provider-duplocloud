resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_storage_share_file" "share_file" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "test-share-file"
  storage_account_name = "test-st-account"
}
