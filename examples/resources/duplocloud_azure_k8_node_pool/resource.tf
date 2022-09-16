resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# supported feature names are "loganalytics", "publicip", "addsjoin", and "aadjoin"
resource "duplocloud_azure_k8_node_pool" "node_pool" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  identifier       = 5
  min_capacity     = 1
  max_capacity     = 1
  desired_capacity = 1
}
