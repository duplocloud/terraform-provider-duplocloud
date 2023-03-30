resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Minimal example
resource "duplocloud_aws_elasticsearch" "sample" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "sample"
}

# Example with hardened settings
resource "duplocloud_aws_elasticsearch" "es-doc" {
  tenant_id                      = duplocloud_tenant.myapp.tenant_id
  name                           = "es-doc"
  enable_node_to_node_encryption = true
  require_ssl                    = true
  use_latest_tls_cipher          = true
}
