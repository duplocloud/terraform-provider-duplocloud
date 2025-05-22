resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Allow the "default" tenant to send HTTPS requests to "myapp"
resource "duplocloud_tenant_network_security_rule" "myrule" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  source_tenant = "default"
  protocol      = "tcp"
  from_port     = 443
  to_port       = 443
  description   = "Allow the default tenant to send HTTPS traffic"
}

resource "duplocloud_tenant_network_security_rule" "myrule" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  source_address = "10.220.46.215/32"
  protocol       = "tcp"
  from_port      = 20
  to_port        = 35
  description    = "Allow the default tenant to send HTTPS traffic"
}