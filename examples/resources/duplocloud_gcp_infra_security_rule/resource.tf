
resource "duplocloud_gcp_infra_security_rule" "irule" {
  infra_name  = "test"
  name        = "test-infra-r14"
  description = "test rule for infra test"
  ports_and_protocols {
    ports            = ["24", "23-89"]
    service_protocol = "tcp"

  }
  ports_and_protocols {
    ports            = ["100"]
    service_protocol = "udp"

  }
  source_ranges = ["0.0.0.0/32"]
  rule_type     = "ALLOW"

}
