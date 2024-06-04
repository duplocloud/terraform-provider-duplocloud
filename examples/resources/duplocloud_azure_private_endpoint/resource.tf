resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_private_endpoint" "mssql_server_private_endpoint" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "pe-duplo-tf"
  subnet_id = "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Network/virtualNetworks/tmerge/subnets/duploinfra-default"
  private_link_service_connection {
    name                           = "pe-mssql"
    private_connection_resource_id = "/subscriptions/<subscription-id>/resourceGroups/duploservices-jee556/providers/Microsoft.Sql/servers/thisistotestprivateendpoint"
    group_ids                      = ["sqlServer"]
  }
}

resource "duplocloud_azure_private_endpoint" "storage_server_private_endpoint" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "pe-duplo-tf-storage"
  subnet_id = "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Network/virtualNetworks/tmerge/subnets/duploinfra-default"
  private_link_service_connection {
    name                           = "pe-storage"
    private_connection_resource_id = "/subscriptions/<subscription-id>/resourceGroups/duploservices-jee556/providers/Microsoft.Storage/storageAccounts/letsfixprivateendpoint"
    group_ids                      = ["Blob"]
  }
}