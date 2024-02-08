resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Example 1:
//
// - Two services in duplocloud (`echo` and `echo2`)
// - Both of the above services configured as kubernetes services of type `NodePort`
// - Both of the services exposed as a kubernetes ingress
//
resource "duplocloud_duplo_service" "echo" {
  tenant_id                            = duplocloud_tenant.myapp.tenant_id
  name                                 = "echo"
  replicas                             = 1
  lb_synced_deployment                 = false
  cloud_creds_from_k8s_service_account = true
  is_daemonset                         = false
  agent_platform                       = 7
  cloud                                = 0
  other_docker_config = jsonencode({
    "Args" : [
      "-text=\"Hello world\""
    ]
    }
  )
  docker_image = "hashicorp/http-echo:latest"
}

resource "duplocloud_duplo_service_lbconfigs" "echo_config" {
  tenant_id                   = duplocloud_duplo_service.echo.tenant_id
  replication_controller_name = duplocloud_duplo_service.echo.name
  lbconfigs {
    lb_type          = 4
    is_native        = false
    is_internal      = false
    port             = 5678
    external_port    = 30001
    protocol         = "tcp"
    health_check_url = "/"
  }
}

resource "duplocloud_duplo_service" "echo2" {
  tenant_id                            = duplocloud_tenant.myapp.tenant_id
  name                                 = "echo2"
  replicas                             = 1
  lb_synced_deployment                 = false
  cloud_creds_from_k8s_service_account = true
  is_daemonset                         = false
  agent_platform                       = 7
  cloud                                = 0
  other_docker_config = jsonencode({
    "Args" : [
      "-text=\"Hello India\""
    ]
    }
  )
  docker_image = "hashicorp/http-echo:latest"
}

resource "duplocloud_duplo_service_lbconfigs" "echo2_config" {
  tenant_id                   = duplocloud_duplo_service.echo2.tenant_id
  replication_controller_name = duplocloud_duplo_service.echo2.name
  lbconfigs {
    lb_type          = 4
    is_native        = false
    is_internal      = false
    port             = 5678
    external_port    = 30002
    protocol         = "tcp"
    health_check_url = "/"
  }
}


resource "time_sleep" "wait_45_seconds" {
  depends_on = [duplocloud_duplo_service.echo, duplocloud_duplo_service.echo2]

  create_duration = "45s"
}

resource "duplocloud_k8_ingress" "ingress" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  name               = "external-echo"
  ingress_class_name = "alb"
  lbconfig {
    is_internal     = false
    dns_prefix      = "external-echo"
    certificate_arn = "put your certificate ARN here"
    https_port      = 443
  }

  rule {
    path         = "/hello-world"
    path_type    = "Prefix"
    service_name = duplocloud_duplo_service.echo.name
    port         = 5678
  }
  rule {
    path         = "/hello-india"
    path_type    = "Prefix"
    service_name = duplocloud_duplo_service.echo2.name
    port         = 5678
  }

  depends_on = [time_sleep.wait_45_seconds]
}



// Example 2:
//
// - Adding a custom redirect to an ingress resource
//
resource "duplocloud_k8_ingress" "ingress" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  name               = "joedemo"
  ingress_class_name = "alb"

  annotations = {
    "alb.ingress.kubernetes.io/actions.redirect-to-new-domain" = jsonencode({
      Type = "redirect",
      RedirectConfig = {
        Host       = "my.example.com",
        Path       = "/#{path}",
        Port       = "#{port}",
        Protocol   = "HTTPS",
        Query      = "#{query}",
        StatusCode = "HTTP_301"
      }
    })
  }

  lbconfig {
    is_internal = false
    dns_prefix  = "joedemo-ingress"
    http_port   = 80
  }

  rule {
    path         = "/"
    path_type    = "Prefix"
    service_name = "redirect-to-new-domain"
    port_name    = "use-annotation"
  }
}
