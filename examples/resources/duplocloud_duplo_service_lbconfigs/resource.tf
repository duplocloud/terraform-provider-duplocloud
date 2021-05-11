resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Deploy NGINX using Duplo's native container agent, and configure a load balancer.
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1
}
resource "duplocloud_duplo_service_lbconfigs" "myservice" {
  tenant_id                   = duplocloud_duplo_service.myservice.tenant_id
  replication_controller_name = duplocloud_duplo_service.myservice.name

  lbconfigs {
    external_port    = 80
    health_check_url = "/"
    is_native        = false
    lb_type          = 1 # Application load balancer
    port             = "80"
    protocol         = "http"
  }
}
