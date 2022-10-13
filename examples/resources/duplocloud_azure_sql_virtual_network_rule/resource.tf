resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_sql_virtual_network_rule" "sql_vnet_rule" {
  tenant_id                            = duplocloud_tenant.myapp.tenant_id
  name                                 = "test-rule"
  server_name                          = "test-server"
  subnet_id                            = "/subscriptions/0c84b91e-95f5-409e-9cff-6c2e60affbb3/resourceGroups/duploinfra-demo/providers/Microsoft.Network/virtualNetworks/demo/subnets/duploinfra-default"
  ignore_missing_vnet_service_endpoint = false
}
