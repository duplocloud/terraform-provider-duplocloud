terraform {
  required_providers {
    duplocloud = {
      version = "0.4.3" # RELEASE VERSION
      source = "registry.terraform.io/duplocloud/duplocloud"
    }
    aws = {
      version = "~> 3.29.1"
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

# AWS region retrieval - straight from Duplo!
data "duplocloud_tenant_aws_region" "test" { tenant_id = var.tenant_id }
output "aws_region" { value = data.duplocloud_tenant_aws_region.test.aws_region }

# Use any AWS terraform resource with just-in-time Duplo credentials!
data "duplocloud_tenant_aws_credentials" "test" { tenant_id = var.tenant_id }
provider "aws" {
  # The following credentials are temporary "just in time" credentials created by Duplo.
  access_key = data.duplocloud_tenant_aws_credentials.test.access_key_id
  secret_key = data.duplocloud_tenant_aws_credentials.test.secret_access_key
  token      = data.duplocloud_tenant_aws_credentials.test.session_token
  region     = data.duplocloud_tenant_aws_credentials.test.region
}

data "aws_caller_identity" "current" {}
output "aws_account_id" { value = data.aws_caller_identity.current.account_id }
