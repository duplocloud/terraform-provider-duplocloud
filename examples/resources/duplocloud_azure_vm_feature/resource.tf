resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# supported feature names are "loganalytics", "publicip", "addsjoin", and "aadjoin"
resource "duplocloud_azure_vm_feature" "vm_feature" {
  tenant_id    = duplocloud_tenant.myapp.tenant_id
  component_id = "p01-host01"
  feature_name = "aadjoin"
  enabled      = true
}
