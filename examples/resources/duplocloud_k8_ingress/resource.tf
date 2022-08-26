locals {
  tenant_id = "72916492-b69f-492e-8f64-957ad211aca1"
  cert_arn  = "arn:aws:acm:us-west-2:708951311304:certificate/829b32dc-d106-4229-a96d-123456789"
}

resource "duplocloud_duplo_service" "echo" {
  tenant_id                            = local.tenant_id
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
  tenant_id                            = local.tenant_id
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
  tenant_id          = local.tenant_id
  name               = "external-echo"
  ingress_class_name = "alb"
  lbconfig {
    is_internal     = false
    dns_prefix      = "external-echo"
    certificate_arn = local.cert_arn
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
