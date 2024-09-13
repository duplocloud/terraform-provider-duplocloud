### Create a S3 bucket named static_assets in DuploCloud

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
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
```

### Provision an S3 bucket within the dev tenant in DuploCloud

```
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
```

### Create an S3 bucket in the dev tenant within DuploCloud, with public access enabled

```
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
```

### Create an S3 bucket in the dev tenant within DuploCloud, with versioning disabled

```
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
```

### Setup an S3 bucket in the qa tenant within duplo, with access logs disabled

```
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
```

### Configure an S3 bucket in the QA tenant within DuploCloud, enabling public access while disabling versioning and access logs

```
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
```

### Create an S3 bucket named data in the preprod tenant within duplo, with tenant kms enabled

```
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
```