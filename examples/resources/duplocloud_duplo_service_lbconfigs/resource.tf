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
    external_port            = 80
    health_check_url         = "/"
    is_native                = false
    lb_type                  = 1 # Application load balancer
    port                     = "80"
    protocol                 = "HTTP"
    backend_protocol_version = "HTTP1"
    health_check {
      healthy_threshold   = 4
      unhealthy_threshold = 4
      timeout             = 50
      interval            = 30
      http_success_codes  = "200-399"
    }
  }
}


resource "duplocloud_duplo_service_lbconfigs" "myservice2" {
  tenant_id                   = duplocloud_duplo_service.myservice.tenant_id
  replication_controller_name = duplocloud_duplo_service.myservice.name

  lbconfigs {
    external_port            = 80
    health_check_url         = "/"
    is_native                = false
    lb_type                  = 1 # Application load balancer
    port                     = "80"
    protocol                 = "HTTPS"
    certificate_arn          = "certificate:arn"
    backend_protocol_version = "HTTP2"
    health_check {
      healthy_threshold   = 4
      unhealthy_threshold = 4
      timeout             = 50
      interval            = 30
      http_success_codes  = "200-399"
    }
  }
}
