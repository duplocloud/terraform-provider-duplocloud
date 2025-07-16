resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


#Example of Azure Cosmos DB Account resource
# This example creates an Azure Cosmos DB Account will create a new Azure Cosmos DB account for provisioned type databases.
resource "duplocloud_azure_cosmos_db_account" "account" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "tfcosmotest"
  kind      = "GlobalDocumentDB"
  consistency_policy {
    default_consistency_level = "Session"
    max_staleness_prefix      = 100
    max_interval_in_seconds   = 5
  }
  disable_key_based_metadata_write_access = false
  public_network_access                   = "Enabled"
  backup_policy {
    backup_interval           = 240
    backup_retention_interval = 8
    backup_storage_redundancy = "Geo"
    type                      = "Periodic"
  }
  enable_free_tier = true #applicable only for provisioned type databases
}

#Example of Azure Cosmos DB Account resource
# This example creates an Azure Cosmos DB Account will create a new Azure Cosmos DB account for serverless type databases.
resource "duplocloud_azure_cosmos_db_account" "account" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "tfcosmotest"
  kind      = "GlobalDocumentDB"
  capabilities {
    name = "EnableServerless"
  }

  consistency_policy {
    default_consistency_level = "Session"
    max_staleness_prefix      = 100
    max_interval_in_seconds   = 5
  }
  disable_key_based_metadata_write_access = false
  public_network_access                   = "Enabled"
  backup_policy {
    backup_interval           = 240
    backup_retention_interval = 8
    backup_storage_redundancy = "Geo"
    type                      = "Periodic"
  }
}


resource "duplocloud_azure_cosmos_db_account" "account" {
  tenant_id = "8ea6f24e-0f13-4ad5-999f-8a2acc18438e"
  name      = "act13"
  kind      = "GlobalDocumentDB"
  consistency_policy {
    default_consistency_level = "Session"
  }
  backup_policy {
    backup_interval           = 240
    backup_retention_interval = 8
    backup_storage_redundancy = "Geo"
    type                      = "Periodic"
  }
  #enable_free_tier = true
  geo_location {
    location_name     = "centralus"
    failover_priority = 0
  }
  is_virtual_network_filter_enabled = true
  virtual_network_rule {
    subnet_id                            = "/subscriptions/143ffc59-9394-4ec6-8f5a-c408a238be62/resourceGroups/duploinfra-denisr-dev/providers/Microsoft.Network/virtualNetworks/denisr-dev/subnets/duploinfra-sql-managed-instance"
    ignore_missing_vnet_service_endpoint = false
  }
  virtual_network_rule {
    subnet_id                            = "/subscriptions/143ffc59-9394-4ec6-8f5a-c408a238be62/resourceGroups/duploinfra-denisr-dev/providers/Microsoft.Network/virtualNetworks/denisr-dev/subnets/duploinfra-sql-managed-instance-01"
    ignore_missing_vnet_service_endpoint = false
  }
}