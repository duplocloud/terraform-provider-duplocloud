terraform {
  required_providers {
    duplocloud = {
      version = "0.4.6" # RELEASE VERSION
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

variable "plan_id" {
  type = string
  default = "default"
}

variable "tenant_id" {
  type = string
}

# # EKS credentials retrieval
# data "duplocloud_eks_credentials" "test" { plan_id = var.plan_id }
# output "eks_creds_name" { value = data.duplocloud_eks_credentials.test.name }
# output "eks_creds_endpoint" { value = data.duplocloud_eks_credentials.test.endpoint }
# output "eks_creds_region" { value = data.duplocloud_eks_credentials.test.region }

# # AWS credentials retrieval
# data "duplocloud_tenant_aws_credentials" "test" { tenant_id = var.tenant_id }
# output "aws_creds_region" { value = data.duplocloud_tenant_aws_credentials.test.region }

# # AWS account ID retrieval
# data "duplocloud_aws_account" "test" {}
# output "aws_account_id" { value = data.duplocloud_aws_account.test.account_id }

# # AWS region retrieval
# data "duplocloud_tenant_aws_region" "test" { tenant_id = var.tenant_id }
# output "aws_region" { value = data.duplocloud_tenant_aws_region.test.aws_region }

# # Tenant secrets retrieval
# data "duplocloud_tenant_secrets" "test" { tenant_id = var.tenant_id }
# output "tenant_secrets" { value = data.duplocloud_tenant_secrets.test.secrets }

resource "duplocloud_tenant_secret" "test" {
  tenant_id = var.tenant_id
  name_suffix = "joetest"
  data = jsonencode({ foo = "bar" })
}
output "tenant_secret_name" { value = duplocloud_tenant_secret.test.name }

# resource "duplocloud_ecs_task_definition" "test" {
#   tenant_id = var.tenant_id
#   family = "duploservices-default-joedemo"
#   container_definitions = jsonencode([{
#     Name = "default"
#     Image = "nginx:latest"
#     Essential = true
#   }])
#   cpu = "256"
#   memory = "1024"
#   requires_compatibilities = [ "FARGATE" ]
# }

# resource "duplocloud_ecs_service" "test" {
#   tenant_id = var.tenant_id
#   name = "joedemo"
#   task_definition = duplocloud_ecs_task_definition.test.arn
#   replicas = 2
#   load_balancer {
#     lb_type = 1
#     port = 8080
#     external_port = 80
#     protocol = "HTTP"
#   }
# }

resource "duplocloud_ecache_instance" "test" {
  tenant_id = var.tenant_id
  name = "joetest"
  cache_type = 0
  replicas = 1
  size = "cache.t2.small"
}

# resource "duplocloud_rds_instance" "test" {
#   tenant_id = var.tenant_id
#   name = "joetest"
#   master_username = "joe"
#   master_password = "test12345!"
#   size = "db.t2.small"
# }

resource "duplocloud_aws_elasticsearch" "test" {
  tenant_id = var.tenant_id
  name = "joe2"
  storage_size = 20
  selected_zone = 1
}
