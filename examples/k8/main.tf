terraform {
  required_providers {
    duplocloud = {
      version = "0.2"
      source = "registry.terraform.io/hashicorp/duplocloud"
    }
  }
}

provider "duplocloud" {
  duplo_host = "https://xxx.duplocloud.net"
  duplo_token = "xxxx"}

resource "duplocloud_tenant" "tenant1" {
  account_name = "tf5"
  plan_id = "devtest"
}

resource "duplocloud_aws_host" "host1" {
  tenant_id = duplocloud_tenant.tenant1.tenant_id
  user_account = duplocloud_tenant.tenant1.account_name
  image_id = "ami-062e7f29a4d477f5a"
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
    certificate_arn = "arn:aws:acm:us-east-1:905621702370:certificate/7b39df5b-2bd8-4445-85a8-ed28efc2461b"
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
    "AWS_DATA_BUCKET_NAME": "customer-nonprod-proof-appdata",
    "AWS_DATA_BUCKET_URL": "https://customer-nonprod-proof-appdata.s3.amazonaws.com/",
    "MERIDIANLINK_URL": "https://demo.mortgagecreditlink.com/inetapi/request_products.aspx",
    "SMTP_ERROR_USER": "customer-sre+proof@brisance.digital",
    "SMTP_HOST": "smtp.gmail.com",
    "SMTP_PORT": "587",
    "TENANT_BRAND_LONG_NAME": "customer proof of trust",
    "TENANT_BRAND_NAME": "customer",
    "TITLE_COMPANY_LOGO": "disabled",
    "TITLE_ORDER_API_EMAIL_FOLLOWUP": "disabled",
    "TITLE_ORDER_API_SEND_ORDER_VIA_API": "disabled",
    "TWILIO_PHONE": "+10000000000",
    "TWILIO_PREFIX": "https://api.proof.nonprod.customer.io"
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