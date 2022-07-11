resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_mssql_elasticpool" "mssql_elasticpool" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "mssql-ep"
  server_name = "mysqlserver"
  sku {
    name     = "StandardPool"
    capacity = 50
  }
}
