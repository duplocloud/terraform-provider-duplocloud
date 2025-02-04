resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_gcp_tenant_security_rule" "trule" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "tenant-rule"
  description = "security rule for target tenant"
  ports_and_protocols {
    ports            = ["24", "23-89"]
    service_protocol = "tcp"

  }
  ports_and_protocols {
    ports            = ["100"]
    service_protocol = "udp"

  }
  source_ranges    = ["0.0.0.0/32"]
  rule_type        = "ALLOW"
  target_tenant_id = "<target-tenant-id>"

}

