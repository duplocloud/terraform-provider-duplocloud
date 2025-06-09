resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_cosmos_db_database" "db" {
  tenant_id    = duplocloud_tenant.myapp.tenant_id
  account_name = duplocloud_azure_cosmos_db_account.account.name
  name         = "db-test"
}