resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_postgresql_database" "mydb" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "postgresql-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  storage_mb                   = 102400
  version                      = 11
  sku_name                     = "B_Gen5_2"
}
