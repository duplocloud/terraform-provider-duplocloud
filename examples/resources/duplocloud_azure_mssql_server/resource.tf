resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_mssql_server" "mssql_server" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "mssql-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  version                      = "12.0"
  minimum_tls_version          = "1.2"
}

resource "duplocloud_azure_mssql_server" "mssql_server" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "mssql-test"
  administrator_login          = "testroot"
  administrator_login_password = "P@ssword12345"
  version                      = "12.0"
  minimum_tls_version          = "1.2"
  public_network_access        = "Disabled"
  active_directory_administrator {
    tenant_id              = "active-directory-tenant-id"
    login                  = "azure user email"
    object_id              = "active-directory-user-object-id"
    ad_authentication_only = true
  }
}