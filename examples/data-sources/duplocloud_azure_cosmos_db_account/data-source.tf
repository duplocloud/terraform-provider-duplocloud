data "duplocloud_azure_cosmos_db_account" "account" {
  tenant_id = "tenant id"
  name      = "cosmos db account name"

}

output "out" {
  value = {
    kind                                    = data.duplocloud_azure_cosmos_db_account.account.kind
    name                                    = data.duplocloud_azure_cosmos_db_account.account.name
    locations                               = data.duplocloud_azure_cosmos_db_account.account.locations
    type                                    = data.duplocloud_azure_cosmos_db_account.account.type
    consistency_policy                      = data.duplocloud_azure_cosmos_db_account.account.consistency_policy
    public_network_access                   = data.duplocloud_azure_cosmos_db_account.account.public_network_access
    disable_key_based_metadata_write_access = data.duplocloud_azure_cosmos_db_account.account.disable_key_based_metadata_write_access
    backup_policy                           = data.duplocloud_azure_cosmos_db_account.account.backup_policy
    enable_free_tier                        = data.duplocloud_azure_cosmos_db_account.account.enable_free_tier
    capabilities                            = data.duplocloud_azure_cosmos_db_account.account.capabilities
    capacity_mode                           = data.duplocloud_azure_cosmos_db_account.account.capacity_mode
  }
}