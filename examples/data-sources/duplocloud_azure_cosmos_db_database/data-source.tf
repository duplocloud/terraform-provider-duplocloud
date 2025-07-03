data "duplocloud_azure_cosmos_db_database" "db" {
  tenant_id    = "tenant id"
  name         = "cosmos db account name"
  account_name = "cosmos db account name"
}

output "out" {
  value = {
    type      = data.duplocloud_azure_cosmos_db_database.db.type
    namespace = data.duplocloud_azure_cosmos_db_database.db.namespace
  }
}