resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Without Elastic Pool
resource "duplocloud_azure_mssql_database" "mssql_database" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "mssql-db"
  server_name = "mysqlserver"
  sku {
    name     = "Free"
    capacity = 5
  }
}

# With Elastic Pool
resource "duplocloud_azure_mssql_elasticpool" "mssql_elasticpool" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "mssql-ep"
  server_name = "mysqlserver"
  sku {
    name     = "StandardPool"
    capacity = 50
  }
}

resource "duplocloud_azure_mssql_database" "mssql_database" {
  tenant_id       = duplocloud_tenant.myapp.tenant_id
  name            = "mssql-db"
  server_name     = "mysqlserver"
  elastic_pool_id = duplocloud_azure_mssql_elasticpool.mssql_elasticpool.elastic_pool_id
}
