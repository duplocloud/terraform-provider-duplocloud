resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}



resource "duplocloud_azure_postgresql_flexible_database_v2" "db" {
  tenant_id                       = duplocloud_tenant.myapp.tenant_id
  name                            = "psqlflex"
  service_tier                    = "Burstable"
  hardware                        = "Standard_B2ms"
  high_availability               = "Disabled"
  storage_gb                      = 64
  version                         = "16"
  minor_version                   = "8"
  delegated_subnet_id             = "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/virtualNetworks/{virtualNetworkName}/subnets/{subnetName}"
  private_dns_zone_id             = "/subscriptions/<subscription-id>/resourceGroups/<resource-group-name>/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com"
  administrator_login             = "tftry"
  administrator_login_password    = "Guide#123"
  backup_retention_days           = 8
  geo_redundant_backup            = "Enabled"
  active_directory_authentication = "Enabled"
  public_network_access           = "Disabled"
}

resource "duplocloud_azure_postgresql_flexible_db_ad_administrator" "adauth" {
  tenant_id       = duplocloud_azure_postgresql_flexible_database_v2.db.tenant_id
  db_name         = duplocloud_azure_postgresql_flexible_database_v2.db.name
  azure_tenant_id = duplocloud_azure_postgresql_flexible_database_v2.db.active_directory_tenant_id
  principal_name  = "nikhil.nambiar@duplocloud.net"
  principal_type  = "User"
  object_id       = "33b6ac35-ec39-4d5a-a42d-3bb4b4f56312"
}
