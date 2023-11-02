terraform {
  required_providers {
    duplocloud = {
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

#variable "tenant_id" {
#  description = "tenant id/subscription id"
#  type = string
#  default = "c7163b39-43ca-4d44-81ce-9a323087039b"
#}
#

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

#resource "duplocloud_tenant" "myapp" {
#  account_name = "myapp"
#  plan_id      = "default"
#  tenant_id = "64813f67-7df1-438d-95ea-c3cc2b151717"
#}

#resource "duplocloud_tenant" "myapp" {
#  account_name = "newtenant"
#  plan_id      = "default"
#}

#resource "duplocloud_tenant_config" "myapp" {
#  tenant_id = "64813f67-7df1-438d-95ea-c3cc2b151717"
#
#  // This turns on a default public access block for S3 buckets.
#  setting {
#    key   = "block_public_access_to_s3"
#    value = "false"
#  }
#
#  // This turns on a default SSL enforcement S3 bucket policy.
#  setting {
#    key   = "enforce_ssl_for_s3"
#    value = "false"
#  }
#}

#resource "duplocloud_tenant_config" "conf" {
#  tenant_id     = "9117f273-03e2-4b1c-8ce0-0039c4c6d2c1"
#}

## Deploy NGINX using Duplo's native container agent, and configure a load balancer.
#resource "duplocloud_duplo_service" "myservice2" {
#  tenant_id = "81386040-9a07-4e76-a195-94191d9d47aa"
#
#  name           = "myservice2"
#  agent_platform = 0 # Duplo native container agent
#  docker_image   = "nginx:latest"
#  replicas       = 1
#}
#resource "duplocloud_duplo_service_lbconfigs" "myservice2" {
#  tenant_id                   = duplocloud_duplo_service.myservice2.tenant_id
#  replication_controller_name = duplocloud_duplo_service.myservice2.name
#
#  lbconfigs {
#    health_check_url = "/"
#    is_native        = false
#    lb_type          = 5 # Application load balancer
#    port             = "80"
#    protocol         = "http"
#  }
#}



#resource "duplocloud_infrastructure" "myinfra" {
#  infra_name        = "dd-eks"
#  cloud             = 0 // AWS
#  region            = "us-east-1"
#  azcount           = 2
#  enable_k8_cluster = true
#  address_prefix    = "10.34.0.0/16"
#  subnet_cidr       = 24
#}

#resource "duplocloud_infrastructure_setting" "settings" {
##  infra_name = duplocloud_infrastructure.myinfra.infra_name
#  infra_name = "dd-eks"
#  setting {
#    key   = "EnableSecretCsiDriver"
#    value = "true"
#  }
#  setting {
#    key   = "EnableAWSEfsVolumes"
#    value = "true"
#  }
#  setting {
#    key   = "EnableAwsAlbIngress"
#    value = "true"
#  }
#}

#resource "duplocloud_aws_lambda_function" "myfunction" {
#
#  tenant_id   = "c7163b39-43ca-4d44-81ce-9a323087039b"
#  name        = "myfunction"
#  description = "A description of my function"
#
#  package_type = "Image"
#  image_uri = "151326636836.dkr.ecr.us-west-2.amazonaws.com/whatever:latest"
#
#  environment {
#    variables = {
#      "FOO" = "BAR"
#    }
#  }
#
#  image_config {
#    command = ["echo", "hello world"]
#    entry_point = ["echo hello workd"]
#    working_directory = "/tmp3"
#  }
#
#  timeout     = 60
#  memory_size = 512
#}
#
#resource "duplocloud_aws_lambda_function" "theirfunction" {
#
#  tenant_id   = "c7163b39-43ca-4d44-81ce-9a323087039b"
#  name        = "theirfunction"
#  description = "A description of my function"
#
#  package_type = "Image"
#  image_uri = "151326636836.dkr.ecr.us-west-2.amazonaws.com/whatever:latest"
#
#  environment {
#    variables = {
#      "FOO" = "BAR"
#    }
#  }
#
#  image_config {
#    command = ["echo", "hello world"]
#    entry_point = ["echo hello workd"]
#    working_directory = "/tmp3"
#  }
#
#  timeout     = 60
#  memory_size = 512
#}

#resource "duplocloud_duplo_service" "lime" {
#  tenant_id = "81386040-9a07-4e76-a195-94191d9d47aa"
#
#  name           = "lime"
#  agent_platform = 0 # Duplo native container agent
#  docker_image   = "nginx:latest"
#  replicas       = 1
#}
#
#resource "duplocloud_duplo_service_lbconfigs" "lime" {
#  tenant_id                   = duplocloud_duplo_service.lime.tenant_id
#  replication_controller_name = duplocloud_duplo_service.lime.name
#  lbconfigs {
#    lb_type          = 7
#    is_native        = false
#    is_internal      = false
#    port             = 3004
#    external_port    = 3004
#    protocol         = "http"
#    health_check_url = "/"
#  }
#}

#// Create an RDS instance.
#resource "duplocloud_rds_instance" "mydb" {
#  tenant_id      = "64813f67-7df1-438d-95ea-c3cc2b151717"
#  name           = "mydb"
#  engine         = 1 // PostgreSQL
#  engine_version = "15.2"
#  size           = "db.t3.medium"
#  storage_type = "io1"
#  allocated_storage = 100
#  iops = 3000
#
#  master_username = "myuser"
#  master_password = random_password.mypassword.result
#
#  encrypt_storage = true
#}

// Generate a random password.
#resource "random_password" "mypassword" {
#  length  = 16
#  special = false
#}
#
#// Create an RDS instance.
#resource "duplocloud_rds_instance" "rds" {
#  tenant_id      = "64813f67-7df1-438d-95ea-c3cc2b151717"
#  name           = "anotherdb"
#  engine         = 1 // AuroraPostgreSQL
#  engine_version = "15.2"
#  size           = "db.t3.medium"
#  storage_type = "io1"
#  allocated_storage = 100
#  iops = 3000
#
#  master_username = "myuser"
#  master_password = random_password.mypassword.result
#
#  encrypt_storage = true
#}
#
##resource "duplocloud_rds_instance" "rds" {
##  tenant_id       = "81386040-9a07-4e76-a195-94191d9d47a4"
##  enable_logging  = false
##  encrypt_storage = true
##  engine          = 8
##  engine_version  = "8.0.mysql_aurora.3.04.0"
##  master_password = "test!!1234"
##  master_username = "masteruser"
##  multi_az        = false
##  name            = "mysqltest"
##  size            = "db.t2.small"
##}
#
#resource "duplocloud_rds_read_replica" "replica" {
#  tenant_id          = "64813f67-7df1-438d-95ea-c3cc2b151716"
#  name               = "read-replica"
#  size               = "db.t2.small"
#  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
#}

#resource "duplocloud_tenant" "myapp" {
#  account_name = "myapp"
#  plan_id      = "default"
#}
#
#resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {
#
#  tenant_id      = duplocloud_tenant.myapp.tenant_id
#  name           = "tst-dynamodb-table"
#  read_capacity  = 10
#  write_capacity = 10
#  #billing_mode = "PAY_PER_REQUEST"
#  tag {
#    key   = "CreatedBy"
#    value = "Duplo"
#  }
#
#  tag {
#    key   = "CreatedFrom"
#    value = "Duplo"
#  }
#
#  attribute {
#    name = "UserId"
#    type = "S"
#  }
#
#  attribute {
#    name = "GameTitle"
#    type = "S"
#  }
#
#  attribute {
#    name = "TopScore"
#    type = "N"
#  }
#
#  key_schema {
#    attribute_name = "UserId"
#    key_type       = "HASH"
#  }
#
#  key_schema {
#    attribute_name = "GameTitle"
#    key_type       = "RANGE"
#  }
#
#  global_secondary_index {
#    name               = "GameTitleIndex"
#    hash_key           = "GameTitle"
#    range_key          = "TopScore"
#    write_capacity     = 10
#    read_capacity      = 10
#    projection_type    = "INCLUDE"
#    non_key_attributes = ["UserId"]
#  }
#}

#resource "duplocloud_infrastructure" "deleteme" {
#  infra_name        = "delete-me-please"
#  enable_ecs_cluster = true
#  enable_k8_cluster = false
#  region = "us-west-2"
#  address_prefix = "10.10.12.0/22"
#  azcount = 2
#  subnet_cidr       = 24
#}

#data "duplocloud_tenant_config" "config" {
#  tenant_id     = "64813f67-7df1-438d-95ea-c3cc2b151717"
#}

#resource "random_password" "mypassword" {
#  length  = 16
#  special = false
#}
#resource "duplocloud_rds_instance" "mydb1" {
#  tenant_id      = "09da93de-cfc0-42bc-9cc5-5f11de485cd2"
#  name           = "mydbstorage-arn"
#  engine         = 0
#  engine_version = "8.0.28"
#  size           = "db.t3.small"
#  master_username = "myuser"
#  master_password = random_password.mypassword.result
#  encrypt_storage = true
#  //multi_az = true
#}

resource "duplocloud_k8s_job" "myjob" {
  tenant_id = "81386040-9a07-4e76-a195-94191d9d47aa"
  metadata {
    name = "ddjob2"
  }
  spec {
    template {
      spec {
        container {
          name = "ddjob2"
          image = "nginx:1.25"
        }
      }
    }
  }
  wait_for_completion = false
}