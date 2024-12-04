resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_azure_postgresql_flexible_database" "db" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "psqlflex"
  service_tier                 = "Burstable"
  hardware                     = "Standard_B2ms"
  high_availability            = "Disabled"
  storage_gb                   = 64
  version                      = "16"
  subnet                       = "subnet"
  administrator_login          = "tftry"
  administrator_login_password = "trynew#1"
  backup_retention_days        = 7
  geo_redundant_backup         = "Enabled"
}
