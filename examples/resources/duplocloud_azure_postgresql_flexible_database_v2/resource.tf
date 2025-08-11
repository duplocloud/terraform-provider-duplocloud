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
  active_directory_authentication = "Disabled"
}

//Example for Entra Authentication 
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
  password_authentication         = "Disabled"
}

//Example for Postgresql and Entra Authentication 
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
