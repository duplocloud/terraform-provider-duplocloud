resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_mssql_database" "mydb" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "mssql-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  version                      = "12.0"
  minimum_tls_version          = "1.2"
}
