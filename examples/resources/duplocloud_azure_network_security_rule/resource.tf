resource "duplocloud_azure_network_security_rule" "security_rule" {
  infra_name                  = "demo"
  network_security_group_name = "duploinfra-sub01"
  name                        = "test-sg-rule"
  source_rule_type            = 0
  destination_rule_type       = 0
  protocol                    = "tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "49.207.178.47/32"
  destination_address_prefix  = "*"
  access                      = "Allow"
  priority                    = 200
  direction                   = "Inbound"
}
