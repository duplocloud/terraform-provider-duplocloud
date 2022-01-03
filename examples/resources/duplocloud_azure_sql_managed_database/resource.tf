resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_sql_managed_database" "mydb" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "db-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  minimum_tls_version          = "1.2"
  sku_name                     = "GP_Gen5"
  vcores                       = 4
  storage_size_in_gb           = 32
  subnet_id                    = "/subscriptions/0c84b91e-95f5-409e-9cff-6c2e60affbb3/resourceGroups/duploinfra-demo/providers/Microsoft.Network/virtualNetworks/demo/subnets/duploinfra-default"
}
