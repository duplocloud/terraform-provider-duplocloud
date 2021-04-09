terraform {
  required_providers {
    duplocloud = {
      version = "0.5.18" # RELEASE VERSION
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

resource "duplocloud_tenant" "tenant1" {
  account_name = "tf5"
  plan_id = "devtest"
}

resource "duplocloud_aws_host" "host1" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  user_account = duplocloud_tenant.tenant1.account_name
  image_id = "ami-062e7f2xxxa"
  capacity = "t2.small"
  agent_platform = 7
  friendly_name = "host1"
  minion_tags {
    key = "AllocationTags"
    value = "apps"
  }
}

resource "duplocloud_duplo_service" "website" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  name = "website"
  agent_platform = 7
  docker_image = "nginx:latest"
  replicas = 1
  other_docker_config = file("other_k8_config_website.json")
}

resource "duplocloud_duplo_service_lbconfigs" "websitelbs" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  replication_controller_name = duplocloud_duplo_service.website.name
  lbconfigs    {
    certificate_arn = "arn:aws:acm:us-east-1:9xxxxx1b"
    external_port =  443
    health_check_url= "/"
    is_native = false
    lb_type= 1
    port ="80"
    protocol = "http"
    replication_controller_name = "website"
  }
}

resource "duplocloud_duplo_service" "mongodb" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  name = "mongodb"
  agent_platform = 7
  docker_image = "docker.io/bitnami/mongodb:4.2.8-debian-10-r47"
  replicas = 1
  other_docker_config = file("other_k8_config_mongodb.json")
  volumes = jsonencode(
  [
    {
      "Name": "datadir",
      "Path": "/bitnami/mongodb",
      "AccessMode": "ReadWriteOnce",
      "Size": "10Gi"
    }
  ]
  )
}

resource "duplocloud_duplo_service_lbconfigs" "mongodblbs" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  replication_controller_name = duplocloud_duplo_service.mongodb.name
  lbconfigs    {
    lb_type= 3
    port ="27017"
    external_port =  27017
    protocol = "tcp"
    replication_controller_name = "mongodb"
  }
}



resource "duplocloud_k8_config_map" "customer-api-configs" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  name = "customer-api-configs"
  data = jsonencode(
  {
    "ABC": "DEF",
    "123": "456"
  }
  )
}

//secret_data  secret_name  secret_version  secret_type
resource "duplocloud_k8_secret" "customer-api-secrets" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  secret_type = "Opaque"
  secret_name = "customer-api-secrets"
  secret_data = file("customer-api-secrets.json")
}
