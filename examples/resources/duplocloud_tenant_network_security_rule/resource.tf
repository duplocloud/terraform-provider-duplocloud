resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Allow the "default" tenant to send HTTPS requests to "myapp"
resource "duplocloud_tenant_network_security_rule" "myrule" {
  tenant_id = duplocloud_tenant.myapp.tenant_idd

  source_tenant = "default"
  protocol      = "tcp"
  from_port     = 443
  to_port       = 443
  description   = "Allow the default tenant to send HTTPS traffic"
}
