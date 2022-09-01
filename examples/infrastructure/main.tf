terraform {
  required_providers {
    duplocloud = {
      version = "0.8.23" # RELEASE VERSION
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

data "duplocloud_plans" "all" {}
data "duplocloud_plan" "default" { plan_id = "default" }

# resource "duplocloud_infrastructure_setting" "test" {
#   infra_name = "nonprod"
#   setting {
#     key   = "foox"
#     value = "barx"
#   }
#}

# resource "duplocloud_infrastructure" "test" {
#   infra_name        = "ecstest"
#   cloud             = 0
#   region            = "us-west-2"
#   azcount           = 2
#   enable_k8_cluster = false
#   enable_ecs_cluster = true
#   address_prefix    = "10.122.0.0/16"
#   subnet_cidr       = 22

#   custom_data {
#     key   = "foox"
#     value = "barx"
#   }
# }

# resource "duplocloud_tenant" "test" {
#   account_name = "t2t1"
#   plan_id      = duplocloud_infrastructure.test.infra_name
# }

# resource "duplocloud_tenant_config" "test" {
#   tenant_id = duplocloud_tenant.test.tenant_id
#   setting {
#     key   = "block_public_access_to_s3"
#     value = "true"
#   }
#   setting {
#     key   = "enforce_ssl_for_s3"
#     value = "true"
#   }
# }
