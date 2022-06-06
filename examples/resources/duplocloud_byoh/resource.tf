resource "duplocloud_tenant" "duplo-app" {
  account_name   = "duplo-app"
  plan_id        = "default"
  allow_deletion = true
}

#Without Credentials
resource "duplocloud_byoh" "byoh" {
  tenant_id      = duplocloud_tenant.duplo-app.tenant_id
  name           = "test-byoh"
  direct_address = "10.99.99.16"
  agent_platform = 0
  allocation_tag = "byoh"
}

#With Credentials
resource "duplocloud_byoh" "byoh" {
  tenant_id      = duplocloud_tenant.duplo-app.tenant_id
  name           = "test-byoh"
  direct_address = "10.99.99.16"
  agent_platform = 0
  allocation_tag = "byoh"
  username       = "byoh-test"
  password       = "By0h@Te$t"
}
