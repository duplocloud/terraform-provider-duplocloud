resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy NGINX using Duplo's native container agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1
}

# Simple Example 2:  Deploy NGINX using Duplo's EKS agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 7 # Duplo EKS agent
  docker_image   = "nginx:latest"
  replicas       = 1
}
