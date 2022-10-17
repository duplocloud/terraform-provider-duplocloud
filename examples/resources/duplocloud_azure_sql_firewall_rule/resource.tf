resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_sql_firewall_rule" "sql_firewall_rule" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "test-rule"
  server_name      = "test-server"
  start_ip_address = "10.0.17.62"
  end_ip_address   = "10.0.17.62"
}
