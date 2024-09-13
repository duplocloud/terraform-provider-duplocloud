### Create a S3 bucket named static_assets in DuploCloud

#### Solution:
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0 # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}

resource "duplocloud_tenant" "tenant" {
  account_name = "prod"
  plan_id      = duplocloud_infrastructure.infra.infra_name
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = duplocloud_tenant.tenant.tenant_id
  name                = "static_assets"
  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Provision an S3 bucket within the dev tenant in DuploCloud

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "bucket"
  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Create an S3 bucket in the dev tenant within DuploCloud, with public access enabled

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "bucket"
  allow_public_access = true
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Create an S3 bucket in the dev tenant within DuploCloud, with versioning disabled

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "bucket"
  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Setup an S3 bucket in the qa tenant within duplo, with access logs disabled

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "qa"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "bucket"
  allow_public_access = false
  enable_access_logs  = false
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Configure an S3 bucket in the QA tenant within DuploCloud, enabling public access while disabling versioning and access logs

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "qa"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "bucket"
  allow_public_access = true
  enable_access_logs  = false
  enable_versioning   = false
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}


### Create an S3 bucket named data in the preprod tenant within duplo, with tenant kms enabled

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "preprod"
}

resource "duplocloud_s3_bucket" "bucket" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "data"
  allow_public_access = false
  enable_access_logs  = false
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "TenantKms"
  }
}

resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}
### Deploy an S3 bucket with hardened security settings

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "test"
}

resource "duplocloud_s3_bucket" "mydata" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "mydata"

  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}

###  Deploy a hardened S3 bucket suitable for public website hosting in test tenant

#### Solution:
data "duplocloud_tenant" "tenant" {
  name = "test"
}

resource "duplocloud_s3_bucket" "www" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "website"

  allow_public_access = true
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse"
  }
}

###  Deploy an S3 bucket to us-east-1 region

#### Solution:
data "aws_region" "region" {
  name = "us-east-1"
}

data "duplocloud_tenant" "tenant" {
  name = "test"
}

resource "duplocloud_s3_bucket" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  # optional, if not provided, tenant region will be used
  region = data.aws_region.region.name
}
