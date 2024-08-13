resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy an S3 bucket with hardened security settings.
resource "duplocloud_gcp_storage_bucket_v2" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  allow_public_access = false
  enable_versioning   = true
  default_encryption {
    method = "Sse" # For even stricter security, use "TenantKms" here.
  }
}

# Simple Example 2:  Deploy a hardened S3 bucket suitable for public website hosting.
resource "duplocloud_gcp_storage_bucket_v2" "www" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "website"
  allow_public_access = true
  enable_versioning   = true
  default_encryption {
    method = "Sse"
  }
}


# Simple Example 3:  Deploy an S3 bucket to dersired region.
resource "duplocloud_gcp_storage_bucket_v2" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  # optional, if not provided, multi-region US will be used
  location = "us-west-2"

}

# Simple Example 4:  Deploy an S3 bucket with multiple region.

resource "duplocloud_gcp_storage_bucket_v2" "mydata" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mydata"

  allow_public_access = true
  enable_versioning   = true
  default_encryption {
    method = "Sse"
  }
  location = "Asia" #pass region value (Asia/EU/US)to location to enable multi region
}
