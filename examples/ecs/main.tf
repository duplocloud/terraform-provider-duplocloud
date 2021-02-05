terraform {
  required_providers {
    duplocloud = {
      version = "0.3.7"
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  //duplo_host = "https://xxx.duplocloud.net"
  //duplo_token = "xxxx"
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
  }])
  cpu = "256"
  memory = "1024"
  requires_compatibilities = [ "FARGATE" ]
}

resource "duplocloud_ecs_service" "test" {
  tenant_id = var.tenant_id
  name = "joedemo"
  task_definition = duplocloud_ecs_task_definition.test.arn
  replicas = 1
  load_balancer {
    lb_type = 1
    port = 80
    external_port = 80
    protocol = "HTTP"
  }
}
