resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_k8_node_pool" "node_pool" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  identifier       = 2
  min_capacity     = 1
  max_capacity     = 1
  desired_capacity = 1
  vm_size          = "Standard_E16_v5"
  wait_until_ready = false
  allocation_tag   = "aks-test"
}
