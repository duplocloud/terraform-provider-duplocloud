resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy an S3 bucket with hardened security settings.
resource "duplocloud_s3_bucket" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  allow_public_access = false
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}

# Simple Example 2:  Deploy a hardened S3 bucket suitable for public website hosting.
resource "duplocloud_s3_bucket" "www" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "website"

  allow_public_access = true
  enable_access_logs  = true
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse"
  }
}


# Simple Example 3:  Deploy an S3 bucket to dersired region.
resource "duplocloud_s3_bucket" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  # optional, if not provided, tenant region will be used
  region = "us-west-2"

}
