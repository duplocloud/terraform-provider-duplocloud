terraform {
  required_providers {
    duplocloud = {
      version = "0.9.13" # RELEASE VERSION
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
    aws = {
      version = "~> 3.29.1"
    }
    kubernetes = {
      version = "~> 2.0"
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

# Use any AWS terraform resource with just-in-time Duplo credentials!
data "duplocloud_tenant_aws_credentials" "test" { tenant_id = var.tenant_id }
provider "aws" {
  # The following credentials are temporary "just in time" credentials created by Duplo.
  access_key = data.duplocloud_tenant_aws_credentials.test.access_key_id
  secret_key = data.duplocloud_tenant_aws_credentials.test.secret_access_key
  token      = data.duplocloud_tenant_aws_credentials.test.session_token
  region     = data.duplocloud_tenant_aws_credentials.test.region
}

# Authenticate to EKS using just-in-time Duplo credentials!
data "duplocloud_eks_credentials" "test" { plan_id = var.plan_id }
data "aws_eks_cluster" "cluster" { name = data.duplocloud_eks_credentials.test.name }
provider "kubernetes" {
  host                   = data.duplocloud_eks_credentials.test.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.duplocloud_eks_credentials.test.token
}

# Use any Kubernetes resource with just-in-time Duplo credentials!
data "kubernetes_all_namespaces" "allns" {}

output "all-ns" { value = data.kubernetes_all_namespaces.allns.namespaces }

output "eks_creds_name" { value = data.duplocloud_eks_credentials.test.name }
output "eks_creds_endpoint" { value = data.duplocloud_eks_credentials.test.endpoint }
output "eks_creds_region" { value = data.duplocloud_eks_credentials.test.region }

data "duplocloud_k8_config_maps" "test" {
  tenant_id = var.tenant_id
}
output "config_maps" { value = data.duplocloud_k8_config_maps.test.config_maps }
data "duplocloud_k8_config_map" "test" {
  tenant_id = var.tenant_id
  name      = "joetest"
}
output "config_map" { value = jsondecode(data.duplocloud_k8_config_map.test.data) }
resource "duplocloud_k8_config_map" "test" {
  tenant_id = var.tenant_id

  name = "joetest"

  data = jsonencode({ foo = "bar2" })
}
