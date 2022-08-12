terraform {
  required_providers {
    duplocloud = {
      version = "0.8.18" # RELEASE VERSION
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

# Tenant information 
data "duplocloud_tenant" "test" { name = "default" }
output "tenant" { value = data.duplocloud_tenant.test }

# Tenant listing
data "duplocloud_tenants" "test" {}
output "tenants" { value = data.duplocloud_tenants.test.tenants.*.name }

resource "duplocloud_tenant_secret" "test" {
  tenant_id   = var.tenant_id
  name_suffix = "joetest2"
  data        = "{ \"foo\" : \"bar4\" }"
}
output "tenant_secret_name" { value = duplocloud_tenant_secret.test.name }

resource "duplocloud_aws_load_balancer" "test" {
  tenant_id            = var.tenant_id
  name                 = "joetest2"
  is_internal          = true
  enable_access_logs   = true
  drop_invalid_headers = true
}

data "duplocloud_aws_lb_listeners" "test" {
  tenant_id = var.tenant_id
  name      = duplocloud_aws_load_balancer.test.name
}
output "test_lb_listeners" { value = data.duplocloud_aws_lb_listeners.test.listeners }
data "duplocloud_aws_lb_target_groups" "test" {
  tenant_id = var.tenant_id
}
output "test_lb_target_groups" { value = data.duplocloud_aws_lb_target_groups.test.target_groups }

resource "duplocloud_ecs_task_definition" "test" {
  tenant_id = var.tenant_id
  family    = "duploservices-default-joedemo"
  container_definitions = jsonencode([{
    Name  = "default"
    Image = "nginx:latest"
    Environment = [
      { Name = "foo", Value = "bar2" },
      { Name = "bar", Value = "foo" }
    ]
  }])
  cpu                      = "256"
  memory                   = "1024"
  requires_compatibilities = ["FARGATE"]
}

resource "duplocloud_ecs_service" "test" {
  tenant_id       = var.tenant_id
  name            = "joedemo-ecs"
  task_definition = duplocloud_ecs_task_definition.test.arn
  replicas        = 2
  load_balancer {
    lb_type              = 1
    port                 = "8080"
    external_port        = 80
    protocol             = "HTTP"
    enable_access_logs   = false
    drop_invalid_headers = true
    webaclid             = ""
  }
}

resource "duplocloud_ecache_instance" "test" {
  tenant_id  = var.tenant_id
  name       = "joetest"
  cache_type = 0
  replicas   = 1
  size       = "cache.t2.small"
}

resource "duplocloud_s3_bucket" "test" {
  tenant_id = var.tenant_id
  name      = "joetestjoetestjoetestjoetestjoetestjoetestjoetestjoetestjoetest"
}

resource "duplocloud_rds_instance" "test" {
  tenant_id       = var.tenant_id
  name            = "joetest2"
  engine          = 0
  master_username = "joe"
  master_password = "test12345"
  size            = "db.t2.small"
}

resource "duplocloud_aws_elasticsearch" "test" {
  tenant_id                      = var.tenant_id
  name                           = "joe2"
  storage_size                   = 20
  selected_zone                  = 1
  enable_node_to_node_encryption = true
  require_ssl                    = true
  use_latest_tls_cipher          = true
}
