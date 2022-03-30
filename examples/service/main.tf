terraform {
  required_providers {
    duplocloud = {
      version = "0.7.8" # RELEASE VERSION
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

variable "plan_id" {
  type    = string
  default = "default"
}

variable "tenant_id" {
  type = string
}

// data "duplocloud_native_host_image" "test" {
//   tenant_id     = var.tenant_id
//   is_kubernetes = true
// }

resource "duplocloud_duplo_service" "test" {
  tenant_id = var.tenant_id

  name           = "joedemo"
  agent_platform = 0
  docker_image   = "nginx:latest"
  replicas       = 1
  /*
  other_docker_config = jsonencode({
    Env = [
      { Name = "NGINX_HOST", Value = "foo" },
      { Name = "NGINX_PORT", Value = "8080" },
    ]
  })
*/
}

resource "duplocloud_duplo_service_lbconfigs" "test" {
  tenant_id                   = var.tenant_id
  replication_controller_name = duplocloud_duplo_service.test.name

  lbconfigs {
    external_port    = 80
    health_check_url = "/"
    is_native        = false
    lb_type          = 1
    port             = "80"
    protocol         = "http"
  }

  # Workaround for AWS:  Even after the ALB is available, there is some short duration where a V2 WAF cannot be attached.
  provisioner "local-exec" {
    command = "sleep 10"
  }
}

resource "duplocloud_duplo_service_params" "test" {
  tenant_id = var.tenant_id

  replication_controller_name = duplocloud_duplo_service_lbconfigs.test.replication_controller_name
  dns_prfx                    = "joedemo-svc"
  drop_invalid_headers        = true
  enable_access_logs          = true
}

output "test" {
  value = duplocloud_duplo_service_params.test
}
