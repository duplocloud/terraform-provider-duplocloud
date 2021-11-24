resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_mysql_database" "mydb" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "mysql-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  storage_mb                   = 102400
  version                      = 5.7
  sku_name                     = "GP_Gen5_4"
}
