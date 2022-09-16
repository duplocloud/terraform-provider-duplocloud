resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_k8_node_pool" "node_pool" {
  tenant_id        = "fda1e050-3168-49b6-be2e-d2be2784f5f7"
  identifier       = 2
  min_capacity     = 1
  max_capacity     = 1
  desired_capacity = 1
  vm_size          = "Standard_E16_v5"
  wait_until_ready = false
  allocation_tag   = "aks-test"
}
