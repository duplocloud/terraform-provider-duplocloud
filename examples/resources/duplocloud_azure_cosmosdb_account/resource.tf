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