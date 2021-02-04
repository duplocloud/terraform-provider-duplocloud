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
  family = "testing"
}
