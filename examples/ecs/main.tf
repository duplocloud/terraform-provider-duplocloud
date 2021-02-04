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
  lifecycle {
    ignore_changes = [task_role_arn, execution_role_arn]
  }
  tenant_id = var.tenant_id
  family = "duploservices-default-joedemo"
  cpu = "256"
  memory = "1024"
  requires_compatibilities = [ "FARGATE" ]
}
