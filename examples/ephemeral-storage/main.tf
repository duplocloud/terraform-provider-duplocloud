terraform {
  required_providers {
    duplocloud = {
      version = "0.10.27" # RELEASE VERSION
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  duplo_host  = "https://xxx.duplocloud.net" # you can also set the duplo_host env var
  duplo_token = "..."                        # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
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

resource "duplocloud_s3_bucket" "s3bucket" {
  tenant_id = var.tenant_id
  name      = "miketest"
}

resource "duplocloud_aws_lambda_function" "test" {
  tenant_id         = var.tenant_id
  name              = "miketest"
  package_type      = "Zip"
  description       = "miketest lambda"
  s3_bucket         = duplocloud_s3_bucket.s3bucket.fullname
  s3_key            = "file.zip"
  memory_size       = 128
  runtime           = "python3.9"
  handler           = "main"
  ephemeral_storage = 1024
}

