resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_tenant_config" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  // This turns on a default public access block for S3 buckets.
  setting {
    key   = "block_public_access_to_s3"
    value = "true"
  }

  // This turns on a default SSL enforcement S3 bucket policy.
  setting {
    key   = "enforce_ssl_for_s3"
    value = "true"
  }
}
