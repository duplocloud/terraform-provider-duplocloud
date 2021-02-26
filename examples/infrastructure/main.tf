terraform {
  required_providers {
    duplocloud = {
      version = "0.4.7" # RELEASE VERSION
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

resource "duplocloud_infrastructure" "tfinfra11" {
  infra_name = "tfinfra11"
  cloud = 0
  region = "us-west-2" 
  azcount = 2
  enable_k8_cluster = true
  address_prefix = "10.23.0.0/16"
  subnet_cidr = 24
}
