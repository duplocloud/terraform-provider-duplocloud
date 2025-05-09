resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_mssqldb_retention_backup" "backup" {
  tenant_id             = duplocloud_tenant.myapp.tenant_id
  server_name           = duplocloud_azure_mssql_server.mssql_server.name
  database_name         = duplocloud_azure_mysql_database.mydb.name
  retention_backup_days = 8
}
