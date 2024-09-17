---

# Resource: duplocloud_s3_bucket



`duplocloud_s3_bucket` manages an s3 bucket in Duplo.


## Example Usage

### Create a S3 bucket named static_assets

```terraform
# Before creating a tenant, you must first set up the infrastructure. Below is the resource for creating the infrastructure.
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0 # AWS Cloud
  region            = "us-west-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}

# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource while creating tenant.
resource "duplocloud_tenant" "tenant" {
  account_name = "prod"
  plan_id      = duplocloud_infrastructure.infra.infra_name
}

# Use the tenant_id from the duplocloud_tenant, which will be populated after the tenant resource is created, when setting up the S3 bucket.
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

### Provision an S3 bucket within the dev tenant

```terraform
# Ensure the 'dev' tenant is already created before setting up the s3 bucket.
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

### Create an S3 bucket in the dev tenant, with public access enabled

```terraform
# Ensure the 'dev' tenant is already created before setting up the s3 bucket.
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

```terraform
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

### Setup an S3 bucket in the qa tenant, with access logs disabled

```terraform
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

### Configure an S3 bucket in the QA tenant, enabling public access while disabling versioning and access logs

```terraform
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

### Create an S3 bucket named data in the preprod tenant, with tenant kms enabled

```terraform
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

### Deploy an S3 bucket with hardened security settings

```terraform
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
```

###  Deploy a hardened S3 bucket suitable for public website hosting in test tenant

```terraform
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
```

###  Deploy an S3 bucket to us-east-1 region

```terraform
# Ensure the 'test' tenant is already created before creating the s3 bucket.
data "duplocloud_tenant" "tenant" {
  name = "test"
}

resource "duplocloud_s3_bucket" "static_assets" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "static_assets"
  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }

  # Optional, if not provided, tenant region will be used
  region = "us-east-1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The short name of the S3 bucket.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.
- `tenant_id` (String) The GUID of the tenant that the S3 bucket will be created in.

### Optional

- `allow_public_access` (Boolean) Whether or not to remove the public access block from the bucket.
- `default_encryption` (Block List, Max: 1) Default encryption settings for objects uploaded to the bucket. (see [below for nested schema](#nestedblock--default_encryption))
- `enable_access_logs` (Boolean) Whether or not to enable access logs.  When enabled, Duplo will send access logs to a centralized S3 bucket per plan.
- `enable_versioning` (Boolean) Whether or not to enable versioning.
- `managed_policies` (List of String) Duplo can manage your S3 bucket policy for you, based on simple list of policy keywords:

 - `"ssl"`: Require SSL / HTTPS when accessing the bucket.
 - `"ignore"`: If this value is present, Duplo will not manage your bucket policy.
- `region` (String) The region of the S3 bucket.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `arn` (String) The ARN of the S3 bucket.
- `domain_name` (String) The domain name of the S3 bucket.
- `fullname` (String) The full name of the S3 bucket.
- `id` (String) The ID of this resource.
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))

<a id="nestedblock--default_encryption"></a>
### Nested Schema for `default_encryption`

Optional:

- `method` (String) Default encryption method.  Must be one of: `None`, `Sse`, `AwsKms`, `TenantKms`. Defaults to `Sse`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)

## Import

Import is supported using the following syntax:

```
# Example: Importing an existing S3 bucket
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the S3 bucket (without the duploservices prefix)
#
terraform import duplocloud_s3_bucket.mybucket *TENANT_ID*/*SHORTNAME*
```
