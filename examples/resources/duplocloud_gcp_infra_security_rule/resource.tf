resource "duplocloud_gcp_infra_security_rule" "irule" {
  infra_name       = "nonprod"
  name             = "nonprod-firewall-rule"
  description      = "firewall rule for infra nonprod"
  ports            = ["24", "23-89"]
  service_protocol = "tcp"
  source_ranges    = ["0.0.0.0/32"]
  rule_type        = "ALLOW"

}
