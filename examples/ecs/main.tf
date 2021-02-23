terraform {
  required_providers {
    duplocloud = {
      version = "0.4.0"
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

variable "tenant_id" {
  type = string
}

resource "duplocloud_ecs_task_definition" "test" {
  tenant_id = var.tenant_id
  family = "duploservices-default-joedemo"
  container_definitions = jsonencode([{
    Name = "default"
    Image = "nginx:latest"
    Essential = true
  }])
  cpu = "256"
  memory = "1024"
  requires_compatibilities = [ "FARGATE" ]
}

resource "duplocloud_ecs_service" "test" {
  tenant_id = var.tenant_id
  name = "joedemo"
  task_definition = duplocloud_ecs_task_definition.test.arn
  replicas = 2
  load_balancer {
    lb_type = 1
    port = 8080
    external_port = 80
    protocol = "HTTP"
  }
}

resource "duplocloud_ecache_instance" "test" {
  tenant_id = var.tenant_id
  name = "joetest"
  cache_type = 0
  size = "cache.t2.small"
}

#resource "duplocloud_rds_instance" "test" {
#  tenant_id = var.tenant_id
#  name = "joetest"
#  master_username = "joe"
#  master_password = "test1234!"
#  size = "db.t2.small"
#}

