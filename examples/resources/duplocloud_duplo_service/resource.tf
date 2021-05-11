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

# Example 3:  Deploy NGINX with host networking and env vars, using Duplo's native container agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    Env = [
      { Name = "NGINX_HOST", Value = "foo" },
    ]
  })

  // Enables host networking, and listening on ports < 1000
  other_docker_host_config = jsonencode({
    NetworkMode = "host",
    CapAdd      = ["NET_ADMIN"]
  })
}

# Example 4:  Deploy NGINX with host networking and env vars, using Duplo's EKS agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 7 # Duplo EKS agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    HostNetwork = true,
    Env = [
      { Name = "NGINX_HOST", Value = "foo" },
    ]
  })
}
