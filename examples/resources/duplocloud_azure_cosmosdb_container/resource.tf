resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_azure_cosmos_db_container" "container" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  account_name       = duplocloud_azure_cosmos_db_account.account.name
  database_name      = duplocloud_azure_cosmos_db_database.db.name
  name               = "container2"
  partition_key_path = "/id"
}