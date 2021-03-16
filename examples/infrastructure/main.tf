terraform {
  required_providers {
    duplocloud = {
      version = "0.5.7" # RELEASE VERSION
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

data "duplocloud_infrastructure" "all" {}

output "all_data" {
  value = data.duplocloud_infrastructure.all.data
}

resource "duplocloud_infrastructure" "test" {
  infra_name = "test"
  cloud = 0
  region = "us-west-2" 
  azcount = 2
  enable_k8_cluster = true
  address_prefix = "10.42.0.0/16"
  subnet_cidr = 24
}

resource "duplocloud_tenant" "test" {
  account_name = "test"
  plan_id = duplocloud_infrastructure.test.infra_name
}

resource "duplocloud_tenant_config" "test" {
  tenant_id = duplocloud_tenant.test.tenant_id
  metadata {
    key = "block_public_access_to_s3"
    value = "true"
  }
}
